package pipeline

import (
	"io"
	"io/ioutil"
	"net/http"
	"gopkg.in/yaml.v2"
	"fmt"
)

// Definition of a pipe
type Definition interface {
	CreatePipe(reqID string) Pipe
}

// DefinePipeFromReader returns a Pipe Definition as defined in the Reader
func DefinePipe(r io.Reader) (Definition, error) {
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	type FittingDef map[string]interface{}
	var pipeConfig []FittingDef
	err = yaml.Unmarshal(bytes, &pipeConfig)
	if err != nil {
		return nil, err
	}

	// todo: validation of structures

	var reqHands []http.HandlerFunc
	for _, fittingDef := range pipeConfig {
		if len(fittingDef) > 1 {
			return nil, fmt.Errorf("bad structure")
		}
		for name, config := range fittingDef{
			fitting := NewFitting(name, config)
			handler := fitting.RequestHandlerFunc()
			if handler != nil {
				reqHands = append(reqHands, handler)
			}
		}
	}

	return NewDefinition(reqHands, nil)
}

// DefinePipe returns a Pipe Definition defined by the passed handlers
func NewDefinition(reqHands []http.HandlerFunc, resHands []ResponseHandlerFunc) (Definition, error) {
	return &definition{reqHands, resHands}, nil
}

type definition struct {
	reqHands []http.HandlerFunc
	resHands []ResponseHandlerFunc
}

// if reqId is nil, will create and use an internal id
func (s *definition) CreatePipe(reqID string) Pipe {
	return newPipe(reqID, s.reqHands, s.resHands)
}
