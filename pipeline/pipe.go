package pipeline

import (
	"net/http"
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
	"strconv"
	"sync/atomic"
)

type Pipe interface {
	ControlHolder
	RequestHandlerFunc() http.HandlerFunc
	ResponseHandlerFunc() ResponseHandlerFunc
}

var reqCounter int64

// create per request - if reqId is empty, will create auto-create an ID to use
func newPipe(reqId string, reqHands []http.HandlerFunc, resHands []ResponseHandlerFunc) Pipe {

	if reqId == "" {
		reqId = string(strconv.FormatInt(atomic.AddInt64(&reqCounter, 1), 10))
	}

	return &pipe{
		reqId: reqId,
		reqHands: reqHands,
		resHands: resHands,
	}
}

type pipe struct {
	reqId      string
	reqHands   []http.HandlerFunc
	resHands   []ResponseHandlerFunc
	control    PipelineControl
	writer     ResponseWriter
}

func (self *pipe) Control() PipelineControl {
	return self.control
}

func (self *pipe) RequestHandlerFunc() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		writer := self.setWriter(w, r)
		defer recoveryFunc(self.control)

		for _, handler := range self.reqHands {
			if self.control.Cancelled() {
				break
			}
			handler(writer, r)
		}
	}
}

func (self *pipe) ResponseHandlerFunc() ResponseHandlerFunc {

	return func(w http.ResponseWriter, r *http.Request, res *http.Response) {

		writer := self.setWriter(w, r)
		defer recoveryFunc(self.control)

		for _, handler := range self.resHands {
			if self.control.Cancelled() {
				break
			}
			handler(writer, r, res)
		}

	}
}

func (self *pipe) setWriter(w http.ResponseWriter, r *http.Request) ResponseWriter {

	writer, ok := w.(ResponseWriter)
	if (!ok) {
		config := GetConfig()

		f := logrus.Fields{
			"id": self.reqId,
			"uri": r.RequestURI,
		}
		log := GetConfig().Log().WithFields(f)

		ctx, cancel := context.WithTimeout(context.Background(), config.Timeout())
		self.control = NewPipelineControl(self.reqId, ctx, w, config, log, cancel)

		writer = NewResponseWriter(w, self.control)
	}
	self.writer = writer
	return writer
}

func recoveryFunc(pc PipelineControl) {
	if r := recover(); r != nil {
		err := errors.New(fmt.Sprintf("%s", r))
		pc.Log().Warn("Panic Recovery Error: ", err)
		pc.SendError(err)
	}
}
