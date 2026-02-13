# boids

## Boids is an artificial life program, which simulates the flocking behaviour of birds.[<sup>wiki<sup>](https://en.wikipedia.org/wiki/Boids)

### Features

- Flocking behavior simulation with alignment, cohesion, and separation rules
- Color-based grouping of boids
- Spatial partitioning using a quadtree for improved performance
- Concurrent processing of boid movements

### Configuration Parameters

All simulation parameters can be configured in the `config.json` file:

| Parameter          | Description |
|--------------------|-------------|
| `width`            | Width of the simulation window in pixels. Defines the horizontal bounds of the simulation space. |
| `height`           | Height of the simulation window in pixels. Defines the vertical bounds of the simulation space. |
| `boids_count`      | Total number of boids to simulate. Higher values create more complex flocking patterns but require more computational resources. |
| `view_radius`      | The radius within which each boid can see other boids. Determines how far a boid can detect neighbors for flocking behaviors. Also used for boundary avoidance calculations. |
| `adj_rate`         | Adjustment rate for steering behaviors. Controls how quickly boids adjust their velocity in response to alignment, cohesion, and separation forces. Higher values make boids more responsive but can lead to erratic movement. |
| `poly_thickness`   | Thickness of the polygon lines used for rendering boids. Affects the visual appearance of boids. |
| `quadtree_max_obj` | Maximum number of objects a quadtree node can contain before it splits into four child nodes. Lower values create more subdivisions, potentially improving query performance at the cost of memory usage. |
| `quadtree_max_lvl` | Maximum depth of the quadtree. Limits how many times the space can be recursively subdivided. Prevents excessive memory usage in dense areas. |
| `update_rate_ms`   | Time in milliseconds between boid updates. Lower values make boids move faster but consume more CPU. Higher values reduce CPU usage but make movement less smooth. |
| `seed`             | Optional random seed for deterministic runs. If omitted or 0, a non-deterministic seed is used. |

Example configuration:
```json
{
  "width": 800,
  "height": 600,
  "boids_count": 1500,
  "view_radius": 7,
  "adj_rate": 0.3,
  "poly_thickness": 1.5,
  "quadtree_max_obj": 10,
  "quadtree_max_lvl": 5,
  "update_rate_ms": 5
}
```

### Requirements:
On Ubuntu/Debian-like Linux distributions, install `libgl1-mesa-dev` and `xorg-dev` packages

### Run:
`make run` for run
`make build` for build
`make test` for tests
`make lint` for linter