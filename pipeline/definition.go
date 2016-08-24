package pipeline

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
)

type PipeDef struct {
	Request  []FittingDef `yaml:"request"`
	Response []FittingDef `yaml:"response"`
}
type FittingDef map[string]interface{}

// Definition of a pipe
type Definition interface {
	CreatePipe() Pipe
}

// DefinePipeFromReader returns a Pipe Definition as defined in the yaml Reader
func DefinePipe(r io.Reader) (Definition, error) {
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var checkKeys map[string]interface{}
	err = yaml.Unmarshal(bytes, &checkKeys)
	if err != nil {
		return nil, err
	}
	for k := range checkKeys {
		if k != "request" && k != "response" {
			return nil, fmt.Errorf("Bad PipeDef key: %s. Valid keys: 'request', 'response'.", k)
		}
	}

	var pipeConfig PipeDef
	err = yaml.Unmarshal(bytes, &pipeConfig)
	if err != nil {
		return nil, err
	}

	return DefinePipeDirectly(pipeConfig)
}

func DefinePipeDirectly(pipeConfig PipeDef) (Definition, error) {

	var reqFittings []FittingWithID
	reqDefs := pipeConfig.Request
	for _, fittingDef := range reqDefs {
		for id, config := range fittingDef {
			fitting, err := NewFitting(id, config)
			if err != nil {
				return nil, err
			}
			handler := fitting.RequestHandlerFunc()
			if handler != nil {
				reqFittings = append(reqFittings, fitting)
			}
		}
	}

	var resFittings []FittingWithID
	resDefs := pipeConfig.Response
	for _, fittingDef := range resDefs {
		for id, config := range fittingDef {
			fitting, err := NewFitting(id, config)
			if err != nil {
				return nil, err
			}
			handler := fitting.ResponseHandlerFunc()
			if handler != nil {
				resFittings = append(resFittings, fitting)
			}
		}
	}

	return NewDefinition(reqFittings, resFittings), nil
}

// DefinePipe returns a Pipe Definition defined by the passed handlers
func NewDefinition(reqFittings []FittingWithID, resFittings []FittingWithID) Definition {
	return &definition{reqFittings, resFittings}
}

type definition struct {
	reqFittings []FittingWithID
	resFittings []FittingWithID
}

func (d *definition) CreatePipe() Pipe {
	return newPipe(d.reqFittings, d.resFittings)
}
