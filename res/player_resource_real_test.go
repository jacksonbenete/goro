package res

import "testing"

func TestPlayerBodyResourceRealWhenConfigured(t *testing.T) {
	job := 0
	sex := byte(1)

	manager := realDataManager(t)
	actSource, actData, ok := manager.ReadFirst(PlayerBodyResourceCandidates(job, sex, "act"))
	if !ok {
		t.Skipf("body act not found for job=%d sex=%d", job, sex)
	}
	sprSource, sprData, ok := manager.ReadFirst(PlayerBodyResourceCandidates(job, sex, "spr"))
	if !ok {
		t.Skipf("body spr not found for job=%d sex=%d", job, sex)
	}
	act, err := ParseACT(actData)
	if err != nil {
		t.Fatalf("parse %s: %v", actSource, err)
	}
	spr, err := ParseSPR(sprData)
	if err != nil {
		t.Fatalf("parse %s: %v", sprSource, err)
	}
	if len(act.Actions) == 0 || len(spr.Frames) == 0 {
		t.Fatalf("empty body resources: actions=%d frames=%d", len(act.Actions), len(spr.Frames))
	}
	t.Logf("body resources act=%s spr=%s actions=%d frames=%d", actSource, sprSource, len(act.Actions), len(spr.Frames))

	headActSource, headActData, ok := manager.ReadFirst(PlayerHeadResourceCandidates(job, 0, sex, "act"))
	if !ok {
		t.Skipf("head act not found for job=%d sex=%d", job, sex)
	}
	headSprSource, headSprData, ok := manager.ReadFirst(PlayerHeadResourceCandidates(job, 0, sex, "spr"))
	if !ok {
		t.Skipf("head spr not found for job=%d sex=%d", job, sex)
	}
	headAct, err := ParseACT(headActData)
	if err != nil {
		t.Fatalf("parse %s: %v", headActSource, err)
	}
	headSpr, err := ParseSPR(headSprData)
	if err != nil {
		t.Fatalf("parse %s: %v", headSprSource, err)
	}
	if len(headAct.Actions) == 0 || len(headSpr.Frames) == 0 {
		t.Fatalf("empty head resources: actions=%d frames=%d", len(headAct.Actions), len(headSpr.Frames))
	}
	t.Logf("head resources act=%s spr=%s actions=%d frames=%d", headActSource, headSprSource, len(headAct.Actions), len(headSpr.Frames))
}
