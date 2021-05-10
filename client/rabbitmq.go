package client

import (
	"context"
	"github.com/streadway/amqp"
	"rabbitmq/naming"
	"sync"
	"time"
)

type Client struct {
	conn           *conn
	addrs          []string
	prefetchCount  int
	prefetchGlobal bool
	mtx            sync.Mutex
	wg             sync.WaitGroup
}

func NewClient(addrs []string) *Client {
	return &Client{addrs: addrs}
}

func (c *Client) Conn(exchange *naming.Exchange) error {
	if c.conn == nil {
		c.conn = newConn(exchange, c.addrs, DefaultPrefetchCount, false)
	}

	conf := defaultAmqpConfig

	return c.conn.connect(false, &conf)
}

func (c *Client) Disconnect() error {
	if c.conn == nil {
		return ErrConnNil
	}
	ret := c.conn.Close()
	c.wg.Wait()
	return ret
}

type subscriber struct {
	mtx    sync.Mutex
	mayRun bool
	queue  *naming.Queue
	topic  *naming.RoutingKey
	ch     *channel
	r      *Client
	fn     func(msg amqp.Delivery)
}

func (s *subscriber) Topic() *naming.RoutingKey {
	return s.topic
}

func (s *subscriber) Unsubscribe() error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.mayRun = false
	if s.ch != nil {
		return s.ch.Close()
	}
	return nil
}

func (s *subscriber) resubscribe() {
	minDelay := 100 * time.Millisecond
	maxDelay := 30 * time.Second
	expFactor := time.Duration(2)
	reDelay := minDelay

	for {
		s.mtx.Lock()
		mayRun := s.mayRun
		s.mtx.Unlock()
		if !mayRun {
			return
		}

		select {
		case <-s.r.conn.close:
			return
		case <-s.r.conn.waitConnection:
		}

		s.r.mtx.Lock()
		if !s.r.conn.connected {
			s.r.mtx.Unlock()
			continue
		}
		ch, sub, err := s.r.conn.Consume(s.queue)
		s.r.mtx.Lock()
		switch err {
		case nil:
			reDelay = minDelay
			s.mtx.Lock()
			s.ch = ch
			s.mtx.Unlock()
		default:
			if reDelay > maxDelay {
				reDelay = maxDelay
			}
			time.Sleep(reDelay)
			reDelay *= expFactor
			continue
		}

		for d := range sub {
			s.r.wg.Add(1)
			s.fn(d)
			s.r.wg.Done()
		}
	}
}

func (c *Client) Publish(routingKey *naming.RoutingKey, msg []byte) error {
	m := amqp.Publishing{ContentType: "text/plain", Body: msg}
	if c.conn == nil {
		return ErrConnNil
	}

	return c.conn.Publish(c.conn.exchange, routingKey, m)
}
func (c *Client) Subscribe(routingKey *naming.RoutingKey, handler func(ctx context.Context) error) error {
	if c.conn == nil {
		return ErrConnNil
	}
	ctx := context.Background()
	fn := func(msg amqp.Delivery) {
		ctx = ContextWithDelivery(ctx, msg)
		err := handler(ctx)
		if err != nil {
			_ = msg.Ack(false)
		} else {
			_ = msg.Nack(false, false)
		}
	}
	sret := &subscriber{
		topic:  routingKey,
		mayRun: true,
		fn:     fn,
		r:      c,
	}
	go sret.resubscribe()
	return nil

}
