package quadtree_test

import (
	"testing"

	"github.com/OutOfStack/boids/quadtree"
	"github.com/gopxl/pixel/v2"
)

func TestBounds(t *testing.T) {
	b := &quadtree.Bounds{X: 0, Y: 0, Width: 10, Height: 10}
	if !b.Contains(pixel.V(5, 5)) {
		t.Fatal("Contains should be true")
	}
	if b.Contains(pixel.V(10, 10)) { // edge is exclusive on high side
		t.Fatal("Contains should be false on high edge")
	}
	c := &quadtree.Bounds{X: 8, Y: 8, Width: 5, Height: 5}
	if !b.Intersects(c) {
		t.Fatal("Intersects should be true")
	}
}

func TestInsertSplitQuery(t *testing.T) {
	qt := quadtree.NewQuadTree(quadtree.Bounds{X: 0, Y: 0, Width: 100, Height: 100}, 0, 4, 5)
	// insert many objects to trigger split
	for i := range 50 {
		qt.Insert(&quadtree.Object{ID: int64(i), Position: pixel.V(float64(i*2), float64(i*2))})
	}
	// query a small range
	r := &quadtree.Bounds{X: 0, Y: 0, Width: 10, Height: 10}
	res := qt.Query(r)
	if len(res) == 0 {
		t.Fatal("expected some results in query")
	}
	// circle query around (0,0)
	cres := qt.QueryCircle(pixel.V(0, 0), 5)
	if len(cres) == 0 {
		t.Fatal("expected circle query results")
	}
}

func TestRemove(t *testing.T) {
	qt := quadtree.NewQuadTree(quadtree.Bounds{X: 0, Y: 0, Width: 10, Height: 10}, 0, 4, 5)
	qt.Insert(&quadtree.Object{ID: 1, Position: pixel.V(5, 5)})
	if !qt.Remove(1) {
		t.Fatal("expected remove to return true")
	}
	if qt.Remove(1) {
		t.Fatal("expected remove to return false when absent")
	}
}
