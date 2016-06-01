package main

import (
	"github.com/30x/gozerian/go_gateway"
	. "github.com/30x/gozerian/handlers"
	. "github.com/30x/gozerian/pipeline"
	"net/http"
	"log"
	"flag"
	"net/url"
)

var sourcePort string
var targetPort string

func init() {
	flag.StringVar(&sourcePort, "sourcePort", "3000", "sourcePort")
	flag.StringVar(&targetPort, "targetPort", "3001", "targetPort")
}

func main() {
	flag.Parse()

	target, err := url.Parse("http://localhost:" + targetPort)
	if err != nil {
		panic(err)
	}

	requestHandlers := []http.HandlerFunc{
		RequestDumper(true),
	}
	responseHandlers := []ResponseHandlerFunc{
		ResponseDumper(true),
	}

	pipeDef, err := DefinePipe(requestHandlers, responseHandlers)
	if err != nil {
		panic(err)
	}

	proxyHandler := go_gateway.ReverseProxyHandler{pipeDef, target}

	log.Printf("Gateway routing port %s => %s", sourcePort, target)

	http.HandleFunc("/", proxyHandler.ServeHTTP)
	err = http.ListenAndServe(":" + sourcePort, nil)

	log.Printf("Server crashed: %v.", err)
}
