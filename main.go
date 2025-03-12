package main

import (
	"github.com/AlanMute/dpm-presets-service/internal/endpoint"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"os"
	"os/signal"
)

var (
	httpHandler *endpoint.HttpHandler
)

func main() {
	httpHandler = endpoint.NewHttpHandler()
	go func() {
		logrus.Info("Server was started")
		err := fasthttp.ListenAndServe(":8000", httpHandler.Handle)
		if err != nil {
			logrus.Fatal("Server error: ", err.Error())
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)
	<-signals
}
