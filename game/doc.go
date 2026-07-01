// Package game orchestrates client behavior: map input, network reactions, skill
// commands, local effects, and calls into render, audio, network, world, and ui.
// Map entity state and movement should live in world; reusable UI composition
// should live in ui; rendering backend details should live in render.
package game
