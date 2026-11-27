package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"sync"
	"time"
)

// Config - all simulation parameters
type Config struct {
	Width        int32   `json:"width"`
	Height       int32   `json:"height"`
	Depth        int32   `json:"depth"`
	BoidsCount   int     `json:"boids_count"`
	ViewRadius   float64 `json:"view_radius"`
	AdjRate      float64 `json:"adj_rate"`
	BoidSize     float32 `json:"boid_size"`
	UpdateRateMs int     `json:"update_rate_ms"`
	// Seed enables deterministic runs; if 0, a random seed is used.
	Seed int64 `json:"seed,omitempty"`
}

const (
	cfgPath = "config.json"
	// maxBoids keeps a typo from turning into a huge allocation.
	maxBoids = 1_000_000
)

var (
	instance *Config
	once     sync.Once
)

// GetConfig returns config instance
func GetConfig() *Config {
	once.Do(func() {
		data, err := os.ReadFile(cfgPath)
		if err != nil {
			log.Fatal(err)
		}
		instance, err = parse(data)
		if err != nil {
			log.Fatalf("%s: %v", cfgPath, err)
		}
	})
	return instance
}

func parse(data []byte) (*Config, error) {
	cfg := Config{BoidSize: 8}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	// Only an omitted depth defaults to height; an explicit 0 is rejected below.
	var depth struct {
		Depth *int32 `json:"depth"`
	}
	if err := json.Unmarshal(data, &depth); err != nil {
		return nil, err
	}
	if depth.Depth == nil {
		cfg.Depth = cfg.Height
	}
	switch {
	case cfg.Width <= 0 || cfg.Height <= 0 || cfg.Depth <= 0:
		return nil, errors.New("width, height, and depth must be positive")
	case cfg.BoidsCount < 0 || cfg.BoidsCount > maxBoids:
		return nil, fmt.Errorf("boids_count must be between 0 and %d", maxBoids)
	case cfg.ViewRadius < 0:
		return nil, errors.New("view_radius must be nonnegative")
	case cfg.BoidSize <= 0:
		return nil, errors.New("boid_size must be positive")
	case cfg.UpdateRateMs <= 0 || time.Duration(cfg.UpdateRateMs) > math.MaxInt64/time.Millisecond:
		return nil, errors.New("update_rate_ms must fit in a positive time.Duration")
	}
	return &cfg, nil
}
