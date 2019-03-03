package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/c0va23/go-proxyprotocol"
)

func main() {
	var addr string
	flag.StringVar(&addr, "bind", ":8080", "Bind address")
	flag.Parse()

	rawList, err := net.Listen("tcp", addr)
	if nil != err {
		log.Fatal(err)
	}

	list := proxyprotocol.NewDefaultListener(rawList).WithLogger(proxyprotocol.LoggerFunc(log.Printf))

	handler := http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		log.Printf("Remote Addr: %s, URI: %s", req.RemoteAddr, req.RequestURI)
		fmt.Fprintf(res, "Hello, %s!\n", req.RemoteAddr)
	})

	log.Printf("Start listen on %s", rawList.Addr())
	if err := http.Serve(list, handler); nil != err {
		log.Fatal(err)
	}
}
