package main

import (
	"math"
	"testing"

	"github.com/OutOfStack/boids/config"
	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestSteeringAcrossDepthBoundary(t *testing.T) {
	t.Parallel()

	sim := newSimulation(&config.Config{
		Width: 100, Height: 100, Depth: 100, BoidsCount: 5, ViewRadius: 10, AdjRate: 0.25, Seed: 1,
	})
	sim.current[0] = boid{position: vec3{50, 50, 1}}
	sim.current[1] = boid{position: vec3{50, 50, 98}, velocity: vec3{1, 0, 0}}
	sim.current[2] = boid{position: vec3{50, 50, 50}, velocity: vec3{1, 1, 1}}
	sim.current[3] = boid{position: vec3{50, 50, 2}, velocity: vec3{1, 1, 1}, flock: blueFlock}
	sim.current[4] = boid{position: vec3{58, 50, 8}}
	sim.grid.rebuild(sim.current)

	want := vec3{0.25, 0, -0.5}
	if got := sim.steer(0); got != want {
		t.Fatalf("acceleration = %v, want %v", got, want)
	}
}

func TestStepCapsSpeedByLength(t *testing.T) {
	t.Parallel()

	sim := newSimulation(&config.Config{Width: 100, Height: 100, Depth: 100, BoidsCount: 1, Seed: 1})
	sim.current[0] = boid{position: vec3{50, 50, 50}, velocity: vec3{1, 1, 1}}
	sim.step()

	var snapshot [1]boid
	sim.snapshot(snapshot[:])
	if speed := snapshot[0].velocity.lengthSquared(); math.Abs(speed-1) > 1e-9 {
		t.Fatalf("speed squared = %v, want 1", speed)
	}
}

func TestStepWrapsDepth(t *testing.T) {
	t.Parallel()

	sim := newSimulation(&config.Config{Width: 100, Height: 100, Depth: 75, BoidsCount: 1, Seed: 1})
	sim.current[0] = boid{position: vec3{50, 50, 0.25}, velocity: vec3{0, 0, -1}, flock: orangeFlock}
	sim.step()

	var snapshot [1]boid
	sim.snapshot(snapshot[:])
	want := boid{position: vec3{50, 50, 74.25}, velocity: vec3{0, 0, -1}, flock: orangeFlock}
	if snapshot[0] != want {
		t.Fatalf("boid after step = %+v, want %+v", snapshot[0], want)
	}
}

func TestSpatialGridVisitsWrappedCellsOnce(t *testing.T) {
	t.Parallel()

	grid := newSpatialGrid(vec3{15, 25, 35}, 10, 3)
	boids := []boid{
		{position: vec3{1, 1, 1}},
		{position: vec3{1, 1, 1}},
		{position: vec3{14, 24, 34}},
	}
	grid.rebuild(boids)

	var visits [3]int
	var cells [27]int
	for _, c := range grid.neighborCells(boids[0].position, &cells) {
		for entry := grid.head[c]; entry != 0; entry = grid.next[entry-1] {
			visits[entry-1]++
		}
	}
	if visits != [3]int{1, 1, 1} {
		t.Fatalf("visits = %v, want one per boid", visits)
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
