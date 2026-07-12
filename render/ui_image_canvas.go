package render

import (
	"image"
	"math"
	"reflect"
	"sync"

	"github.com/gogpu/gg/scene"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
	xdraw "golang.org/x/image/draw"
)

type scaledImageCanvas struct {
	widget.Canvas
	scale float32
}

func (c scaledImageCanvas) DrawImage(img image.Image, at geometry.Point) {
	if c.Canvas == nil || img == nil {
		return
	}
	bounds := img.Bounds()
	logicalW, logicalH := bounds.Dx(), bounds.Dy()
	if logicalW <= 0 || logicalH <= 0 || c.scale <= 1 {
		c.Canvas.DrawImage(img, at)
		return
	}
	scImg := scaledCanvasSceneImage(img, logicalW, logicalH, c.scale)
	if scImg == nil || scImg.Width <= 0 || scImg.Height <= 0 {
		c.Canvas.DrawImage(img, at)
		return
	}
	s := scene.NewScene()
	s.DrawImage(scImg, scene.NewAffine(
		float32(logicalW)/float32(scImg.Width), 0, at.X,
		0, float32(logicalH)/float32(scImg.Height), at.Y,
	))
	c.Canvas.ReplayScene(s)
}

type scaledCanvasImageKey struct {
	ptr                          uintptr
	srcMinX, srcMinY, srcW, srcH int
	dstW, dstH                   int
}

var scaledCanvasImageCache = struct {
	sync.Mutex
	images map[scaledCanvasImageKey]*scene.Image
}{images: make(map[scaledCanvasImageKey]*scene.Image)}

func scaledCanvasSceneImage(img image.Image, logicalW, logicalH int, scale float32) *scene.Image {
	bounds := img.Bounds()
	srcW, srcH := bounds.Dx(), bounds.Dy()
	if srcW <= 0 || srcH <= 0 || logicalW <= 0 || logicalH <= 0 {
		return nil
	}
	dstW := scaledCanvasPhysicalSize(logicalW, scale)
	dstH := scaledCanvasPhysicalSize(logicalH, scale)
	key := scaledCanvasImageKey{
		ptr:     imagePointer(img),
		srcMinX: bounds.Min.X,
		srcMinY: bounds.Min.Y,
		srcW:    srcW,
		srcH:    srcH,
		dstW:    dstW,
		dstH:    dstH,
	}
	if key.ptr != 0 {
		scaledCanvasImageCache.Lock()
		cached := scaledCanvasImageCache.images[key]
		scaledCanvasImageCache.Unlock()
		if cached != nil {
			return cached
		}
	}

	rgba := image.NewRGBA(image.Rect(0, 0, dstW, dstH))
	xdraw.ApproxBiLinear.Scale(rgba, rgba.Bounds(), img, bounds, xdraw.Src, nil)
	scImg := &scene.Image{Width: dstW, Height: dstH, Data: rgba.Pix}
	if key.ptr != 0 {
		scaledCanvasImageCache.Lock()
		scaledCanvasImageCache.images[key] = scImg
		scaledCanvasImageCache.Unlock()
	}
	return scImg
}

func scaledCanvasPhysicalSize(logical int, scale float32) int {
	if logical <= 0 {
		return 0
	}
	size := int(math.Round(float64(logical) * float64(scale)))
	if size < 1 {
		return 1
	}
	return size
}

func imagePointer(img image.Image) uintptr {
	v := reflect.ValueOf(img)
	if v.Kind() != reflect.Pointer && v.Kind() != reflect.UnsafePointer {
		return 0
	}
	if v.IsNil() {
		return 0
	}
	return v.Pointer()
}
