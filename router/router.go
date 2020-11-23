package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gprcTest/reverse_proxy"
	"log"
	"time"

	"google.golang.org/grpc/health/grpc_health_v1"
	//"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/metadata"
	"net"
)

type healthServer struct{}

var stuckDuration time.Duration

func (h *healthServer) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	log.Println("recv health check for service:", req.Service)
	if stuckDuration == time.Second {
		return &grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_NOT_SERVING}, nil
	}
	return &grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_SERVING}, nil
}

func (h *healthServer) Watch(req *grpc_health_v1.HealthCheckRequest, stream grpc_health_v1.Health_WatchServer) error {
	log.Println("recv health watch for service:", req.Service)
	resp := new(grpc_health_v1.HealthCheckResponse)
	if stuckDuration == time.Second {
		resp.Status = grpc_health_v1.HealthCheckResponse_NOT_SERVING
	} else {
		resp.Status = grpc_health_v1.HealthCheckResponse_SERVING
	}
	for range time.NewTicker(time.Second).C {
		err := stream.Send(resp)
		if err != nil {
			return status.Error(codes.Canceled, "Stream has ended.")
		}
	}
	return nil
}

func main() {
	director := func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
		fmt.Println(fullMethodName)
		md, ok := metadata.FromIncomingContext(ctx)
		outCtx, _ := context.WithCancel(ctx)
		outCtx = metadata.NewOutgoingContext(outCtx, md.Copy())
		if ok {
			if val, exists := md[":authority"]; exists && val[0] == reverse_proxy.SERVICE_ID {

				conn, err := grpc.DialContext(ctx, "127.0.0.1:8080",
					//grpc.WithDefaultCallOptions(grpc.ForceCodec(codec)),
					grpc.WithCodec(reverse_proxy.NewGrpcCodec()),
					//grpc.WithCodec(grpcCod.Codec()),
					grpc.WithInsecure())

				return outCtx, conn, err
			}
		}
		return nil, nil, status.Errorf(codes.Unimplemented, "Unknown method")
	}
	server := grpc.NewServer(
		grpc.CustomCodec(reverse_proxy.NewGrpcCodec()),
		grpc.UnknownServiceHandler(reverse_proxy.ReverseProxyHandler(director)))
	grpc_health_v1.RegisterHealthServer(server, &healthServer{})
	lis, err := net.Listen("tcp", ":10000")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	server.Serve(lis)
}
