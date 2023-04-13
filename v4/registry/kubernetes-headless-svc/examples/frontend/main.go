package main

import (
	"fmt"
	"github.com/go-micro-v4-demo/frontend/handler"
	helloworldPb "github.com/go-micro-v4-demo/helloworld/proto"
	userPb "github.com/go-micro-v4-demo/user/proto"
	mgrpc "github.com/go-micro/plugins/v4/client/grpc"
	mhttp "github.com/go-micro/plugins/v4/server/http"
	"github.com/gorilla/mux"
	k8sHeadlessSvc "github.com/gsmini/k8s-headless-svc"
	"go-micro.dev/v4/logger"
	"net/http"
)

var (
	service = "frontend"
	version = "latest"
)

const K8sSvcName = "user-svc"

const UserSvcName = "user-svc"        //user模块在k8s service中的metadata.name的名字
const HelloWordSvcName = "helloworld" //user模块在k8s service中的metadata.name的名字
func main() {
	UserSvc := &k8sHeadlessSvc.Service{Namespace: "default", SvcName: UserSvcName, PodPort: 8080}
	//HelloWordSvc := &k8sHeadlessSvc.Service{Namespace: "default", SvcName: HelloWordSvcName, PodPort: 9090}
	reg := k8sHeadlessSvc.NewRegistry([]*k8sHeadlessSvc.Service{UserSvc})
	// 当前frontend调用依赖多个grpc 上游服务器的情况
	//reg := k8sHeadlessSvc.NewRegistry([]*k8sHeadlessSvc.Service{UserSvc},[]*k8sHeadlessSvc.Service{HelloWordSvcName})
	srv := micro.NewService(
		micro.Server(mhttp.NewServer()), //当前服务的类型 http 对外提供http
		micro.Client(mgrpc.NewClient())) //当前client的类型grpc 对内调用grpc
	srv.Init(
		micro.Name(service),
		micro.Version(version),
		micro.Address("0.0.0.0:8080"), //对外暴漏8000端口
		micro.Registry(reg),
	)
	client := srv.Client()
	svc := &handler.Frontend{
		UserService:       userPb.NewUserService(UserSvcName, client),
		HelloworldService: helloworldPb.NewHelloworldService(HelloWordSvcName, client),
	}
	r := mux.NewRouter()
	r.HandleFunc("/index", svc.HomeHandler).Methods(http.MethodGet)
	//反爬虫
	r.HandleFunc("/robots.txt", func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, "User-agent: *\nDisallow: /") })
	//健康检查
	r.HandleFunc("/_healthz", func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, "ok") })

	var httpHandler http.Handler = r
	// Register handler
	if err := micro.RegisterHandler(srv.Server(), httpHandler); err != nil {
		logger.Fatal(err)
	}
	// Run service
	if err := srv.Run(); err != nil {
		logger.Fatal(err)
	}
}
