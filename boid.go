package main

import (
	"math"
	"math/rand"
	"time"

	"github.com/OutOfStack/boids/config"
	v "github.com/OutOfStack/boids/vector"
	"github.com/faiface/pixel"
)

//Boid represents boid
type Boid struct {
	position pixel.Vec
	velocity pixel.Vec
	id       int
}

func init() {

}

func createBoid(bid int) *Boid {
	boid := &Boid{
		position: pixel.V(rand.Float64()*float64(config.Width), rand.Float64()*float64(config.Height)),
		velocity: pixel.V(rand.Float64()*2-1.0, rand.Float64()*2-1.0),
		id:       bid,
	}
	boidMap[int(boid.position.X)][int(boid.position.Y)] = boid.id
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
	boidMap[int(b.position.X)][int(b.position.Y)] = -1
	b.position = b.position.Add(b.velocity)
	boidMap[int(b.position.X)][int(b.position.Y)] = b.id
	rwLock.Unlock()
}

func (b *Boid) calcAcceleration() pixel.Vec {
	upper, lower := v.AddV(b.position, config.ViewRadius), v.AddV(b.position, -config.ViewRadius)
	avgPosition, avgVelocity, separation := pixel.V(0, 0), pixel.V(0, 0), pixel.V(0, 0)
	count := 0.0

	rwLock.RLock()
	for i := math.Max(lower.X, 0); i <= math.Min(upper.X, config.Width); i++ {
		for j := math.Max(lower.Y, 0); j <= math.Min(upper.Y, config.Height); j++ {
			if otherBoidID := boidMap[int(i)][int(j)]; otherBoidID != -1 && otherBoidID != b.id {
				if dist := v.Distance(boids[otherBoidID].position, b.position); dist < config.ViewRadius {
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
		accellCohesion := avgPosition.Sub(b.position).Scaled(config.AdjRate)
		accelSeparation := separation.Scaled(config.AdjRate)
		accel = accel.Add(accelAlignment).Add(accellCohesion).Add(accelSeparation)
	}

	return accel
}

func (*Boid) borderBounce(pos, maxBorderPos float64) float64 {
	if pos < config.ViewRadius {
		return 1 / pos
	} else if pos > maxBorderPos-config.ViewRadius {
		return 1 / (pos - maxBorderPos)
	}
	return 0
}
