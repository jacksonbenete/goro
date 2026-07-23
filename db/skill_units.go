package db

const (
	SkillUnitUsedTraps uint16 = 140
)

type SkillUnitModelSpec struct {
	ModelPath       string
	TriggerEffectID int
}

// Mirrors roBrowser DB/Skills/SkillUnit.js ef_trap_* aliases. The path is
// relative to data\model\, matching RSW model filenames.
var SkillUnitModels = map[uint16]SkillUnitModelSpec{
	143: {ModelPath: "외부소품\\트랩03_3.rsm", TriggerEffectID: effectBlastMineBomb}, // UNT_BLASTMINE -> ef_trap_03_3
	144: {ModelPath: "외부소품\\트랩02.rsm"},                                         // UNT_SKIDTRAP -> ef_trap_02
	145: {ModelPath: "외부소품\\트랩01.rsm"},                                         // UNT_ANKLESNARE -> ef_trap_01
	147: {ModelPath: "외부소품\\트랩03.rsm"},                                         // UNT_LANDMINE -> ef_trap_03
	148: {ModelPath: "외부소품\\트랩03_6.rsm"},                                       // UNT_SHOCKWAVE -> ef_trap_03_6
	149: {ModelPath: "외부소품\\트랩03_4.rsm", TriggerEffectID: effectSandman},       // UNT_SANDMAN -> ef_trap_03_4
	150: {ModelPath: "외부소품\\트랩03_5.rsm", TriggerEffectID: effectFlasher},       // UNT_FLASHER -> ef_trap_03_5
	151: {ModelPath: "외부소품\\트랩03_2.rsm", TriggerEffectID: effectFreezingTrap},  // UNT_FREEZINGTRAP -> ef_trap_03_2
	152: {ModelPath: "외부소품\\트랩04.rsm", TriggerEffectID: effectClaymore},        // UNT_CLAYMORETRAP -> ef_trap_04
	153: {ModelPath: "외부소품\\트랩05.rsm"},                                         // UNT_TALKIEBOX -> ef_trap_05
}
