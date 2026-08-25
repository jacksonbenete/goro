package db

import "testing"

func TestAdoptableChildJobs(t *testing.T) {
	for _, job := range []int{JobNovice, JobSwordman, JobMagician, JobArcher, JobAcolyte, JobMerchant, JobThief, JobSuperNovice} {
		if !IsAdoptableChildJob(job) {
			t.Fatalf("job %d should be adoptable", job)
		}
	}
	for _, job := range []int{JobKnight, JobNoviceH, JobNoviceB, JobGunslinger, JobNinja} {
		if IsAdoptableChildJob(job) {
			t.Fatalf("job %d should not be adoptable", job)
		}
	}
}

func TestBabyJobs(t *testing.T) {
	for _, job := range []int{JobNoviceB, JobSwordmanB, JobSuperNoviceB, JobNinjaB, JobTaekwonB, JobGunslingerB} {
		if !IsBabyJob(job) {
			t.Fatalf("job %d should be a baby job", job)
		}
	}
	for _, job := range []int{JobNovice, JobNoviceH, JobKnight, JobNinja} {
		if IsBabyJob(job) {
			t.Fatalf("job %d should not be a baby job", job)
		}
	}
}
