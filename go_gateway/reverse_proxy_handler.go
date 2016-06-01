package go_gateway

// This is used for a standalone Go gateway utilizing http.ReverseProxy.
// It is used for unit tests and such, but it's not expected to be a production implementation.

import (
	"net/url"
	"net/http"
	"path"
	"net/http/httputil"
	"github.com/30x/gozerian/pipeline"
)

type ReverseProxyHandler struct {
	Definition pipeline.Definition
	Target     *url.URL
}

func (self *ReverseProxyHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	req.URL.Host = self.Target.Host

	//writer := NewResponseWriter(w)
	pipe := self.Definition.CreatePipe("")

	// call request handlers
	pipe.RequestHandlerFunc()(w, req)

	// abort pipeline as needed
	if pipe.Control().Cancelled() {
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

	resHandler := pipe.ResponseHandlerFunc()
	transport := &targetTransport{http.DefaultTransport, pipe.Control(), w, req, resHandler}
	targetProxy := &httputil.ReverseProxy{Director: director, Transport: transport}
	targetProxy.ServeHTTP(w, req)
}
