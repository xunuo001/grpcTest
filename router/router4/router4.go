package main

import "gprcTest/reverse_proxy"

const port = ":40000"

func main() {
	reverse_proxy.StartProxy(port)
}
