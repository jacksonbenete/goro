package game

import "github.com/kivutar/goro/db"

func effectTableSize(value float64) float64 {
	return value * effectPixelRatio
}

type effectCoverage struct {
	Implemented     int
	ReferenceActive int
	ReferenceAll    int
	ActivePercent   float64
	AllPercent      float64
}

func effectCoverageSnapshot() effectCoverage {
	implemented := len(db.EffectSpecs)
	return effectCoverage{
		Implemented:     implemented,
		ReferenceActive: db.ReferenceActiveEffectTableEntries,
		ReferenceAll:    db.ReferenceNumericEffectConstants,
		ActivePercent:   float64(implemented) / float64(db.ReferenceActiveEffectTableEntries) * 100,
		AllPercent:      float64(implemented) / float64(db.ReferenceNumericEffectConstants) * 100,
	}
}
