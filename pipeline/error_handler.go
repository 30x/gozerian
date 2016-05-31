package pipeline

import (
	"net/http"
	"runtime/debug"
)

type ErrorHandlerFunc func(writer http.ResponseWriter, err error) error

func DefaultErrorHanderFunc(writer http.ResponseWriter, err error) error {
	writer.WriteHeader(500)
	_, err = writer.Write([]byte(err.Error() + "\n"))
	if err != nil {
		_, err = writer.Write(debug.Stack())
	}
	return err
}
