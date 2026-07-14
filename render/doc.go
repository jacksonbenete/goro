// Package render wraps the gogpu backend and provides drawing primitives used by game.
//
// Image is CPU texture data. Frame is a per-frame command buffer consumed by
// the gogpu renderer. The package should draw submitted data without owning
// Ragnarok map, actor, input, or UI behavior.
package render
