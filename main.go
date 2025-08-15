package main

import (
	"context"
	"log"
	"math"
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

func main() {
	cfg := config.GetConfig()
	setSeed(cfg.Seed)

	// initialize boids
	boids = make([]*Boid, cfg.BoidsCount)
	for i := range cfg.BoidsCount {
		boids[i] = createBoid(i)
	}

	// build initial quadtree from snapshot positions
	rebuildQuadTreeSnapshot()

	// run simulation in a separate goroutine at fixed update rate
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go simulationLoop(ctx)

	// start the rendering loop
	opengl.Run(func() { run(cancel) })
}

// handles the rendering of boids
func run(cancel context.CancelFunc) {
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

	imd := imdraw.New(nil)

	// main render loop
	for !win.Closed() {
		win.Clear(colornames.Black)
		rwLock.RLock()
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
		rwLock.RUnlock()
		imd.Draw(win)
		imd.Clear()

		win.Update()
	}

	// request simulation shutdown when window closes
	cancel()
}

// simulationLoop runs the two-phase update at a fixed tick rate
func simulationLoop(ctx context.Context) {
	cfg := config.GetConfig()
	ticker := time.NewTicker(time.Duration(cfg.UpdateRateMs) * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tickOnce()
		}
	}
}

// tickOnce performs: snapshot -> build qtree -> compute -> apply -> rebuild qtree for next queries
func tickOnce() {
	cfg := config.GetConfig()
	// snapshot positions and velocities
	rwLock.RLock()
	positions := make([]pixel.Vec, len(boids))
	velocities := make([]pixel.Vec, len(boids))
	for i, b := range boids {
		positions[i] = b.position
		velocities[i] = b.velocity
	}
	rwLock.RUnlock()

	// build quadtree from snapshot for neighbor queries, with wrap-around ghosts
	buildQuadTreeWithGhosts(positions)

	// compute accelerations and apply updates
	newPositions := make([]pixel.Vec, len(boids))
	newVelocities := make([]pixel.Vec, len(boids))

	for i := range boids {
		accel := calcAccelerationFor(i, positions, velocities)
		// limit velocity and integrate
		nv := velocities[i].Add(accel)
		if nv.X > 1 {
			nv.X = 1
		} else if nv.X < -1 {
			nv.X = -1
		}
		if nv.Y > 1 {
			nv.Y = 1
		} else if nv.Y < -1 {
			nv.Y = -1
		}
		np := positions[i].Add(nv)
		// wrap around
		width, height := float64(cfg.Width), float64(cfg.Height)
		if np.X < 0 {
			np.X += width
		} else if np.X >= width {
			np.X -= width
		}
		if np.Y < 0 {
			np.Y += height
		} else if np.Y >= height {
			np.Y -= height
		}
		newPositions[i] = np
		newVelocities[i] = nv
	}

	// apply
	rwLock.Lock()
	for i, b := range boids {
		b.position = newPositions[i]
		b.velocity = newVelocities[i]
	}
	rwLock.Unlock()

	// rebuild quadtree from updated positions for next frame queries
	rebuildQuadTreeSnapshot()
}

// rebuildQuadTreeSnapshot builds a quadtree with current boid positions (no ghosts)
func rebuildQuadTreeSnapshot() {
	rwLock.RLock()
	positions := make([]pixel.Vec, len(boids))
	for i, b := range boids {
		positions[i] = b.position
	}
	rwLock.RUnlock()
	buildQuadTreeWithGhosts(positions)
}

// buildQuadTreeWithGhosts builds qtree from given positions and inserts ghost objects near borders to emulate toroidal space
func buildQuadTreeWithGhosts(positions []pixel.Vec) {
	cfg := config.GetConfig()
	qt := quadtree.NewQuadTree(quadtree.Bounds{X: 0, Y: 0, Width: float64(cfg.Width), Height: float64(cfg.Height)}, 0, cfg.QuadtreeMaxObj, cfg.QuadtreeMaxLvl)
	width := float64(cfg.Width)
	height := float64(cfg.Height)
	r := cfg.ViewRadius

	for i := range positions {
		p := positions[i]
		// insert original
		qt.Insert(&quadtree.Object{ID: int64(i), Position: p})

		// ghosts when within view radius of edges
		nearLeft := p.X < r
		nearRight := p.X > width-r
		nearBottom := p.Y < r
		nearTop := p.Y > height-r

		if nearLeft {
			qt.Insert(&quadtree.Object{ID: int64(i), Position: pixel.V(p.X+width, p.Y)})
		}
		if nearRight {
			qt.Insert(&quadtree.Object{ID: int64(i), Position: pixel.V(p.X-width, p.Y)})
		}
		if nearBottom {
			qt.Insert(&quadtree.Object{ID: int64(i), Position: pixel.V(p.X, p.Y+height)})
		}
		if nearTop {
			qt.Insert(&quadtree.Object{ID: int64(i), Position: pixel.V(p.X, p.Y-height)})
		}
		// corners
		if nearLeft && nearBottom {
			qt.Insert(&quadtree.Object{ID: int64(i), Position: pixel.V(p.X+width, p.Y+height)})
		}
		if nearLeft && nearTop {
			qt.Insert(&quadtree.Object{ID: int64(i), Position: pixel.V(p.X+width, p.Y-height)})
		}
		if nearRight && nearBottom {
			qt.Insert(&quadtree.Object{ID: int64(i), Position: pixel.V(p.X-width, p.Y+height)})
		}
		if nearRight && nearTop {
			qt.Insert(&quadtree.Object{ID: int64(i), Position: pixel.V(p.X-width, p.Y-height)})
		}
	}

	// swap global qtree
	rwLock.Lock()
	qtree = qt
	rwLock.Unlock()
}
