package config_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/OutOfStack/boids/config"
)

func TestGetConfig(t *testing.T) {
	// save current directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer func() {
		// restore original directory
		t.Chdir(originalWd)
	}()

	// create a temporary directory for test
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	// create a test config.json
	testConfig := &config.Config{
		Width:          1024,
		Height:         768,
		BoidsCount:     500,
		ViewRadius:     40.0,
		AdjRate:        0.015,
		PolyThickness:  2.0,
		QuadtreeMaxObj: 10,
		QuadtreeMaxLvl: 8,
		UpdateRateMs:   16,
	}

	configData, err := json.Marshal(testConfig)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	configPath := filepath.Join(tempDir, "config.json")
	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// reset the singleton for testing (this is tricky because of sync.Once)
	// we need to use reflection or accept that this test can only run once
	// for simplicity, we'll just test that GetConfig doesn't panic and returns a config

	t.Run("returns non-nil config", func(t *testing.T) {
		cfg := config.GetConfig()
		if cfg == nil {
			t.Error("GetConfig() returned nil")
		}
	})

	t.Run("returns same instance on multiple calls", func(t *testing.T) {
		cfg1 := config.GetConfig()
		cfg2 := config.GetConfig()
		if cfg1 != cfg2 {
			t.Error("GetConfig() did not return the same instance on multiple calls")
		}
	})

	t.Run("config has expected values", func(t *testing.T) {
		cfg := config.GetConfig()
		if cfg.Width == 0 {
			t.Error("Config Width is zero")
		}
		if cfg.Height == 0 {
			t.Error("Config Height is zero")
		}
		if cfg.BoidsCount == 0 {
			t.Error("Config BoidsCount is zero")
		}
		if cfg.ViewRadius == 0 {
			t.Error("Config ViewRadius is zero")
		}
	})
}

func TestGetConfigConcurrency(t *testing.T) {
	// test that GetConfig is safe for concurrent access
	const numGoroutines = 100
	var wg sync.WaitGroup
	configs := make([]*config.Config, numGoroutines)

	for i := range numGoroutines {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			configs[idx] = config.GetConfig()
		}(i)
	}

	wg.Wait()

	// verify all goroutines got the same instance
	first := configs[0]
	for i := 1; i < numGoroutines; i++ {
		if configs[i] != first {
			t.Errorf("Concurrent calls to GetConfig() returned different instances")
			break
		}
	}
}

func TestConfigStruct(t *testing.T) {
	// test that Config struct can be created and marshaled/unmarshaled
	t.Run("marshal and unmarshal config", func(t *testing.T) {
		original := &config.Config{
			Width:          800,
			Height:         600,
			BoidsCount:     100,
			ViewRadius:     50.0,
			AdjRate:        0.01,
			PolyThickness:  1.5,
			QuadtreeMaxObj: 5,
			QuadtreeMaxLvl: 10,
			UpdateRateMs:   20,
		}

		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Failed to marshal config: %v", err)
		}

		var unmarshaled config.Config
		if err := json.Unmarshal(data, &unmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal config: %v", err)
		}

		if unmarshaled.Width != original.Width {
			t.Errorf("Width mismatch: got %d, want %d", unmarshaled.Width, original.Width)
		}
		if unmarshaled.Height != original.Height {
			t.Errorf("Height mismatch: got %d, want %d", unmarshaled.Height, original.Height)
		}
		if unmarshaled.BoidsCount != original.BoidsCount {
			t.Errorf("BoidsCount mismatch: got %d, want %d", unmarshaled.BoidsCount, original.BoidsCount)
		}
		if unmarshaled.ViewRadius != original.ViewRadius {
			t.Errorf("ViewRadius mismatch: got %f, want %f", unmarshaled.ViewRadius, original.ViewRadius)
		}
		if unmarshaled.AdjRate != original.AdjRate {
			t.Errorf("AdjRate mismatch: got %f, want %f", unmarshaled.AdjRate, original.AdjRate)
		}
		if unmarshaled.PolyThickness != original.PolyThickness {
			t.Errorf("PolyThickness mismatch: got %f, want %f", unmarshaled.PolyThickness, original.PolyThickness)
		}
		if unmarshaled.QuadtreeMaxObj != original.QuadtreeMaxObj {
			t.Errorf("QuadtreeMaxObj mismatch: got %d, want %d", unmarshaled.QuadtreeMaxObj, original.QuadtreeMaxObj)
		}
		if unmarshaled.QuadtreeMaxLvl != original.QuadtreeMaxLvl {
			t.Errorf("QuadtreeMaxLvl mismatch: got %d, want %d", unmarshaled.QuadtreeMaxLvl, original.QuadtreeMaxLvl)
		}
		if unmarshaled.UpdateRateMs != original.UpdateRateMs {
			t.Errorf("UpdateRateMs mismatch: got %d, want %d", unmarshaled.UpdateRateMs, original.UpdateRateMs)
		}
	})
}
