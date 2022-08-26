module github.com/go-micro/plugins/v3/server/grpc

go 1.16

require (
	github.com/asim/go-micro/v3 v3.7.1
	github.com/go-micro/plugins/v3/broker/memory v1.0.0
	github.com/go-micro/plugins/v3/client/grpc v1.0.0
	github.com/go-micro/plugins/v3/transport/grpc v1.0.0
	github.com/golang/protobuf v1.5.2
	golang.org/x/net v0.0.0-20220822230855-b0a4917ee28c
	google.golang.org/genproto v0.0.0-20220822174746-9e6da59bd2fc
	google.golang.org/grpc v1.49.0
)

replace (
	github.com/go-micro/plugins/v3/broker/memory => ../../broker/memory
	github.com/go-micro/plugins/v3/client/grpc => ../../client/grpc
	github.com/go-micro/plugins/v3/transport/grpc => ../../transport/grpc
)
