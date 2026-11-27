package main

import (
	"context"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/OutOfStack/boids/config"
)

type vec3 [3]float64

func (v vec3) add(w vec3) vec3 {
	return vec3{v[0] + w[0], v[1] + w[1], v[2] + w[2]}
}

func (v vec3) sub(w vec3) vec3 {
	return vec3{v[0] - w[0], v[1] - w[1], v[2] - w[2]}
}

func (v vec3) scale(s float64) vec3 {
	return vec3{v[0] * s, v[1] * s, v[2] * s}
}

func (v vec3) lengthSquared() float64 {
	return v[0]*v[0] + v[1]*v[1] + v[2]*v[2]
}

const (
	minSpeed = 0.5
	maxSpeed = 1.0
)

const (
	grayFlock = iota
	orangeFlock
	blueFlock
	greenFlock
	flockCount
)

type boid struct {
	position vec3
	velocity vec3
	flock    int
}

type simulation struct {
	cfg    *config.Config
	extent vec3
	grid   spatialGrid

	mu      sync.RWMutex
	current []boid
	next    []boid
}

func newSimulation(cfg *config.Config) *simulation {
	extent := vec3{float64(cfg.Width), float64(cfg.Height), float64(cfg.Depth)}
	count := cfg.BoidsCount
	sim := &simulation{
		cfg:     cfg,
		extent:  extent,
		grid:    newSpatialGrid(extent, cfg.ViewRadius, count),
		current: make([]boid, count),
		next:    make([]boid, count),
	}

	random := rand.Float64
	if cfg.Seed != 0 {
		rng := rand.New(rand.NewSource(cfg.Seed)) //nolint:gosec // Simulation only.
		random = rng.Float64
	}
	for i := range sim.current {
		sim.current[i] = newBoid(i, extent, random)
	}
	return sim
}

func newBoid(i int, extent vec3, random func() float64) boid {
	flock := grayFlock
	switch {
	case i%7 == 0:
		flock = orangeFlock
	case i%11 == 0:
		flock = blueFlock
	case i%17 == 0:
		flock = greenFlock
	}

	b := boid{flock: flock}
	for axis, length := range extent {
		b.position[axis] = random() * length
		b.velocity[axis] = random()*2 - 1
	}
	return b
}

func (s *simulation) run(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(s.cfg.UpdateRateMs) * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.step()
		}
	}
}

func (s *simulation) step() {
	s.grid.rebuild(s.current)
	for i, b := range s.current {
		velocity := b.velocity.add(s.steer(i))
		// Keep speed in [minSpeed, maxSpeed] by length, so diagonal motion is
		// not faster than axial and groups whose velocities cancel keep moving.
		// Steering that cancels exactly keeps the previous heading.
		speed := math.Sqrt(velocity.lengthSquared())
		if speed == 0 {
			velocity, speed = b.velocity, math.Sqrt(b.velocity.lengthSquared())
		}
		if speed > 0 {
			velocity = velocity.scale(min(maxSpeed, max(minSpeed, speed)) / speed)
		}
		for axis, length := range s.extent {
			// Border steering turns boids back; a boid that still reaches a
			// wall is clamped to it and bounces off it.
			position := b.position[axis] + velocity[axis]
			if position <= 0 || position >= length {
				velocity[axis] = -velocity[axis]
			}
			b.position[axis] = min(length, max(0, position))
		}
		b.velocity = velocity
		s.next[i] = b
	}

	s.mu.Lock()
	s.current, s.next = s.next, s.current
	s.mu.Unlock()
}

func (s *simulation) snapshot(dst []boid) {
	s.mu.RLock()
	copy(dst, s.current)
	s.mu.RUnlock()
}

func (s *simulation) steer(i int) vec3 {
	self := s.current[i]
	radius := s.cfg.ViewRadius
	var alignment, cohesion, separation vec3
	var neighbors float64

	var cells [27]int
	for _, c := range s.grid.neighborCells(self.position, &cells) {
		for entry := s.grid.head[c]; entry != 0; entry = s.grid.next[entry-1] {
			j := entry - 1
			other := &s.current[j]
			if i == j || self.flock != other.flock {
				continue
			}

			delta := other.position.sub(self.position)
			distanceSquared := delta.lengthSquared()
			if distanceSquared == 0 || distanceSquared >= radius*radius {
				continue
			}
			neighbors++
			alignment = alignment.add(other.velocity)
			cohesion = cohesion.add(delta)
			separation = separation.sub(delta.scale(1 / math.Sqrt(distanceSquared)))
		}
	}

	var acceleration vec3
	for axis, length := range s.extent {
		acceleration[axis] = borderForce(self.position[axis], length, radius)
	}
	if neighbors > 0 {
		flocking := alignment.scale(1 / neighbors).sub(self.velocity).
			add(cohesion.scale(1 / neighbors)).add(separation)
		acceleration = acceleration.add(flocking.scale(s.cfg.AdjRate))
	}
	return acceleration
}

// borderForce pushes a boid away from the nearer face once it is within
// radius of it, growing as 1/distance.
func borderForce(position, length, radius float64) float64 {
	opposite := length - position
	switch {
	case position < opposite && position < radius:
		return 1 / max(1, position)
	case opposite < position && opposite < radius:
		return -1 / max(1, opposite)
	default:
		return 0
	}
}
