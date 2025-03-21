package main

import (
	"fmt"
	"github.com/AlanMute/dpm-presets-service/internal/endpoint"
	"github.com/AlanMute/dpm-presets-service/internal/service"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"os"
	"os/signal"
)

var (
	httpHandler *endpoint.HttpHandler
)

func main() {
	storage := service.NewStorage()
	storage.AddPreset(service.Preset{
		ID:    0,
		Query: "платье",
	})

	storage.AddPreset(service.Preset{ID: 1, Query: "Красное платье"})

	query := "плать"
	presetID, found := storage.FindClosestPreset(query)
	if found {
		fmt.Printf("Найден пресет: %d\n", presetID)
	}

	query = "криснае платьё"
	presetID, found = storage.FindClosestPreset(query)
	if found {
		fmt.Printf("Найден пресет: %d\n", presetID)
	} else {
		fmt.Printf("А где\n")
	}

	query = "платье"
	presetID, found = storage.FindClosestPreset(query)
	if found {
		fmt.Printf("Найден пресет: %d\n", presetID)
	}

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
