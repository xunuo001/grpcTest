package reverse_proxy

import (
	"google.golang.org/grpc/resolver"
)

const REVERSE_NAME = "my_schema_name"
const (
	ROUTER_1   = "router1"
	ROUTER_2   = "router2"
	SERVICE_ID = "service"
)

//设置转发规则，两层反向代理
var addrsStore = map[string][]string{
	ROUTER_1:   {"localhost:10000", "localhost:20000"},//第一层反向代理地址
	ROUTER_2:   {"localhost:30000", "localhost:40000"},//第二层反向代理地址
	SERVICE_ID: {"localhost:8080", "localhost:8081"},//具体服务地址
}
//设置转发规则，两层反向代理。client->router1->router2-server
var routeMapping = map[string]string{
	ROUTER_1: SERVICE_ID,
}

//func NewBuilder() resolver.Builder {
//	return &mySoftBuilder{}
//}

type mySoftBuilder struct {
}

func (b *mySoftBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := &mySoftResolver{
		target:     target,
		cc:         cc,
		addrsStore: addrsStore,
	}
	r.start()
	return r, nil
}
func (b *mySoftBuilder) Scheme() string {
	return REVERSE_NAME
}

func (b *mySoftBuilder) watcher() {

}

type mySoftResolver struct {
	target     resolver.Target
	cc         resolver.ClientConn
	addrsStore map[string][]string
}

func (*mySoftResolver) ResolveNow(o resolver.ResolveNowOptions) {
}

// Close closes the resolver.
func (*mySoftResolver) Close() {}
func (r *mySoftResolver) start() {
	addrStrs := r.addrsStore[r.target.Endpoint]
	addrs := make([]resolver.Address, len(addrStrs))
	for i, s := range addrStrs {
		addrs[i] = resolver.Address{Addr: s}
	}
	r.cc.UpdateState(resolver.State{Addresses: addrs})
}
