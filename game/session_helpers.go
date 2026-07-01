package game

import "github.com/kivutar/goro/session"

func selectedCharacter(s *session.Session) session.Character {
	if s.Selected.ID != 0 {
		return s.Selected
	}
	for _, character := range s.Characters {
		if character.ID == s.CharID {
			return character
		}
	}
	if len(s.Characters) > 0 {
		return s.Characters[0]
	}
	return session.Character{ID: s.CharID, Name: "Player", Job: 0}
}

func sessionVitalsFromCharacter(character session.Character) session.Vitals {
	return session.Vitals{
		HP:    int(character.HP),
		MaxHP: int(character.MaxHP),
		SP:    int(character.SP),
		MaxSP: int(character.MaxSP),
	}
}

func sessionProgressFromCharacter(character session.Character) session.Progress {
	return session.Progress{
		BaseLevel: int(character.Level),
		JobLevel:  int(character.JobLevel),
	}
}
