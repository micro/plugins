package handler

import (
	"context"
	"io"
	"net/http"
	"time"

	"go-micro.dev/v4/logger"

	pb "github.com/go-micro-v4-demo/frontend/proto"
	helloworldPb "github.com/go-micro-v4-demo/helloworld/proto"
	userPb "github.com/go-micro-v4-demo/user/proto"
)

type Frontend struct {
	UserService       userPb.UserService
	HelloworldService helloworldPb.HelloworldService
}

func (e *Frontend) Call(ctx context.Context, req *pb.CallRequest, rsp *pb.CallResponse) error {
	logger.Infof("Received Frontend.Call request: %v", req)
	rsp.Msg = "Hello " + req.Name
	return nil
}

func (e *Frontend) ClientStream(ctx context.Context, stream pb.Frontend_ClientStreamStream) error {
	var count int64
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			logger.Infof("Got %v pings total", count)
			return stream.SendMsg(&pb.ClientStreamResponse{Count: count})
		}
		if err != nil {
			return err
		}
		logger.Infof("Got ping %v", req.Stroke)
		count++
	}
}

func (e *Frontend) ServerStream(ctx context.Context, req *pb.ServerStreamRequest, stream pb.Frontend_ServerStreamStream) error {
	logger.Infof("Received Frontend.ServerStream request: %v", req)
	for i := 0; i < int(req.Count); i++ {
		logger.Infof("Sending %d", i)
		if err := stream.Send(&pb.ServerStreamResponse{
			Count: int64(i),
		}); err != nil {
			return err
		}
		time.Sleep(time.Millisecond * 250)
	}
	return nil
}

func (e *Frontend) BidiStream(ctx context.Context, stream pb.Frontend_BidiStreamStream) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		logger.Infof("Got ping %v", req.Stroke)
		if err := stream.Send(&pb.BidiStreamResponse{Stroke: req.Stroke}); err != nil {
			return err
		}
	}
}

func (e *Frontend) HomeHandler(w http.ResponseWriter, r *http.Request) {
	res, err := e.UserService.Call(r.Context(), &userPb.CallRequest{Name: "gsmini@sina.cn"})
	if err != nil {
		logger.Infof("Received userService.Call request: %v", err)
		w.Write([]byte("gsmini@sina.cn"))
	}

	w.Write([]byte(res.Msg))
}
