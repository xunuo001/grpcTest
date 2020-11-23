package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	_ "google.golang.org/grpc/health"
	"google.golang.org/grpc/keepalive"
	"gprcTest/domain"
	"io"
	"time"
)

const serverPort = "10000"

func main() {
	//resolver.SetDefaultScheme(reverse_proxy.REVERSE_NAME)
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

	conn, err := grpc.Dial("127.0.0.1:8080",opts...)
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	defer conn.Close()

	client := domain.NewHelloWorldServiceClient(conn)

	//
	fmt.Println("一元 RPC")
	resp, err := client.Hello(context.Background(), &domain.HelloRequest{
		Name: "xxxx",
	})
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(resp)
	//
	fmt.Println("request stream ")
	stream, err := client.HelloReqStream(context.Background())
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	r := &domain.HelloRequest{
		Name: "stream req test",
	}
	for i := 0; i < 10; i++ {
		stream.Send(r)
	}
	resp, err = stream.CloseAndRecv()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("HelloReqStream",resp)
	//
	fmt.Println("response stream ")

	res, err :=client.HelloResStream(context.Background(), &domain.HelloRequest{Name: "response stream"})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	for {
		resp, err := res.Recv()
		if err == io.EOF {
			break
		}
		fmt.Println("HelloResStream",resp)
	}
	//
	fmt.Println("all stream ")
	alres,err:=client.HelloStream(context.Background())
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	for i:=0;i<10;i++{
		alres.Send(&domain.HelloRequest{Name: "all"})
		respa,err:=alres.Recv()
		if err==io.EOF{
			break
		}
		fmt.Println("all",respa)
	}
	alres.CloseSend()
}
