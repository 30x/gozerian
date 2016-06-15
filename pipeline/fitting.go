package pipeline

import (
	"net/http"
	"sync"
)

// Dies is a mapping of id to Die
type Dies map[string]Die

// Die is a factory to create Fittings
type Die func(dieID string, config interface{}) (Fitting, error)

var dies Dies
var diesMutex sync.RWMutex

// RequestHandlerFactory creates a http.HandlerFunc
type RequestHandlerFactory func(interface{}) http.HandlerFunc

// ResponseHandlerFactory creates a ResponseHandlerFunc
type ResponseHandlerFactory func(interface{}) ResponseHandlerFunc

// Fitting is a function in a Pipe
type Fitting interface {
	ID() string
	RequestHandlerFunc() http.HandlerFunc
	ResponseHandlerFunc() ResponseHandlerFunc
}

// RegisterDie associates a Die with an id
func RegisterDie(id string, die Die) {
	diesMutex.Lock()
	defer diesMutex.Unlock()
	if dies == nil {
		dies = make(Dies)
	}
	dies[id] = die
}

// RegisterDies associates multiple Dies with ids
func RegisterDies(d Dies) {
	diesMutex.Lock()
	defer diesMutex.Unlock()
	if dies == nil {
		dies = make(Dies)
	}
	for id, die := range d {
		dies[id] = die
	}
}

// NewFitting creates a fitting from a registered Die
func NewFitting(dieID string, config interface{}) (Fitting, error) {
	diesMutex.RLock()
	defer diesMutex.RUnlock()
	return dies[dieID](dieID, config)
}

// NewFittingFromHandlers creates a fitting from a request and respose handler
func NewFittingFromHandlers(id string, reqHandler http.HandlerFunc, resHandler ResponseHandlerFunc) Fitting {
	return &fitting{id, reqHandler, resHandler}
}

type fitting struct {
	id         string
	reqHandler http.HandlerFunc
	resHandler ResponseHandlerFunc
}

func (f *fitting) ID() string {
	return f.id
}

func (f *fitting) RequestHandlerFunc() http.HandlerFunc {
	return f.reqHandler
}

func (f *fitting) ResponseHandlerFunc() ResponseHandlerFunc {
	return f.resHandler
}
