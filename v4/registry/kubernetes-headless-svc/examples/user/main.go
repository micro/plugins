package main

import (
	"github.com/go-micro-v4-demo/user/handler"
	pb "github.com/go-micro-v4-demo/user/proto"

	"go-micro.dev/v4"
	"go-micro.dev/v4/logger"
	regs "go-micro.dev/v4/registry"
)

var (
	service = "user"
	version = "latest"
)

func main() {
	// 因为我们不用第第三方服务注册发现
	// 所以这里是用内存注册，也就是不注册，完全靠k8s service实现多pod的动态发现
	reg := regs.NewMemoryRegistry()
	srv := micro.NewService()
	srv.Init(
		micro.Name(service),
		micro.Version(version),
		micro.Address("0.0.0.0:8080"),
		micro.Registry(reg),
	)

	// Register handler
	if err := pb.RegisterUserHandler(srv.Server(), new(handler.User)); err != nil {
		logger.Fatal(err)
	}
	// Run service
	if err := srv.Run(); err != nil {
		logger.Fatal(err)
	}
}
