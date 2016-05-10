package pipeline

// accept a request
// setup context
//	error handler
// 	timeout
// run request pipeline
// process request pipeline result
// 	update target request method, uri, headers
//	or respond to client and end
// make request to target
// run response pipeline
// respond to client
// establish upgraded socket if needed
//	run upgraded pipeline handlers
//	close pipeline on error, timeout
// close context

type Container interface {
	RunTarget()
}
