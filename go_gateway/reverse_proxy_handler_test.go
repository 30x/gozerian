package go_gateway_test

import (
	. "github.com/30x/gozerian/pipeline"
	"github.com/30x/gozerian/go_gateway"
	"github.com/30x/gozerian/test_util"
	"net/http/httptest"
	"net/http"
	"net/url"
)

// Test framework: http://onsi.github.io/ginkgo/

func makeGateway(targetURL string, reqHands []http.HandlerFunc, resHands []ResponseHandlerFunc) *httptest.Server {

	reqHands = append(reqHands, test_util.RequestDumper(true))
	resHands = append(resHands, test_util.ResponseDumper(true))

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
var _ = test_util.TestPipelineSocketUpgradesAgainst(makeGateway)
