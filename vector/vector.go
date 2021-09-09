package vector

import (
	"math"

	"github.com/faiface/pixel"
)

func Limit(vector pixel.Vec, lower, upper float64) pixel.Vec {
	return pixel.V(math.Min(math.Max(vector.X, lower), upper),
		math.Min(math.Max(vector.Y, lower), upper))
}

func AddV(vector pixel.Vec, value float64) pixel.Vec {
	return pixel.V(vector.X+value, vector.Y+value)
}

func Distance(v1 pixel.Vec, v2 pixel.Vec) float64 {
	return math.Sqrt(math.Pow(v1.X-v2.X, 2) + math.Pow(v1.Y-v2.Y, 2))
}

func DivisionV(vector pixel.Vec, d float64) pixel.Vec {
	return pixel.V(vector.X/d, vector.Y/d)
}
