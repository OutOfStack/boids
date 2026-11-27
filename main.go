package main

import (
	"context"
	"log"

	"github.com/OutOfStack/boids/config"
)

func main() {
	if err := run(config.GetConfig()); err != nil {
		log.Fatal(err)
	}
}

func run(cfg *config.Config) error {
	sim := newSimulation(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go sim.run(ctx)
	return render(sim)
}
