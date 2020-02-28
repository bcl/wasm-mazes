package main

import (
	"flag"
	"fmt"
	"net/http"
)

func main() {
	ip := "127.0.0.1"
	port := 8000

	flag.StringVar(&ip, "ip", ip, "IP Address to Listen to")
	flag.IntVar(&port, "port", port, "Port to listen to")

	http.Handle("/", http.FileServer(http.Dir("./web")))
	listen := fmt.Sprintf("%s:%d", ip, port)

	if err := http.ListenAndServe(listen, nil); err != nil {
		panic(err)
	}
}
