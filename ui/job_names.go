package ui

import (
	"github.com/kivutar/goro/db"
	"github.com/kivutar/goro/session"
)

func CharacterJobName(character session.Character) string {
	return db.JobDisplayName(int(character.Job))
}
