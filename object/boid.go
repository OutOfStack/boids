package object

import (
	"math"
	"math/rand"
	"time"

	"github.com/OutOfStack/boids/config"
	"github.com/faiface/pixel"
)

type Boid struct {
	Position pixel.Vec
	velocity pixel.Vec
	id       int
}

var boidMap [config.Width + 1][config.Height + 1]int
var boids *[config.BoidsCount]*Boid

func init() {
	for i, row := range boidMap {
		for j := range row {
			boidMap[i][j] = -1
		}
	}
}

func CreateBoid(bid int) *Boid {
	boid := &Boid{
		Position: pixel.V(rand.Float64()*float64(config.Width), rand.Float64()*float64(config.Height)),
		velocity: pixel.V(rand.Float64()*2-1.0, rand.Float64()*2-1.0),
		id:       bid,
	}
	boidMap[int(boid.Position.X)][int(boid.Position.Y)] = boid.id
	return boid
}

func (b *Boid) Start(bs *[config.BoidsCount]*Boid) {
	boids = bs
	for {
		b.moveOne()
		time.Sleep(10 * time.Millisecond)
	}
}

func (b *Boid) moveOne() {
	b.velocity = limit(b.velocity.Add(b.calcAcceleration()), -1, 1)
	boidMap[int(b.Position.X)][int(b.Position.Y)] = -1
	b.Position = b.Position.Add(b.velocity)
	boidMap[int(b.Position.X)][int(b.Position.Y)] = b.id
	next := b.Position.Add(b.velocity)
	if next.X >= float64(config.Width) || next.X < 0 {
		b.velocity = pixel.V(-b.velocity.X, b.velocity.Y)
	}
	if next.Y >= float64(config.Height) || next.Y < 0 {
		b.velocity = pixel.V(b.velocity.X, -b.velocity.Y)
	}
}

func (b *Boid) calcAcceleration() pixel.Vec {
	upper, lower := addV(b.Position, config.ViewRadius), addV(b.Position, -config.ViewRadius)
	avgVelocity := pixel.V(0, 0)
	count := 0.0
	for i := math.Max(lower.X, 0); i <= math.Min(upper.X, config.Width); i++ {
		for j := math.Max(lower.Y, 0); j <= math.Min(upper.Y, config.Height); j++ {
			if otherBoidId := boidMap[int(i)][int(j)]; otherBoidId != -1 && otherBoidId != b.id {
				if dist := distance(boids[otherBoidId].Position, b.Position); dist < config.ViewRadius {
					count++
					avgVelocity = avgVelocity.Add(boids[otherBoidId].velocity)
				}
			}
		}
	}
	accel := pixel.V(0, 0)
	if count > 0 {
		avgVelocity = divisionV(avgVelocity, count)
		accel = avgVelocity.Sub(b.velocity).Scaled(config.AdjRate)
	}

	return accel
}
