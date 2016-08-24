package go_gateway

// This is used for a standalone Go gateway utilizing http.ReverseProxy.
// It is used for unit tests and such, but it's not expected to be a production implementation.

import (
	"github.com/30x/gozerian/pipeline"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
)

type ReverseProxyHandler struct {
	Definition pipeline.Definition
	Target     *url.URL
}

func (self *ReverseProxyHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	req.URL.Host = self.Target.Host

	// todo: Set up pipes, select pipe by host & path
	pipe := self.Definition.CreatePipe()

	req = pipe.PrepareRequest("", req) // leaving reqID blank will auto-assign ID

	// call request handlers
	pipe.RequestHandlerFunc()(w, req)

	// abort pipeline as needed

	control := pipeline.ControlFromContext(req.Context())
	if control.Cancelled() {
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

	// provide a logger that avoids extraneous log messages on error
	maskingLogger := log.New(&logWriter{control, w}, "", 0)

	transport := &targetTransport{http.DefaultTransport, control, w, req, resHandler}
	targetProxy := &httputil.ReverseProxy{Director: director, Transport: transport, ErrorLog: maskingLogger}


	targetProxy.ServeHTTP(w, req)
}

type logWriter struct {
	control pipeline.Control
	w       http.ResponseWriter
}

func (lw *logWriter) Write(p []byte) (n int, err error) {
	if !lw.control.Cancelled() {
		lw.control.Log().Debug(string(p))
	}
	return 0, nil
}
