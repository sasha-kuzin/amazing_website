package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func initCache() error {
	// Initialize the cache by loading data from the file
	err := wCache.init()
	if err != nil {
		return fmt.Errorf("failed to initialize cache: %v", err)
	}

	now := time.Now()
	needsSave := false

	// First, check and update offsets if needed
	updatedOffset, err := updateOffsetIfNeeded(now)
	if err != nil {
		return err
	}

	// Then, check and update weather data if needed
	updatedWeather, err := updateWeatherIfNeeded(now)
	if err != nil {
		return err
	}

	// Mark the cache for saving if either offsets or weather were updated
	if updatedWeather || updatedOffset {
		needsSave = true
	}

	// Save the cache if there were any updates
	if needsSave {
		log.Println("Saving cache due to data updates")
		err := saveCacheToFile()
		if err != nil {
			return fmt.Errorf("error saving cache: %v", err)
		}
	}

	return nil
}

// Updates weather data for all cities if it's outdated.
func updateWeatherIfNeeded(now time.Time) (bool, error) {
	if now.Sub(wCache.LastWeatherUpdate) > time.Hour {
		log.Println("Weather data is outdated, updating...")
		for i := range wCache.AvailableCities {
			err := wCache.updateWeather(i)
			if err != nil {
				return false, fmt.Errorf("error updating weather for city %s: %v", wCache.AvailableCities[i].String(), err)
			}
		}
		wCache.LastWeatherUpdate = time.Now()
		return true, nil
	}
	return false, nil
}

// Updates offset data for all cities if it's outdated.
func updateOffsetIfNeeded(now time.Time) (bool, error) {
	if now.Sub(wCache.LastOffsetUpdate) > 24*time.Hour {
		log.Println("Offset data is outdated, updating...")
		for i := range wCache.AvailableCities {
			err := wCache.updateOffset(i)
			if err != nil {
				return false, fmt.Errorf("error updating offset for city %s: %v", wCache.AvailableCities[i].String(), err)
			}
		}
		wCache.LastOffsetUpdate = time.Now()
		return true, nil
	}
	return false, nil
}

// Runs the scheduler for periodic updates of weather, offsets, and cache saving.
func RunScheduler(ctx context.Context, wg *sync.WaitGroup) error {
	wg.Add(3)

	errChan := make(chan error, 3)

	// Launch periodic update goroutines for weather, offsets, and saving the cache.
	go func() {
		errChan <- weatherUpdateJob(ctx, wg)
	}()
	go func() {
		errChan <- offsetUpdateJob(ctx, wg)
	}()
	go func() {
		errChan <- cacheSaveJob(ctx, wg)
	}()

	// Wait for either context cancellation or an error.
	select {
	case <-ctx.Done():
		return nil
	case err := <-errChan:
		return err
	}
}

// Periodically updates weather data for all cities.
func weatherUpdateJob(ctx context.Context, wg *sync.WaitGroup) error {
	defer wg.Done()
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping weather update job")
			return nil
		case <-ticker.C:
			log.Println("Updating weather data")
			// Update offsets first if needed before updating weather.
			_, err := updateOffsetIfNeeded(time.Now())
			if err != nil {
				return err
			}
			// Now update the weather data.
			_, err = updateWeatherIfNeeded(time.Now())
			if err != nil {
				return err
			}
		}
	}
}

// Periodically updates timezone offsets for all cities.
func offsetUpdateJob(ctx context.Context, wg *sync.WaitGroup) error {
	defer wg.Done()
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping offset update job")
			return nil
		case <-ticker.C:
			log.Println("Updating offsets")
			_, err := updateOffsetIfNeeded(time.Now())
			if err != nil {
				return err
			}
		}
	}
}

// Periodically saves the cache to the file.
func cacheSaveJob(ctx context.Context, wg *sync.WaitGroup) error {
	defer wg.Done()
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping cache save job")
			return nil
		case <-ticker.C:
			log.Println("Saving cache to file")
			err := saveCacheToFile()
			if err != nil {
				return fmt.Errorf("error saving cache: %w", err)
			}
		}
	}
}

// Saves the current cache state to the weatherdata.json file.
func saveCacheToFile() error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	citiesPath := filepath.Join(filepath.Dir(execPath), "weatherdata", "weatherdata.json")

	file, err := os.Create(citiesPath)
	if err != nil {
		return fmt.Errorf("failed to create cities file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(wCache); err != nil {
		return fmt.Errorf("failed to encode cache to file: %w", err)
	}

	return nil
}

func StartWeatherService(ctx context.Context, masterWg *sync.WaitGroup) {
	defer masterWg.Done()

	// Initialize the cache and check for outdated data
	err := initCache()
	if err != nil {
		log.Fatalf("Failed to initialize cache: %v", err)
	}

	wg := sync.WaitGroup{}

	// Start the scheduler for updating weather and offsets
	err = RunScheduler(ctx, &wg)
	if err != nil {
		log.Printf("Scheduler stopped due to an error: %v", err)
	}

	<-(ctx).Done()
	log.Println("Shutting down weather service...")
	wg.Wait()

	log.Println("Weather service gracefully stopped")
}

func LoadWeather() []string {
	// Return a message along with a list of cities from the cache
	result := []string{"Вы охуеете от этой погоды.", "Но вроде дальше будет получше.", "Мы очень надеемся."}
	result = append(result, wCache.getCities()...)
	return result
}
