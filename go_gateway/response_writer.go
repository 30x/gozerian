package go_gateway

import (
	"net/http"
	"golang.org/x/net/context"
	"net"
	"bufio"
	"github.com/30x/gozerian/pipeline"
)

type ResponseWriter interface {
	http.ResponseWriter
	http.Hijacker
	pipeline.ContextHolder
	pipeline.PipelineControl
}

func NewResponseWriter(writer http.ResponseWriter) ResponseWriter {
	ctx := context.Background()
	return responseWriter{pipeline.NewResponseWriter(writer, ctx), writer}
}

type responseWriter struct {
	pipeline.ResponseWriter
	embeddedWriter http.ResponseWriter
}

func (self responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return self.embeddedWriter.(http.Hijacker).Hijack()
}
