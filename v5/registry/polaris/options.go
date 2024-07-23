package polaris

import (
	"context"

	"go-micro.dev/v5/registry"
)

type authKey struct{}

type nameSpaceKey struct{}

type serverTokenKey struct{}

type getOneInstanceKey struct{}

type authCreds struct {
	Username string
	Password string
}

// Auth allows you to specify username/password.
// just no effect.
func Auth(username, password string) registry.Option {
	return func(o *registry.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}

		o.Context = context.WithValue(o.Context, authKey{}, &authCreds{Username: username, Password: password})
	}
}

// NameSpace sets the namespace.
func NameSpace(namespace string) registry.Option {
	return func(o *registry.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}

		o.Context = context.WithValue(o.Context, nameSpaceKey{}, namespace)
	}
}

// ServerToken sets the server token.
func ServerToken(token string) registry.Option {
	return func(o *registry.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}

		o.Context = context.WithValue(o.Context, serverTokenKey{}, token)
	}
}

// GetOneInstance will fetch only one instance for use with Polaris's loadbalancer
// and disable cache by: github.com/micro/plugins/v5/selector/registry, option TTF(0).
func GetOneInstance(flag bool) registry.Option {
	return func(o *registry.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}

		o.Context = context.WithValue(o.Context, getOneInstanceKey{}, flag)
	}
}
