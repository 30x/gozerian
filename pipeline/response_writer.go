package pipeline

import (
	"golang.org/x/net/context"
	"net/http"
)

func NewResponseWriter(writer http.ResponseWriter, ctx context.Context) ResponseWriter {

	// todo: set config elsewhere
	config := NewDefaultConfig()

	ctx, cancel := context.WithTimeout(ctx, config.Timeout())

	control := NewPipelineControl(ctx, writer, config, cancel)

	return &responseWriter{writer, control}
}

type ResponseWriter interface {
	http.ResponseWriter
	ControlHolder
}

type responseWriter struct {
	writer  http.ResponseWriter
	control PipelineControl
}

func (self *responseWriter) Header() http.Header {
	return self.writer.Header()
}

func (self *responseWriter) Write(bytes []byte) (int, error) {
	self.control.Logger().Debug("Write:", string(bytes))
	self.control.Cancel()
	return self.writer.Write(bytes)
}

func (self *responseWriter) WriteHeader(status int) {
	self.control.Logger().Debug("WriteHeader:", status)
	self.control.Cancel()
	self.writer.WriteHeader(status)
}

func (self *responseWriter) Control() PipelineControl {
	return self.control
}
