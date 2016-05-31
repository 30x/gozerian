package pipeline

import (
	"net/http"
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"github.com/Sirupsen/logrus"
)

type Pipeline struct {
	ReqHandlers []http.HandlerFunc
	ResHandlers []ResponseHandlerFunc
}

func (self *Pipeline) RequestHandlerFunc() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		writer, ok := w.(ResponseWriter)
		if (!ok) {
			writer = NewResponseWriter(w, context.Background())
		}
		pc := writer.Control()
		defer recoveryFunc(pc)

		pc.SetLogger(loggerWithFields(r))

		for _, handler := range self.ReqHandlers {
			if pc.Cancelled() {
				break
			}
			handler(writer, r)
		}
	}
}

func (self *Pipeline) ResponseHandlerFunc() ResponseHandlerFunc {

	return func(w http.ResponseWriter, r *http.Request, res *http.Response) {

		writer, ok := w.(ResponseWriter)
		if (!ok) {
			writer = NewResponseWriter(w, context.Background())
		}
		pc := writer.Control()
		defer recoveryFunc(pc)

		pc.SetLogger(loggerWithFields(r))

		for _, handler := range self.ResHandlers {
			if pc.Cancelled() {
				break
			}
			handler(writer, r, res)
		}

	}
}

func recoveryFunc(pc PipelineControl) {
	if r := recover(); r != nil {
		err := errors.New(fmt.Sprintf("%s", r))
		pc.Logger().Warn("Panic Recovery Error: ", err)
		pc.SendError(err)
	}
}

func loggerWithFields(r *http.Request) Logger {
	id := -1 // todo: get ID from lib-gozerian
	f := logrus.Fields{
		"id": id,
		"URI": r.RequestURI,
	}
	return logrus.WithFields(f)
}