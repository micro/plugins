module github.com/go-micro/plugins/v2/micro/metrics

go 1.17

require (
	github.com/go-micro/plugins/v2/micro/metrics/prometheus v0.0.0-20220823021029-ced7724b5d15
	github.com/micro/cli/v2 v2.1.2
	github.com/micro/go-micro/v2 v2.9.1
	github.com/micro/micro/v2 v2.9.3
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/prometheus/client_golang v1.13.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	golang.org/x/sys v0.0.0-20220825204002-c680a09ffe64 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
)

replace (
	github.com/go-micro/plugins/v2/micro/metrics/prometheus => ./prometheus
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)
