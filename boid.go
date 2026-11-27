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
	half   vec3 // extent / 2, for shortest wrapped displacement
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
		half:    extent.scale(0.5),
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
		// Cap speed by length so diagonal motion is not faster than axial.
		if speedSquared := velocity.lengthSquared(); speedSquared > 1 {
			velocity = velocity.scale(1 / math.Sqrt(speedSquared))
		}
		for axis, length := range s.extent {
			b.position[axis] = wrap(b.position[axis]+velocity[axis], length)
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
			for axis, length := range s.extent {
				// Use the shortest displacement across each periodic boundary.
				switch half := s.half[axis]; {
				case delta[axis] > half:
					delta[axis] -= length
				case delta[axis] < -half:
					delta[axis] += length
				}
			}
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

	if neighbors == 0 {
		return vec3{}
	}
	flocking := alignment.scale(1 / neighbors).sub(self.velocity).
		add(cohesion.scale(1 / neighbors)).add(separation)
	return flocking.scale(s.cfg.AdjRate)
}

func wrap(position, length float64) float64 {
	if position < 0 {
		position += length
	}
	if position >= length {
		return position - length
	}
	return position
}
