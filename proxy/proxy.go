package main

import (
	"crypto/tls"
	"fmt"
	"golang.org/x/net/http2"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	url, err := url.Parse("http://127.0.0.1:8080")
	if err!=nil{
		fmt.Println(err.Error())
		return
	}
	proxy := &Upstream{
		target: url,
		proxy:  httputil.NewSingleHostReverseProxy(url),
	}
	mux := http.NewServeMux()
	mux.HandleFunc(".", proxy.handle)
	log.Fatal(http.ListenAndServe(":10000", mux))
}

type Upstream struct {
	target *url.URL
	proxy  *httputil.ReverseProxy
}

func (p *Upstream) handle(w http.ResponseWriter, r *http.Request) {
	fmt.Print("here")
	w.Header().Set("X-Forwarded-For", r.Host)
	p.proxy.Transport = &http2.Transport{
		AllowHTTP: true,
		DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
			ta, err := net.ResolveTCPAddr(network, addr)
			if err != nil {
				return nil, err
			}
			return net.DialTCP(network, nil, ta)
		},
	}
	p.proxy.ServeHTTP(w, r)
}
