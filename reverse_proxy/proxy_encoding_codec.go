package reverse_proxy

import (
	"fmt"
	"google.golang.org/grpc/encoding"
)

func NewEncodingCodec() encoding.Codec {
	return encodingCodecWithParent(&protoCodec{})
}

// EncodingCodecWithParent returns a proxying grpc.Codec with a user provided codec as parent.
//
// This codec is *crucial* to the functioning of the proxy. It allows the proxy server to be oblivious
// to the schema of the forwarded messages. It basically treats a gRPC message frame as raw bytes.
// However, if the server handler, or the client caller are not proxy-internal functions it will fall back
// to trying to decode the message using a fallback codec.
func encodingCodecWithParent(fallback encoding.Codec) encoding.Codec {
	return &encordingRawCodec{fallback}
}

type encordingRawCodec struct {
	parentCodec encoding.Codec
}

func (c *encordingRawCodec) Marshal(v interface{}) ([]byte, error) {
	out, ok := v.(*frame)
	if !ok {
		return c.parentCodec.Marshal(v)
	}
	return out.payload, nil

}

func (c *encordingRawCodec) Unmarshal(data []byte, v interface{}) error {
	dst, ok := v.(*frame)
	if !ok {
		return c.parentCodec.Unmarshal(data, v)
	}
	dst.payload = data
	return nil
}

func (c *encordingRawCodec) Name() string {
	return fmt.Sprintf("proxy>%s", c.parentCodec.Name())
}

