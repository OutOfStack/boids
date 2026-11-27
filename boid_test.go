package main

import (
	"math"
	"slices"
	"testing"

	"github.com/OutOfStack/boids/config"
	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestSteeringIgnoresBoidsBeyondWall(t *testing.T) {
	t.Parallel()

	sim := newSimulation(&config.Config{
		Width: 100, Height: 100, Depth: 100, BoidsCount: 5, ViewRadius: 10, AdjRate: 0.25, Seed: 1,
	})
	sim.current[0] = boid{position: vec3{50, 50, 1}}
	sim.current[1] = boid{position: vec3{50, 50, 98}, velocity: vec3{1, 0, 0}} // near, but across the wall
	sim.current[2] = boid{position: vec3{50, 50, 5}, velocity: vec3{0, 1, 0}}
	sim.current[3] = boid{position: vec3{50, 50, 2}, velocity: vec3{1, 1, 1}, flock: blueFlock}
	sim.current[4] = boid{position: vec3{58, 50, 8}}
	sim.grid.rebuild(sim.current)

	// Flocking with boid 2 only, (0, 1, 3) * 0.25, plus a push of 1 off the z=0 face.
	want := vec3{0, 0.25, 1.75}
	if got := sim.steer(0); got != want {
		t.Fatalf("acceleration = %v, want %v", got, want)
	}
}

func TestBorderForceWithOverlappingBands(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		position float64
		want     float64
	}{
		{2.5, 0.4},
		{5, 0},
		{7.5, -0.4},
	} {
		if got := borderForce(test.position, 10, 8); got != test.want {
			t.Errorf("force at %v = %v, want %v", test.position, got, test.want)
		}
	}
}

func TestStepKeepsSpeedInRange(t *testing.T) {
	t.Parallel()

	sim := newSimulation(&config.Config{Width: 100, Height: 100, Depth: 100, BoidsCount: 1, Seed: 1})
	for _, test := range []struct {
		velocity vec3
		want     float64
	}{
		{vec3{1, 1, 1}, maxSpeed},
		{vec3{0.1, 0, 0.1}, minSpeed},
	} {
		sim.current[0] = boid{position: vec3{50, 50, 50}, velocity: test.velocity}
		sim.step()

		var snapshot [1]boid
		sim.snapshot(snapshot[:])
		if speed := math.Sqrt(snapshot[0].velocity.lengthSquared()); math.Abs(speed-test.want) > 1e-9 {
			t.Errorf("speed from %v = %v, want %v", test.velocity, speed, test.want)
		}
	}
}

func TestStepClampsToWalls(t *testing.T) {
	t.Parallel()

	sim := newSimulation(&config.Config{Width: 100, Height: 100, Depth: 75, BoidsCount: 1, Seed: 1})
	sim.current[0] = boid{position: vec3{50, 50, 0.25}, velocity: vec3{0, 0, -1}, flock: orangeFlock}
	sim.step()

	var snapshot [1]boid
	sim.snapshot(snapshot[:])
	want := boid{position: vec3{50, 50, 0}, velocity: vec3{0, 0, 1}, flock: orangeFlock}
	if snapshot[0] != want {
		t.Fatalf("boid after step = %+v, want %+v", snapshot[0], want)
	}

	// With no border steering (view_radius 0) the bounce alone brings it back.
	sim.step()
	sim.snapshot(snapshot[:])
	if z := snapshot[0].position[2]; z != 1 {
		t.Fatalf("z after bounce = %v, want 1", z)
	}

	// Landing exactly on a wall bounces too, instead of waiting a step there.
	sim.current[0] = boid{position: vec3{50, 50, 1}, velocity: vec3{0, 0, -1}}
	sim.step()
	sim.snapshot(snapshot[:])
	if got := snapshot[0]; got.position[2] != 0 || got.velocity[2] != 1 {
		t.Fatalf("boid after exact contact = %+v, want z 0 moving +1", got)
	}
	sim.current[0] = boid{position: vec3{50, 50, 74}, velocity: vec3{0, 0, 1}}
	sim.step()
	sim.snapshot(snapshot[:])
	if got := snapshot[0]; got.position[2] != 75 || got.velocity[2] != -1 {
		t.Fatalf("boid after exact contact = %+v, want z 75 moving -1", got)
	}
}

func TestStepKeepsSpeedWhenSteeringCancelsVelocity(t *testing.T) {
	t.Parallel()

	// Alignment, cohesion and separation exactly cancel each boid's velocity.
	sim := newSimulation(&config.Config{
		Width: 100, Height: 100, Depth: 100, BoidsCount: 2, ViewRadius: 10, AdjRate: 1, Seed: 1,
	})
	sim.current[0] = boid{position: vec3{49, 50, 50}, velocity: vec3{1, 0, 0}}
	sim.current[1] = boid{position: vec3{51, 50, 50}, velocity: vec3{-1, 0, 0}}
	start := [2]vec3{sim.current[0].position, sim.current[1].position}
	sim.grid.rebuild(sim.current)
	if got := sim.current[0].velocity.add(sim.steer(0)); got != (vec3{}) {
		t.Fatalf("fixture no longer cancels: velocity + steering = %v", got)
	}
	sim.step()

	var snapshot [2]boid
	sim.snapshot(snapshot[:])
	for i, b := range snapshot {
		if speed := math.Sqrt(b.velocity.lengthSquared()); speed < minSpeed {
			t.Errorf("boid %d speed = %v, want at least %v", i, speed, minSpeed)
		}
		if b.position == start[i] {
			t.Errorf("boid %d did not move from %v", i, b.position)
		}
	}
}

func TestSpatialGridStopsAtWalls(t *testing.T) {
	t.Parallel()

	// 1000 boids in a 100^3 world give 10 cells of width 10 per axis.
	boids := make([]boid, 1000)
	for i := range boids {
		boids[i].position = vec3{50, 50, 50}
	}
	boids[0].position = vec3{1, 1, 1}
	boids[1].position = vec3{99, 99, 99} // opposite corner
	boids[2].position = vec3{15, 15, 15} // diagonal neighbor cell
	grid := newSpatialGrid(vec3{100, 100, 100}, 10, len(boids))
	grid.rebuild(boids)

	var cells [27]int
	var visited []int
	for _, c := range grid.neighborCells(boids[0].position, &cells) {
		for entry := grid.head[c]; entry != 0; entry = grid.next[entry-1] {
			visited = append(visited, entry-1)
		}
	}
	slices.Sort(visited)
	if !slices.Equal(visited, []int{0, 2}) {
		t.Fatalf("visited = %v, want [0 2]", visited)
	}
}

func TestBoidTransformPointsAlongVelocity(t *testing.T) {
	t.Parallel()

	position := vec3{40, 50, 60}
	for _, velocity := range []vec3{{1, 0, 0}, {0, -1, 0}, {0, 0, 1}, {0, 0, -1}, {0, 0, 0}} {
		transform := boidTransform(boid{position: position, velocity: velocity})
		want := renderVector(position.add(velocity))
		if velocity == (vec3{}) {
			want.Y++
		}
		if tip := rl.Vector3Transform(rl.Vector3{Y: 1}, transform); tip != want {
			t.Errorf("tip for velocity %v = %v, want %v", velocity, tip, want)
		}
	}
}
