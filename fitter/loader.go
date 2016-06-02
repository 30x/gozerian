package fitter

//import (
//	"github.com/30x/gozerian/pipeline"
//	"net/http"
//	"github.com/30x/gozerian/handlers"
//	"net/url"
//)
//
//func CreateFromURL(uri url.URL) pipeline.Pipeline {
//
//	//res, err := http.Get(configUrl.String())
//	//defer res.Body.Close()
//}
//
//func CreateFromConfig(config string) pipeline.Pipeline {
//	//decoder := json.NewDecoder(res.Body)
//	//config := decoder.Decode()
//}
//
//type PipelineConfig struct {
//
//}
//
//type RequestHandler http.HandlerFunc
//type HandlerConfig struct {
//	handler string
//	config interface{}
//}
//type PipeDef struct {
//	factory *Factory
//	configs []HandlerConfig
//}
//func (self *PipeDef) append(config HandlerConfig) {
//	self.configs = append(self.configs, config)
//}
//
//type HandlerFactory interface {
//}
//type RequestHandlerFactory interface {
//	NewRequestHandler(config HandlerConfig) *http.HandlerFunc
//}
//type ResponseHandlerFactory interface {
//	NewResponseHandler(config HandlerConfig) *pipeline.ResponseHandlerFunc
//}
//
//func main() {
//
//	dump := DumperFactory{}
//	var pipe []http.HandlerFunc
//
//	FittingFactory.addType("dump", DumperFactory{})
//
//	// register all handler factories
//	f := Factory{}
//	f.RegisterRequestHandler("dump", DumperFactory{})
//	f.RegisterResponseHandler("dump", DumperFactory{})
//
//	reqPipe := f.NewPipeline()
//	reqPipe.append(HandlerConfig{handler: "dump",
//		config: map[string]bool{
//			"body": true,
//		},
//	})
//
//	var pipeConfig map[HandlerFactory]HandlerConfig
//
//	pipeConfig["dump"] = {}
//	append(configs, HandlerConfig{
//		"a": "B",
//	})
//
//	reqPipe := CreateRequestPipeline(f, configs)
//
//	var reqPipe http.HandlerFunc
//	var resPipe pipeline.ResponseHandlerFunc
//}
//
//type Factory struct {
//	reqFactories map[string]RequestHandlerFactory
//	resFactories map[string]ResponseHandlerFactory
//}
//
//func (self *Factory) RegisterRequestHandler(id string, factory RequestHandlerFactory) {
//	self.reqFactories[id] = factory
//}
//
//func (self *Factory) RegisterResponseHandler(id string, factory ResponseHandlerFactory) {
//	self.resFactories[id] = factory
//}
//
//func (self *Factory) NewPipeline() *PipeDef {
//	return &PipeDef{self}
//}
//
//func (self *Factory) NewRequestHandler(factoryId string, config HandlerConfig) (http.HandlerFunc, error) {
//	f, ok := self.reqFactories[factoryId]
//	if !ok {
//		return error("not found")
//	}
//	return f.NewRequestHandler(config), nil
//}
//
//func (self *Factory) NewResponseHandler(factoryId string, config HandlerConfig) (pipeline.ResponseHandlerFunc, error) {
//	f, ok := self.resFactories[factoryId]
//	if !ok {
//		return error("not found")
//	}
//	return f.NewResponseHandler(config), nil
//}
//
//func (self *Factory) CreateRequestPipeline(pd PipeDef) http.HandlerFunc {
//	var reqHandlers []http.HandlerFunc
//	var resHandlers []pipeline.ResponseHandlerFunc
//
//	for i, config := range configs {
//		h := f.create(config)
//		if reqH := h.getRequestHandler(); reqH != nil {
//			append(reqHandlers, reqH)
//		}
//
//		if resH := h.getResponseHandler(); resH != nil {
//			append(resHandlers, resH)
//		}
//	}
//	p := pipeline.Pipeline{
//		reqHandlers,
//		resHandlers,
//	}
//	return p
//}
//
//func (self *Factory) CreateResponsePipeline(pd PipeDef) pipeline.ResponseHandlerFunc {
//
//	pd range
//
//	var resHandlers []pipeline.ResponseHandlerFunc
//	for i, config := range configs {
//		h := f.create(config)
//		if reqH := h.getRequestHandler(); reqH != nil {
//			append(reqHandlers, reqH)
//		}
//
//		if resH := h.getResponseHandler(); resH != nil {
//			append(resHandlers, resH)
//		}
//	}
//	p := pipeline.Pipeline{
//		reqHandlers,
//		resHandlers,
//	}
//	return p
//}
//
//type Handler interface {
//	getRequestHandler() http.HandlerFunc
//	getResponseHandler() pipeline.ResponseHandlerFunc
//}
//
//
//// set up factories
//
//type DumperFactory struct {
//}
//
//func (self *DumperFactory) NewRequestHandler(config HandlerConfig) http.HandlerFunc {
//	return handlers.RequestDumper(config["dump_body"] == true)
//}
//
//func (self *DumperFactory) NewResponseHandler() pipeline.ResponseHandlerFunc {
//	return handlers.ResponseDumper(config["dump_body"] == true)
//}
