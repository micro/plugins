package stream

import "crypto/tls"

// Options which are used to configure the redis stream
type Options struct {
	Address   string
	User      string
	Password  string
	TLSConfig *tls.Config
}

// Option is a function which configures options
type Option func(o *Options)

// Address sets the Redis address option.
// Needs to be a full URL with scheme (redis://, rediss://, unix://).
// (eg. redis://user:password@localhost:6789/3?dial_timeout=3).
// Alternatively, the address can simply be the `host:port` format
// where User, Password, TLSConfig are defined with their respective options.
func Address(addr string) Option {
	return func(o *Options) {
		o.Address = addr
	}
}

func User(user string) Option {
	return func(o *Options) {
		o.User = user
	}
}

func Password(password string) Option {
	return func(o *Options) {
		o.Password = password
	}
}

func TLSConfig(tlsConfig *tls.Config) Option {
	return func(o *Options) {
		o.TLSConfig = tlsConfig
	}
}
