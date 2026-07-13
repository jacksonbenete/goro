package game

import (
	"log"

	"github.com/kivutar/goro/client"
)

func sendLessEffectPreference(ctx client.Context) {
	if ctx.Session == nil || ctx.Network == nil || !ctx.Session.LessEffects {
		return
	}
	if err := ctx.Network.SendLessEffect(true); err != nil {
		log.Printf("send less effect preference failed: %v", err)
	}
}
