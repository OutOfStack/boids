package config

import (
	"strings"
	"testing"
)

func TestParseRejectsInvalidConfig(t *testing.T) {
	t.Parallel()

	const valid = `"width": 100, "height": 100, "boids_count": 10, "view_radius": 10, "update_rate_ms": 10`
	for _, test := range []struct{ field, want string }{
		{`"width": 0`, "width, height, and depth"},
		{`"height": -1`, "width, height, and depth"},
		{`"depth": 0`, "width, height, and depth"},
		{`"boids_count": -1`, "boids_count"},
		{`"boids_count": 2000000000`, "boids_count"},
		{`"view_radius": -1`, "view_radius"},
		{`"boid_size": 0`, "boid_size"},
		{`"boid_size": 1e300`, "float32"},
		{`"update_rate_ms": 0`, "update_rate_ms"},
	} {
		// Later duplicate keys override earlier ones in encoding/json.
		_, err := parse([]byte("{" + valid + ", " + test.field + "}"))
		if err == nil || !strings.Contains(err.Error(), test.want) {
			t.Errorf("%s: error = %v, want it to mention %q", test.field, err, test.want)
		}
	}
	if _, err := parse([]byte("{" + valid + "}")); err != nil {
		t.Errorf("valid config: %v", err)
	}
}
