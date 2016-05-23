package ext_gateway_test

import (
	"github.com/30x/gozerian/pipeline"
	. "github.com/30x/gozerian/handlers"
	. "github.com/30x/gozerian/ext_gateway"
	"github.com/30x/gozerian/test_util"
	"net/http/httptest"
	"net/http"
	"net/url"
	"path"
	"io/ioutil"
	"bytes"
	"io"
)

// Test framework: http://onsi.github.io/ginkgo/

func makeGateway(targetURL string, reqHands []http.HandlerFunc, resHands []pipeline.ResponseHandlerFunc) *httptest.Server {

	reqHands = append(reqHands, RequestDumper(true))
	resHands = append(resHands, ResponseDumper(true))

	target, _:= url.Parse(targetURL)
	pipeline := pipeline.Pipeline{reqHands, resHands}
	writer := NewResponseWriter(nil)
	proxyHandler := ExtProxyHandler{pipeline, writer}

	handler := server{target, &proxyHandler}

	return httptest.NewServer(handler)
}

type server struct {
	target	*url.URL
	proxyHandler *ExtProxyHandler
}

func (self server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	writer := self.proxyHandler.Writer
	ctx := writer.Context()

	// for testing control, create entirely new request with new body
	body, err := copyReader(r.Body)
	if err != nil {
		panic(err)  // todo: handle error
	}
	req, err := http.NewRequest(r.Method, r.URL.String(), body)
	if err != nil {
		panic(err)  // todo: handle error
	}
	req.Header = r.Header

	// fix req.URL to point at target
	req.URL.Scheme = self.target.Scheme
	req.URL.Host = self.target.Host
	req.URL.Path = path.Join(self.target.Path, req.URL.Path)
	targetQuery := self.target.RawQuery
	if targetQuery == "" || req.URL.RawQuery == "" {
		req.URL.RawQuery = targetQuery + req.URL.RawQuery
	} else {
		req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
	}

	// call request handlers
	res := self.proxyHandler.HandleRequest(req)
	if ctx.Err() != nil {
		writeResponse(writer.GetResponse(), w)
		return
	}

	// call proxy target
	res, err = http.DefaultTransport.RoundTrip(req)
	if err != nil {
		// maintains compatibility with the go_gateway tests
		writer.SendError(err)
		writeResponse(writer.GetResponse(), w)
		return
	}

	// todo: this copy doesn't work... the headers won't show up in the response dump for some reason
	// for testing control, create entirely new Response
	//body, err = copyReader(res.Body)
	//if err != nil {
	//	panic(err)  // todo: handle error
	//}
	//res = &http.Response{
	//	StatusCode: res.StatusCode,
	//	Header: res.Header,
	//	Body: ioutil.NopCloser(body),
	//}

	// call response handlers
	res = self.proxyHandler.HandleResponse(req, res)

	// write status, header, and body back to client
	writeResponse(res, w)
}

func copyReader(reader io.ReadCloser) (io.Reader, error) {
	byteData, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	reader.Close()
	return bytes.NewReader(byteData), nil
}

func writeResponse(res *http.Response, w http.ResponseWriter) {
	copyHeader(w.Header(), res.Header)
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)  // todo: handle error
	}
	w.WriteHeader(res.StatusCode)
	if len(bytes) > 0 {
		_, err = w.Write(bytes)
		if err != nil {
			panic(err)  // todo: handle error
		}
	}
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

var _ = test_util.TestPipelineAgainst(makeGateway)
