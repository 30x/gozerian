package pipeline

import (
	"net/http"
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
)

type Pipeline struct {
	ReqHandlers []http.HandlerFunc
	ResHandlers []ResponseHandlerFunc
}

func (self *Pipeline) RequestHandlerFunc(reqId string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		writer := wrapWriter(w)
		pc := writer.Control()
		defer recoveryFunc(pc)

		pc.SetRequestId(reqId)
		pc.SetLog(loggerWithFields(reqId, r))

		for _, handler := range self.ReqHandlers {
			if pc.Cancelled() {
				break
			}
			handler(writer, r)
		}
	}
}

func (self *Pipeline) ResponseHandlerFunc(reqId string) ResponseHandlerFunc {

	return func(w http.ResponseWriter, r *http.Request, res *http.Response) {

		writer := wrapWriter(w)
		pc := writer.Control()
		defer recoveryFunc(pc)

		pc.SetRequestId(reqId)
		pc.SetLog(loggerWithFields(reqId, r))

		for _, handler := range self.ResHandlers {
			if pc.Cancelled() {
				break
			}
			handler(writer, r, res)
		}

	}
}

func wrapWriter(w http.ResponseWriter) ResponseWriter {
	writer, ok := w.(ResponseWriter)
	if (!ok) {
		writer = NewResponseWriter(w)
	}
	return writer
}

func recoveryFunc(pc PipelineControl) {
	if r := recover(); r != nil {
		err := errors.New(fmt.Sprintf("%s", r))
		pc.Log().Warn("Panic Recovery Error: ", err)
		pc.SendError(err)
	}
}

func loggerWithFields(reqId string, r *http.Request) Logger {
	f := logrus.Fields{
		"id": reqId,
		"uri": r.RequestURI,
	}
	return GetConfig().Log().WithFields(f)
}