package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8088"
	}

	fmt.Printf("Server will start on port %s\n", port)
	log.Println("うんちんぐすぽっと API server initialized")
}
