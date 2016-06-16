package pipeline

import (
	"net/http"
	"sync"
)

// Die is a factory to create Fittings
type Die func(config interface{}) (Fitting, error)

var dies map[string]Die
var diesMutex sync.RWMutex

// RequestHandlerFactory creates a http.HandlerFunc
type RequestHandlerFactory func(interface{}) http.HandlerFunc

// ResponseHandlerFactory creates a ResponseHandlerFunc
type ResponseHandlerFactory func(interface{}) ResponseHandlerFunc

// Fitting is a function in a Pipe
type Fitting interface {
	RequestHandlerFunc() http.HandlerFunc
	ResponseHandlerFunc() ResponseHandlerFunc
}

type FittingWithID interface {
	Fitting
	ID() string
}

// RegisterDie associates a Die with an id
func RegisterDie(id string, die Die) {
	diesMutex.Lock()
	defer diesMutex.Unlock()
	if dies == nil {
		dies = make(map[string]Die)
	}
	dies[id] = die
}

// RegisterDies associates multiple Dies with ids
func RegisterDies(m map[string]Die) {
	diesMutex.Lock()
	defer diesMutex.Unlock()
	if dies == nil {
		dies = make(map[string]Die)
	}
	for id, die := range m {
		dies[id] = die
	}
}

// NewFitting creates a fitting from a registered Die
func NewFitting(dieID string, config interface{}) (FittingWithID, error) {
	diesMutex.RLock()
	defer diesMutex.RUnlock()
	internalDie, err := dies[dieID](config)
	if err != nil {
		return nil, err
	}
	return &dieFitting{internalDie, dieID}, nil
}

type dieFitting struct {
	Fitting
	id         string
}

func (f *dieFitting) ID() string {
	return f.id
}
