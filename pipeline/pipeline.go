package pipeline

import (
	"net/http"
	"errors"
	"fmt"
)

type Pipeline struct {
	ReqHandlers []http.HandlerFunc
	ResHandlers []ResponseHandlerFunc
}

func (self *Pipeline) RequestHandlerFunc() http.HandlerFunc {

	// ResponseWriter must be a ContextHolder
	return func(w http.ResponseWriter, r *http.Request) {

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

	// ResponseWriter must be a ContextHolder
	return func(w http.ResponseWriter, r *http.Request, res *http.Response) {

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
