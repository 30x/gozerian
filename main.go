package main

import (
	"github.com/30x/gozerian/go_gateway"
	"github.com/30x/gozerian/pipeline"
	"net/http"
	"os"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"fmt"
	"net/url"
	"bytes"
	"github.com/30x/gozerian/test_util"
)

// This is just an example using go_gateway. Config via main.yaml.
// todo: create a better example and refactor a better config loader
//       eventually, I'll just move all this to another repo

func main() {
	pipeline.RegisterDie("dump", test_util.CreateDumpFitting)

	yamlReader, err := os.Open("main.yaml")
	if err != nil {
		panic(err)
	}

	yamlIn, err := ioutil.ReadAll(yamlReader)
	if err != nil {
		panic(err)
	}
	gateway := gatewayDef{}
	yaml.Unmarshal(yamlIn, &gateway)

	pipeDefinitions := make(map[string]pipeline.Definition)
	for id, def := range gateway.Pipes {
		pipeDefYaml, err := yaml.Marshal(def)
		if err != nil {
			panic(err)
		}

		pipeDef, err := pipeline.DefinePipe(bytes.NewReader(pipeDefYaml))
		if err != nil {
			panic(err)
		}

		pipeDefinitions[id] = pipeDef
	}

	target, err := url.Parse(gateway.Target)
	if err != nil {
		panic(err)
	}

	for _, proxy := range gateway.Proxies {
		proxyHandler := go_gateway.ReverseProxyHandler{pipeDefinitions[proxy.Pipe], target}
		http.HandleFunc(proxy.Path, proxyHandler.ServeHTTP)
	}

	fmt.Printf("Gateway routing port %s => %s\n", gateway.Port, target)
	err = http.ListenAndServe(":" + gateway.Port, nil)

	fmt.Printf("Server crashed: %v\n", err)
}

type fittingDef map[string]interface{}

type pipeDef struct {
	Request  []fittingDef `json:"request"`
	Response []fittingDef `json:"response"`
}

type proxyDef struct {
	Host string `json:"host"`
	Path string `json:"path"`
	Pipe string `json:"pipe"`
}

type gatewayDef struct {
	Port   string             `json:"port"`
	Target string             `json:"target"`
	Pipes  map[string]pipeDef `json:"pipes"`
	Proxies []proxyDef        `json:"proxies"`
}
