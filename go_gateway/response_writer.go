package go_gateway

import (
	"net/http"
	"net"
	"bufio"
	"github.com/30x/gozerian/pipeline"
)

type ResponseWriter interface {
	http.ResponseWriter
	http.Hijacker
	pipeline.ControlHolder
}

func NewResponseWriter(writer http.ResponseWriter) ResponseWriter {
	return responseWriter{pipeline.NewResponseWriter(writer), writer}
}

type responseWriter struct {
	pipeline.ResponseWriter
	embeddedWriter http.ResponseWriter
}

func (self responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return self.embeddedWriter.(http.Hijacker).Hijack()
}
