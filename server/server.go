package server

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"gprcTest/domain"
	"gprcTest/reverse_proxy"
	"io"
	"net"
)

type HelloworldService struct {
	address string
	domain.UnimplementedHelloWorldServiceServer
}

func (h *HelloworldService) Hello(c context.Context, d *domain.HelloRequest) (*domain.HelloResponse, error) {
	fmt.Println("get hell request name is :", d.Name)
	return &domain.HelloResponse{Msg: &domain.Message{Name: "response:" + d.Name, Msg: "message:" + d.Name + " server address: " + h.address}}, nil
}

func (*HelloworldService) HelloReqStream(stream domain.HelloWorldService_HelloReqStreamServer) error {
	for {
		r, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&domain.HelloResponse{
				Msg: &domain.Message{Name: "HelloReqStream", Msg: "closed stream"},
			})
		}
		fmt.Println("HelloReqStream", r)
	}
}
func (*HelloworldService) HelloResStream(req *domain.HelloRequest, stream domain.HelloWorldService_HelloResStreamServer) error {
	for i := 0; i < 10; i++ {
		stream.Send(&domain.HelloResponse{
			Msg: &domain.Message{Name: "HelloResStream", Msg: "loop send"},
		})
	}
	return nil
}
func (*HelloworldService) HelloStream(stream domain.HelloWorldService_HelloStreamServer) error {
	for {
		stream.Send(&domain.HelloResponse{Msg: &domain.Message{Name: "all",Msg: "all res msg"}})
		r, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		fmt.Println("HelloReqStream", r)
	}
	return nil
}

func StartServer(address string) {
	server := grpc.NewServer()
	domain.RegisterHelloWorldServiceServer(server, &HelloworldService{address: address})
	grpc_health_v1.RegisterHealthServer(server, &reverse_proxy.HealthServer{})
	lis, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	server.Serve(lis)
}
