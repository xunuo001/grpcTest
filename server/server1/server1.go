package main

import "gprcTest/server"

const serverPort  = ":8080"
func main() {
	server.StartServer(serverPort)
}
