package pipeline

import "net/http"

type ErrorHandlerFunc func(writer http.ResponseWriter, err error) error
