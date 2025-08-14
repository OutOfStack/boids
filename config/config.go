package config

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

// Config - all simulation parameters
type Config struct {
	Width          int32   `json:"width"`
	Height         int32   `json:"height"`
	BoidsCount     int64   `json:"boids_count"`
	ViewRadius     float64 `json:"view_radius"`
	AdjRate        float64 `json:"adj_rate"`
	PolyThickness  float64 `json:"poly_thickness"`
	QuadtreeMaxObj int     `json:"quadtree_max_obj"`
	QuadtreeMaxLvl int     `json:"quadtree_max_lvl"`
	UpdateRateMs   int     `json:"update_rate_ms"`
	// Seed enables deterministic runs; if 0, a random seed is used.
	Seed int64 `json:"seed,omitempty"`
}

const (
	cfgPath = "config.json"
)

var (
	instance *Config
	once     sync.Once
)

// GetConfig returns config instance
func GetConfig() *Config {
	once.Do(func() {
		// load config from file
		instance = &Config{}
		data, err := os.ReadFile(cfgPath)
		if err != nil {
			log.Fatal(err)
		}
		_ = json.Unmarshal(data, instance)
	})
	return instance
}
