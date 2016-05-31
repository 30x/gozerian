package go_gateway

// This is used for a standalone Go gateway utilizing http.ReverseProxy.
// It is used for unit tests and such, but it's not expected to be a production implementation.

import (
	"net/url"
	"net/http"
	"path"
	"net/http/httputil"
	"github.com/30x/gozerian/pipeline"
	"golang.org/x/net/context"
)

type ReverseProxyHandler struct {
	Pipeline pipeline.Pipeline
	Target   *url.URL
}

func (self ReverseProxyHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	req.URL.Host = self.Target.Host

	writer := NewResponseWriter(w, context.Background())

	// call request handlers
	self.Pipeline.RequestHandlerFunc()(writer, req)

	// abort pipeline as needed
	if writer.Control().Cancelled() {
		return
	}

	// call target & response handlers
	targetQuery := self.Target.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = self.Target.Scheme
		req.URL.Path = path.Join(self.Target.Path, req.URL.Path)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
	}

	resHandler := self.Pipeline.ResponseHandlerFunc()
	transport := &targetTransport{http.DefaultTransport, writer, req, resHandler}
	targetProxy := &httputil.ReverseProxy{Director: director, Transport: transport}
	targetProxy.ServeHTTP(writer, req)
}
