package weather

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

type city struct {
	Latitude    float64     `json:"latitude"`
	Longitude   float64     `json:"longitude"`
	Title       string      `json:"title"`
	Offset      int         `json:"dstOffset"`
	Time        []time.Time `json:"time"`
	Temperature []float64   `json:"temperature_2m"`
}

func (c city) String() string {
	return fmt.Sprintf("%s (%.4f, %.4f)", c.Title, c.Latitude, c.Longitude)
}

type weatherCache struct {
	AvailableCities   []city `json:"available_cities"`
	cityMt            []sync.Mutex
	LastWeatherUpdate time.Time `json:"last_weather_update"`
	LastOffsetUpdate  time.Time `json:"last_offset_update"`
}

const geoNamesBaseURL = "http://api.geonames.org/timezoneJSON"
const geoNamesUsername = "amazing.website"
const openMeteoBaseURL = "https://api.open-meteo.com/v1/forecast"

type geoResponse struct {
	GMTOffset int `json:"gmtOffset"`
	DSTOffset int `json:"dstOffset"`
}

type hourly struct {
	Time        []string  `json:"time"`
	Temperature []float64 `json:"temperature_2m"`
}

type hourlyUnits struct {
	Time        string `json:"time"`
	Temperature string `json:"temperature_2m"`
}

type weatherResponce struct {
	Hourly      hourly      `json:"hourly"`
	HourlyUnits hourlyUnits `json:"hourly_units"`
}

func (cache *weatherCache) updateOffset(cityIndex int) error {
	if cityIndex < 0 || cityIndex >= len(cache.AvailableCities) {
		return errors.New("city index out of range")
	}

	params := url.Values{}
	params.Add("lat", strconv.FormatFloat(cache.AvailableCities[cityIndex].Latitude, 'f', 4, 64))
	params.Add("lng", strconv.FormatFloat(cache.AvailableCities[cityIndex].Longitude, 'f', 4, 64))
	params.Add("username", geoNamesUsername)

	fullURL := geoNamesBaseURL + "?" + params.Encode()
	response, err := httpClient.Get(fullURL)
	if err != nil {
		return fmt.Errorf("failed to get response from the geoNames API: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get timezone data: status code %d, city: %s, url: %s", response.StatusCode, cache.AvailableCities[cityIndex].String(), fullURL)
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed to read responce: status code %d, city: %s, url: %s", response.StatusCode, cache.AvailableCities[cityIndex].String(), fullURL)
	}

	var geoResp geoResponse
	err = json.Unmarshal(data, &geoResp)
	if err != nil {
		return fmt.Errorf("error unmarshalling geoResponse for city %s: %w", cache.AvailableCities[cityIndex].String(), err)
	}

	cache.cityMt[cityIndex].Lock()
	defer cache.cityMt[cityIndex].Unlock()

	cache.AvailableCities[cityIndex].Offset = geoResp.DSTOffset
	log.Printf("for the city %s new offset is %d", cache.AvailableCities[cityIndex].String(), cache.AvailableCities[cityIndex].Offset)

	return nil
}

func timeConverter(rawData []string, timeFormat string, offset time.Duration) ([]time.Time, error) {
	result := make([]time.Time, len(rawData))
	var timeLayout string

	// Define time format
	switch timeFormat {
	case "iso8601":
		timeLayout = "2006-01-02T15:04"
	default:
		return nil, fmt.Errorf("unsupported time format: %s", timeFormat)
	}

	for i := range rawData {
		parsedTime, err := time.Parse(timeLayout, rawData[i])
		if err != nil {
			return nil, fmt.Errorf("error parsing time value %s: %w", rawData[i], err)
		}

		result[i] = parsedTime //.Add(offset) TODO: make sure that we don't need to add offset
	}

	return result, nil
}

func (cache *weatherCache) updateWeather(cityIndex int) error {
	if cityIndex < 0 || cityIndex >= len(cache.AvailableCities) {
		return errors.New("city index out of range")
	}

	params := url.Values{}
	params.Add("latitude", strconv.FormatFloat(cache.AvailableCities[cityIndex].Latitude, 'f', 4, 64))
	params.Add("longitude", strconv.FormatFloat(cache.AvailableCities[cityIndex].Longitude, 'f', 4, 64))
	params.Add("hourly", "temperature_2m")

	fullURL := openMeteoBaseURL + "?" + params.Encode()
	log.Printf("asking weather: %s", fullURL)

	response, err := httpClient.Get(fullURL)
	if err != nil {
		return fmt.Errorf("failed to get response from the openWeather API: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get weather data: status code %d, city: %s, url: %s", response.StatusCode, cache.AvailableCities[cityIndex].String(), fullURL)
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed to read responce: status code %d, city: %s, url: %s", response.StatusCode, cache.AvailableCities[cityIndex].String(), fullURL)
	}

	var weatherResp weatherResponce
	err = json.Unmarshal(data, &weatherResp)
	if err != nil {
		return fmt.Errorf("error unmarshalling weatherResponce for city %s: %w", cache.AvailableCities[cityIndex].String(), err)
	}

	cache.cityMt[cityIndex].Lock()
	defer cache.cityMt[cityIndex].Unlock()

	cache.AvailableCities[cityIndex].Time, err = timeConverter(weatherResp.Hourly.Time, weatherResp.HourlyUnits.Time, time.Hour*time.Duration(cache.AvailableCities[cityIndex].Offset))
	if err != nil {
		return fmt.Errorf("error parsing time values for city %s: %w", cache.AvailableCities[cityIndex].String(), err)
	}
	cache.AvailableCities[cityIndex].Temperature = weatherResp.Hourly.Temperature
	log.Printf("for the city %s temperature is updated", cache.AvailableCities[cityIndex].String())

	return nil
}

func (cache *weatherCache) init() error {
	// Получаем путь к исполняемому файлу
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("[ERROR] Failed to get executable path: %v", err)
	}

	citiesPath := filepath.Join(filepath.Dir(execPath), "weatherdata", "weatherdata.json")

	data, err := os.ReadFile(citiesPath)
	if err != nil {
		return fmt.Errorf("[ERROR] Error reading weatherdata.json: %v", err)
	}
	err = json.Unmarshal(data, &cache)
	if err != nil {
		return fmt.Errorf("[ERROR] Error unmarshalling weatherdata.json: %v", err)
	}

	if len(cache.AvailableCities) == 0 {
		return errors.New("no cities found in weatherdata.json")
	}

	for _, e := range cache.AvailableCities {
		log.Printf("Loaded city: %s\n", e)
	}

	log.Printf("Last weather update: %s\n", cache.LastWeatherUpdate.Format(time.UnixDate))
	log.Printf("Last offset update: %s\n", cache.LastOffsetUpdate.Format(time.UnixDate))

	cache.cityMt = make([]sync.Mutex, len(cache.AvailableCities))

	return nil
}

func (cache *weatherCache) getCities() []string {
	result := make([]string, 0)
	for i := range cache.AvailableCities {
		data := ""
		cache.cityMt[i].Lock()

		for j := range cache.AvailableCities[i].Time {
			if time.Now().Before(cache.AvailableCities[i].Time[j]) {
				timeAsString := (cache.AvailableCities[i].Time[j].Add(time.Hour * time.Duration(cache.AvailableCities[i].Offset)).Format(time.DateTime))
				data = fmt.Sprintf("%s: %.1f (%s GMT+%d)", cache.AvailableCities[i].Title, cache.AvailableCities[i].Temperature[j], timeAsString, cache.AvailableCities[i].Offset)
				break
			}
		}

		cache.cityMt[i].Unlock()

		result = append(result, data)
	}
	result = append(result, "У нас еще есть и прогноз, но пока что мы его вам не покажем, мы не умеем делать красивые странички")
	result = append(result, "Кто мы, Саша, ту эту хрень один пишешь!")
	return result
}

var wCache weatherCache
