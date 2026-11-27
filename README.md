# Boids 3D

A desktop flocking simulation in Go. Boids align, cohere, and separate within their color group in a bounded 3D world. A CPU spatial grid finds nearby boids, and raylib draws the flock as instanced meshes.

## Run

From the project directory, run `go run .` with Go 1.27 or later. On Windows amd64, `go run -tags nocgo .` uses the embedded raylib DLL and works without a C compiler. For a cgo build on Debian or Ubuntu, install `libgl1-mesa-dev xorg-dev libegl-dev libgbm-dev libdrm-dev libwayland-dev libxkbcommon-dev`.

Drag with the left mouse button to orbit, scroll to zoom, and press Escape to close.

`make build` builds the app, `make test` runs tests (with the race detector outside Windows), and `make lint` runs the linter.

## Configuration

Edit `config.json` in the working directory.

| Setting | Description |
| --- | --- |
| `width`, `height` | World dimensions along X and Y, and initial window dimensions in pixels. Resizing the window leaves the world dimensions alone. |
| `depth` | World dimension along Z. Defaults to `height` when omitted. |
| `boids_count` | Number of boids, up to 1,000,000. |
| `view_radius` | Neighbor detection radius in world units. Also controls how far from a face border steering starts. |
| `adj_rate` | Strength of the alignment, cohesion, and separation forces. |
| `boid_size` | Length of the rendered boid in world units. Defaults to 8. |
| `update_rate_ms` | Interval between simulation steps in milliseconds. |
| `seed` | Set a nonzero seed for repeatable initial positions and velocities. Zero or omission uses a random seed. |

Boids steer away from the six faces of the box as they approach them, and are kept inside it if they overshoot.
