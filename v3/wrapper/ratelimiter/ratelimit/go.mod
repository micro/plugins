module github.com/go-micro/plugins/v3/wrapper/ratelimiter/ratelimit/v3

go 1.16

require (
	github.com/asim/go-micro/v3 v3.7.1
	github.com/go-micro/plugins/v3/broker/memory v1.0.0
	github.com/go-micro/plugins/v3/registry/memory v1.0.0
	github.com/go-micro/plugins/v3/transport/memory v1.0.0
	github.com/juju/ratelimit v1.0.2
)

replace (
	github.com/go-micro/plugins/v3/broker/memory => ../../../broker/memory
	github.com/go-micro/plugins/v3/registry/memory => ../../../registry/memory
	github.com/go-micro/plugins/v3/transport/memory => ../../../transport/memory
)
