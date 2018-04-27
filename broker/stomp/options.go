package stomp

import (
	"context"
	"time"

	"github.com/micro/go-micro/broker"
)

// setBrokerPublishOptionsContextValueFunc returns a function to setup a context with given value
func setBrokerSubscribeOptionsContextValueFunc(k, v interface{}) broker.SubscribeOption {
	return func(o *broker.SubscribeOptions) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, k, v)
	}
}

type subscribeHeaderKey struct{}

// SubscribeHeaders sets headers for subscriptions
func SubscribeHeaders(h map[string]string) broker.SubscribeOption {
	return setBrokerSubscribeOptionsContextValueFunc(subscribeHeaderKey{}, h)
}

type durableQueueKey struct{}

// Durable sets a durable subscription
func Durable() broker.SubscribeOption {
	return setBrokerSubscribeOptionsContextValueFunc(durableQueueKey{}, true)
}

// setBrokerPublishOptionsContextValueFunc returns a function to setup a context with given value
func setBrokerPublishOptionsContextValueFunc(k, v interface{}) broker.PublishOption {
	return func(o *broker.PublishOptions) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, k, v)
	}
}

type receiptKey struct{}

// Receipt requests a receipt for the delivery should be received
func Receipt(ct time.Duration) broker.PublishOption {
	return setBrokerPublishOptionsContextValueFunc(receiptKey{}, true)
}

type suppressContentLengthKey struct{}

// SuppressContentLength requests that send does not include a content length
func SuppressContentLength(ct time.Duration) broker.PublishOption {
	return setBrokerPublishOptionsContextValueFunc(suppressContentLengthKey{}, true)
}

// setBrokerOptionsContextValueFunc returns a function to setup a context with given value
func setBrokerOptionsContextValueFunc(k, v interface{}) broker.Option {
	return func(o *broker.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, k, v)
	}
}

// Connection Timeout
type connectTimeoutKey struct{}

// ConnectTimeout sets the connection timeout duration
func ConnectTimeout(ct time.Duration) broker.Option {
	return setBrokerOptionsContextValueFunc(connectTimeoutKey{}, ct)
}

// Authentication
type authKey struct{}
type authRecord struct {
	username string
	password string
}

// Auth sets the authentication information
func Auth(username string, password string) broker.Option {
	return setBrokerOptionsContextValueFunc(authKey{}, &authRecord{
		username: username,
		password: password,
	})
}

// Connect header
type connectHeaderKey struct{}

// ConnectHeaders adds headers for the connection
func ConnectHeaders(h map[string]string) broker.Option {
	return setBrokerOptionsContextValueFunc(connectHeaderKey{}, h)
}

// vHost header
type vHostKey struct{}

// vHost adds host header to define the vhost
func VirtualHost(h string) broker.Option {
	return setBrokerOptionsContextValueFunc(vHostKey{}, h)
}
