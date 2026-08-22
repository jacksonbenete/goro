package ui

import (
	"image"
	"strconv"
	"strings"

	"github.com/kivutar/goro/glog"
	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/res"
)

const (
	pvpRankBaseline = 52
	pvpRankStep     = 28
)

type pvpRankSpriteSet struct {
	manager *res.Manager
	act     *res.ACT
	spr     *res.SPR
	frames  map[pvpRankFrameKey]*render.Image
	cached  image.Image
	rank    int
	total   int
	miss    bool
}

type pvpRankFrameKey struct {
	index   int32
	sprType int32
}

func (s *pvpRankSpriteSet) image(manager *res.Manager, rank, total int) image.Image {
	if s.manager != manager {
		*s = pvpRankSpriteSet{manager: manager}
	}
	if manager == nil || rank <= 0 || total <= 0 {
		return nil
	}
	if s.cached != nil && s.rank == rank && s.total == total {
		return s.cached
	}
	if !s.ensure(manager) {
		return nil
	}
	composed := s.compose(rank, total)
	if composed == nil {
		s.cached = nil
		s.miss = true
		return nil
	}
	s.rank = rank
	s.total = total
	s.cached = composed
	return s.cached
}

func (s *pvpRankSpriteSet) ensure(manager *res.Manager) bool {
	if s.miss {
		return false
	}
	if s.act != nil && s.spr != nil {
		return true
	}
	actSource, actData, actOK := manager.ReadFirst(pvpRankSpriteResourceCandidates("act"))
	if !actOK {
		s.miss = true
		glog.Warnf("PvP rank font act unavailable")
		return false
	}
	sprSource, sprData, sprOK := manager.ReadFirst(pvpRankSpriteResourceCandidates("spr"))
	if !sprOK {
		s.miss = true
		glog.Warnf("PvP rank font spr unavailable")
		return false
	}
	act, err := res.ParseACT(actData)
	if err != nil {
		s.miss = true
		glog.Warnf("PvP rank font act parse %s: %v", actSource, err)
		return false
	}
	spr, err := res.ParseSPR(sprData)
	if err != nil {
		s.miss = true
		glog.Warnf("PvP rank font spr parse %s: %v", sprSource, err)
		return false
	}
	if !validPvPRankSpriteSet(act, spr) {
		s.miss = true
		glog.Warnf("PvP rank font resources incompatible act=%s spr=%s", actSource, sprSource)
		return false
	}
	s.act = act
	s.spr = spr
	s.frames = make(map[pvpRankFrameKey]*render.Image)
	glog.Debugf("PvP rank font resources act=%s spr=%s actions=%d frames=%d", actSource, sprSource, len(act.Actions), len(spr.Frames))
	return true
}

func validPvPRankSpriteSet(act *res.ACT, spr *res.SPR) bool {
	if act == nil || spr == nil || len(act.Actions) < 11 {
		return false
	}
	for actionIndex := 0; actionIndex <= 10; actionIndex++ {
		action := act.Actions[actionIndex]
		if len(action.Animations) == 0 {
			return false
		}
		animation := action.Animations[len(action.Animations)/2]
		validLayer := false
		for _, layer := range animation.Layers {
			frameIndex := int(layer.Index)
			if layer.SPRType == res.SPRFrameRGBA {
				frameIndex += spr.RGBAIndex
			}
			if frameIndex < 0 || frameIndex >= len(spr.Frames) {
				continue
			}
			frame := spr.Frames[frameIndex]
			if frame.Width > 0 && frame.Height > 0 {
				validLayer = true
				break
			}
		}
		if !validLayer {
			return false
		}
	}
	return true
}

func pvpRankSpriteResourceCandidates(ext string) []string {
	path := "data\\sprite\\이팩트\\rankfont." + ext
	return []string{path, strings.ReplaceAll(path, "\\", "/")}
}

func (s *pvpRankSpriteSet) compose(rank, total int) image.Image {
	rankText := strconv.Itoa(rank)
	totalText := strconv.Itoa(total)
	totalWidth := (len(rankText) + 1 + len(totalText)) * pvpRankStep
	x := (pvpCounterWidth - totalWidth) / 2
	target := render.NewImage(pvpCounterWidth, pvpCounterHeight)

	for _, ch := range rankText {
		if !s.drawAction(target, int(ch-'0'), x, pvpRankBaseline-6) {
			return nil
		}
		x += pvpRankStep
	}
	if !s.drawAction(target, 10, x, pvpRankBaseline) {
		return nil
	}
	x += pvpRankStep
	for _, ch := range totalText {
		if !s.drawAction(target, int(ch-'0'), x, pvpRankBaseline+6) {
			return nil
		}
		x += pvpRankStep
	}
	return target.RGBA()
}

func (s *pvpRankSpriteSet) drawAction(target *render.Image, actionIndex, x, y int) bool {
	if s.act == nil || actionIndex < 0 || actionIndex >= len(s.act.Actions) {
		return false
	}
	action := s.act.Actions[actionIndex]
	if len(action.Animations) == 0 {
		return false
	}
	animation := action.Animations[len(action.Animations)/2]
	drawn := false
	for _, layer := range animation.Layers {
		if layer.Index < 0 {
			continue
		}
		frame := s.frameImage(layer.Index, layer.SPRType)
		if frame == nil {
			continue
		}
		drawUISpriteLayer(target, frame, layer, float64(x), float64(y)-uiSpriteRendererMidCellY)
		drawn = true
	}
	return drawn
}

func (s *pvpRankSpriteSet) frameImage(index, sprType int32) *render.Image {
	key := pvpRankFrameKey{index: index, sprType: sprType}
	if frame, ok := s.frames[key]; ok {
		return frame
	}
	frame, ok := s.spr.FrameImage(int(index), int(sprType))
	if !ok {
		return nil
	}
	img := render.NewImageFromStraightAlpha(frame)
	s.frames[key] = img
	return img
}
