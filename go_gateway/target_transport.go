package go_gateway

import (
	"bufio"
	"errors"
	"github.com/30x/gozerian/pipeline"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

type targetTransport struct {
	http.RoundTripper
	control    pipeline.Control
	writer     http.ResponseWriter
	origReq    *http.Request
	resHandler pipeline.ResponseHandlerFunc
}

func (tt *targetTransport) RoundTrip(req *http.Request) (res *http.Response, err error) {

	// upgrade to hijacked connection
	upgrade := tt.origReq.Header.Get("Connection") == "Upgrade"
	if upgrade {
		return tt.upgradedRoundTrip(req)
	}

	// call target
	res, err = tt.RoundTripper.RoundTrip(req)

	if err != nil {
		tt.control.Cancel()
		tt.control.Log().Debug(err)
		return res, err
	}

	// run response handlers
	tt.resHandler(tt.writer, tt.origReq, res)

	return res, tt.control.Error()
}

func (tt *targetTransport) upgradedRoundTrip(req *http.Request) (res *http.Response, err error) {

	req.Header.Set("Connection", tt.origReq.Header.Get("Connection"))
	req.Header.Set("Upgrade", tt.origReq.Header.Get("Upgrade"))

	targetConn, err := net.Dial("tcp", req.URL.Host)
	if err != nil {
		return nil, err
	}
	defer targetConn.Close()

	if err := req.Write(targetConn); err != nil {
		return nil, err
	}

	targetReader := bufio.NewReader(targetConn)
	res, err = http.ReadResponse(targetReader, req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 101 {
		return nil, errors.New("Upgrade status code (101) required")
	}

	// Run Response Handlers
	tt.resHandler(tt.writer, tt.origReq, res)

	con, clientRW, err := tt.writer.(http.Hijacker).Hijack()
	if err != nil {
		return nil, err
	}
	defer con.Close()

	res.Write(clientRW)
	clientRW.Flush()

	done := make(chan error)

	// set timeout timer
	// todo: can we get the elapsed time left instead?
	timeout := tt.control.Config().GetDuration(pipeline.ConfigTimeout)
	timer := time.AfterFunc(timeout, func() { close(done) })

	// todo: pipe data through upgraded stream handlers!

	go copyData(clientRW, targetConn, done, timer, timeout)
	go copyData(targetConn, clientRW, done, timer, timeout)

	err = <-done

	return nil, err
}

func copyData(writer io.Writer, reader io.Reader, done chan error, timer *time.Timer, timeout time.Duration) {

	// avoid "send on closed channel" panic
	defer func() {
		if r := recover(); r != nil {
			log.Printf("recovered panic: %v", r)
		}
	}()

	buf := make([]byte, 100, 100)
	c := time.Tick(time.Millisecond)
	for range c {
		n, err := io.CopyBuffer(writer, reader, buf)
		if err != nil {
			done <- err
			break
		}
		if n > 0 {
			timer.Reset(timeout)
		}
	}
}
