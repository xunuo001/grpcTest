package main

import (
	"gprcTest/reverse_proxy"
)

const port = ":10000"

func main() {
	reverse_proxy.StartProxy(port)
}
