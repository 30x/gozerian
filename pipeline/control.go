package pipeline

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"golang.org/x/net/context"
)

// ControlHolder holds a Control
type ControlHolder interface {
	Control() Control
}

// ExtraData is just an open map for storing stuff
type FlowData map[string]interface{}

// Control contains control and context for a Pipe instance
type Control interface {
	RequestID() string

	SetErrorHandler(eh ErrorHandlerFunc)
	ErrorHandler() ErrorHandlerFunc
	SendError(err interface{}) error
	Error() error

	Cancel()
	Cancelled() bool

	Log() Logger

	Config() Config

	FlowData() FlowData
}

// NewControl creates a new Control
func NewControl(reqID string, ctx context.Context, w http.ResponseWriter,
	config Config, log Logger, cancel context.CancelFunc) Control {
	return &control{
		ctx:          ctx,
		writer:       w,
		errorHandler: DefaultErrorHanderFunc,
		config:       config,
		logger:       log,
		cancel:       cancel,
		reqID:        reqID,
	}
}

type control struct {
	ctx          context.Context
	writer       http.ResponseWriter
	errorHandler ErrorHandlerFunc
	config       Config
	logger       Logger
	cancel       context.CancelFunc
	reqID        string
	flowData     FlowData
}

func (pc *control) FlowData() FlowData {
	return pc.flowData
}

func (pc *control) RequestID() string {
	return pc.reqID
}

func (pc *control) Config() Config {
	return pc.config
}

func (pc *control) Log() Logger {
	return pc.logger
}

func (pc *control) ErrorHandler() ErrorHandlerFunc {
	if pc.errorHandler == nil {
		return DefaultErrorHanderFunc
	}
	return pc.errorHandler
}

func (pc *control) SetErrorHandler(eh ErrorHandlerFunc) {
	pc.Log().Debug("SetErrorHandler", eh)
	pc.errorHandler = eh
}

func (pc *control) SendError(r interface{}) error {
	if pc.Cancelled() {
		return errors.New("Cancelled response, unable to send")
	}
	pc.Log().Debug("SendError: ", r)
	var err error
	if reflect.TypeOf(r).String() != "error" {
		err = r.(error)
	} else {
		err = fmt.Errorf("%s", r)
	}
	writeErr := pc.ErrorHandler()(pc.writer, err)
	if writeErr != nil {
		return writeErr
	}
	pc.Cancel()
	return nil
}

func (pc *control) Cancel() {
	pc.Log().Debug("Cancel pipe")
	if pc.Error() == nil {
		pc.cancel()
	}
}

func (pc *control) Error() error {
	return pc.ctx.Err()
}

func (pc *control) Cancelled() bool {
	pc.Log().Debug("Pipe cancelled check = ", pc.Error() != nil)
	return pc.Error() != nil
}
