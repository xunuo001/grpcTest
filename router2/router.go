package main

import (
	"context"
	"fmt"
	"github.com/mwitkow/grpc-proxy/proxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	//"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/metadata"
	"net"
	"strings"
)

func main() {
	//codec := newCodec()
	director := func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
		fmt.Println(fullMethodName)
		if strings.HasPrefix(fullMethodName, "/com.example.internal.") {
			return nil, nil, grpc.Errorf(codes.Unimplemented, "Unknown method")
		}
		md, ok := metadata.FromIncomingContext(ctx)
		outCtx, _ := context.WithCancel(ctx)
		outCtx = metadata.NewOutgoingContext(outCtx, md.Copy())
		if ok {
			if val, exists := md[":authority"]; exists && val[0] == "127.0.0.1:20000" {

				conn, err := grpc.DialContext(ctx, "127.0.0.1:8080",
					//grpc.WithDefaultCallOptions(grpc.ForceCodec(codec)),
					grpc.WithCodec(proxy.Codec()),
					//grpc.WithCodec(grpcCod.Codec()),
					grpc.WithInsecure())

				return outCtx, conn, err
			} else if val, exists := md[":authority"]; exists && val[0] == "api.example.com" {
				conn, err := grpc.DialContext(ctx, "api-service.prod.svc.local", grpc.WithCodec(proxy.Codec()))
				return outCtx, conn, err
			}
		}
		return nil, nil, grpc.Errorf(codes.Unimplemented, "Unknown method")
	}
	server := grpc.NewServer(
		//encoding.RegisterCodec(proxy.Codec()),
		grpc.CustomCodec(proxy.Codec()),
		grpc.UnknownServiceHandler(proxy.TransparentHandler(director)))
	lis, err := net.Listen("tcp", ":20000")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	server.Serve(lis)
}
