package pipeline

import (
	"net/url"
	"net/http"
	"io/ioutil"
	"io"
)

type Definition interface {
	CreatePipe(reqId string) Pipe
}

func DefinePipeFromURL(definitionURL url.URL) (Definition, error) {
	res, err := http.Get(definitionURL.String())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	return DefinePipeFromReader(res.Body)
}

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

func DefinePipe(reqHands []http.HandlerFunc, resHands []ResponseHandlerFunc) (Definition, error) {
	return &definition{reqHands, resHands}, nil
}

type definition struct {
	reqHands []http.HandlerFunc
	resHands []ResponseHandlerFunc
}

// if reqId is nil, will create and use an internal id
func (self *definition) CreatePipe(reqId string) Pipe {
	return newPipe(reqId, self.reqHands, self.resHands)
}
