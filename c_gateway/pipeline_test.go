package c_gateway_test

import (
	. "github.com/30x/gozerian/pipeline"
	"github.com/30x/gozerian/test_util"
	"net/http/httptest"
	"net/url"
	"github.com/30x/gozerian/go_gateway"
	"net/http"
)

// Test framework: http://onsi.github.io/ginkgo/

func makeGateway(targetURL string, reqHands []http.HandlerFunc, resHands []ResponseHandlerFunc) *httptest.Server {

	reqHands = append(reqHands, test_util.RequestDumper(false))
	resHands = append(resHands, test_util.ResponseDumper(false))

	var reqFittings []Fitting
	for _, h := range reqHands {
		reqFittings = append(reqFittings, NewFittingFromHandlers("test", h, nil))
	}

	var resFittings []Fitting
	for _, h := range resHands {
		resFittings = append(resFittings, NewFittingFromHandlers("test", nil, h))
	}

	pipeDef, err := NewDefinition(reqFittings, resFittings)
	if err != nil {
		panic(err)
	}
	target, _:= url.Parse(targetURL)
	proxyHandler := &go_gateway.ReverseProxyHandler{pipeDef, target}

	return httptest.NewServer(proxyHandler)
}


var _ = test_util.TestPipelineAgainst(makeGateway)
