package main

import (
	"image/color"
	"math/rand"

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
			rand.Float64()*float64(config.GetConfig().Width),   //nolint:gosec
			rand.Float64()*float64(config.GetConfig().Height)), //nolint:gosec
		// random initial velocity in the range [-1, 1] for both X and Y
		velocity: pixel.V(
			rand.Float64()*2-1.0,  //nolint:gosec
			rand.Float64()*2-1.0), //nolint:gosec
		color: c,
	}

	return boid
}

// Updates the boid's velocity and position based on the calculated acceleration
func (b *Boid) moveOne() {
	// calculate acceleration
	acceleration := b.calcAcceleration()
	rwLock.Lock()
	defer rwLock.Unlock()

	// update velocity with acceleration and limit to [-1, 1]
	b.velocity = v.Limit(b.velocity.Add(acceleration), -1, 1)

	// update position based on new velocity
	oldPosition := b.position
	b.position = b.position.Add(b.velocity)

	// wrap around screen edges if needed
	cfg := config.GetConfig()
	width, height := float64(cfg.Width), float64(cfg.Height)

	if b.position.X < 0 {
		b.position.X += width
	} else if b.position.X >= width {
		b.position.X -= width
	}

	if b.position.Y < 0 {
		b.position.Y += height
	} else if b.position.Y >= height {
		b.position.Y -= height
	}

	// update the boid's position in the quadtree
	if oldPosition != b.position {
		qtree.Update(b.id, b.position)
	}
}

// Computes the steering acceleration based on nearby boids.
// It takes into account alignment, cohesion, separation, and border avoidance
func (b *Boid) calcAcceleration() pixel.Vec {
	cfg := config.GetConfig()

	// query the quadtree for nearby boids
	rwLock.RLock()
	nearbyObjects := qtree.QueryCircle(b.position, cfg.ViewRadius)
	rwLock.RUnlock()

	avgPosition, avgVelocity, separation := pixel.V(0, 0), pixel.V(0, 0), pixel.V(0, 0)
	count := 0.0

	width, height := float64(cfg.Width), float64(cfg.Height)

	// process nearby boids
	for _, obj := range nearbyObjects {
		if obj.ID == b.id {
			continue // skip self
		}

		otherBoid := boids[obj.ID]

		// consider only boids with matching color group
		if otherBoid.color == b.color {
			dist := v.Distance(otherBoid.position, b.position)

			// consider only boids within view radius
			if dist < cfg.ViewRadius {
				count++
				avgVelocity = avgVelocity.Add(otherBoid.velocity)
				avgPosition = avgPosition.Add(otherBoid.position)
				// calculate separation: steer away from neighbors
				separation = separation.Add(v.DivisionV(b.position.Sub(otherBoid.position), dist))
			}
		}
	}

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

// Provides a force to steer the boid away from boundaries
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
