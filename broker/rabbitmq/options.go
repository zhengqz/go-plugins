package rabbitmq

import (
	"context"
	"github.com/micro/go-micro/broker"
)

type durableQueueKey struct{}
type headersKey struct{}
type prefetchCountKey struct{}
type prefetchGlobalKey struct{}
type exchangeKey struct{}
type requeueOnErrorKey struct{}
type deliveryMode struct{}

// DurableQueue creates a durable queue when subscribing.
func DurableQueue() broker.SubscribeOption {
	return func(o *broker.SubscribeOptions) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, durableQueueKey{}, true)
	}
}

// Headers adds headers used by the headers exchange
func Headers(h map[string]interface{}) broker.SubscribeOption {
	return func(o *broker.SubscribeOptions) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, headersKey{}, h)
	}
}

// RequeueOnError calls Nack(muliple:false, requeue:true) on amqp delivery when handler returns error
func RequeueOnError() broker.SubscribeOption {
	return func(o *broker.SubscribeOptions) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, requeueOnErrorKey{}, true)
	}
}

// Exchange is an option to set the Exchange
func Exchange(e string) broker.Option {
	return func(o *broker.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, exchangeKey{}, e)
	}
}

// PrefetchCount ...
func PrefetchCount(c int) broker.Option {
	return func(o *broker.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, prefetchCountKey{}, c)
	}
}

// PrefetchGlobal creates a durable queue when subscribing.
func PrefetchGlobal() broker.Option {
	return func(o *broker.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, prefetchGlobalKey{}, true)
	}
}

// DeliveryMode sets a delivery mode for publishing
func DeliveryMode(value uint8) broker.PublishOption {
	return func(o *broker.PublishOptions) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, deliveryMode{}, value)
	}
}
