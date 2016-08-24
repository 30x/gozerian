package pipeline

import (
	"context"
	"fmt"
	"github.com/Sirupsen/logrus"
	"net/http"
	"strconv"
	"sync/atomic"
)

// Pipe runs a series of request and response handlers
type Pipe interface {
	PrepareRequest(reqID string, req *http.Request) *http.Request

	// Be sure to call this with a prepared http.Request!
	RequestHandlerFunc() http.HandlerFunc

	// Be sure to call this with a prepared http.Request!
	ResponseHandlerFunc() ResponseHandlerFunc
}

var reqCounter int64

func newPipe(reqFittings []FittingWithID, resFittings []FittingWithID) Pipe {

	return &pipe{
		reqFittings: reqFittings,
		resFittings: resFittings,
	}
}

type pipe struct {
	reqFittings []FittingWithID
	resFittings []FittingWithID
}

func (p *pipe) PrepareRequest(reqID string, r *http.Request) *http.Request {

	if reqID == "" {
		reqID = string(strconv.FormatInt(atomic.AddInt64(&reqCounter, 1), 10))
	}

	f := logrus.Fields{
		"reqID": reqID,
		"uri":   r.RequestURI,
	}
	log := logger.WithFields(f)

	ctx, cancel := context.WithTimeout(r.Context(), getConfig().GetDuration(ConfigTimeout))
	ctl := &control{
		reqID:    reqID,
		ctx:      ctx,
		conf:     conf,
		logger:   log,
		cancel:   cancel,
		flowData: make(map[string]interface{}),
	}
	ctx = NewControlContext(ctx, ctl)

	return r.WithContext(ctx)
}

func (p *pipe) RequestHandlerFunc() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		control := ControlFromContext(r.Context())
		if control == nil {
			panic("You must run PrepareRequest() on the request!")
		}
		w = &resWriter{w, control}
		defer recoveryFunc(w, control)


		for _, fitting := range p.reqFittings {
			if control.Cancelled() {
				break
			}
			control.Log().Debugf("enter req handler %s", fitting.ID())
			fitting.RequestHandlerFunc()(w, r)
			control.Log().Debugf("exit req handler %s", fitting.ID())
		}

		if r.Context().Err() == context.DeadlineExceeded {
			w.WriteHeader(http.StatusRequestTimeout)
		}
	}
}

func (p *pipe) ResponseHandlerFunc() ResponseHandlerFunc {

	return func(w http.ResponseWriter, r *http.Request, res *http.Response) {

		control := ControlFromContext(r.Context())
		if control == nil {
			panic("You must run PrepareRequest() on the request!")
		}
		w = &resWriter{w, control}
		defer recoveryFunc(w, control)


		for _, fitting := range p.resFittings {
			if control.Cancelled() {
				break
			}
			control.Log().Debugf("enter res handler %s", fitting.ID())
			fitting.ResponseHandlerFunc()(w, r, res)
			control.Log().Debugf("exit res handler %s", fitting.ID())
		}

		if r.Context().Err() == context.DeadlineExceeded {
			w.WriteHeader(http.StatusRequestTimeout)
		}
	}
}

func recoveryFunc(w http.ResponseWriter, pc Control) {
	if r := recover(); r != nil {
		err := fmt.Errorf("%s", r)
		pc.Log().Warn("Panic Recovery Error: ", err)
		pc.HandleError(w, err)
	}
}
