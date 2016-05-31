package pipeline

import (
	"reflect"
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"net/http"
)

type ControlHolder interface {
	Control() PipelineControl
}

type PipelineControl interface {
	SetErrorHandler(eh ErrorHandlerFunc)
	ErrorHandler() ErrorHandlerFunc
	SendError(err interface{}) error
	Error() error

	Cancel()
	Cancelled() bool

	Logger() Logger
	SetLogger(logger Logger)

	Config() Config
}

func NewPipelineControl(ctx context.Context, w http.ResponseWriter, conf Config, cancel context.CancelFunc) PipelineControl {
	return &pipelineControl{ctx, w, DefaultErrorHanderFunc, conf, conf.Logger(), cancel}
}

type pipelineControl struct {
	ctx          context.Context
	writer       http.ResponseWriter
	errorHandler ErrorHandlerFunc
	config       Config
	logger       Logger
	cancel       context.CancelFunc
}

func (self *pipelineControl) Config() Config {
	return self.config
}

func (self *pipelineControl) Logger() Logger {
	return self.logger
}

func (self *pipelineControl) SetLogger(logger Logger) {
	self.logger = logger
}

func (self *pipelineControl) ErrorHandler() ErrorHandlerFunc {
	if self.errorHandler == nil {
		return DefaultErrorHanderFunc
	}
	return self.errorHandler
}

func (self *pipelineControl) SetErrorHandler(eh ErrorHandlerFunc) {
	self.Logger().Debug("SetErrorHandler", eh)
	self.errorHandler = eh
}

func (self *pipelineControl) SendError(r interface{}) error {
	if self.Cancelled() {
		return errors.New("Cancelled response, unable to send")
	}
	self.Logger().Debug("SendError: ", r)
	var err error
	if reflect.TypeOf(r).String() != "error" {
		err = r.(error)
	} else {
		err = errors.New(fmt.Sprintf("%s", r))
	}
	writeErr := self.ErrorHandler()(self.writer, err)
	if writeErr != nil {
		return writeErr
	}
	self.Cancel()
	return nil
}

func (self *pipelineControl) Cancel() {
	self.Logger().Debug("Cancel pipe")
	if self.Error() == nil {
		self.cancel()
	}
}

func (self *pipelineControl) Error() error {
	return self.ctx.Err()
}

func (self *pipelineControl) Cancelled() bool {
	self.Logger().Debug("Pipe cancelled check = ", self.Error() != nil)
	return self.Error() != nil
}
