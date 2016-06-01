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

type ExtraData map[interface{}]interface{}

type PipelineControl interface {
	RequestId() string

	SetErrorHandler(eh ErrorHandlerFunc)
	ErrorHandler() ErrorHandlerFunc
	SendError(err interface{}) error
	Error() error

	Cancel()
	Cancelled() bool

	Log() Logger

	Config() Config

	UserData() ExtraData
}

type pipelineAdmin interface {
	systemData() ExtraData
}

func NewPipelineControl(reqId string, ctx context.Context, w http.ResponseWriter,
			config Config, log Logger, cancel context.CancelFunc) PipelineControl {
	return &pipelineControl{
		ctx: ctx,
		writer: w,
		errorHandler: DefaultErrorHanderFunc,
		config: config,
		logger: log,
		cancel: cancel,
		reqId: reqId,
	}
}

type pipelineControl struct {
	ctx          context.Context
	writer       http.ResponseWriter
	errorHandler ErrorHandlerFunc
	config       Config
	logger       Logger
	cancel       context.CancelFunc
	reqId        string
	userData     ExtraData
	systemData   ExtraData
}

func (self *pipelineControl) UserData() ExtraData {
	return self.userData
}

func (self *pipelineControl) RequestId() string {
	return self.reqId
}

func (self *pipelineControl) Config() Config {
	return self.config
}

func (self *pipelineControl) Log() Logger {
	return self.logger
}

func (self *pipelineControl) ErrorHandler() ErrorHandlerFunc {
	if self.errorHandler == nil {
		return DefaultErrorHanderFunc
	}
	return self.errorHandler
}

func (self *pipelineControl) SetErrorHandler(eh ErrorHandlerFunc) {
	self.Log().Debug("SetErrorHandler", eh)
	self.errorHandler = eh
}

func (self *pipelineControl) SendError(r interface{}) error {
	if self.Cancelled() {
		return errors.New("Cancelled response, unable to send")
	}
	self.Log().Debug("SendError: ", r)
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
	self.Log().Debug("Cancel pipe")
	if self.Error() == nil {
		self.cancel()
	}
}

func (self *pipelineControl) Error() error {
	return self.ctx.Err()
}

func (self *pipelineControl) Cancelled() bool {
	self.Log().Debug("Pipe cancelled check = ", self.Error() != nil)
	return self.Error() != nil
}
