package reverse_proxy

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/codes"
	_ "google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"net"
	"time"
)

var opts = []grpc.DialOption{
	grpc.WithInsecure(),
	// 设置阻塞等待连接成功
	grpc.WithBlock(),
	grpc.WithCodec(NewGrpcCodec()),
	//grpc.WithDefaultCallOptions(grpc.ForceCodec(codec)),
	grpc.WithDefaultCallOptions(
		// 设置最大接收/发送消息的大小, 当前为10M
		grpc.MaxCallRecvMsgSize(100*1024*1024),
		grpc.MaxCallSendMsgSize(100*1024*1024),
		// 设置使用snappy压缩
		//grpc.UseCompressor(Snappy),
	),
	// 设置keepalive相关
	grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:                time.Duration(1) * time.Minute,
		Timeout:             time.Duration(20) * time.Second,
		PermitWithoutStream: true,
	}),
	grpc.WithDefaultServiceConfig(
		fmt.Sprintf(`{"LoadBalancingPolicy": "%s",
				"MethodConfig": [{
				"RetryPolicy": {"MaxAttempts":2, "InitialBackoff": "0.1s", "MaxBackoff": "1s", "BackoffMultiplier": 2.0, 
					"RetryableStatusCodes": ["UNAVAILABLE", "CANCELLED"]}}], 
				"HealthCheckConfig": {"ServiceName": "router"}}`, roundrobin.Name),
	),
}

func StartProxy(address string) {
	director := func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
		fmt.Println(fullMethodName)
		md, ok := metadata.FromIncomingContext(ctx)
		outCtx, _ := context.WithCancel(ctx)
		outCtx = metadata.NewOutgoingContext(outCtx, md.Copy())
		if ok {
			if val, exists := md[":authority"]; exists {
				if val, exists := routeMapping[val[0]]; exists {
					conn, err := grpc.DialContext(ctx, val, opts...)
					return outCtx, conn, err
				}
				return nil, nil, status.Errorf(codes.Unimplemented, val[0])
			}
		}
		return nil, nil, status.Errorf(codes.Unimplemented, "Unknown method")
	}
	server := grpc.NewServer(
		grpc.CustomCodec(NewGrpcCodec()),
		grpc.UnknownServiceHandler(ReverseProxyHandler(director)))

	grpc_health_v1.RegisterHealthServer(server, &HealthServer{})

	lis, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	server.Serve(lis)
}
