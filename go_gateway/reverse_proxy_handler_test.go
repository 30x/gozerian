package go_gateway_test

import (
	. "github.com/30x/gozerian/pipeline"
	. "github.com/30x/gozerian/handlers"
	"github.com/30x/gozerian/go_gateway"
	"github.com/30x/gozerian/test_util"
	"net/http/httptest"
	"net/http"
	"net/url"
)

// Test framework: http://onsi.github.io/ginkgo/

func makeGateway(targetURL string, reqHands []http.HandlerFunc, resHands []ResponseHandlerFunc) *httptest.Server {

	reqHands = append(reqHands, RequestDumper(true))
	resHands = append(resHands, ResponseDumper(true))

	target, _:= url.Parse(targetURL)
	pipeDef, err := DefinePipe(reqHands, resHands)
	if err != nil {
		panic(err)
	}
	proxyHandler := &go_gateway.ReverseProxyHandler{pipeDef, target}

	return httptest.NewServer(proxyHandler)
}

var _ = test_util.TestPipelineAgainst(makeGateway)
var _ = test_util.TestPipelineSocketUpgradesAgainst(makeGateway)
