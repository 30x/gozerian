package handlers

import (
	"net/http"
	"strconv"
	"sync/atomic"
	"log"
	"net/http/httputil"
	"github.com/30x/gozerian/pipeline"
)

func ResponseDumper(dumpBody bool) pipeline.ResponseHandlerFunc {
	return responseDumper{dumpBody, 0}.handleResponse
}

type responseDumper struct {
	dumpBody   bool
	resCounter int64
}

func (self responseDumper) handleResponse(w http.ResponseWriter, r *http.Request, res *http.Response) {
	id := strconv.FormatInt(atomic.AddInt64(&self.resCounter, 1), 10)
	if self.dumpBody {
		res.Body = loggingReadCloser{res.Body, id + "<<"}
	}
	log.Printf("======================== response %s ========================", id)
	dump, _ := httputil.DumpResponse(res, false)
	log.Print(string(dump))
}
