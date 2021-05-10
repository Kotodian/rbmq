package client

import (
	"errors"
	"github.com/streadway/amqp"
	"rabbitmq/naming"
)

type channel struct {
	connection *amqp.Connection
	channel    *amqp.Channel
}

func newChannel(conn *amqp.Connection, prefetchCount int, prefetchGlobal bool) (*channel, error) {
	ch := &channel{connection: conn}
	if err := ch.Connect(prefetchCount, prefetchGlobal); err != nil {
		return nil, ErrConnect
	}
	return ch, nil
}

func (c *channel) Connect(prefetchCount int, prefetchGlobal bool) error {
	var err error
	c.channel, err = c.connection.Channel()
	if err != nil {
		return err
	}
	err = c.channel.Qos(prefetchCount, 0, prefetchGlobal)
	if err != nil {
		return err
	}
	return nil
}

func (c *channel) Close() error {
	if c.channel == nil {
		return errors.New("channel is nil")
	}
	return c.channel.Close()
}

func (c *channel) Publish(
	exchange *naming.Exchange,
	key *naming.RoutingKey,
	message amqp.Publishing) error {
	if c.channel == nil {
		return ErrChannelNil
	}
	if exchange == nil {
		return ErrExchangeNil
	}

	if key == nil {
		return ErrRoutingKeyNil
	}
	return amqpError("publish", c.channel.Publish(exchange.String(true), key.String(true), false, false, message))
}

func (c *channel) DeclareExchange(exchange *naming.Exchange) error {
	return amqpError("declare_exchange", c.channel.ExchangeDeclare(
		exchange.String(true),
		exchange.Kind(),
		true,
		false,
		false,
		false,
		nil,
	))
}

func (c *channel) DeclareQueue(queue *naming.Queue) error {
	_, err := c.channel.QueueDeclare(
		queue.String(true),
		true,
		false,
		false,
		false,
		nil)
	return amqpError("declare_queue", err)
}

func (c *channel) BindQueue(queue *naming.Queue, exchange *naming.Exchange) error {
	return amqpError("bind_queue", c.channel.QueueBind(
		queue.String(true),
		queue.RoutingKey().String(true),
		exchange.String(true),
		false,
		nil))
}

func (c *channel) ConsumeQueue(queue *naming.Queue, autoAck bool) (<-chan amqp.Delivery, error) {
	delivery, err := c.channel.Consume(
		queue.String(true),
		queue.RoutingKey().Consumer().String(),
		autoAck,
		false,
		false,
		false,
		nil)
	return delivery, amqpError("consume", err)
}
