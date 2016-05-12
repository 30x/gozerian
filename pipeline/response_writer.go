package pipeline

import (
	"golang.org/x/net/context"
	"net/http"
	"time"
	"errors"
	"fmt"
	"reflect"
)

const (
	CancelFuncKey = "gozerian:cancel"
)

func NewResponseWriter(writer http.ResponseWriter, ctx context.Context) ResponseWriter {

	config := NewDefaultConfig()
	ctx = context.WithValue(ctx, "config", config)

	timeout := config.Get("timeout").(time.Duration)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	ctx = context.WithValue(ctx, CancelFuncKey, cancel)

	return responseWriter{writer, ctx, DefaultErrorHanderFunc}
}

type ResponseWriter interface {
	http.ResponseWriter
	ContextHolder
	PipelineControl
}

type responseWriter struct {
	writer     http.ResponseWriter
	ctx        context.Context
	errHandler ErrorHandlerFunc
}

func (self responseWriter) Header() http.Header {
	return self.writer.Header()
}

func (self responseWriter) Write(bytes []byte) (int, error) {
	self.Cancel()
	return self.writer.Write(bytes)
}

func (self responseWriter) WriteHeader(status int) {
	self.Cancel()
	self.writer.WriteHeader(status)
}

func (self responseWriter) Context() context.Context {
	return self.ctx
}

func (self responseWriter) SendError(r interface{}) error {
	var err error
	if reflect.TypeOf(r).String() != "error" {
		err = r.(error)
	} else {
		err = errors.New(fmt.Sprintf("%s", r))
	}
	return self.errHandler(self, err)
}

func (self responseWriter) Cancel() {
	if self.ctx.Err() == nil {
		self.Context().Value(CancelFuncKey).(context.CancelFunc)()
	}
}

func DefaultErrorHanderFunc(writer http.ResponseWriter, err error) error {
	writer.WriteHeader(500)
	_, err = writer.Write([]byte(err.Error()))
	return err
}
