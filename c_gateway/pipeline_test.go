package c_gateway_test

import (
	. "github.com/30x/gozerian/pipeline"
	. "github.com/30x/gozerian/handlers"
	"github.com/30x/gozerian/test_util"
	"net/http/httptest"
	"net/url"
	"github.com/30x/gozerian/go_gateway"
	"net/http"
)

// Test framework: http://onsi.github.io/ginkgo/

func makeGateway(targetURL string, reqHands []http.HandlerFunc, resHands []ResponseHandlerFunc) *httptest.Server {

	reqHands = append(reqHands, RequestDumper(false))
	resHands = append(resHands, ResponseDumper(false))

	target, _:= url.Parse(targetURL)
	pipeline := Pipeline{reqHands, resHands}
	proxyHandler := go_gateway.ReverseProxyHandler{pipeline, target}

	return httptest.NewServer(proxyHandler)
}


var _ = test_util.TestPipelineAgainst(makeGateway)
