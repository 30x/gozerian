package handlers

import (
	"net/http"
	"log"
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
	id := w.(pipeline.ControlHolder).Control().RequestId()
	if self.dumpBody {
		r.Body = loggingReadCloser{r.Body, id + ">>"}
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
	indicator string
}

func (self loggingReadCloser) Read(buf []byte) (n int, err error) {

	n, err = self.ReadCloser.Read(buf)
	if n > 0 {
		log.Printf("%s%q\n", self.indicator, string(buf[:n]))
	}
	return n, err
}
