package game

import (
	"github.com/gogpu/ui/widget"
)

type worldModeTestUIManager struct {
	overlays []widget.Widget
}

func (m *worldModeTestUIManager) AddOverlay(root widget.Widget) {
	m.overlays = append(m.overlays, root)
}

func (m *worldModeTestUIManager) RemoveOverlay(root widget.Widget) {
	for i, overlay := range m.overlays {
		if overlay == root {
			m.overlays = append(m.overlays[:i], m.overlays[i+1:]...)
			return
		}
	}
}

func (m *worldModeTestUIManager) Clear() {
	m.overlays = nil
}
