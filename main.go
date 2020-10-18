package main

import (
	"log"

	"golang.org/x/image/colornames"

	"github.com/OutOfStack/boids/object"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
)

const (
	width, height = 800, 480
	boidsCount    = 500
)

var (
	boids [boidsCount]*object.Boid
)

func main() {
	for i := 0; i < boidsCount; i++ {
		boid := object.CreateBoid(i, width, height)
		boids[i] = boid
		go boid.Start()
	}
	pixelgl.Run(run)
}

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Boids",
		Bounds: pixel.R(0, 0, width, height),
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
			imd.Color = colornames.Green
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
