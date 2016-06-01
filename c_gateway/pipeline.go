package c_gateway

import (
	"net/http"
	"github.com/30x/gozerian/pipeline"
	"net/url"
	"io/ioutil"
	"bytes"
	"fmt"
	"time"
	"github.com/30x/gozerian/handlers"
)

// external interface for gozerian-c
func DefinePipe(configUrl *url.URL) (pipeline.Definition, error) {

	// todo: temporary: hard-coded to create a test pipeline for lib-gozerian tests

	var reqHands []http.HandlerFunc
	reqHands = append(reqHands, handlers.RequestDumper(false))
	reqHands = append(reqHands, testHandleRequest)

	var resHands []pipeline.ResponseHandlerFunc
	resHands = append(resHands, handlers.ResponseDumper(false))
	resHands = append(resHands, testHandleResponse)

	return pipeline.DefinePipe(reqHands, resHands)
}

func testHandleRequest(resp http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	case "/pass":
		// Nothing to do

	case "/slowpass":
		time.Sleep(time.Second)

	case "/readbody":
		_, err := ioutil.ReadAll(req.Body)
		if err != nil {
			fmt.Printf("Error reading body: %v\n", err)
		}
		req.Body.Close()

	case "/readbodyslow":
		tmp := make([]byte, 2)
		buf := &bytes.Buffer{}
		len, _ := req.Body.Read(tmp)
		for len > 0 {
			buf.Write(tmp[0:len])
			len, _ = req.Body.Read(tmp)
		}
		req.Body.Close()

	case "/readanddiscard":
		tmp := make([]byte, 2)
		req.Body.Read(tmp)
		req.Body.Close()

	case "/replacebody":
		req.Body = ioutil.NopCloser(bytes.NewReader([]byte("Hello! I am the server!")))

	case "/writeheaders":
		req.Header.Add("Server", "Go Test Stuff")
		req.Header.Add("X-Apigee-Test", "HeaderTest")

	case "/writepath":
		newURL, _ := url.Parse("/newpath")
		req.URL = newURL

	case "/return201":
		resp.WriteHeader(http.StatusCreated)

	case "/returnheaders":
		resp.Header().Add("X-Apigee-Test", "Return Header Test")
		resp.WriteHeader(http.StatusOK)

	case "/returnbody":
		resp.Write([]byte("Hello! I am the server!"))

	case "/completerequest":
		newURL, _ := url.Parse("/totallynewurl")
		req.URL = newURL
		req.Header.Add("X-Apigee-Test", "Complete")
		// TODO would like reader to return in two chunks
		req.Body = ioutil.NopCloser(
			bytes.NewReader([]byte("Hello Again! Time for a complete rewrite!")))
	//ctx.ProxyRequest().Write([]byte("Hello Again! "))
	//ctx.ProxyRequest().Write([]byte("Time for a complete rewrite!"))

	case "/completeresponse":
		ioutil.ReadAll(req.Body)
		req.Body.Close()
		resp.Header().Add("X-Apigee-Test", "Complete")
		resp.WriteHeader(http.StatusCreated)
		resp.Write([]byte("Hello Again! "))
		resp.Write([]byte("Time for a complete rewrite!"))

	case "/writeresponseheaders":
	case "/transformbody":
	case "/transformbodychunks":
	case "/responseerror":
	case "/responseerror2":

	default:
		resp.WriteHeader(http.StatusNotFound)
	}
}

func testHandleResponse(w http.ResponseWriter, req *http.Request, resp *http.Response) {
	switch req.URL.Path {
	case "/writeresponseheaders":
		resp.Header.Set("X-Apigee-ResponseHeader", "yes")

	case "/transformbody":
		resp.Body = ioutil.NopCloser(
			bytes.NewReader([]byte("We have transformed the response!")))

	case "/responseerror":
		resp.StatusCode = http.StatusInternalServerError
		resp.Body = ioutil.NopCloser(
			bytes.NewReader([]byte("Error in the server!")))

	case "/responseerror2":
		w.Header().Set("X-Apigee-Response", "error")
		w.WriteHeader(http.StatusGatewayTimeout)
		w.Write([]byte("Response Error"))

	case "/transformbodychunks":
		resp.Header.Set("X-Apigee-Transformed", "yes")
		defer resp.Body.Close()

		buf := &bytes.Buffer{}
		rb := make([]byte, 128)
		len, _ := resp.Body.Read(rb)
		for len > 0 {
			s := fmt.Sprintf("{ %v }\n", rb[:len])
			buf.WriteString(s)
			len, _ = resp.Body.Read(rb)
		}
		resp.Body = ioutil.NopCloser(buf)

		resp.Header.Set("X-Apigee-Invisible", "yes")
	}
}
