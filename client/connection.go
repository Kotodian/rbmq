package client

import (
	"crypto/tls"
	"github.com/streadway/amqp"
	"rabbitmq/naming"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	DefaultExchange       = *naming.NewExchange("csms", naming.Direct)
	DefaultURL            = "amqp://guest:guest@127.0.0.1:5672"
	DefaultPrefetchCount  = 0
	DefaultPrefetchGlobal = false
	DefaultRequeueOnError = false

	defaultHeartbeat = 10 * time.Second
	defaultLocale    = "zh_CN"

	defaultAmqpConfig = amqp.Config{
		Heartbeat: defaultHeartbeat,
		Locale:    defaultLocale,
	}

	dial       = amqp.Dial
	dialTLS    = amqp.DialTLS
	dialConfig = amqp.DialConfig
)

type conn struct {
	Connection      *amqp.Connection
	Channel         *channel
	ExchangeChannel *channel
	exchange        *naming.Exchange
	url             string
	prefetchCount   int
	prefetchGlobal  bool

	sync.Mutex
	connected bool
	close     chan bool

	waitConnection chan struct{}
}

func newConn(ex *naming.Exchange, urls []string, prefetchCount int, prefetchGlobal bool) *conn {
	var url string
	if len(urls) > 0 && validUrl(urls[0]) {
		url = urls[0]
	} else {
		url = DefaultURL
	}
	ret := &conn{
		exchange:       ex,
		url:            url,
		prefetchCount:  prefetchCount,
		prefetchGlobal: prefetchGlobal,
		close:          make(chan bool),
		waitConnection: make(chan struct{}),
	}
	close(ret.waitConnection)
	return ret
}

func (c *conn) connect(secure bool, config *amqp.Config) error {
	if err := c.tryConnect(secure, config); err != nil {
		return err
	}
	c.Lock()
	c.connected = true
	c.Unlock()

	go c.reconnect(secure, config)
	return nil
}

func (c *conn) tryConnect(secure bool, config *amqp.Config) error {
	var err error
	if config == nil {
		config = &defaultAmqpConfig
	}

	url := c.url

	if secure || config.TLSClientConfig != nil || strings.HasPrefix(c.url, "amqps://") {
		if config.TLSClientConfig == nil {
			config.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		}

		url = strings.Replace(c.url, "amqp://", "amqps://", 1)
	}

	c.Connection, err = dialConfig(url, *config)
	if err != nil {
		return err
	}

	if c.Channel, err = newChannel(c.Connection, c.prefetchCount, c.prefetchGlobal); err != nil {
		return err
	}

	err = c.Channel.DeclareExchange(c.exchange)
	if err != nil {
		return err
	}
	c.ExchangeChannel, err = newChannel(c.Connection, c.prefetchCount, c.prefetchGlobal)
	return err
}

func (c *conn) reconnect(secure bool, config *amqp.Config) {
	var connect bool
	for {
		if connect {
			if err := c.tryConnect(secure, config); err != nil {
				time.Sleep(1 * time.Second)
				continue
			}
			c.Lock()
			c.connected = true
			c.Unlock()
			close(c.waitConnection)
		}

		connect = true
		notifyClose := make(chan *amqp.Error)
		c.Connection.NotifyClose(notifyClose)
		select {
		case <-notifyClose:
			c.Lock()
			c.connected = false
			c.waitConnection = make(chan struct{})
			c.Unlock()
		case <-c.close:
			return
		}
	}
}

func (c *conn) Consume(queue *naming.Queue) (*channel, <-chan amqp.Delivery, error) {
	consumeChannel, err := newChannel(c.Connection, c.prefetchCount, c.prefetchGlobal)
	if err != nil {
		return nil, nil, err
	}
	err = consumeChannel.DeclareQueue(queue)
	if err != nil {
		return nil, nil, err
	}
	deliveries, err := consumeChannel.ConsumeQueue(queue, true)
	if err != nil {
		return nil, nil, err
	}
	err = consumeChannel.BindQueue(queue, c.exchange)
	if err != nil {
		return nil, nil, err
	}
	return consumeChannel, deliveries, nil
}

func (c *conn) Publish(exchange *naming.Exchange, key *naming.RoutingKey, msg amqp.Publishing) error {
	return c.ExchangeChannel.Publish(exchange, key, msg)
}

func (c *conn) Close() error {
	c.Lock()
	defer c.Unlock()

	select {
	case <-c.close:
		return nil
	default:
		close(c.close)
		c.connected = false
	}
	return c.Connection.Close()
}

func validUrl(url string) bool {
	return regexp.MustCompile("^amqp(s)?://.*").MatchString(url)
}
