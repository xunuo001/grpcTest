package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	_ "google.golang.org/grpc/health"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/resolver"
	"gprcTest/domain"
	"gprcTest/reverse_proxy"
	"log"
	"time"
)

const serverPort = "10000"

func main() {
	resolver.SetDefaultScheme(reverse_proxy.REVERSE_NAME)
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		// 设置阻塞等待连接成功
		grpc.WithBlock(),
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
				"HealthCheckConfig": {"ServiceName": "client"}}`, roundrobin.Name),
		),
	}

	conn, err := grpc.Dial(reverse_proxy.ROUTER_1, opts...)
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	defer conn.Close()

	client := domain.NewHelloWorldServiceClient(conn)

	context.Background()

	for range time.Tick(time.Second) {
		md:=metadata.Pairs("testMetadata","test")
		ctx:=metadata.NewOutgoingContext(context.Background(),md)
		ctx, cancel := context.WithTimeout(ctx, time.Duration(10)*time.Second)
		resp, err := client.Hello(ctx, &domain.HelloRequest{
			Name: "xxxx",
		})
		if err != nil {
			log.Printf("could not greet: %v\n", err)
		} else {
			log.Printf("Greeting: %s", resp.Msg)
		}
		cancel()
	}

	//resp, err := client.Hello(context.Background(), &domain.HelloRequest{
	//	Name: "xxxx",
	//})
	//if err != nil {
	//	fmt.Println(err.Error())
	//	return
	//}
	//
	//fmt.Println(resp)
}
