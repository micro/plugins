package nacos

import (
	"context"

	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"go-micro.dev/v5/registry"
)

type addressKey struct{}
type configKey struct{}

// WithAddress sets the nacos address.
func WithAddress(addrs []string) registry.Option {
	return func(o *registry.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, addressKey{}, addrs)
	}
}

// WithClientConfig sets the nacos config.
func WithClientConfig(cc constant.ClientConfig) registry.Option {
	return func(o *registry.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, configKey{}, cc)
	}
}
