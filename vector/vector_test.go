package vector_test

import (
	"math"
	"testing"

	"github.com/OutOfStack/boids/vector"
	"github.com/gopxl/pixel/v2"
)

func TestLimit(t *testing.T) {
	tests := []struct {
		name     string
		vector   pixel.Vec
		lower    float64
		upper    float64
		expected pixel.Vec
	}{
		{
			name:     "vector within bounds",
			vector:   pixel.V(0.5, 0.5),
			lower:    -1.0,
			upper:    1.0,
			expected: pixel.V(0.5, 0.5),
		},
		{
			name:     "vector exceeds upper bound",
			vector:   pixel.V(2.0, 1.5),
			lower:    -1.0,
			upper:    1.0,
			expected: pixel.V(1.0, 1.0),
		},
		{
			name:     "vector below lower bound",
			vector:   pixel.V(-2.0, -1.5),
			lower:    -1.0,
			upper:    1.0,
			expected: pixel.V(-1.0, -1.0),
		},
		{
			name:     "mixed bounds violation",
			vector:   pixel.V(-2.0, 2.0),
			lower:    -1.0,
			upper:    1.0,
			expected: pixel.V(-1.0, 1.0),
		},
		{
			name:     "at lower bound",
			vector:   pixel.V(-1.0, -1.0),
			lower:    -1.0,
			upper:    1.0,
			expected: pixel.V(-1.0, -1.0),
		},
		{
			name:     "at upper bound",
			vector:   pixel.V(1.0, 1.0),
			lower:    -1.0,
			upper:    1.0,
			expected: pixel.V(1.0, 1.0),
		},
		{
			name:     "zero vector",
			vector:   pixel.V(0, 0),
			lower:    -1.0,
			upper:    1.0,
			expected: pixel.V(0, 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vector.Limit(tt.vector, tt.lower, tt.upper)
			if result.X != tt.expected.X || result.Y != tt.expected.Y {
				t.Errorf("Limit(%v, %v, %v) = %v, want %v",
					tt.vector, tt.lower, tt.upper, result, tt.expected)
			}
		})
	}
}

func TestDistance(t *testing.T) {
	tests := []struct {
		name     string
		v1       pixel.Vec
		v2       pixel.Vec
		expected float64
	}{
		{
			name:     "identical vectors",
			v1:       pixel.V(0, 0),
			v2:       pixel.V(0, 0),
			expected: 0.0,
		},
		{
			name:     "horizontal distance",
			v1:       pixel.V(0, 0),
			v2:       pixel.V(3, 0),
			expected: 3.0,
		},
		{
			name:     "vertical distance",
			v1:       pixel.V(0, 0),
			v2:       pixel.V(0, 4),
			expected: 4.0,
		},
		{
			name:     "diagonal distance (3-4-5 triangle)",
			v1:       pixel.V(0, 0),
			v2:       pixel.V(3, 4),
			expected: 5.0,
		},
		{
			name:     "negative coordinates",
			v1:       pixel.V(-1, -1),
			v2:       pixel.V(2, 3),
			expected: 5.0,
		},
		{
			name:     "symmetric distance",
			v1:       pixel.V(1, 1),
			v2:       pixel.V(4, 5),
			expected: 5.0,
		},
		{
			name:     "unit distance diagonal",
			v1:       pixel.V(0, 0),
			v2:       pixel.V(1, 1),
			expected: math.Sqrt(2),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vector.Distance(tt.v1, tt.v2)
			if math.Abs(result-tt.expected) > 1e-10 {
				t.Errorf("Distance(%v, %v) = %v, want %v",
					tt.v1, tt.v2, result, tt.expected)
			}

			// test symmetry: Distance(v1, v2) should equal Distance(v2, v1)
			reverse := vector.Distance(tt.v2, tt.v1)
			if math.Abs(reverse-result) > 1e-10 {
				t.Errorf("Distance is not symmetric: Distance(%v, %v) = %v, Distance(%v, %v) = %v",
					tt.v1, tt.v2, result, tt.v2, tt.v1, reverse)
			}
		})
	}
}

func TestDivisionV(t *testing.T) {
	tests := []struct {
		name     string
		vector   pixel.Vec
		divisor  float64
		expected pixel.Vec
	}{
		{
			name:     "divide by 1",
			vector:   pixel.V(10, 20),
			divisor:  1.0,
			expected: pixel.V(10, 20),
		},
		{
			name:     "divide by 2",
			vector:   pixel.V(10, 20),
			divisor:  2.0,
			expected: pixel.V(5, 10),
		},
		{
			name:     "divide by 0.5 (multiply by 2)",
			vector:   pixel.V(10, 20),
			divisor:  0.5,
			expected: pixel.V(20, 40),
		},
		{
			name:     "divide by zero (returns original)",
			vector:   pixel.V(10, 20),
			divisor:  0.0,
			expected: pixel.V(10, 20),
		},
		{
			name:     "divide negative vector",
			vector:   pixel.V(-10, -20),
			divisor:  2.0,
			expected: pixel.V(-5, -10),
		},
		{
			name:     "divide by negative divisor",
			vector:   pixel.V(10, 20),
			divisor:  -2.0,
			expected: pixel.V(-5, -10),
		},
		{
			name:     "divide zero vector",
			vector:   pixel.V(0, 0),
			divisor:  5.0,
			expected: pixel.V(0, 0),
		},
		{
			name:     "divide small vector",
			vector:   pixel.V(1.5, 3.0),
			divisor:  3.0,
			expected: pixel.V(0.5, 1.0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vector.DivisionV(tt.vector, tt.divisor)
			if math.Abs(result.X-tt.expected.X) > 1e-10 || math.Abs(result.Y-tt.expected.Y) > 1e-10 {
				t.Errorf("DivisionV(%v, %v) = %v, want %v",
					tt.vector, tt.divisor, result, tt.expected)
			}
		})
	}
}
