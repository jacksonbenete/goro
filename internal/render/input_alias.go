package render

import "github.com/kivutar/goro/internal/input"

type Key = input.Key

const (
	KeyEnter      = input.KeyEnter
	KeyEscape     = input.KeyEscape
	KeyTab        = input.KeyTab
	KeyArrowUp    = input.KeyArrowUp
	KeyArrowDown  = input.KeyArrowDown
	KeyArrowLeft  = input.KeyArrowLeft
	KeyArrowRight = input.KeyArrowRight
	KeyQ          = input.KeyQ
	KeyE          = input.KeyE
	KeyR          = input.KeyR
)

type MouseButton = input.MouseButton

const (
	MouseButtonLeft  = input.MouseButtonLeft
	MouseButtonRight = input.MouseButtonRight
)

const CursorModeHidden = 1

func SetCursorMode(int) {}
