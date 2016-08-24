package pipeline

import (
	"net/http"
)

/*
This wrapper is used to ensure that when the response is written to, the pipeline gets canceled.
 */
type resWriter struct {
	writer  http.ResponseWriter
	control Control
}

func (w *resWriter) Header() http.Header {
	return w.writer.Header()
}

func (w *resWriter) Write(bytes []byte) (int, error) {
	w.control.Log().Debugf("write: %s", string(bytes))
	return w.writer.Write(bytes)
}

func (w *resWriter) WriteHeader(status int) {
	w.control.Log().Debugf("writeHeader: %d", status)
	w.control.Cancel()
	w.writer.WriteHeader(status)
}

func (w *resWriter) Control() Control {
	return w.control
}
