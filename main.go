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
	boids   [config.BoidsCount]*Boid
	boidMap [config.Width + 1][config.Height + 1]int
	rwLock  = sync.RWMutex{}
)

func main() {
	for i, row := range boidMap {
		for j := range row {
			boidMap[i][j] = -1
		}
	}

	for i := 0; i < config.BoidsCount; i++ {
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
		for i, boid := range boids {
			switch {
			case i%11 == 0:
				imd.Color = colornames.Darkorange
			case i%17 == 0:
				imd.Color = colornames.Cornflowerblue
			case i%29 == 0:
				imd.Color = colornames.Yellowgreen
			case i%47 == 0:
				imd.Color = colornames.Whitesmoke
			default:
				imd.Color = colornames.Gray
			}
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
