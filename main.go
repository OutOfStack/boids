package main

import (
	"log"
	"math"
	"runtime"
	"sync"
	"time"

	"github.com/OutOfStack/boids/config"
	"github.com/OutOfStack/boids/quadtree"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/gopxl/pixel/v2/ext/imdraw"
	"golang.org/x/image/colornames"
)

var (
	boids  []*Boid
	qtree  *quadtree.QuadTree
	rwLock = &sync.RWMutex{}
)

func init() {
	cfg := config.GetConfig()
	boids = make([]*Boid, cfg.BoidsCount)

	// initialize the quadtree with the simulation bounds
	qtree = quadtree.NewQuadTree(quadtree.Bounds{
		X:      0,
		Y:      0,
		Width:  float64(cfg.Width),
		Height: float64(cfg.Height),
	}, 0)
}

func main() {
	cfg := config.GetConfig()

	// create boids
	for i := range cfg.BoidsCount {
		boid := createBoid(i)
		boids[i] = boid

		// insert the boid into the quadtree
		qtree.Insert(&quadtree.Object{
			ID:       boid.id,
			Position: boid.position,
		})
	}

	// start a fixed number of worker goroutines to update boids.
	// this prevents excessive goroutine creation with large boid counts
	numWorkers := runtime.NumCPU() * 2
	boidsPerWorker := int(cfg.BoidsCount) / numWorkers

	for w := range numWorkers {
		startIdx := w * boidsPerWorker
		endIdx := startIdx + boidsPerWorker
		if w == numWorkers-1 {
			endIdx = int(cfg.BoidsCount) // the last worker gets any remaining boids
		}

		go func(start, end int) {
			for {
				for i := start; i < end; i++ {
					boids[i].moveOne()
				}
				// sleep according to the configured update rate
				time.Sleep(time.Duration(cfg.UpdateRateMs) * time.Millisecond)
			}
		}(startIdx, endIdx)
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
			imd.Polygon(cfg.PolyThickness) // filled triangle
		}

		imd.Draw(win)

		win.Update()
	}
}
