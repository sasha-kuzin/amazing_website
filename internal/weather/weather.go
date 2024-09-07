package weather

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

func LoadWeather() []string {
	loadCities()
	result := []string{"Вы охуеете от этой погоды.", "Но вроде дальше будет получше.", "Мы очень надеемся. ", "Скоро покажем прогноз в этих городах:"}
	for _, e := range availableCities {
		result = append(result, e.Title)
	}
	return result
}

func tmp() {
	baseURL := "https://api.open-meteo.com/v1/forecast"
	params := url.Values{}
	params.Add("latitude", "43.3247,44.804,53.2521,55.7522")
	params.Add("longitude", "21.9033,20.4651,34.3717,37.6156")
	params.Add("hourly", "temperature_2m")

	fullURL := baseURL + "?" + params.Encode()
	response, err := http.Get(fullURL)
	if err != nil {
		fmt.Println("Error after sending a GET-request:", err)
		return
	}
	defer response.Body.Close()

	data, err := io.ReadAll(response.Body) // read responce
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("%s", data) //print as string
}
