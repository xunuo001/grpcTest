package reverse_proxy

import "github.com/golang/protobuf/proto"

type frame struct {
	payload []byte
}

const PROTO_NAME = "proto"

// protoCodec is a Codec implementation with protobuf. It is the default rawCodec for gRPC.
type protoCodec struct{}

func (protoCodec) Marshal(v interface{}) ([]byte, error) {
	return proto.Marshal(v.(proto.Message))
}

func (protoCodec) Unmarshal(data []byte, v interface{}) error {
	return proto.Unmarshal(data, v.(proto.Message))
}

func (protoCodec) String() string {
	return PROTO_NAME
}
func (protoCodec) Name() string {
	return PROTO_NAME
}
