package main

import (
	"math"
	"math/rand"
	"time"

	"github.com/OutOfStack/boids/config"
	"github.com/OutOfStack/boids/v"
	"github.com/faiface/pixel"
)

type Boid struct {
	position pixel.Vec
	velocity pixel.Vec
	id       int
}

func init() {

}

func CreateBoid(bid int) *Boid {
	boid := &Boid{
		position: pixel.V(rand.Float64()*float64(config.Width), rand.Float64()*float64(config.Height)),
		velocity: pixel.V(rand.Float64()*2-1.0, rand.Float64()*2-1.0),
		id:       bid,
	}
	boidMap[int(boid.position.X)][int(boid.position.Y)] = boid.id
	return boid
}

func (b *Boid) Start() {
	for {
		b.moveOne()
		time.Sleep(10 * time.Millisecond)
	}
}

func (b *Boid) moveOne() {
	acceleration := b.calcAcceleration()
	lock.Lock()
	b.velocity = v.Limit(b.velocity.Add(acceleration), -1, 1)
	boidMap[int(b.position.X)][int(b.position.Y)] = -1
	b.position = b.position.Add(b.velocity)
	boidMap[int(b.position.X)][int(b.position.Y)] = b.id
	next := b.position.Add(b.velocity)
	if next.X >= float64(config.Width) || next.X < 0 {
		b.velocity = pixel.V(-b.velocity.X, b.velocity.Y)
	}
	if next.Y >= float64(config.Height) || next.Y < 0 {
		b.velocity = pixel.V(b.velocity.X, -b.velocity.Y)
	}
	lock.Unlock()
}

func (b *Boid) calcAcceleration() pixel.Vec {
	upper, lower := v.AddV(b.position, config.ViewRadius), v.AddV(b.position, -config.ViewRadius)
	avgPosition, avgVelocity := pixel.V(0, 0), pixel.V(0, 0)
	count := 0.0

	lock.Lock()
	for i := math.Max(lower.X, 0); i <= math.Min(upper.X, config.Width); i++ {
		for j := math.Max(lower.Y, 0); j <= math.Min(upper.Y, config.Height); j++ {
			if otherBoidId := boidMap[int(i)][int(j)]; otherBoidId != -1 && otherBoidId != b.id {
				if dist := v.Distance(boids[otherBoidId].position, b.position); dist < config.ViewRadius {
					count++
					avgVelocity = avgVelocity.Add(boids[otherBoidId].velocity)
					avgPosition = avgPosition.Add(boids[otherBoidId].position)
				}
			}
		}
	}
	lock.Unlock()
	accel := pixel.V(0, 0)
	if count > 0 {
		avgPosition, avgVelocity = v.DivisionV(avgPosition, count), v.DivisionV(avgVelocity, count)
		accelAlignment := avgVelocity.Sub(b.velocity).Scaled(config.AdjRate)
		accellCohesion := avgPosition.Sub(b.position).Scaled(config.AdjRate)
		accel = accel.Add(accelAlignment).Add(accellCohesion)
	}

	return accel
}
