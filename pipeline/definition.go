package pipeline

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Definition of a pipe
type Definition interface {
	CreatePipe(reqID string) Pipe
}

// DefinePipeFromURL returns a Pipe Definition as defined in the URL
func DefinePipeFromURL(definitionURL url.URL) (Definition, error) {
	res, err := http.Get(definitionURL.String())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	return DefinePipeFromReader(res.Body)
}

// DefinePipeFromReader returns a Pipe Definition as defined in the Reader
func DefinePipeFromReader(r io.Reader) (Definition, error) {
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// todo: create handlers
	if bytes != nil {
	}

	return DefinePipe(nil, nil)
}

// DefinePipe returns a Pipe Definition defined by the passed handlers
func DefinePipe(reqHands []http.HandlerFunc, resHands []ResponseHandlerFunc) (Definition, error) {
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
