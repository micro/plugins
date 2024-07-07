// Package mucp provides an mucp server
package mucp

import (
	"go-micro.dev/v5/server"
	"go-micro.dev/v5/cmd"
)

func init() {
	cmd.DefaultServers["mucp"] = NewServer
}

// NewServer returns a micro server interface.
func NewServer(opts ...server.Option) server.Server {
	return server.NewServer(opts...)
}
