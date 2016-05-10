package pipeline

import "net/http"

type ResponseHandlerFunc func(w http.ResponseWriter, r *http.Request, res *http.Response)
