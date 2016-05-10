package pipeline

type PipelineControl interface {
	SendError(err interface{}) error
	Cancel()
}
