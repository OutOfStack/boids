package quadtree

import (
	"github.com/gopxl/pixel/v2"
)

const (
	// NumQuadrants defines how many quadrants each node has
	NumQuadrants = 4
)

// Bounds represents a rectangular area
type Bounds struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

// Contains checks if a point is within the bounds
func (b *Bounds) Contains(p pixel.Vec) bool {
	return p.X >= b.X && p.X < b.X+b.Width &&
		p.Y >= b.Y && p.Y < b.Y+b.Height
}

// Intersects checks if two bounds intersect
func (b *Bounds) Intersects(other *Bounds) bool {
	return !(other.X > b.X+b.Width ||
		other.X+other.Width < b.X ||
		other.Y > b.Y+b.Height ||
		other.Y+other.Height < b.Y)
}

// Object represents an object in the quadtree
type Object struct {
	ID       int64
	Position pixel.Vec
}

// QuadTree represents a quadtree node
type QuadTree struct {
	bounds  Bounds
	objects []*Object
	nodes   [4]*QuadTree
	level   int
	divided bool
	maxObj  int
	maxLvl  int
}

// NewQuadTree creates a new quadtree
func NewQuadTree(bounds Bounds, level int, maxObj, maxLvl int) *QuadTree {
	return &QuadTree{
		bounds:  bounds,
		objects: make([]*Object, 0, maxObj),
		level:   level,
		divided: false,
		maxObj:  maxObj,
		maxLvl:  maxLvl,
	}
}

// Clear removes all objects from the quadtree
func (qt *QuadTree) Clear() {
	qt.objects = make([]*Object, 0, qt.maxObj)

	if qt.divided {
		for i := range NumQuadrants {
			qt.nodes[i].Clear()
			qt.nodes[i] = nil
		}
		qt.divided = false
	}
}

// Split divides the node into four quadrants
func (qt *QuadTree) Split() {
	subWidth := qt.bounds.Width / 2
	subHeight := qt.bounds.Height / 2
	x := qt.bounds.X
	y := qt.bounds.Y

	// create four children nodes
	qt.nodes[0] = NewQuadTree(Bounds{X: x + subWidth, Y: y + subHeight, Width: subWidth, Height: subHeight}, qt.level+1, qt.maxObj, qt.maxLvl) // northeast
	qt.nodes[1] = NewQuadTree(Bounds{X: x, Y: y + subHeight, Width: subWidth, Height: subHeight}, qt.level+1, qt.maxObj, qt.maxLvl)            // northwest
	qt.nodes[2] = NewQuadTree(Bounds{X: x, Y: y, Width: subWidth, Height: subHeight}, qt.level+1, qt.maxObj, qt.maxLvl)                        // southwest
	qt.nodes[3] = NewQuadTree(Bounds{X: x + subWidth, Y: y, Width: subWidth, Height: subHeight}, qt.level+1, qt.maxObj, qt.maxLvl)             // southeast

	qt.divided = true

	// redistribute existing objects to children
	for _, obj := range qt.objects {
		for i := range NumQuadrants {
			if qt.nodes[i].bounds.Contains(obj.Position) {
				qt.nodes[i].Insert(obj)
				break
			}
		}
	}

	// clear the parent's objects
	qt.objects = make([]*Object, 0, qt.maxObj)
}

// GetIndex determines which node the object belongs to
func (qt *QuadTree) GetIndex(obj *Object) int {
	idx := -1
	midX := qt.bounds.X + qt.bounds.Width/2
	midY := qt.bounds.Y + qt.bounds.Height/2

	// object can completely fit within the top quadrants
	topQuadrant := obj.Position.Y >= midY
	// object can completely fit within the bottom quadrants
	bottomQuadrant := obj.Position.Y < midY

	// object can completely fit within the left quadrants
	if obj.Position.X < midX {
		if topQuadrant {
			idx = 1 // northwest
		} else if bottomQuadrant {
			idx = 2 // southwest
		}
	} else if obj.Position.X >= midX { // object can completely fit within the right quadrants
		if topQuadrant {
			idx = 0 // northeast
		} else if bottomQuadrant {
			idx = 3 // southeast
		}
	}

	return idx
}

// Insert adds an object to the quadtree
func (qt *QuadTree) Insert(obj *Object) {
	// if this node is divided, insert the object into the appropriate child
	if qt.divided {
		idx := qt.GetIndex(obj)
		if idx != -1 {
			qt.nodes[idx].Insert(obj)
			return
		}
	}

	// add the object to this node
	qt.objects = append(qt.objects, obj)

	// check if we need to split the node
	if len(qt.objects) > qt.maxObj && qt.level < qt.maxLvl {
		if !qt.divided {
			qt.Split()
		}

		// redistribute objects to children
		i := 0
		for i < len(qt.objects) {
			idx := qt.GetIndex(qt.objects[i])
			if idx != -1 {
				// move the object to the child
				qt.nodes[idx].Insert(qt.objects[i])
				// remove from this node (swap with last element and truncate)
				qt.objects[i] = qt.objects[len(qt.objects)-1]
				qt.objects = qt.objects[:len(qt.objects)-1]
			} else {
				// this object stays in this node
				i++
			}
		}
	}
}

// Update updates an object's position in the quadtree
func (qt *QuadTree) Update(id int64, newPosition pixel.Vec) {
	// remove the object
	qt.Remove(id)

	// insert it with the new position
	qt.Insert(&Object{
		ID:       id,
		Position: newPosition,
	})
}

// Remove removes an object from the quadtree
func (qt *QuadTree) Remove(id int64) bool {
	// check if the object is in this node
	for i, obj := range qt.objects {
		if obj.ID == id {
			// remove the object (swap with last element and truncate)
			qt.objects[i] = qt.objects[len(qt.objects)-1]
			qt.objects = qt.objects[:len(qt.objects)-1]
			return true
		}
	}

	// if this node is divided, check the children
	if qt.divided {
		for i := range NumQuadrants {
			if qt.nodes[i].Remove(id) {
				return true
			}
		}
	}

	return false
}

// Query returns all objects in the specified range
func (qt *QuadTree) Query(rang *Bounds) []*Object {
	result := make([]*Object, 0)

	// if the range doesn't intersect this node, return empty
	if !qt.bounds.Intersects(rang) {
		return result
	}

	// add objects from this node that are within the range
	for _, obj := range qt.objects {
		if rang.Contains(obj.Position) {
			result = append(result, obj)
		}
	}

	// if this node is divided, query the children
	if qt.divided {
		for i := range NumQuadrants {
			childResults := qt.nodes[i].Query(rang)
			result = append(result, childResults...)
		}
	}

	return result
}

// QueryCircle returns all objects within a circular range
func (qt *QuadTree) QueryCircle(center pixel.Vec, radius float64) []*Object {
	// create a bounding box for the circle
	rang := &Bounds{
		X:      center.X - radius,
		Y:      center.Y - radius,
		Width:  radius * 2,
		Height: radius * 2,
	}

	// get all objects in the bounding box
	candidates := qt.Query(rang)
	result := make([]*Object, 0, len(candidates))

	// filter objects by distance
	for _, obj := range candidates {
		dx := obj.Position.X - center.X
		dy := obj.Position.Y - center.Y
		distSquared := dx*dx + dy*dy

		if distSquared <= radius*radius {
			result = append(result, obj)
		}
	}

	return result
}
