package main

import (
	"log"
	"math"
	"sync"

	"github.com/OutOfStack/boids/config"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/gopxl/pixel/v2/ext/imdraw"
	"golang.org/x/image/colornames"
)

var (
	boids       []*Boid
	boidsMatrix [][]int64
	rwLock      = &sync.RWMutex{}
)

func init() {
	cfg := config.GetConfig()
	boids = make([]*Boid, cfg.BoidsCount)
	boidsMatrix = make([][]int64, cfg.Width+1)
	for i := range boidsMatrix {
		boidsMatrix[i] = make([]int64, cfg.Height+1)
	}
}

func main() {
	cfg := config.GetConfig()
	// initialize boidsMatrix cells to -1 to indicate empty cells
	for i, row := range boidsMatrix {
		for j := range row {
			boidsMatrix[i][j] = -1
		}
	}

	// create boids and start their movement concurrently
	for i := range cfg.BoidsCount {
		boid := createBoid(i)
		boids[i] = boid
		go boid.start()
	}

	// start the rendering loop
	opengl.Run(run)
}

// handles the rendering of boids
func run() {
	cfg := config.GetConfig()

	windowCfg := opengl.WindowConfig{
		Title:  "Boids",
		Bounds: pixel.R(0, 0, float64(cfg.Width), float64(cfg.Height)),
		VSync:  true,
	}
	win, err := opengl.NewWindow(windowCfg)
	if err != nil {
		log.Fatal(err)
	}

	// main render loop
	for !win.Closed() {
		win.Clear(colornames.Black)

		imd := imdraw.New(nil)
		for _, b := range boids {
			// compute the angle of the boid's velocity for directional rendering
			angle := math.Atan2(b.velocity.Y, b.velocity.X)

			// calculate triangle vertices to represent the boid's direction
			size := float64(4)
			tip := pixel.V(b.position.X+size*math.Cos(angle),
				b.position.Y+size*math.Sin(angle))
			left := pixel.V(b.position.X+size*math.Cos(angle-2.3),
				b.position.Y+size*math.Sin(angle-2.3))
			right := pixel.V(b.position.X+size*math.Cos(angle+2.3),
				b.position.Y+size*math.Sin(angle+2.3))

			imd.Color = b.color
			imd.Push(tip, left, right)
			imd.Polygon(0) // filled triangle
		}

		imd.Draw(win)

		win.Update()
	}
}
