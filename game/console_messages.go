package game

import (
	"fmt"
	"strings"

	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/res"
	gameui "github.com/kivutar/goro/ui"
)

func formatConsoleMessage(manager *res.Manager, chat network.ChatMessage) string {
	if chat.Text != "" {
		return chat.Text
	}
	if chat.MessageID < 0 {
		return ""
	}
	text := ""
	if manager != nil {
		text, _ = manager.MsgString(chat.MessageID)
	}
	if text == "" {
		text = fmt.Sprintf("message #%d", chat.MessageID)
	}
	if chat.Value != 0 {
		if strings.Contains(text, "%") {
			text = fmt.Sprintf(text, chat.Value)
		} else {
			text = fmt.Sprintf("%s %d", text, chat.Value)
		}
	}
	if chat.SkillID != 0 {
		text = fmt.Sprintf("skill %d: %s", chat.SkillID, text)
	}
	return text
}

func addConsoleMessage(console *gameui.ChatConsole, manager *res.Manager, chat network.ChatMessage) {
	if console == nil {
		return
	}
	text := formatConsoleMessage(manager, chat)
	if text == "" {
		return
	}
	if chat.Text == "" || !strings.Contains(text, " : ") {
		console.AddSystemMessage("%s", text)
		return
	}
	console.AddMessage("%s", text)
}

func addWhisperMessage(console *gameui.ChatConsole, whisper network.WhisperMessage) {
	if console == nil {
		return
	}
	if whisper.Sender == "" || whisper.Message == "" {
		return
	}
	console.AddBlueMessage("[ From %s ] : %s", whisper.Sender, whisper.Message)
}

func addWhisperAck(console *gameui.ChatConsole, manager *res.Manager, ack network.WhisperAck) {
	if console == nil || ack.Result == 0 {
		return
	}
	console.AddErrorMessage("%s", whisperAckMessage(manager, ack))
}

func addWhisperIgnoreAck(console *gameui.ChatConsole, ack network.WhisperIgnoreAck) {
	if console == nil {
		return
	}
	if ack.Result != 0 {
		console.AddErrorMessage("%s", whisperIgnoreAckFailure(ack))
		return
	}
	console.AddSystemMessage("%s", whisperIgnoreAckSuccess(ack))
}

func whisperIgnoreAckSuccess(ack network.WhisperIgnoreAck) string {
	if ack.TargetAll {
		if ack.Allow {
			return "Whispers from everyone are allowed."
		}
		return "Whispers from everyone are blocked."
	}
	if ack.Allow {
		return "Whispers from that player are allowed."
	}
	return "Whispers from that player are blocked."
}

func whisperIgnoreAckFailure(ack network.WhisperIgnoreAck) string {
	if !ack.TargetAll && ack.Result == 2 {
		return "Whisper block list is full."
	}
	if ack.TargetAll {
		if ack.Allow {
			return "Allow all whispers failed."
		}
		return "Block all whispers failed."
	}
	if ack.Allow {
		return "Allow whisper failed."
	}
	return "Block whisper failed."
}

func whisperAckMessage(manager *res.Manager, ack network.WhisperAck) string {
	message := ""
	if manager != nil {
		message, _ = manager.MsgString(int(147 + ack.Result))
	}
	if message == "" {
		message = "Whisper failed."
	}
	return message
}

func formatPickupConsoleMessage(manager *res.Manager, pickup network.ItemPickupAck) string {
	itemName := fmt.Sprintf("item %d", pickup.ItemID)
	if manager != nil {
		if name, ok := manager.ItemDisplayName(int(pickup.ItemID), pickup.Identified); ok && name != "" {
			itemName = name
		}
	}
	amount := int(pickup.Amount)
	if amount <= 0 {
		amount = 1
	}
	template := ""
	if manager != nil {
		template, _ = manager.MsgString(153)
	}
	if template == "" {
		template = "You got %s %d."
	}
	if strings.Contains(template, "%s") {
		template = strings.Replace(template, "%s", itemName, 1)
	} else {
		template = strings.TrimSpace(template + " " + itemName)
	}
	if strings.Contains(template, "%d") {
		template = strings.Replace(template, "%d", fmt.Sprintf("%d", amount), 1)
	} else if amount != 1 {
		template = strings.TrimSpace(fmt.Sprintf("%s %d", template, amount))
	}
	return template
}
