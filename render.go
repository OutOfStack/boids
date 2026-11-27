package main

import (
	"errors"
	"image/color"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const boidVertexShader = `#version 330
in vec3 vertexPosition;
in vec3 vertexNormal;
in mat4 instanceTransform;
uniform mat4 mvp;
out vec3 normal;
void main() {
    normal = mat3(instanceTransform) * vertexNormal;
    gl_Position = mvp * instanceTransform * vec4(vertexPosition, 1.0);
}`

const boidFragmentShader = `#version 330
in vec3 normal;
uniform vec4 colDiffuse;
out vec4 finalColor;
void main() {
    vec3 light = normalize(vec3(0.4, 1.0, 0.7));
    float shade = 0.55 + 0.45 * max(dot(normalize(normal), light), 0.0);
    finalColor = vec4(colDiffuse.rgb * shade, colDiffuse.a);
}`

func render(sim *simulation) error {
	rl.SetConfigFlags(rl.FlagWindowResizable | rl.FlagVsyncHint | rl.FlagMsaa4xHint)
	rl.InitWindow(sim.cfg.Width, sim.cfg.Height, "Boids 3D")
	if !rl.IsWindowReady() {
		return errors.New("could not open the raylib window")
	}
	defer rl.CloseWindow()

	size := sim.cfg.BoidSize
	// Release through the owning model to avoid a crash in v0.63.0's nocgo UnloadMesh.
	model := rl.LoadModelFromMesh(rl.GenMeshCone(size*0.3, size, 4))
	defer rl.UnloadModel(model)
	mesh := model.GetMeshes()[0]
	material := model.GetMaterials()[0]
	material.Shader = rl.LoadShaderFromMemory(boidVertexShader, boidFragmentShader)
	defer rl.UnloadShader(material.Shader)
	transformLocation := rl.GetShaderLocationAttrib(material.Shader, "instanceTransform")
	if transformLocation < 0 {
		return errors.New("could not load the boid instancing shader")
	}
	material.Shader.UpdateLocation(rl.ShaderLocMatrixModel, transformLocation)

	worldSize := renderVector(sim.extent)
	center := rl.Vector3Scale(worldSize, 0.5)
	worldLength := max(sim.extent[0], sim.extent[1], sim.extent[2])
	rl.SetClipPlanes(worldLength/1000, worldLength*8)
	camera := rl.Camera3D{
		Target: center, Up: rl.Vector3{Y: 1}, Fovy: 48, Projection: rl.CameraPerspective,
	}
	yaw, pitch, distance := 0.65, 0.35, worldLength*1.75
	background := color.RGBA{R: 10, G: 14, B: 23, A: 255}
	boundsColor := color.RGBA{R: 57, G: 72, B: 95, A: 255}
	controlsColor := color.RGBA{R: 160, G: 174, B: 192, A: 255}
	flockColors := [flockCount]color.RGBA{
		grayFlock:   {R: 180, G: 188, B: 200, A: 255},
		orangeFlock: {R: 255, G: 140, B: 0, A: 255},
		blueFlock:   {R: 100, G: 149, B: 237, A: 255},
		greenFlock:  {R: 154, G: 205, B: 50, A: 255},
	}
	snapshot := make([]boid, sim.cfg.BoidsCount)
	var batches [flockCount][]rl.Matrix
	diffuse := material.GetMap(rl.MapDiffuse)

	for !rl.WindowShouldClose() {
		if rl.IsMouseButtonDown(rl.MouseButtonLeft) {
			delta := rl.GetMouseDelta()
			yaw -= float64(delta.X) * 0.006
			pitch = min(1.5, max(-1.5, pitch+float64(delta.Y)*0.006))
		}
		distance *= math.Pow(0.9, float64(rl.GetMouseWheelMove()))
		distance = min(worldLength*4, max(worldLength*0.45, distance))
		horizontal := distance * math.Cos(pitch)
		camera.Position = rl.Vector3Add(center, renderVector(vec3{
			horizontal * math.Sin(yaw), distance * math.Sin(pitch), horizontal * math.Cos(yaw),
		}))

		sim.snapshot(snapshot)
		for flock := range batches {
			batches[flock] = batches[flock][:0]
		}
		for _, b := range snapshot {
			batches[b.flock] = append(batches[b.flock], boidTransform(b))
		}

		rl.BeginDrawing()
		rl.ClearBackground(background)
		rl.BeginMode3D(camera)
		for flock, transforms := range batches {
			if len(transforms) > 0 {
				diffuse.Color = flockColors[flock]
				rl.DrawMeshInstanced(mesh, material, transforms, len(transforms))
			}
		}
		rl.DrawCubeWiresV(center, worldSize, boundsColor)
		rl.EndMode3D()
		rl.DrawText("Drag to orbit   /   Scroll to zoom", 16, 16, 16, controlsColor)
		rl.EndDrawing()
	}
	return nil
}

func renderVector(v vec3) rl.Vector3 {
	return rl.Vector3{X: float32(v[0]), Y: float32(v[1]), Z: float32(v[2])}
}

func boidTransform(b boid) rl.Matrix {
	position := renderVector(b.position)
	forward := rl.Vector3Normalize(renderVector(b.velocity))
	if forward == (rl.Vector3{}) {
		forward.Y = 1
	}
	// The cone points along its local +Y axis. Pick a reference axis that
	// stays away from parallel so vertical and near-zero motion are safe.
	reference := rl.Vector3{Z: 1}
	if forward.Z > 0.95 || forward.Z < -0.95 {
		reference = rl.Vector3{Y: 1}
	}
	right := rl.Vector3Normalize(rl.Vector3CrossProduct(forward, reference))
	up := rl.Vector3CrossProduct(right, forward)
	return rl.Matrix{
		M0: right.X, M4: forward.X, M8: up.X, M12: position.X,
		M1: right.Y, M5: forward.Y, M9: up.Y, M13: position.Y,
		M2: right.Z, M6: forward.Z, M10: up.Z, M14: position.Z,
		M15: 1,
	}
}
