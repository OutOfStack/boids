package main

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/OutOfStack/boids/config"
	"github.com/OutOfStack/boids/vector"
	"github.com/gopxl/pixel/v2"
	"golang.org/x/image/colornames"
)

// Boid - boid model
type Boid struct {
	id       int64      // Unique identifier for the boid
	position pixel.Vec  // Current position in 2D space
	velocity pixel.Vec  // Current velocity vector
	color    color.RGBA // Color used for rendering
}

var rng = rand.New(rand.NewSource(1)) //nolint:gosec

func setSeed(seed int64) {
	if seed == 0 {
		return
	}
	rng = rand.New(rand.NewSource(seed)) //nolint:gosec
}

// Initializes a new boid with random position and velocity.
// The boid's color is chosen based on its id for visual variety
func createBoid(bID int64) *Boid {
	c := colornames.Gray
	switch {
	case bID%7 == 0:
		c = colornames.Darkorange
	case bID%11 == 0:
		c = colornames.Cornflowerblue
	case bID%17 == 0:
		c = colornames.Yellowgreen
	}

	boid := &Boid{
		id: bID,
		// random initial position within simulation bounds.
		position: pixel.V(
			rng.Float64()*float64(config.GetConfig().Width),
			rng.Float64()*float64(config.GetConfig().Height)),
		// random initial velocity in the range [-1, 1] for both X and Y
		velocity: pixel.V(
			rng.Float64()*2-1.0,
			rng.Float64()*2-1.0),
		color: c,
	}

	return boid
}

// Computes the steering acceleration for boid i based on snapshots and the quadtree built from snapshots.
func calcAccelerationFor(i int, positions, velocities []pixel.Vec) pixel.Vec {
	cfg := config.GetConfig()
	selfPos := positions[i]
	selfVel := velocities[i]

	// query the quadtree for nearby boids (including ghosts)
	nearbyObjects := qtree.QueryCircle(selfPos, cfg.ViewRadius)

	avgPosition, avgVelocity, separation := pixel.V(0, 0), pixel.V(0, 0), pixel.V(0, 0)
	count := 0.0

	seen := make(map[int64]struct{})

	// process nearby boids, deduping ghosts by ID
	for _, obj := range nearbyObjects {
		if obj.ID == int64(i) {
			continue
		}
		if _, ok := seen[obj.ID]; ok {
			continue
		}
		seen[obj.ID] = struct{}{}

		otherPos := positions[int(obj.ID)]
		otherVel := velocities[int(obj.ID)]

		// consider only boids with matching color group
		if boids[int(obj.ID)].color == boids[i].color {
			dx := otherPos.X - selfPos.X
			dy := otherPos.Y - selfPos.Y
			dist2 := dx*dx + dy*dy
			r2 := cfg.ViewRadius * cfg.ViewRadius
			if dist2 < r2 && dist2 > 0 {
				d := math.Sqrt(dist2)
				count++
				avgVelocity = avgVelocity.Add(otherVel)
				avgPosition = avgPosition.Add(otherPos)
				// separation: steer away from neighbors
				sep := vector.DivisionV(selfPos.Sub(otherPos), d)
				separation = separation.Add(sep)
			}
		}
	}

	width, height := float64(cfg.Width), float64(cfg.Height)
	// start with border bounce acceleration to avoid edges
	accel := pixel.V(borderBounce(selfPos.X, width), borderBounce(selfPos.Y, height))
	if count > 0 {
		avgPosition, avgVelocity = vector.DivisionV(avgPosition, count), vector.DivisionV(avgVelocity, count)
		accelAlignment := avgVelocity.Sub(selfVel).Scaled(cfg.AdjRate)
		accelCohesion := avgPosition.Sub(selfPos).Scaled(cfg.AdjRate)
		accelSeparation := separation.Scaled(cfg.AdjRate)
		accel = accel.Add(accelAlignment).Add(accelCohesion).Add(accelSeparation)
	}

	return accel
}

// Provides a force to steer the boid away from boundaries with clamping to avoid infinities
func borderBounce(pos, maxBorderPos float64) float64 {
	cfg := config.GetConfig()
	eps := 1e-3
	maxForce := 1.0
	if pos < cfg.ViewRadius {
		v := 1.0 / math.Max(pos, eps)
		if v > maxForce {
			v = maxForce
		}
		return v
	}
	if pos > maxBorderPos-cfg.ViewRadius {
		v := 1.0 / math.Max(maxBorderPos-pos, eps)
		if v > maxForce {
			v = maxForce
		}
		return -v
	}
	return 0
}
