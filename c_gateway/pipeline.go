package c_gateway

import (
	"net/http"
	"github.com/30x/gozerian/pipeline"
	"net/url"
)

// external interface for gozerian-c
func DefinePipe(configUrl *url.URL) (pipeline.Definition, error) {

	res, err := http.Get(configUrl.String())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	return pipeline.DefinePipe(res.Body)
}
