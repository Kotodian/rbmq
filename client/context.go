package client

import (
	"context"
	"github.com/streadway/amqp"
)

type delivery struct{}

func DeliveryFromContext(ctx context.Context) amqp.Delivery {
	return ctx.Value(delivery{}).(amqp.Delivery)
}

func ContextWithDelivery(ctx context.Context, d amqp.Delivery) context.Context {
	ctx = context.WithValue(ctx, delivery{}, d)
	return ctx
}
