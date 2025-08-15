package vector_test

import (
	"testing"

	"github.com/OutOfStack/boids/vector"
	"github.com/gopxl/pixel/v2"
)

func TestLimit(t *testing.T) {
	v := pixel.V(2, -3)
	got := vector.Limit(v, -1, 1)
	if got.X != 1 || got.Y != -1 {
		t.Fatalf("Limit failed: got=%v", got)
	}
}

func TestDistanceSquared(t *testing.T) {
	a := pixel.V(0, 0)
	b := pixel.V(3, 4)
	if ds := vector.DistanceSquared(a, b); ds != 25 {
		t.Fatalf("DistanceSquared expected 25, got %v", ds)
	}
}

func TestDivisionV(t *testing.T) {
	v := pixel.V(10, -5)
	got := vector.DivisionV(v, 5)
	if got.X != 2 || got.Y != -1 {
		t.Fatalf("DivisionV failed: got=%v", got)
	}
	// d=0 should return input
	got = vector.DivisionV(v, 0)
	if got != v {
		t.Fatalf("DivisionV with d=0 should return input: got=%v", got)
	}
}
