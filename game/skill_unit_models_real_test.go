package game

import (
	"testing"

	"github.com/kivutar/goro/db"
)

func TestSkillUnitTrapRSMModelsLoadFromRealDataWhenConfigured(t *testing.T) {
	manager := realDataManager(t)
	for unitID, spec := range db.SkillUnitModels {
		rsm, err := loadRSMModel(manager, spec.ModelPath)
		if err != nil {
			t.Fatalf("load trap model unit=%d model=%s: %v", unitID, spec.ModelPath, err)
		}
		if len(rsm.Nodes) == 0 {
			t.Fatalf("trap model unit=%d model=%s has no nodes", unitID, spec.ModelPath)
		}
	}
}
