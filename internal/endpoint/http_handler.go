package endpoint

import (
	"github.com/AlanMute/dpm-presets-service/pkg/util/cast"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

type route struct {
	path    string
	handler func(ctx *fasthttp.RequestCtx, h *HttpHandler)
}

var routingMap = map[string]route{
	"/status": {handler: func(ctx *fasthttp.RequestCtx, h *HttpHandler) {
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetBodyString("OK")
	}},
}

func init() {
	for p, r := range routingMap {
		r.path = p
		routingMap[p] = r
	}
}

type HttpHandler struct {
	fsHandler fasthttp.RequestHandler
}

func NewHttpHandler() *HttpHandler {
	return &HttpHandler{
		fsHandler: (&fasthttp.FS{Root: "./"}).NewRequestHandler(),
	}
}

func (h *HttpHandler) Handle(ctx *fasthttp.RequestCtx) {
	defer func() {
		err := recover()
		if err != nil {
			logrus.Error("Critical error during handling: ", err)
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		}
	}()

	if r, ok := routingMap[cast.ByteArrayToString(ctx.Path())]; ok {
		r.handler(ctx, h)
	} else {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
	}
}
