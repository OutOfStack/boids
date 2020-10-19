package main

import (
	"log"

	"golang.org/x/image/colornames"

	"github.com/OutOfStack/boids/config"
	"github.com/OutOfStack/boids/object"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
)

var (
	boids [config.BoidsCount]*object.Boid
)

func main() {
	for i := 0; i < config.BoidsCount; i++ {
		boid := object.CreateBoid(i)
		boids[i] = boid
		go boid.Start(&boids)
	}
	pixelgl.Run(run)
}

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Boids",
		Bounds: pixel.R(0, 0, config.Width, config.Height),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		log.Fatal(err)
	}

	for !win.Closed() {
		win.Clear(colornames.Black)

		imd := imdraw.New(nil)
		for _, boid := range boids {
			imd.Color = colornames.Gray
			imd.Push(pixel.V(boid.Position.X+2, boid.Position.Y),
				pixel.V(boid.Position.X-2, boid.Position.Y),
				pixel.V(boid.Position.X, boid.Position.Y-2),
				pixel.V(boid.Position.X, boid.Position.Y+2))
			imd.Polygon(1)
		}

		imd.Draw(win)

		win.Update()
	}
}
