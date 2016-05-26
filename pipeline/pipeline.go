package pipeline

import (
	"net/http"
	"errors"
	"fmt"
	"golang.org/x/net/context"
)

type Pipeline struct {
	ReqHandlers []http.HandlerFunc
	ResHandlers []ResponseHandlerFunc
}

func (self *Pipeline) RequestHandlerFunc() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// ResponseWriter must be a ContextHolder
		_, ok := w.(ContextHolder)
		if (!ok) {
			w = NewResponseWriter(w, context.Background())
		}

		ctx := w.(ContextHolder).Context()

		defer func() {
			if r := recover(); r != nil {
				err := errors.New(fmt.Sprintf("%s", r))
				w.(PipelineControl).SendError(err)
			}
		}()

		for _, handler := range self.ReqHandlers {

			if ctx.Err() != nil {
				break
			}
			handler(w, r)
		}
	}
}

func (self *Pipeline) ResponseHandlerFunc() ResponseHandlerFunc {

	return func(w http.ResponseWriter, r *http.Request, res *http.Response) {

		// ResponseWriter must be a ContextHolder
		_, ok := w.(ContextHolder)
		if (!ok) {
			w = NewResponseWriter(w, context.Background())
		}

		ctx := w.(ContextHolder).Context()

		defer func() {
			if r := recover(); r != nil {
				err := errors.New(fmt.Sprintf("%s", r))
				w.(PipelineControl).SendError(err)
			}
		}()

		for _, handler := range self.ResHandlers {
			if ctx.Err() != nil {
				break
			}
			handler(w, r, res)
		}

	}
}
