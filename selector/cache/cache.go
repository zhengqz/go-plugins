// Package cache is a caching selector (the default selector)
package cache

import (
	"github.com/micro/go-micro/selector"
)

// NewSelector returns a new caching selector using the go-micro registry
func NewSelector(opts ...selector.Option) selector.Selector {
	return selector.NewSelector(opts...)
}
