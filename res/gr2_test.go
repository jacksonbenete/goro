package res

import "testing"

func TestGR2ResourceHelpers(t *testing.T) {
	if !IsGR2ResourceName(`data/model/3dmob/Guildflag90_1.gr2`) {
		t.Fatal("expected GR2 resource name")
	}
	if IsGR2ResourceName("PORING.spr") {
		t.Fatal("sprite resource should not be GR2")
	}
	if got := GR2ModelResourceCandidates("Empelium90_0.gr2")[0]; got != `data\model\3dmob\empelium90_0.gr2` {
		t.Fatalf("first model candidate = %q", got)
	}
	if bone, ok := GR2BoneTypeFromName("Kguardian90_7.gr2"); !ok || bone != 7 {
		t.Fatalf("bone type = %d ok=%t, want 7 true", bone, ok)
	}
	if got := GR2AnimationResourceCandidates(7, GR2ActionAttack)[0]; got != `data\model\3dmob_bone\7_attack.gr2` {
		t.Fatalf("attack animation candidate = %q", got)
	}
}

func TestGR2RealDataModels(t *testing.T) {
	manager := realDataManager(t)
	tests := []struct {
		name        string
		minVerts    int
		minIndices  int
		minBatches  int
		minTextures int
	}{
		{name: "Empelium90_0.gr2", minVerts: 100, minIndices: 300, minBatches: 1, minTextures: 1},
		{name: "Guildflag90_1.gr2", minVerts: 20, minIndices: 30, minBatches: 1, minTextures: 1},
		{name: "Kguardian90_7.gr2", minVerts: 100, minIndices: 300, minBatches: 2, minTextures: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, data, ok := manager.ReadFirst(GR2ModelResourceCandidates(tt.name))
			if !ok {
				t.Skipf("%s not present in real data", tt.name)
			}
			file, err := ParseGR2(data)
			if err != nil {
				t.Fatal(err)
			}
			if len(file.Models) == 0 || len(file.Skeletons) == 0 {
				t.Fatalf("models=%d skeletons=%d, want non-empty", len(file.Models), len(file.Skeletons))
			}
			geometry, err := BuildGR2Geometry(file, 0)
			if err != nil {
				t.Fatal(err)
			}
			if len(geometry.Vertices) < tt.minVerts || len(geometry.Indices) < tt.minIndices || len(geometry.Batches) < tt.minBatches || len(file.Textures) < tt.minTextures {
				t.Fatalf("geometry vertices=%d indices=%d batches=%d textures=%d, minimum %d/%d/%d/%d", len(geometry.Vertices), len(geometry.Indices), len(geometry.Batches), len(file.Textures), tt.minVerts, tt.minIndices, tt.minBatches, tt.minTextures)
			}
			for i, texture := range file.Textures {
				if texture.Width <= 0 || texture.Height <= 0 {
					t.Fatalf("texture %d size = %dx%d", i, texture.Width, texture.Height)
				}
				if _, err := texture.RGBA(); err != nil {
					t.Fatalf("texture %d decode failed: %v", i, err)
				}
			}
		})
	}
}
