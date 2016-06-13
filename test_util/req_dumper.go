package test_util

import (
	"net/http"
	"net/http/httputil"
	"io"
	"github.com/30x/gozerian/pipeline"
)


func RequestDumper(dumpBody bool) http.HandlerFunc {
	return requestDumper{dumpBody}.handleRequest
}

type requestDumper struct {
	dumpBody   bool
}

func (self requestDumper) handleRequest(w http.ResponseWriter, r *http.Request) {
	control := w.(pipeline.ControlHolder).Control()
	id := control.RequestID()
	log := control.Log()
	if self.dumpBody {
		r.Body = loggingReadCloser{r.Body, log, id + ">>"}
	}
	log.Printf("======================== request %s ========================", id)
	dump, err := httputil.DumpRequest(r, false)
	if err != nil {
		w.(pipeline.ControlHolder).Control().SendError(err)
	}
	log.Print(string(dump))
}


type loggingReadCloser struct {
	io.ReadCloser
	log       pipeline.Logger
	indicator string
}

func (l loggingReadCloser) Read(buf []byte) (n int, err error) {

	n, err = l.ReadCloser.Read(buf)
	if n > 0 {
		l.log.Printf("%s%q\n", l.indicator, string(buf[:n]))
	}
	return n, err
}
