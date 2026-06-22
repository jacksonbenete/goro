package gamemode

import (
	"math"
	"os"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	worldstate "github.com/kivutar/goro/internal/world"
)

const (
	sceneTileW = 32.0
	sceneTileH = 16.0
)

type sceneProjection struct {
	playerX        float64
	playerY        float64
	playerZ        float64
	centerX        float64
	centerY        float64
	screenW        float64
	screenH        float64
	tileW          float64
	tileH          float64
	heightScale    float64
	camera         bool
	viewProjection mat4
}

func newSceneProjection(screen *ebiten.Image, playerX, playerY int, playerZ float64) sceneProjection {
	return newSceneProjectionForSize(screen.Bounds().Dx(), screen.Bounds().Dy(), playerX, playerY, playerZ)
}

func newSceneProjectionForSize(width, height, playerX, playerY int, playerZ float64) sceneProjection {
	return newSceneProjectionForTarget(width, height, cellCenter(float64(playerX)), cellCenter(float64(playerY)), playerZ)
}

func newSceneProjectionForTarget(width, height int, targetX, targetY, targetZ float64) sceneProjection {
	return newSceneProjectionForTargetYaw(width, height, targetX, targetY, targetZ, sceneCameraYaw())
}

func newSceneProjectionForTargetYaw(width, height int, targetX, targetY, targetZ, yaw float64) sceneProjection {
	projection := sceneProjection{
		playerX:     targetX,
		playerY:     targetY,
		playerZ:     targetZ,
		centerX:     float64(width) * 0.5,
		centerY:     float64(height) * 0.5,
		screenW:     float64(width),
		screenH:     float64(height),
		tileW:       sceneTileW,
		tileH:       sceneTileH,
		heightScale: sceneHeightScale(),
	}
	if os.Getenv("GORO_SCENE_PROJECTION") != "flat" {
		projection.camera = true
		projection.viewProjection = sceneCameraMatrixWithYaw(float64(width), float64(height), targetX, targetY, targetZ, yaw)
	}
	return projection
}

func (p sceneProjection) Project(x, y, z float64) screenPoint {
	if p.camera {
		return p.projectCamera(x, y, z)
	}
	rx := x - p.playerX
	ry := y - p.playerY
	return screenPoint{
		x: float32(p.centerX + (rx-ry)*p.tileW*0.5),
		y: float32(p.centerY + (rx+ry)*p.tileH*0.5 - z*p.heightScale),
	}
}

func (p sceneProjection) Depth(x, y, z float64) float64 {
	if p.camera {
		_, _, _, clipW := mat4TransformVec4(p.viewProjection, x, z, y, 1)
		if clipW <= 0 || !isFinite(clipW) {
			return math.Inf(-1)
		}
		return clipW
	}
	return -(x + y)
}

func (p sceneProjection) VisibleForTriangle(x, y, z float64) bool {
	if !p.camera {
		return true
	}
	clipX, clipY, clipZ, clipW := mat4TransformVec4(p.viewProjection, x, z, y, 1)
	if clipW <= 1 || !finite4(clipX, clipY, clipZ, clipW) {
		return false
	}
	return clipZ >= -clipW && clipZ <= clipW
}

func (p sceneProjection) projectCamera(x, y, z float64) screenPoint {
	clipX, clipY, _, clipW := mat4TransformVec4(p.viewProjection, x, z, y, 1)
	if clipW <= 0.001 || !finite4(clipX, clipY, clipW, 1) {
		return screenPoint{x: -1 << 20, y: -1 << 20}
	}
	ndcX := clipX / clipW
	ndcY := clipY / clipW
	return screenPoint{
		x: float32((ndcX + 1) * p.screenW * 0.5),
		y: float32((1 - ndcY) * p.screenH * 0.5),
	}
}

func sceneCameraMatrix(width, height, targetX, targetY, targetZ float64) mat4 {
	return sceneCameraMatrixWithYaw(width, height, targetX, targetY, targetZ, sceneCameraYaw())
}

func sceneCameraMatrixWithYaw(width, height, targetX, targetY, targetZ, yawDegrees float64) mat4 {
	distance := sceneCameraZoom() * 0.5
	pitch := sceneCameraPitch()
	if pitch > 180 {
		pitch -= 180
	}
	pitch = degreesToRadians(pitch)
	yaw := degreesToRadians(yawDegrees)
	horizontal := math.Cos(pitch) * distance
	target := modelPoint3{x: targetX, y: targetZ, z: targetY}
	eye := modelPoint3{
		x: target.x + math.Sin(yaw)*horizontal,
		y: target.y + math.Sin(pitch)*distance,
		z: target.z - math.Cos(yaw)*horizontal,
	}
	view := mat4LookAt(eye, target, modelPoint3{y: 1})
	aspect := 1.0
	if height > 0 {
		aspect = width / height
	}
	return mat4Multiply(mat4Perspective(degreesToRadians(sceneCameraFOV()), aspect, 1, 1000), view)
}

func mat4LookAt(eye, target, up modelPoint3) mat4 {
	forward := normalize3(modelPoint3{x: target.x - eye.x, y: target.y - eye.y, z: target.z - eye.z})
	right := normalize3(cross3(up, forward))
	if right == (modelPoint3{}) {
		right = modelPoint3{x: 1}
	}
	cameraUp := cross3(forward, right)
	out := mat4Identity()
	out[0], out[1], out[2] = right.x, cameraUp.x, -forward.x
	out[4], out[5], out[6] = right.y, cameraUp.y, -forward.y
	out[8], out[9], out[10] = right.z, cameraUp.z, -forward.z
	out[12] = -dot3(right, eye)
	out[13] = -dot3(cameraUp, eye)
	out[14] = dot3(forward, eye)
	return out
}

func dot3(a, b modelPoint3) float64 {
	return a.x*b.x + a.y*b.y + a.z*b.z
}

func mat4Perspective(fovy, aspect, near, far float64) mat4 {
	f := 1 / math.Tan(fovy*0.5)
	out := mat4{}
	out[0] = f / aspect
	out[5] = f
	out[10] = (far + near) / (near - far)
	out[11] = -1
	out[14] = (2 * far * near) / (near - far)
	return out
}

func mat4TransformVec4(matrix mat4, x, y, z, w float64) (float64, float64, float64, float64) {
	return matrix[0]*x + matrix[4]*y + matrix[8]*z + matrix[12]*w,
		matrix[1]*x + matrix[5]*y + matrix[9]*z + matrix[13]*w,
		matrix[2]*x + matrix[6]*y + matrix[10]*z + matrix[14]*w,
		matrix[3]*x + matrix[7]*y + matrix[11]*z + matrix[15]*w
}

func finite4(a, b, c, d float64) bool {
	return isFinite(a) && isFinite(b) && isFinite(c) && isFinite(d)
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func sceneHeightScale() float64 {
	raw := os.Getenv("GORO_SCENE_HEIGHT_SCALE")
	if raw == "" {
		raw = os.Getenv("GORO_RSM_HEIGHT_SCALE")
	}
	if raw == "" {
		return 2.8
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil || value <= 0 {
		return 2.8
	}
	return value
}

func sceneCameraPitch() float64 {
	return sceneFloatEnv("GORO_CAMERA_PITCH", 230)
}

func sceneCameraYaw() float64 {
	return sceneFloatEnv("GORO_CAMERA_YAW", 0)
}

func sceneCameraZoom() float64 {
	return sceneFloatEnv("GORO_CAMERA_ZOOM", 150)
}

func sceneCameraFOV() float64 {
	return sceneFloatEnv("GORO_CAMERA_FOV", 15)
}

func sceneFloatEnv(name string, fallback float64) float64 {
	raw := os.Getenv(name)
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return fallback
	}
	return value
}

func terrainHeightAt(world *worldstate.World, x, y float64) float64 {
	if world == nil {
		return 0
	}
	if world.GAT != nil {
		cellX := int(math.Floor(x + 0.5))
		cellY := int(math.Floor(y + 0.5))
		cell, ok := world.GAT.Cell(cellX, cellY)
		if ok {
			return bilinearHeight(cell.Heights, x+0.5-float64(cellX), y+0.5-float64(cellY))
		}
	}
	if world.GND != nil {
		gridX := (x + 0.5) * 0.5
		gridY := (y + 0.5) * 0.5
		cellX := int(math.Floor(gridX))
		cellY := int(math.Floor(gridY))
		cell, ok := world.GND.Cell(cellX, cellY)
		if ok {
			return bilinearHeight(cell.Heights, gridX-float64(cellX), gridY-float64(cellY))
		}
	}
	return 0
}

func cellCenter(value float64) float64 {
	return value + 0.5
}

func bilinearHeight(heights [4]float32, x, y float64) float64 {
	if x < 0 {
		x = 0
	} else if x > 1 {
		x = 1
	}
	if y < 0 {
		y = 0
	} else if y > 1 {
		y = 1
	}
	x1 := float64(heights[0]) + (float64(heights[1])-float64(heights[0]))*x
	x2 := float64(heights[2]) + (float64(heights[3])-float64(heights[2]))*x
	return x1 + (x2-x1)*y
}
