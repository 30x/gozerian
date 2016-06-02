package pipeline

import (
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"

	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

// Pipe runs a series of request and response handlers - one created per request
type Pipe interface {
	ControlHolder
	RequestHandlerFunc() http.HandlerFunc
	ResponseHandlerFunc() ResponseHandlerFunc
}

var reqCounter int64

func newPipe(reqID string, reqHands []http.HandlerFunc, resHands []ResponseHandlerFunc) Pipe {

	if reqID == "" {
		reqID = string(strconv.FormatInt(atomic.AddInt64(&reqCounter, 1), 10))
	}

	return &pipe{
		reqID:    reqID,
		reqHands: reqHands,
		resHands: resHands,
	}
}

type pipe struct {
	reqID    string
	reqHands []http.HandlerFunc
	resHands []ResponseHandlerFunc
	control  Control
	writer   responseWriter
}

func (p *pipe) Control() Control {
	return p.control
}

func (p *pipe) RequestHandlerFunc() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		writer := p.setWriter(w, r)
		defer recoveryFunc(p.control)

		for _, handler := range p.reqHands {
			if p.control.Cancelled() {
				break
			}
			handler(writer, r)
		}
	}
}

func (p *pipe) ResponseHandlerFunc() ResponseHandlerFunc {

	return func(w http.ResponseWriter, r *http.Request, res *http.Response) {

		writer := p.setWriter(w, r)
		defer recoveryFunc(p.control)

		for _, handler := range p.resHands {
			if p.control.Cancelled() {
				break
			}
			handler(writer, r, res)
		}

	}
}

func (p *pipe) setWriter(w http.ResponseWriter, r *http.Request) responseWriter {

	writer, ok := w.(responseWriter)
	if !ok {
		config := GetConfig()

		f := logrus.Fields{
			"id":  p.reqID,
			"uri": r.RequestURI,
		}
		log := GetConfig().Log().WithFields(f)

		ctx, cancel := context.WithTimeout(context.Background(), config.Timeout())
		p.control = NewControl(p.reqID, ctx, w, config, log, cancel)

		writer = newResponseWriter(w, p.control)
	}
	p.writer = writer
	return writer
}

func recoveryFunc(pc Control) {
	if r := recover(); r != nil {
		err := fmt.Errorf("%s", r)
		pc.Log().Warn("Panic Recovery Error: ", err)
		pc.SendError(err)
	}
}
