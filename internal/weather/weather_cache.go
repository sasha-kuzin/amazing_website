package weather

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type city struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Title     string  `json:"title"`
}

func (c city) String() string {
	return fmt.Sprintf("%s (%.4f, %.4f)", c.Title, c.Latitude, c.Longitude)
}

var availableCities []city

func loadCities() {
	// Получаем путь к исполняемому файлу
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("[ERROR] Failed to get executable path: %v", err)
	}

	citiesPath := filepath.Join(filepath.Dir(execPath), "weatherdata", "cities.json")

	data, err := os.ReadFile(citiesPath)
	if err != nil {
		log.Printf("[ERROR] Error reading cities.json: %v", err)
		return
	}
	err = json.Unmarshal(data, &availableCities)
	if err != nil {
		log.Printf("[ERROR] Error unmarshalling cities.json: %v", err)
		return
	}

	for _, e := range availableCities {
		log.Printf("Loaded city: %s\n", e)
	}
}
