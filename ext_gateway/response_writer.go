package ext_gateway

import (
	"net/http"
	"golang.org/x/net/context"
	"github.com/30x/gozerian/pipeline"
	"bytes"
	"io/ioutil"
)

type ResponseWriter interface {
	pipeline.ResponseWriter
	GetResponse() *http.Response
}

// the returned Response is written to by the returned ResponseWriter
func NewResponseWriter() ResponseWriter {
	res := &http.Response{
		Header: http.Header{},
	}

	buffer := &bytes.Buffer{}
	res.Body = ioutil.NopCloser(buffer)

	w := &internalResponseWriter{buffer, res}

	ctx := context.Background()

	writer := pipeline.NewResponseWriter(w, ctx)

	return &responseWriter{writer, res}
}

type responseWriter struct {
	pipeline.ResponseWriter
	response *http.Response
}

func (self *responseWriter) GetResponse() *http.Response {
	return self.response
}


type internalResponseWriter struct {
	buffer   *bytes.Buffer
	response *http.Response
}

func (self *internalResponseWriter) Header() http.Header {
	return self.response.Header
}

func (self *internalResponseWriter) Write(bytes []byte) (int, error) {
	return self.buffer.Write(bytes)
}

func (self *internalResponseWriter) WriteHeader(statusCode int) {
	self.response.StatusCode = statusCode
}
