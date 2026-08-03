package ui

import (
	"image"

	"github.com/kivutar/goro/res"
)

func cachedIdentifiedItemIcon(manager *res.Manager, itemID uint16, icons *map[identifyItemIconKey]image.Image, misses *map[identifyItemIconKey]struct{}) image.Image {
	if manager == nil || itemID == 0 {
		return nil
	}
	key := identifyItemIconKey{itemID: itemID, identified: true}
	if *icons != nil {
		if img := (*icons)[key]; img != nil {
			return img
		}
	}
	if _, ok := (*misses)[key]; ok {
		return nil
	}
	resourceName, ok := manager.ItemResourceName(int(itemID), true)
	if !ok {
		markIdentifiedItemIconMiss(key, misses)
		return nil
	}
	img, _, err := res.LoadImage(manager, res.ItemIconTextureCandidates(resourceName))
	if err != nil {
		markIdentifiedItemIconMiss(key, misses)
		return nil
	}
	if *icons == nil {
		*icons = make(map[identifyItemIconKey]image.Image)
	}
	(*icons)[key] = img
	return img
}

func markIdentifiedItemIconMiss(key identifyItemIconKey, misses *map[identifyItemIconKey]struct{}) {
	if *misses == nil {
		*misses = make(map[identifyItemIconKey]struct{})
	}
	(*misses)[key] = struct{}{}
}
