package ext_gateway

import (
	"net/http"
	"github.com/30x/gozerian/pipeline"
	"net/url"
	"io"
)

type ExtResponse struct {
	Status int
	Header http.Header
	Body io.Reader
}

type ExtProxyHandler struct {
	pipeline pipeline.Pipeline
}

// changes to Request Method, URL, Header, and Body are intended to affect proxying to target
// however, if Response is not nil, send Response to client and do not continue target request
func (self ExtProxyHandler) HandleRequest(req *http.Request) (res http.Response) {

	writer := NewResponseWriter{}

	ctx := writer.(pipeline.ContextHolder).Context()

	reqHandler := self.pipeline.RequestHandlerFunc()
	reqHandler(writer, req)

	if ctx.Err() != nil {
		res = http.Response{}
	}

	return res
}

func (self ExtProxyHandler) HandleResponse(res ExtResponse) ExtResponse {
	return nil
}
