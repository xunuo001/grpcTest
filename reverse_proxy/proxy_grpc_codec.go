package reverse_proxy

import (
	"fmt"
	"google.golang.org/grpc"
)

func NewGrpcCodec() grpc.Codec {
	return grpcCodecWithParent(&protoCodec{})
}
func grpcCodecWithParent(fallback grpc.Codec) grpc.Codec {
	return &grpcRawCodec{fallback}
}

type grpcRawCodec struct {
	parentCodec grpc.Codec
}
// 序列化函数，
// 尝试将消息转换为*frame类型，并返回frame的payload实现序列化
// 若失败，则采用变量parentCodec中的Marshal进行序列化
func (c *grpcRawCodec) Marshal(v interface{}) ([]byte, error) {
	out, ok := v.(*frame)
	if !ok {
		return c.parentCodec.Marshal(v)
	}
	return out.payload, nil
}
// 反序列化函数，
// 尝试通过将消息转为*frame类型，提取出payload到[]byte，实现反序列化
// 若失败，则采用变量parentCodec中的Unmarshal进行反序列化
func (c *grpcRawCodec) Unmarshal(data []byte, v interface{}) error {
	dst, ok := v.(*frame)
	if !ok {
		return c.parentCodec.Unmarshal(data, v)
	}
	dst.payload = data
	return nil
}

func (c *grpcRawCodec) String() string {
	return fmt.Sprintf("proxy>%s", c.parentCodec.String())
}
