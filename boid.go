package main

import (
	"image/color"
	"math"
	"math/rand"
	"time"

	"github.com/OutOfStack/boids/config"
	v "github.com/OutOfStack/boids/vector"
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

// initializes a new boid with random position and velocity.
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
		// Random initial position within simulation bounds.
		position: pixel.V(
			rand.Float64()*float64(config.GetConfig().Width),   //nolint:gosec
			rand.Float64()*float64(config.GetConfig().Height)), //nolint:gosec
		// Random initial velocity in the range [-1, 1] for both X and Y
		velocity: pixel.V(
			rand.Float64()*2-1.0,  //nolint:gosec
			rand.Float64()*2-1.0), //nolint:gosec
		color: c,
	}

	boidsMatrix[int(boid.position.X)][int(boid.position.Y)] = boid.id
	return boid
}

// begins the boid's movement loop in its own goroutine
func (b *Boid) start() {
	for {
		b.moveOne()
		time.Sleep(10 * time.Millisecond)
	}
}

// updates the boid's velocity and position based on the calculated acceleration
func (b *Boid) moveOne() {
	acceleration := b.calcAcceleration()
	rwLock.Lock()
	defer rwLock.Unlock()
	// update velocity with acceleration and limit to [-1, 1]
	b.velocity = v.Limit(b.velocity.Add(acceleration), -1, 1)
	// clear the boid's previous position in the boidsMatrix
	boidsMatrix[int(b.position.X)][int(b.position.Y)] = -1
	// update position based on new velocity
	b.position = b.position.Add(b.velocity)
	boidsMatrix[int(b.position.X)][int(b.position.Y)] = b.id
}

// computes the steering acceleration based on nearby boids.
// It takes into account alignment, cohesion, separation, and border avoidance
func (b *Boid) calcAcceleration() pixel.Vec {
	cfg := config.GetConfig()

	upper, lower := v.AddV(b.position, cfg.ViewRadius), v.AddV(b.position, -cfg.ViewRadius)
	avgPosition, avgVelocity, separation := pixel.V(0, 0), pixel.V(0, 0), pixel.V(0, 0)
	count := 0.0

	width, height := float64(cfg.Width), float64(cfg.Height)

	rwLock.RLock()
	// iterate over grid cells in the neighborhood
	for i := math.Max(lower.X, 0); i <= math.Min(upper.X, width); i++ {
		for j := math.Max(lower.Y, 0); j <= math.Min(upper.Y, height); j++ {
			if otherBoidID := boidsMatrix[int(i)][int(j)]; otherBoidID != -1 && otherBoidID != b.id {
				otherBoid := boids[otherBoidID]
				// consider only boids within view radius and matching color group
				if dist := v.Distance(otherBoid.position, b.position); dist < cfg.ViewRadius && otherBoid.color == b.color {
					count++
					avgVelocity = avgVelocity.Add(boids[otherBoidID].velocity)
					avgPosition = avgPosition.Add(boids[otherBoidID].position)
					// calculate separation: steer away from neighbors
					separation = separation.Add(v.DivisionV(b.position.Sub(boids[otherBoidID].position), dist))
				}
			}
		}
	}
	rwLock.RUnlock()

	// start with border bounce acceleration to avoid edges
	accel := pixel.V(b.borderBounce(b.position.X, width), b.borderBounce(b.position.Y, height))
	if count > 0 {
		// compute average position and velocity
		avgPosition, avgVelocity = v.DivisionV(avgPosition, count), v.DivisionV(avgVelocity, count)
		// alignment: steer towards average velocity
		accelAlignment := avgVelocity.Sub(b.velocity).Scaled(cfg.AdjRate)
		// cohesion: steer towards average position
		accelCohesion := avgPosition.Sub(b.position).Scaled(cfg.AdjRate)
		// separation: steer away to avoid crowding
		accelSeparation := separation.Scaled(cfg.AdjRate)
		// combine steering behaviors
		accel = accel.Add(accelAlignment).Add(accelCohesion).Add(accelSeparation)
	}

	return accel
}

// provides a force to steer the boid away from boundaries
func (*Boid) borderBounce(pos, maxBorderPos float64) float64 {
	cfg := config.GetConfig()
	// apply force when close to the left/top boundary.
	if pos < cfg.ViewRadius {
		return 1 / pos
	}
	// apply a force when close to the right/bottom boundary
	if pos > maxBorderPos-cfg.ViewRadius {
		return 1 / (pos - maxBorderPos)
	}
	return 0
}
