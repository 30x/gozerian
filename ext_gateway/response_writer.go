package ext_gateway

import (
	"net/http"
	"golang.org/x/net/context"
	"net"
	"bufio"
	"github.com/30x/gozerian/pipeline"
	"bytes"
	"io/ioutil"
)

func NewResponseWriter() pipeline.ResponseWriter {
	response := new(http.Response)
	buffer := new(bytes.Buffer)
	response.Body = ioutil.NopCloser(buffer)

	writer := responseWriter{buffer, response}

	ctx := context.Background()
	return pipeline.NewResponseWriter(writer, ctx)
}

type responseWriter struct {
	Buffer bytes.Buffer
	Response http.Response
}

func (self responseWriter) Header() http.Header {
	return self.Response.Header
}

func (self responseWriter) Write(bytes []byte) (int, error) {
	return self.Buffer.Write(bytes)
}

func (self responseWriter) WriteHeader(status int) {
	self.Response.Status = status
}

type Request struct {
	http.Request
}
