module github.com/go-micro/plugins/v3/wrapper/monitoring/prometheus

go 1.16

require (
	github.com/asim/go-micro/v3 v3.7.1
	github.com/go-micro/plugins/v3/broker/memory v0.0.0
	github.com/go-micro/plugins/v3/registry/memory v0.0.0
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/client_model v0.2.0
	github.com/stretchr/testify v1.7.0
)

replace (
	github.com/go-micro/plugins/v3/broker/memory => ../../../broker/memory
	github.com/go-micro/plugins/v3/registry/memory => ../../../registry/memory
)
