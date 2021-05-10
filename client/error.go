package client

import (
	"errors"
	"fmt"
)

var (
	ErrChannelNil    = errors.New("channel is nil")
	ErrExchangeNil   = errors.New("exchange is nil")
	ErrRoutingKeyNil = errors.New("routingKey is nil")
	ErrConnect       = errors.New("connect failed")
	ErrConnNil       = errors.New("connection is nil")
)

func amqpError(action string, err error) error {
	if err != nil {
		return fmt.Errorf("%s error: %v", action, err)
	}
	return nil
}
