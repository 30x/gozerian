package pipeline

type PipelineControl interface {
	SetErrorHandler(eh ErrorHandlerFunc)
	SendError(err interface{}) error
	Cancel()
}
