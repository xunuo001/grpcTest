package reverse_proxy

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
)

var (
	clientStreamDescForProxying = &grpc.StreamDesc{
		ServerStreams: true,
		ClientStreams: true,
	}
)

type proxyDirect func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error)

func ReverseProxyHandler(director proxyDirect) grpc.StreamHandler {
	streamer := &handler{director: director}
	return streamer.handler
}

type handler struct {
	director proxyDirect
}

func (h *handler) handler(srv interface{}, serverStream grpc.ServerStream) error {

	fullMethodName, ok := grpc.MethodFromServerStream(serverStream)
	if !ok {
		return status.Errorf(codes.Internal, "lowLevelServerStream not exists in context")
	}
	outgoingCtx, backendConn, err := h.director(serverStream.Context(), fullMethodName)
	if err != nil {
		return err
	}
	clientCtx, clientCancel := context.WithCancel(outgoingCtx)
	clientStream, err := grpc.NewClientStream(clientCtx, clientStreamDescForProxying, backendConn, fullMethodName)
	if err != nil {
		return err
	}
	s2cErrChan := h.forwardServerToClient(serverStream, clientStream)
	c2sErrChan := h.forwardClientToServer(clientStream, serverStream)
	for i := 0; i < 2; i++ {
		select {
		case s2cErr := <-s2cErrChan:
			if s2cErr == io.EOF {
				//当数据发送完成后，关闭客户端流
				clientStream.CloseSend()
				break
			} else {
				//当收到其他错误时，则取消对远端的请求，并释放goroutines
				clientCancel()
				return status.Errorf(codes.Internal, "failed proxying s2c: %v", s2cErr)
			}
		case c2sErr := <-c2sErrChan:
			//如果发生错误，则把trailer返回给服务端
			serverStream.SetTrailer(clientStream.Trailer())
			// 如果不是eof错误则返回对应错误信息
			if c2sErr != io.EOF {
				return c2sErr
			}
			return nil
		}
	}
	return grpc.Errorf(codes.Internal, "gRPC proxying should never reach this stage.")
}
//数据从客户端到服务端，请求方的数据到目的方
func (h *handler) forwardClientToServer(src grpc.ClientStream, dst grpc.ServerStream) chan error {
	ret := make(chan error, 1)
	go func() {
		f := &frame{}
		for i := 0; ; i++ {
			if err := src.RecvMsg(f); err != nil {
				ret <- err
				break
			}
			if i == 0 {
				// grpc中客户端到服务器的header只能在第一个客户端消息后才可以读取到，
				// 同时又必须在flush第一个msg之前写入到流中。
				md, err := src.Header()
				if err != nil {
					ret <- err
					break
				}
				if err := dst.SetHeader(md); err != nil {
					ret <- err
					break
				}
			}
			if err := dst.SendMsg(f); err != nil {
				ret <- err
				break
			}
		}

	}()
	return ret
}
//数据从服务端到客户端，目的方的数据到请求方
func (h *handler) forwardServerToClient(src grpc.ServerStream, dst grpc.ClientStream) chan error {
	ret := make(chan error, 1)
	go func() {
		f := &frame{}
		for i := 0; ; i++ {
			if err := src.RecvMsg(f); err != nil {
				ret <- err
				break
			}
			if err := dst.SendMsg(f); err != nil {
				ret <- err
				break
			}
		}
	}()
	return ret
}
