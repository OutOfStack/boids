package config

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

// Config - all simulation parameters
type Config struct {
	Width      int32   `json:"width"`
	Height     int32   `json:"height"`
	BoidsCount int64   `json:"boids_count"`
	ViewRadius float64 `json:"view_radius"`
	AdjRate    float64 `json:"adj_rate"`
}

var (
	instance *Config
	once     sync.Once
)

// GetConfig returns config instance
func GetConfig() *Config {
	once.Do(func() {
		// load from config file
		instance = &Config{}
		data, err := os.ReadFile("config.json")
		if err != nil {
			log.Fatal(err)
		}
		_ = json.Unmarshal(data, instance)
	})
	return instance
}
