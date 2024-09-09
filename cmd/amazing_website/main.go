package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sasha-kuzin/amazing_website/internal/httpgen"
	"github.com/sasha-kuzin/amazing_website/internal/weather"
)

func mainHandler(w http.ResponseWriter, r *http.Request) {
	httpgen.GenerateHttp(w, &httpgen.Data{Message: []string{"Вы охуеете, какой тут будет сайт"}, Header: "Главная", WhereToGo: "/weather", WhereToGoCapture: "Посмотреть погоду"})
}

func weatherHandler(w http.ResponseWriter, r *http.Request) {
	httpgen.GenerateHttp(w, &httpgen.Data{Message: weather.LoadWeather(), Header: "Погода", WhereToGo: "/", WhereToGoCapture: "На главную"})
}

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	wg := sync.WaitGroup{}
	wg.Add(1)
	go weather.StartWeatherService(ctx, &wg)

	http.HandleFunc("/", mainHandler)
	http.HandleFunc("/weather", weatherHandler)

	server := &http.Server{Addr: ":8080"}
	go func() {
		log.Println("Server started on port 8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println("Error starting server:", err)
		}
	}()

	<-stop

	log.Println("Shutting down server...")
	cancel()

	wg.Wait()

	ctxShutDown, shutDownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutDownCancel()

	if err := server.Shutdown(ctxShutDown); err != nil {
		log.Println("Error shutting down server:", err)
	}

	log.Println("Server stopped")
}
