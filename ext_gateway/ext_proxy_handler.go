package ext_gateway

import (
	"net/http"
	"github.com/30x/gozerian/pipeline"
)

type ExtProxyHandler struct {
	Pipeline pipeline.Pipeline
	Writer ResponseWriter
}

// changes to passed Request Method, URL, Header, and Body are intended to affect proxying to target
// however, if response is not nil, send Response to client and do not continue target request
func (self ExtProxyHandler) HandleRequest(req *http.Request) *http.Response {

	ctx := self.Writer.Context()

	reqHandler := self.Pipeline.RequestHandlerFunc()
	reqHandler(self.Writer, req)

	if ctx.Err() != nil {
		return self.Writer.GetResponse()
	}

	return nil
}

// the passed Request should be the original request from the client
// the passed Response should include Header and Body from the target
// the Response returned will include the final Header and Body to send to the client
func (self ExtProxyHandler) HandleResponse(req *http.Request, res *http.Response) *http.Response {

	ctx := self.Writer.Context()

	resHandler := self.Pipeline.ResponseHandlerFunc()
	resHandler(self.Writer, req, res)

	if ctx.Err() != nil {
		return self.Writer.GetResponse()
	}

	return res
}
