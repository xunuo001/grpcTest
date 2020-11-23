package main

import "gprcTest/reverse_proxy"

const port = ":20000"

func main() {
	reverse_proxy.StartProxy(port)
}
