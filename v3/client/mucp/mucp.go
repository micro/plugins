// Package mucp provides an mucp client
package mucp

import (
	"github.com/asim/go-micro/v3/client"
	"github.com/asim/go-micro/v3/cmd"
)

func init() {
	cmd.DefaultClients["mucp"] = NewClient
}

// NewClient returns a new micro client interface
func NewClient(opts ...client.Option) client.Client {
	return client.NewClient(opts...)
}
