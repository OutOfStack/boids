package main

import (
	"log"
	"sync"

	"github.com/OutOfStack/boids/config"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

var (
	boids    [config.BoidsCount]*Boid
	boidsMap [config.Width + 1][config.Height + 1]int
	rwLock   = sync.RWMutex{}
)

func main() {
	for i, row := range boidsMap {
		for j := range row {
			boidsMap[i][j] = -1
		}
	}

	for i := range config.BoidsCount {
		boid := createBoid(i)
		boids[i] = boid
		go boid.start()
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
			imd.Color = boid.color
			imd.Push(pixel.V(boid.position.X+2, boid.position.Y),
				pixel.V(boid.position.X-2, boid.position.Y),
				pixel.V(boid.position.X, boid.position.Y-2),
				pixel.V(boid.position.X, boid.position.Y+2))
			imd.Polygon(1.5)
		}

		imd.Draw(win)

		win.Update()
	}
}
