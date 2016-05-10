package pipeline_test

import (
	"github.com/30x/gozerian/go_gateway"
	. "github.com/30x/gozerian/pipeline"
	. "github.com/30x/gozerian/handlers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http/httptest"
	"net/http"
	"fmt"
	"io/ioutil"
	"strings"
	"io"
	"bytes"
	"errors"
	"net"
	"bufio"
	"net/url"
	"strconv"
	"log"
	"github.com/gorilla/websocket"
)

var _ = log.Print // avoid having to add/remove the log import so much

// Test framework: http://onsi.github.io/ginkgo/

var noRequestHandlers = []http.HandlerFunc{}
var noResponseHandlers = []ResponseHandlerFunc{}

func pipelineHandler(targetURL string, reqHands []http.HandlerFunc, resHands []ResponseHandlerFunc) http.HandlerFunc {

	reqHands = append(reqHands, RequestDumper(true))
	resHands = append(resHands, ResponseDumper(true))

	target, _:= url.Parse(targetURL)
	pipeline := Pipeline{reqHands, resHands}
	proxyHandler := go_gateway.ReverseProxyHandler{pipeline, target}
	return proxyHandler.ServeHTTP
}

var _ = Describe("Pipeline", func() {

	It("should pass request and response untouched", func() {

		clientHeaders := http.Header{"Foo": {"Bar"}}
		clientBody := "Hello, Whomever"

		// create the target
		targetBody := "Hello, client\n"
		target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// check the received request
			body, _ := ioutil.ReadAll(r.Body)
			for key, value := range clientHeaders {
				Expect(r.Header.Get(key)).To(Equal(value[0]))
			}
			Expect(string(body)).To(Equal(clientBody))

			// send response
			w.Header().Set("Bar", "Baz")
			w.WriteHeader(200)
			fmt.Fprint(w, targetBody)
		}))
		defer target.Close()

		// create the gateway
		gateway := httptest.NewServer(pipelineHandler(target.URL, noRequestHandlers, noResponseHandlers))
		defer gateway.Close()

		// send the request
		client := &http.Client{}
		req, _ := http.NewRequest("POST", gateway.URL, strings.NewReader(clientBody))
		req.Header = clientHeaders
		res, _ := client.Do(req)
		defer res.Body.Close()

		// check the received response
		body, _ := ioutil.ReadAll(res.Body)
		Expect(res.StatusCode).To(Equal(200))
		Expect(res.Header.Get("Bar")).To(Equal("Baz"))
		Expect(string(body)).To(Equal(targetBody))
	})

	It("should deal with target connection error", func() {

		// create the gateway
		gateway := httptest.NewServer(pipelineHandler("http://localhost:9999", noRequestHandlers, noResponseHandlers))
		defer gateway.Close()

		// send the request
		res, _ := http.Get(gateway.URL)
		defer res.Body.Close()

		// check the received response
		body, _ := ioutil.ReadAll(res.Body)
		Expect(res.StatusCode).To(Equal(500))
		Expect(string(body)).To(Equal("dial tcp [::1]:9999: getsockopt: connection refused"))
	})

	PIt("should timeout request")

	Context("request handler", func() {

		It("should be able to modify request URL", func() {

			newPath := "/test"

			// create the target
			target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// check the received request
				Expect(r.URL.Path).To(Equal(newPath))

				// send response
				w.WriteHeader(200)
			}))
			defer target.Close()

			// create the gateway
			requestHandlers := []http.HandlerFunc{func(w http.ResponseWriter, r *http.Request) {
				r.URL.Path = newPath
			}}
			gateway := httptest.NewServer(pipelineHandler(target.URL, requestHandlers, noResponseHandlers))
			defer gateway.Close()

			// send the request
			res, _ := http.Get(gateway.URL)
			defer res.Body.Close()
		})

		It("should be able to hit a different target", func() {

			// create the original target
			origTarget := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// send response
				w.WriteHeader(404)
			}))
			defer origTarget.Close()

			// create the new target
			newTarget := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// send response
				w.WriteHeader(200)
			}))
			defer newTarget.Close()

			// create the gateway
			requestHandlers := []http.HandlerFunc{func(w http.ResponseWriter, r *http.Request) {
				// update the target
				newUrl, _ := url.Parse(newTarget.URL)
				r.URL.Host = newUrl.Host
			}}
			gateway := httptest.NewServer(pipelineHandler(origTarget.URL, requestHandlers, noResponseHandlers))
			defer gateway.Close()

			// send the request
			res, _ := http.Get(gateway.URL)
			defer res.Body.Close()

			Expect(res.StatusCode).To(Equal(200))
		})

		It("should be able to modify request method", func() {

			newMethod := "PUT"

			// create the target
			target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// check the received request
				Expect(r.Method).To(Equal(newMethod))

				// send response
				w.WriteHeader(200)
			}))
			defer target.Close()

			// create the gateway
			requestHandlers := []http.HandlerFunc{func(w http.ResponseWriter, r *http.Request) {
				r.Method = newMethod
			}}
			gateway := httptest.NewServer(pipelineHandler(target.URL, requestHandlers, noResponseHandlers))
			defer gateway.Close()

			// send the request
			res, _ := http.Get(gateway.URL)
			defer res.Body.Close()
		})

		It("should be able to modify request headers", func() {

			clientHeaders := http.Header{
				"Foo": {"Bar"},
				"Add": {"Bar"},
				"Change": {"Me"},
				"Del": {"Me"},
			}

			// create the target
			target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// check the received request
				Expect(r.Header.Get("Foo")).To(Equal("Bar")) // unchanged
				Expect(r.Header.Get("Del")).To(Equal("")) // deleted
				Expect(r.Header.Get("New")).To(Equal("Test")) // new
				Expect(r.Header.Get("Change")).To(Equal("Test")) // changed
				Expect(r.Header["Add"]).To(Equal([]string{"Bar", "Test"})) // added

				// send response
				w.WriteHeader(200)
			}))
			defer target.Close()

			// create the gateway
			requestHandlers := []http.HandlerFunc{func(w http.ResponseWriter, r *http.Request) {
				r.Header.Del("Del") // delete existing
				r.Header.Set("New", "Test") // add new
				r.Header.Set("Change", "Test") // change existing
				r.Header.Add("Add", "Test") // add new
			}}
			gateway := httptest.NewServer(pipelineHandler(target.URL, requestHandlers, noResponseHandlers))
			defer gateway.Close()

			// send the request
			client := &http.Client{}
			req, _ := http.NewRequest("GET", gateway.URL, nil)
			req.Header = clientHeaders
			res, _ := client.Do(req)
			defer res.Body.Close()
		})

		It("should be able to filter request body", func() {

			clientBody := "Hello, Whomever"
			targetBody := "Hello, Scott"

			// create the target
			target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// check the received request
				body, _ := ioutil.ReadAll(r.Body)
				Expect(string(body)).To(Equal(targetBody))

				w.WriteHeader(200)
			}))
			defer target.Close()

			// create the gateway
			requestHandlers := []http.HandlerFunc{func(w http.ResponseWriter, r *http.Request) {
				r.Body = testFilter{r.Body}
			}}
			gateway := httptest.NewServer(pipelineHandler(target.URL, requestHandlers, noResponseHandlers))
			defer gateway.Close()

			// send the request
			client := &http.Client{}
			req, _ := http.NewRequest("POST", gateway.URL, strings.NewReader(clientBody))
			res, _ := client.Do(req)
			defer res.Body.Close()
		})

		It("should be able to replace request body", func() {

			clientBody := "Hello, Whomever"
			targetBody := "Hello, Scott"

			// create the target
			target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// check the received request
				body, _ := ioutil.ReadAll(r.Body)
				Expect(string(body)).To(Equal(targetBody))

				w.WriteHeader(200)
			}))
			defer target.Close()

			// create the gateway
			requestHandlers := []http.HandlerFunc{func(w http.ResponseWriter, r *http.Request) {
				// read original body
				b, _ := ioutil.ReadAll(r.Body)
				r.Body.Close()

				// replace it
				b = bytes.Replace(b, []byte("Whomever"), []byte("Scott"), -1)
				r.Body = ioutil.NopCloser(bytes.NewReader(b))
			}}
			gateway := httptest.NewServer(pipelineHandler(target.URL, requestHandlers, noResponseHandlers))
			defer gateway.Close()

			// send the request
			client := &http.Client{}
			req, _ := http.NewRequest("POST", gateway.URL, strings.NewReader(clientBody))
			res, _ := client.Do(req)
			defer res.Body.Close()
		})

		It("should be able to cancel the pipeline (and request)", func() {
			// create the target
			target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Fail("Should not reach")
			}))
			defer target.Close()

			// create the gateway
			requestHandlers := []http.HandlerFunc{func(w http.ResponseWriter, r *http.Request) {
				w.(PipelineControl).Cancel()
			}}
			gateway := httptest.NewServer(pipelineHandler(target.URL, requestHandlers, noResponseHandlers))
			defer gateway.Close()

			// send the request
			res, _ := http.Get(gateway.URL)
			defer res.Body.Close()
		})

		It("should be able to handle an error using default error handler", func() {
			errMsg := "What's going on?"

			// create the target
			target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Fail("Should not reach")
			}))
			defer target.Close()

			// create the gateway
			requestHandlers := []http.HandlerFunc{func(w http.ResponseWriter, r *http.Request) {
				w.(PipelineControl).SendError(errors.New(errMsg))
			}}
			gateway := httptest.NewServer(pipelineHandler(target.URL, requestHandlers, noResponseHandlers))
			defer gateway.Close()

			// send the request
			res, _ := http.Get(gateway.URL)
			defer res.Body.Close()

			Expect(res.StatusCode).To(Equal(500))
			body, _ := ioutil.ReadAll(res.Body)
			Expect(string(body)).To(Equal(errMsg))
		})

		It("should be able to handle a panic using default error handler", func() {
			errMsg := "What's going on?"

			// create the target
			target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Fail("Should not reach")
			}))
			defer target.Close()

			// create the gateway
			requestHandlers := []http.HandlerFunc{func(w http.ResponseWriter, r *http.Request) {
				panic(errMsg)
			}}
			gateway := httptest.NewServer(pipelineHandler(target.URL, requestHandlers, noResponseHandlers))
			defer gateway.Close()

			// send the request
			res, _ := http.Get(gateway.URL)
			defer res.Body.Close()

			Expect(res.StatusCode).To(Equal(500))
			body, _ := ioutil.ReadAll(res.Body)
			Expect(string(body)).To(Equal(errMsg))
		})
	})

	Context("response handler", func() {

		It("should be able to modify response status", func() {
			// create the target
			target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// send response
				w.WriteHeader(200)
			}))
			defer target.Close()

			// create the gateway
			responseHandlers := []ResponseHandlerFunc{func(w http.ResponseWriter, r *http.Request, res *http.Response) {
				// rewrite status
				w.WriteHeader(404)
			}}
			gateway := httptest.NewServer(pipelineHandler(target.URL, noRequestHandlers, responseHandlers))
			defer gateway.Close()

			// send the request
			res, _ := http.Get(gateway.URL)
			defer res.Body.Close()

			// check the response
			Expect(res.StatusCode).To(Equal(404))
		})

		It("should be able to modify response headers", func() {

			responseHeaders := http.Header{
				"Foo": {"Bar"},
				"Add": {"Bar"},
				"Change": {"Me"},
				"Del": {"Me"},
			}

			// create the target
			target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// send response
				for key, values := range responseHeaders {
					for _, value := range values {
						w.Header().Add(key, value)
					}
				}
				w.WriteHeader(200)
			}))
			defer target.Close()

			// create the gateway
			responseHandlers := []ResponseHandlerFunc{func(w http.ResponseWriter, r *http.Request, res *http.Response) {
				// rewrite headers
				res.Header.Del("Del") // delete existing
				res.Header.Set("New", "Test") // add new
				res.Header.Set("Change", "Test") // change existing
				res.Header.Add("Add", "Test") // add new
			}}
			gateway := httptest.NewServer(pipelineHandler(target.URL, noRequestHandlers, responseHandlers))
			defer gateway.Close()

			// send the request
			res, _ := http.Get(gateway.URL)
			defer res.Body.Close()

			// check the response
			Expect(res.Header.Get("Foo")).To(Equal("Bar")) // unchanged
			Expect(res.Header.Get("Del")).To(Equal("")) // deleted
			Expect(res.Header.Get("New")).To(Equal("Test")) // new
			Expect(res.Header.Get("Change")).To(Equal("Test")) // changed
			Expect(res.Header["Add"]).To(Equal([]string{"Bar", "Test"})) // added
		})

		It("should be able to filter response body", func() {

			targetBody := "Hello, Whomever"
			testBody := "Hello, Scott"

			// create the target
			target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// send response
				w.Write([]byte(targetBody))
			}))
			defer target.Close()

			// create the gateway
			responseHandlers := []ResponseHandlerFunc{func(w http.ResponseWriter, r *http.Request, res *http.Response) {
				// filter body
				res.Body = testFilter{res.Body}
			}}
			gateway := httptest.NewServer(pipelineHandler(target.URL, noRequestHandlers, responseHandlers))
			defer gateway.Close()

			// send the request
			res, _ := http.Get(gateway.URL)
			defer res.Body.Close()

			// check the response
			b, _ := ioutil.ReadAll(res.Body)
			Expect(string(b)).To(Equal(testBody))
		})

		It("should be able to replace response body", func() {

			targetBody := "Hello, Whomever"
			testBody := "Hello, Scott"

			// create the target
			target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// send response
				w.Write([]byte(targetBody))
			}))
			defer target.Close()

			// create the gateway
			responseHandlers := []ResponseHandlerFunc{func(w http.ResponseWriter, r *http.Request, res *http.Response) {
				// read original body
				b, _ := ioutil.ReadAll(res.Body)
				res.Body.Close()

				// replace it
				b = bytes.Replace(b, []byte("Whomever"), []byte("Scott"), -1)
				res.Body = ioutil.NopCloser(bytes.NewReader(b))
			}}
			gateway := httptest.NewServer(pipelineHandler(target.URL, noRequestHandlers, responseHandlers))
			defer gateway.Close()

			// send the request
			res, _ := http.Get(gateway.URL)
			defer res.Body.Close()

			// check the response
			b, _ := ioutil.ReadAll(res.Body)
			Expect(string(b)).To(Equal(testBody))
		})

		It("should be able to handle an error using default error handler", func() {
			errMsg := "What's going on?"

			// create the target
			target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("Should not reach client"))
			}))
			defer target.Close()

			// create the gateway
			responseHandlers := []ResponseHandlerFunc{func(w http.ResponseWriter, r *http.Request, res *http.Response) {
				w.(PipelineControl).SendError(errors.New(errMsg))
			}}
			gateway := httptest.NewServer(pipelineHandler(target.URL, noRequestHandlers, responseHandlers))
			defer gateway.Close()

			// send the request
			res, _ := http.Get(gateway.URL)
			defer res.Body.Close()

			// check the response
			Expect(res.StatusCode).To(Equal(500))
			body, _ := ioutil.ReadAll(res.Body)
			Expect(string(body)).To(Equal(errMsg))
		})

		It("should be able to handle a panic using default error handler", func() {
			errMsg := "What's going on?"

			// create the target
			target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("Should not reach client"))
			}))
			defer target.Close()

			// create the gateway
			responseHandlers := []ResponseHandlerFunc{func(w http.ResponseWriter, r *http.Request, res *http.Response) {
				panic(errMsg)
			}}
			gateway := httptest.NewServer(pipelineHandler(target.URL, noRequestHandlers, responseHandlers))
			defer gateway.Close()

			// send the request
			res, _ := http.Get(gateway.URL)
			defer res.Body.Close()

			// check the response
			Expect(res.StatusCode).To(Equal(500))
			body, _ := ioutil.ReadAll(res.Body)
			Expect(string(body)).To(Equal(errMsg))
		})
	})

	Context("web socket upgrade", func() {

		It("should be able to pass through", func() {

			// create the target
			var upgrader = websocket.Upgrader{}
			target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

				c, err := upgrader.Upgrade(w, r, nil)
				Expect(err).NotTo(HaveOccurred())
				defer c.Close()

				for {
					mt, message, err := c.ReadMessage()
					Expect(err).NotTo(HaveOccurred())

					err = c.WriteMessage(mt, message)
					Expect(err).NotTo(HaveOccurred())
				}
			}))
			defer target.Close()

			// create the gateway
			gateway := httptest.NewServer(pipelineHandler(target.URL, noRequestHandlers, noResponseHandlers))
			defer gateway.Close()

			// communicate with server
			targetUrl, _ := url.Parse(gateway.URL)
			targetUrl.Scheme = "ws"
			c, _, err := websocket.DefaultDialer.Dial(targetUrl.String(), nil)
			Expect(err).NotTo(HaveOccurred())
			defer c.Close()

			// write
			err = c.WriteMessage(websocket.TextMessage, []byte("Hello"))
			Expect(err).NotTo(HaveOccurred())

			// read
			_, message, err := c.ReadMessage()
			Expect(err).NotTo(HaveOccurred())
			Expect(string(message)).To(Equal("Hello"))

			// write
			err = c.WriteMessage(websocket.TextMessage, []byte("Goodbye"))
			Expect(err).NotTo(HaveOccurred())

			// read
			_, message, err = c.ReadMessage()
			Expect(err).NotTo(HaveOccurred())
			Expect(string(message)).To(Equal("Goodbye"))

			target.Close()
		})

		//PIt("should be able to filter requests and responses - maybe")
	})

	Context("custom protocol upgrade", func() {

		It("shold be able to pass through", func() {

			// create the target
			target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {

				w.WriteHeader(101)

				if hj, ok := w.(http.Hijacker); ok {

					con, bufrw, err := hj.Hijack()
					Expect(err).NotTo(HaveOccurred())

					bufrw.Flush()
					defer con.Close()

					send := func(input string) {
						str := input + "\r\n"
						_, err := bufrw.WriteString(str)
						//n, err := bufrw.WriteString(str)
						//log.Printf("s send (%d): %q", n, str)
						Expect(err).NotTo(HaveOccurred())
						bufrw.Flush()
					}

					receive := func() string {
						line, err := bufrw.ReadBytes('\r')
						if err != nil && err == io.EOF {
							return "EOF"
						}
						Expect(err).NotTo(HaveOccurred())
						str := string(bytes.TrimSpace(line))
						//log.Printf("s rece: %q", str)
						return str
					}

					out:
					for {
						received := receive()
						switch received {
						case "ADD":
							op1, err := strconv.Atoi(receive())
							Expect(err).NotTo(HaveOccurred())
							op2, err := strconv.Atoi(receive())
							Expect(err).NotTo(HaveOccurred())
							result := op1 + op2
							send(strconv.Itoa(result))
						default:
							break out
						}
					}
				}
			}))
			defer target.Close()

			// create the gateway
			gateway := httptest.NewServer(pipelineHandler(target.URL, noRequestHandlers, noResponseHandlers))
			defer gateway.Close()
			targetUrl, _:= url.Parse(gateway.URL)


			// communicate with server
			tcpConn, _ := net.Dial("tcp", targetUrl.Host)

			reader := bufio.NewReader(tcpConn)
			receive := func() string {
				line, err := reader.ReadBytes('\r')
				Expect(err).NotTo(HaveOccurred())
				str := string(bytes.TrimSpace(line))
				//log.Printf("c rece: %q", str)
				return str
			}

			send := func(line string) {
				str := line + "\r\n"
				//log.Printf("c send: %q", str)
				_, err := tcpConn.Write([]byte(str))
				Expect(err).NotTo(HaveOccurred())
			}

			// http handshake
			req, _ := http.NewRequest("GET", gateway.URL, strings.NewReader(""))
			req.Header.Set("Connection", "Upgrade")
			err := req.Write(tcpConn)
			Expect(err).NotTo(HaveOccurred())

			targetReader := bufio.NewReader(tcpConn)
			res, err := http.ReadResponse(targetReader, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(res.StatusCode).To(Equal(101))

			send("ADD")
			send("1")
			send("2")
			Expect(receive()).To(Equal("3"))
			send("QUIT")
		})

		//PIt("should be able to filter requests and responses - maybe")
	})

	Context("http 2.0 upgrade", func() {

		PIt("should be able to pass through")

		//PIt("should be able to filter requests and responses - maybe")
	})
})

type testFilter struct {
	io.ReadCloser
}

func (self testFilter) Read(buf []byte) (n int, err error) {

	n, err = self.ReadCloser.Read(buf)
	if n > 0 {
		replaced := bytes.Replace(buf[:n], []byte("Whomever"), []byte("Scott"), -1)
		newLen := len(replaced)
		if newLen > cap(buf) {
			return -1, errors.New("Dragons be here!") // todo: test this before using as example
		}
		// update buf and n
		n = newLen
		copy(buf, replaced)
	}
	return n, err
}

type echoFilter struct {
	io.ReadCloser
}

func (self echoFilter) Read(buf []byte) (n int, err error) {

	n, err = self.ReadCloser.Read(buf)
	if n > 0 {
		fmt.Println(string(buf[:n]))
	}
	return n, err
}
