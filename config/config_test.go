package config_test

import (
	"os"
	"sync"
	"testing"

	"github.com/OutOfStack/boids/config"
)

func TestGetConfig(t *testing.T) {
	t.Chdir(t.TempDir())
	legacy := `{
		"width": 1024, "height": 768, "boids_count": 500, "view_radius": 40,
		"adj_rate": -0.2, "update_rate_ms": 16, "seed": 2,
		"poly_thickness": 2, "quadtree_max_obj": 10, "quadtree_max_lvl": 8
	}`
	if err := os.WriteFile("config.json", []byte(legacy), 0600); err != nil {
		t.Fatal(err)
	}
	want := config.Config{
		Width: 1024, Height: 768, Depth: 768, BoidsCount: 500, ViewRadius: 40,
		AdjRate: -0.2, BoidSize: 8, UpdateRateMs: 16, Seed: 2,
	}

	var loaded [16]*config.Config
	var readers sync.WaitGroup
	for i := range loaded {
		readers.Go(func() { loaded[i] = config.GetConfig() })
	}
	readers.Wait()

	if got := *loaded[0]; got != want {
		t.Fatalf("config = %+v, want %+v", got, want)
	}
	for _, got := range loaded[1:] {
		if got != loaded[0] {
			t.Fatal("GetConfig returned different instances")
		}
	}
}
