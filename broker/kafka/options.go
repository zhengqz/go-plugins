package kafka

import (
	"context"

	"github.com/Shopify/sarama"
	"github.com/micro/go-micro/broker"
	sc "gopkg.in/bsm/sarama-cluster.v2"
)

var (
	DefaultBrokerConfig  = sarama.NewConfig()
	DefaultClusterConfig = sc.NewConfig()
)

type brokerConfigKey struct{}
type clusterConfigKey struct{}

func BrokerConfig(c *sarama.Config) broker.Option {
	return func(o *broker.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, brokerConfigKey{}, c)
	}
}

func ClusterConfig(c *sc.Config) broker.Option {
	return func(o *broker.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, clusterConfigKey{}, c)
	}
}
