// Package mucp provides an mucp client
package mucp

import (
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/util/cmd"
)

func init() {
	cmd.DefaultClients["mucp"] = NewClient
}

// NewClient returns a new micro client interface
func NewClient(opts ...client.Option) client.Client {
	return client.NewClient(opts...)
}
