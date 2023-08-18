package main

import (
	"image/color"
	"math"
	"math/rand"
	"time"

	"github.com/OutOfStack/boids/config"
	v "github.com/OutOfStack/boids/vector"
	"github.com/faiface/pixel"
	"golang.org/x/image/colornames"
)

type Boid struct {
	id       int
	position pixel.Vec
	velocity pixel.Vec
	color    color.RGBA
}

func createBoid(bID int) *Boid {
	c := colornames.Gray
	switch {
	case bID%7 == 0:
		c = colornames.Darkorange
	case bID%11 == 0:
		c = colornames.Cornflowerblue
	case bID%13 == 0:
		c = colornames.Yellowgreen
	}
	boid := &Boid{
		id:       bID,
		position: pixel.V(rand.Float64()*float64(config.Width), rand.Float64()*float64(config.Height)),
		velocity: pixel.V(rand.Float64()*2-1.0, rand.Float64()*2-1.0),
		color:    c,
	}
	boidsMap[int(boid.position.X)][int(boid.position.Y)] = boid.id
	return boid
}

func (b *Boid) start() {
	for {
		b.moveOne()
		time.Sleep(10 * time.Millisecond)
	}
}

func (b *Boid) moveOne() {
	acceleration := b.calcAcceleration()
	rwLock.Lock()
	b.velocity = v.Limit(b.velocity.Add(acceleration), -1, 1)
	boidsMap[int(b.position.X)][int(b.position.Y)] = -1
	b.position = b.position.Add(b.velocity)
	boidsMap[int(b.position.X)][int(b.position.Y)] = b.id
	rwLock.Unlock()
}

func (b *Boid) calcAcceleration() pixel.Vec {
	upper, lower := v.AddV(b.position, config.ViewRadius), v.AddV(b.position, -config.ViewRadius)
	avgPosition, avgVelocity, separation := pixel.V(0, 0), pixel.V(0, 0), pixel.V(0, 0)
	count := 0.0

	rwLock.RLock()
	for i := math.Max(lower.X, 0); i <= math.Min(upper.X, config.Width); i++ {
		for j := math.Max(lower.Y, 0); j <= math.Min(upper.Y, config.Height); j++ {
			if otherBoidID := boidsMap[int(i)][int(j)]; otherBoidID != -1 && otherBoidID != b.id {
				otherBoid := boids[otherBoidID]
				if dist := v.Distance(otherBoid.position, b.position); dist < config.ViewRadius && otherBoid.color == b.color {
					count++
					avgVelocity = avgVelocity.Add(boids[otherBoidID].velocity)
					avgPosition = avgPosition.Add(boids[otherBoidID].position)
					separation = separation.Add(v.DivisionV(b.position.Sub(boids[otherBoidID].position), dist))
				}
			}
		}
	}
	rwLock.RUnlock()
	accel := pixel.V(b.borderBounce(b.position.X, config.Width), b.borderBounce(b.position.Y, config.Height))
	if count > 0 {
		avgPosition, avgVelocity = v.DivisionV(avgPosition, count), v.DivisionV(avgVelocity, count)
		accelAlignment := avgVelocity.Sub(b.velocity).Scaled(config.AdjRate)
		accelCohesion := avgPosition.Sub(b.position).Scaled(config.AdjRate)
		accelSeparation := separation.Scaled(config.AdjRate)
		accel = accel.Add(accelAlignment).Add(accelCohesion).Add(accelSeparation)
	}

	return accel
}

func (*Boid) borderBounce(pos, maxBorderPos float64) float64 {
	if pos < config.ViewRadius {
		return 1 / pos
	}
	if pos > maxBorderPos-config.ViewRadius {
		return 1 / (pos - maxBorderPos)
	}
	return 0
}
