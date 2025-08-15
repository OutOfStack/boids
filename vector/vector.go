package vector

import (
	"math"

	"github.com/gopxl/pixel/v2"
)

// Limit restricts vector's X and Y to range [lower, upper]
func Limit(vector pixel.Vec, lower, upper float64) pixel.Vec {
	return pixel.V(
		math.Min(math.Max(vector.X, lower), upper),
		math.Min(math.Max(vector.Y, lower), upper))
}

// DistanceSquared calculates squared distance between two vectors.
func DistanceSquared(v1, v2 pixel.Vec) float64 {
	dx := v1.X - v2.X
	dy := v1.Y - v2.Y
	return dx*dx + dy*dy
}

// DivisionV divides X and Y of vector by a specified divisor.
func DivisionV(vector pixel.Vec, d float64) pixel.Vec {
	if d == 0 {
		return vector
	}
	return pixel.V(vector.X/d, vector.Y/d)
}
