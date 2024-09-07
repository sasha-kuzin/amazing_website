package main

import (
	"fmt"
	"net/http"

	"github.com/sasha-kuzin/amazing_website/internal/httpgen"
	"github.com/sasha-kuzin/amazing_website/internal/weather"
)

func mainHandler(w http.ResponseWriter, r *http.Request) {
	httpgen.GenerateHttp(w, &httpgen.Data{Message: []string{"Вы охуеете, какой тут будет сайт"}, Header: "Главная", WhereToGo: "/weather"})
}

func weatherHandler(w http.ResponseWriter, r *http.Request) {
	httpgen.GenerateHttp(w, &httpgen.Data{Message: weather.LoadWeather(), Header: "Главная", WhereToGo: "/"})
}

func main() {
	http.HandleFunc("/", mainHandler)
	http.HandleFunc("/weather", weatherHandler)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Ошибка запуска сервера:", err)
	}
}
