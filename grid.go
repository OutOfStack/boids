package main

import "math"

type cell [3]int

type spatialGrid struct {
	size  cell
	width vec3
	head  []int // first boid+1 in each cell, indexed by x+nx*(y+ny*z); 0 is empty
	next  []int
}

func newSpatialGrid(extent vec3, radius float64, count int) spatialGrid {
	grid := spatialGrid{next: make([]int, count)}
	// Cells at least as wide as the radius let us check one cell in each
	// direction, including across wrapped boundaries. Cells no smaller than the
	// volume per boid keep the cell count near the boid count for any radius.
	volume := extent[0] * extent[1] * extent[2]
	cellWidth := max(radius, math.Cbrt(volume/float64(max(1, count))))
	for axis, length := range extent {
		grid.size[axis] = max(1, int(length/cellWidth))
		grid.width[axis] = length / float64(grid.size[axis])
	}
	grid.head = make([]int, grid.size[0]*grid.size[1]*grid.size[2])
	return grid
}

func (g *spatialGrid) index(x, y, z int) int {
	return x + g.size[0]*(y+g.size[1]*z)
}

func (g *spatialGrid) cellAt(position vec3) cell {
	var index cell
	for axis, width := range g.width {
		index[axis] = min(g.size[axis]-1, int(position[axis]/width))
	}
	return index
}

func (g *spatialGrid) rebuild(boids []boid) {
	clear(g.head)
	for i, b := range boids {
		c := g.cellAt(b.position)
		index := g.index(c[0], c[1], c[2])
		g.next[i] = g.head[index]
		g.head[index] = i + 1
	}
}

// neighborCells fills dst with the indices of the cells around position and
// returns them. Iterating head/next directly in the caller keeps the hot loop
// free of callbacks.
func (g *spatialGrid) neighborCells(position vec3, dst *[27]int) []int {
	index := g.cellAt(position)
	nx, ny, nz := g.size[0], g.size[1], g.size[2]
	cells := dst[:0]
	// A wrapped axis with fewer than three cells only needs one visit per cell.
	for dz := range min(3, nz) {
		z := (index[2] + dz + nz - 1) % nz
		for dy := range min(3, ny) {
			y := (index[1] + dy + ny - 1) % ny
			for dx := range min(3, nx) {
				x := (index[0] + dx + nx - 1) % nx
				cells = append(cells, g.index(x, y, z))
			}
		}
	}
	return cells
}
