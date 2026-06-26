package network

import (
	"encoding/binary"
	"testing"
)

func TestParseSkillInfoList(t *testing.T) {
	data := make([]byte, 4+skillInfoEntryLen)
	binary.LittleEndian.PutUint16(data[0:2], 0x010F)
	binary.LittleEndian.PutUint16(data[2:4], uint16(len(data)))
	writeSkillInfoEntry(data[4:], 1, 1001, 3, 14, 9, "First Aid", true)

	list, ok, err := ParseSkillInfoList(Packet{ID: 0x010F, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("skill list not parsed")
	}
	if len(list.Skills) != 1 {
		t.Fatalf("skill count = %d", len(list.Skills))
	}
	skill := list.Skills[0]
	if skill.ID != 1001 || skill.Type != 1 || skill.Level != 3 || skill.SPCost != 14 || skill.Range != 9 || skill.Name != "First Aid" || !skill.Upgradable {
		t.Fatalf("skill = %+v", skill)
	}
}

func TestParseSkillInfoUpdate(t *testing.T) {
	data := make([]byte, 11)
	binary.LittleEndian.PutUint16(data[0:2], 0x010E)
	binary.LittleEndian.PutUint16(data[2:4], 1001)
	binary.LittleEndian.PutUint16(data[4:6], 4)
	binary.LittleEndian.PutUint16(data[6:8], 15)
	binary.LittleEndian.PutUint16(data[8:10], 10)
	data[10] = 1

	update, ok, err := ParseSkillInfoUpdate(Packet{ID: 0x010E, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("skill update not parsed")
	}
	if update.Skill.ID != 1001 || update.Skill.Level != 4 || update.Skill.SPCost != 15 || update.Skill.Range != 10 || !update.Skill.Upgradable {
		t.Fatalf("update = %+v", update)
	}
}

func TestParseAddSkill(t *testing.T) {
	data := make([]byte, 39)
	binary.LittleEndian.PutUint16(data[0:2], 0x0111)
	writeSkillInfoEntry(data[2:], 0, 2, 1, 8, 1, "Heal", false)

	update, ok, err := ParseSkillInfoUpdate(Packet{ID: 0x0111, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("add skill not parsed")
	}
	if update.Skill.ID != 2 || update.Skill.Name != "Heal" || update.Skill.Upgradable {
		t.Fatalf("update = %+v", update)
	}
}

func TestBuildSkillLevelUpPacket(t *testing.T) {
	packet := BuildSkillLevelUpPacket(1001)
	if got := ID(packet); got != 0x0112 {
		t.Fatalf("opcode = 0x%04X", got)
	}
	if len(packet) != 4 {
		t.Fatalf("len = %d", len(packet))
	}
	if got := binary.LittleEndian.Uint16(packet[2:4]); got != 1001 {
		t.Fatalf("skill id = %d", got)
	}
}

func writeSkillInfoEntry(data []byte, typ uint32, id uint16, level, sp, attackRange int, name string, upgradable bool) {
	binary.LittleEndian.PutUint16(data[0:2], id)
	binary.LittleEndian.PutUint32(data[2:6], typ)
	binary.LittleEndian.PutUint16(data[6:8], uint16(level))
	binary.LittleEndian.PutUint16(data[8:10], uint16(sp))
	binary.LittleEndian.PutUint16(data[10:12], uint16(attackRange))
	copy(data[12:36], []byte(name))
	if upgradable {
		data[36] = 1
	}
}
