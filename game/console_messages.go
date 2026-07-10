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
