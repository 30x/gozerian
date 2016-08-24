package pipeline

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

// ExtraData is just an open map for storing stuff
type FlowData map[string]interface{}

// Control contains control and context for a Pipe instance
type Control interface {
	RequestID() string

	SetErrorHandler(eh ErrorHandlerFunc)
	ErrorHandler() ErrorHandlerFunc
	HandleError(w http.ResponseWriter, err interface{}) error

	Error() error
	Cancel()
	Cancelled() bool

	Log() Logger

	Config() config

	FlowData() FlowData
}

func NewControlContext(ctx context.Context, ctl Control) context.Context {
	return context.WithValue(ctx, "control", ctl)
}

func ControlFromContext(ctx context.Context) Control {
	return ctx.Value("control").(Control)
}

type control struct {
	ctx          context.Context
	errorHandler ErrorHandlerFunc
	conf         config
	logger       Logger
	cancel       context.CancelFunc
	reqID        string
	flowData     FlowData
}

func (c *control) FlowData() FlowData {
	return c.flowData
}

func (c *control) RequestID() string {
	return c.reqID
}

func (c *control) Config() config {
	return c.conf
}

func (c *control) Log() Logger {
	return c.logger
}

func (c *control) ErrorHandler() ErrorHandlerFunc {
	if c.errorHandler == nil {
		return DefaultErrorHanderFunc
	}
	return c.errorHandler
}

func (c *control) SetErrorHandler(eh ErrorHandlerFunc) {
	c.Log().Debug("set error handler", eh)
	c.errorHandler = eh
}

func (c *control) HandleError(w http.ResponseWriter, r interface{}) error {
	if c.Cancelled() {
		return errors.New("Cancelled response, unable to send")
	}
	c.Log().Debug("send error: ", r)
	var err error
	err, ok := r.(error)
	if !ok {
		err = fmt.Errorf("%s", r)
	}
	return c.ErrorHandler()(w, err)
}

func (c *control) Cancel() {
	c.Log().Debug("cancel pipe")
	if c.Error() == nil {
		c.cancel()
	}
}

func (c *control) Error() error {
	return c.ctx.Err()
}

func (c *control) Cancelled() bool {
	return c.Error() != nil
}
