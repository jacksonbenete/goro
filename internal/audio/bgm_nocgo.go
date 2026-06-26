//go:build !cgo

package audio

import "github.com/kivutar/goro/internal/res"

type BGM struct{}

func NewBGM(*res.Manager, bool, float64) *BGM {
	return &BGM{}
}

func (b *BGM) PlayMap(mapName string) (string, error) {
	return "", nil
}

func (b *BGM) Play(path string) error {
	return nil
}

func (b *BGM) PlaySFX(path string) (string, error) {
	return "", nil
}

func (b *BGM) ResolveMapBGM(mapName string) string {
	return ""
}

func (b *BGM) Stop() {}
