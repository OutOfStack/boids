package object

import (
	"math/rand"
	"time"

	"github.com/faiface/pixel"
)

type Boid struct {
	Position pixel.Vec
	velocity pixel.Vec
	id       int
}

var width, height int

func CreateBoid(bid, maxX, maxY int) *Boid {
	width = maxX
	height = maxY
	boid := &Boid{
		Position: pixel.V(rand.Float64()*float64(width), rand.Float64()*float64(height)),
		velocity: pixel.V(rand.Float64()*2-1.0, rand.Float64()*2-1.0),
		id:       bid,
	}
	return boid
}

func (b *Boid) Start() {
	for {
		b.moveOne()
		time.Sleep(10 * time.Millisecond)
	}
}

func (b *Boid) moveOne() {
	b.Position = b.Position.Add(b.velocity)
	next := b.Position.Add(b.velocity)
	if next.X >= float64(width) || next.X < 0 {
		b.velocity = pixel.V(-b.velocity.X, b.velocity.Y)
	}
	if next.Y >= float64(height) || next.Y < 0 {
		b.velocity = pixel.V(b.velocity.X, -b.velocity.Y)
	}
}
