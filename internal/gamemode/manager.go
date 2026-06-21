package gamemode

import "github.com/hajimehoshi/ebiten/v2"

type Mode interface {
	Name() string
	Enter(Context)
	Update(Context) (Mode, error)
	Draw(Context, *ebiten.Image)
}

type Manager struct {
	ctx  Context
	mode Mode
}

func NewManager(ctx Context, mode Mode) *Manager {
	m := &Manager{ctx: ctx, mode: mode}
	if m.mode != nil {
		m.mode.Enter(ctx)
	}
	return m
}

func (m *Manager) UpdateContext(ctx Context) {
	m.ctx = ctx
}

func (m *Manager) Update() error {
	if m.mode == nil {
		return nil
	}

	next, err := m.mode.Update(m.ctx)
	if err != nil {
		return err
	}
	if next != nil {
		m.mode = next
		m.mode.Enter(m.ctx)
	}
	return nil
}

func (m *Manager) Draw(screen *ebiten.Image) {
	if m.mode != nil {
		m.mode.Draw(m.ctx, screen)
	}
}
