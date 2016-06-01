package handlers

import (
	"net/http"
	"log"
	"net/http/httputil"
	"github.com/30x/gozerian/pipeline"
)

func ResponseDumper(dumpBody bool) pipeline.ResponseHandlerFunc {
	return responseDumper{dumpBody}.handleResponse
}

type responseDumper struct {
	dumpBody   bool
}

func (self responseDumper) handleResponse(w http.ResponseWriter, r *http.Request, res *http.Response) {
	id := w.(pipeline.ControlHolder).Control().RequestId()
	if self.dumpBody {
		res.Body = loggingReadCloser{res.Body, id + "<<"}
	}
	log.Printf("======================== response %s ========================", id)
	dump, err := httputil.DumpResponse(res, false)
	if err != nil {
		w.(pipeline.ControlHolder).Control().SendError(err)
	}
	log.Print(string(dump))
}
