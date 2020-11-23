package main

import "gprcTest/server"

const serverPort  = ":8081"
func main() {
	server.StartServer(serverPort)
}
