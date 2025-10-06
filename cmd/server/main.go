package main

import (
	"net/http"

	"github.com/iPatrushevSergey/metrics/internal/handler"
	"github.com/iPatrushevSergey/metrics/internal/repository/inmemory"
	"github.com/iPatrushevSergey/metrics/internal/service"
)

func main() {
	repo := inmemory.NewMemStorageMetricRepository()
	metricService := service.NewMetricService(repo)
	metricHandler := handler.NewMetricHandler(metricService)

	mux := http.NewServeMux()
	mux.HandleFunc(`/update/{type}/{name}/{value}`, metricHandler.Update)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
