package c_gateway

import (
	"net/http"
	"github.com/30x/gozerian/pipeline"
	"net/url"
	"errors"
)

/*
Note: Register Dies for pipeline before calling ListenAndServe.

Example of expected YAML at URL:

    request:                # request pipeline
    - dump:                 # name of plugin
        dumpBody: true      # plugin-specific configuration
    response:               # response pipeline
    - dump:                 # name of plugin
        dumpBody: true      # plugin-specific configuration
 */

func DefinePipe(configUrl *url.URL) (pipeline.Definition, error) {

	res, err := http.Get(configUrl.String())
	if err != nil {
		return nil, err
	}
	if res.Body == nil {
		return nil, errors.New("Invalid URL, no body")
	}
	defer res.Body.Close()

	return pipeline.DefinePipe(res.Body)
}
