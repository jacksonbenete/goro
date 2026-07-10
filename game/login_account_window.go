package game

import (
	"github.com/kivutar/goro/client"
	gameui "github.com/kivutar/goro/ui"
)

func (m *LoginMode) updateFormInput(ctx client.Context) {
	m.updateLoginWindow(ctx)
	if m.loginWindow != nil {
		m.loginWindow.Update(ctx)
		m.loginWindow.Publish(ctx)
	}
}

func (m *LoginMode) drawLoginWindow(ctx client.Context) {
	if m.loginWindow == nil {
		m.updateLoginWindow(ctx)
	}
	if m.loginWindow != nil {
		m.loginWindow.Publish(ctx)
	}
}

func (m *LoginMode) updateLoginWindow(ctx client.Context) {
	if m.loginWindow == nil {
		m.loginWindow = gameui.NewLoginWindow(ctx, m.username, m.password, gameui.LoginWindowCallbacks{
			OnSubmit: func() {
				m.username = m.loginWindow.Username
				m.password = m.loginWindow.Password
				if len(ctx.Resources.ClientInfo.Connections) > 0 {
					m.connectAndMaybeLogin(ctx, ctx.Resources.ClientInfo.Connections[0], true)
				}
			},
		})
		m.loginWindow.Publish(ctx)
		return
	}
	m.loginWindow.SetContext(ctx)
	m.username = m.loginWindow.Username
	m.password = m.loginWindow.Password
	m.loginWindow.Publish(ctx)
}
