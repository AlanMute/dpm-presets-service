package fhttp

import (
	"bufio"
	"encoding/json"
	"github.com/valyala/fasthttp"
	"io"
)

const (
	RecommendedMaxResponseSizeForStreamingThreshold = 5 << 20  // 5mb
	MaxResponseSizeForStreamingThreshold            = 10 << 20 // 10mb, recommended by fasthttp max size
)

type ResponseWriter struct {
	streamer    func(w io.Writer)
	bigResponse bool
}

// NewResponseWriter returns a ResponseWriter with the given streamer. If isBigResponse is true - the writer will send the response via fasthttp.SetBodyStreamWriter,
// otherwise the response will be written directly to the fasthttp.RequestCtx. fashttp recommends response size >= 10mb for streaming, but we use soft limit of 5mb.
// Using fasthttp.RequestCtx directly may improve your program performance
func NewResponseWriter(streamer func(w io.Writer), isBigResponse bool) ResponseWriter {
	return ResponseWriter{
		streamer:    streamer,
		bigResponse: isBigResponse,
	}
}

func (r ResponseWriter) WriteResponse(ctx *fasthttp.RequestCtx) {
	if r.streamer == nil {
		panic("must be unreachable")
	}

	if r.bigResponse {
		ctx.SetBodyStreamWriter(func(w *bufio.Writer) {
			r.streamer(w)
		})
	} else {
		r.streamer(ctx)
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func WriteError(ctx *fasthttp.RequestCtx, message string, status int) {
	response := ErrorResponse{Error: message}
	raw, err := json.Marshal(&response)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetStatusCode(status)
	ctx.Response.Header.Add(fasthttp.HeaderContentType, "application/json")
	_, _ = ctx.Write(raw)
}

func WriteObject(ctx *fasthttp.RequestCtx, obj any, status int) {
	raw, err := json.Marshal(obj)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetStatusCode(status)
	ctx.Response.Header.Add(fasthttp.HeaderContentType, "application/json")
	_, _ = ctx.Write(raw)
}
