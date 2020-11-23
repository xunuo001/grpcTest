package reverse_proxy

import (
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/resolver"
)

func init() {
	resolver.Register(&mySoftBuilder{})
	resolver.SetDefaultScheme(REVERSE_NAME)
	encoding.RegisterCodec(NewEncodingCodec())
}
