package game

import (
	"log"
	"strings"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/session"
	gameui "github.com/kivutar/goro/ui"
)

func (m *WorldMode) updateWhisperWindow(ctx client.Context) bool {
	consumed := m.whisperWindow.Update(ctx)
	if action := m.whisperWindow.PopAction(); action.Target != "" && action.Message != "" {
		m.sendWhisperWindowMessage(ctx, action)
		return true
	}
	return consumed
}

func (m *WorldMode) sendWhisperWindowMessage(ctx client.Context, action gameui.WhisperWindowAction) {
	target := strings.TrimSpace(action.Target)
	message := strings.TrimSpace(action.Message)
	if target == "" || message == "" {
		return
	}
	if ctx.Network == nil {
		m.whisperWindow.AddError(ctx, "send failed: not connected")
		m.console.AddErrorMessage("send failed: not connected")
		return
	}
	if err := ctx.Network.SendWhisper(target, message); err != nil {
		m.whisperWindow.AddError(ctx, "send failed: "+err.Error())
		m.console.AddErrorMessage("send failed: %s", err)
		log.Printf("whisper window send failed target=%q: %v", target, err)
		return
	}
	m.whisperWindow.AddOutgoing(ctx, message)
	m.console.AddBlueMessage("[ To %s ] : %s", target, message)
}

func (m *WorldMode) addWhisperWindowIncoming(ctx client.Context, whisper network.WhisperMessage) {
	sender := strings.TrimSpace(whisper.Sender)
	message := strings.TrimSpace(whisper.Message)
	if sender == "" || message == "" {
		return
	}
	if !m.whisperWindow.IsOpen() && !shouldOpenWhisperWindow(ctx.Session, sender) {
		return
	}
	m.whisperWindow.Open(ctx, sender)
	m.whisperWindow.AddIncoming(ctx, sender, message)
}

func (m *WorldMode) addWhisperWindowAck(ctx client.Context, ack network.WhisperAck) {
	if ack.Result == 0 || !m.whisperWindow.IsOpen() {
		return
	}
	m.whisperWindow.AddError(ctx, whisperAckMessage(ctx.Resources, ack))
}

func shouldOpenWhisperWindow(s *session.Session, sender string) bool {
	settings := session.DefaultWhisperSettings()
	if s != nil && s.Whisper.Configured {
		settings = s.Whisper
	}
	if friendNameInSession(s, sender) {
		return settings.OpenFriends
	}
	return settings.OpenStrangers
}
