package db

import "testing"

func TestMountJobMapsPecoJobs(t *testing.T) {
	tests := []struct {
		job  int
		want int
	}{
		{JobKnight, JobKnight2},
		{JobKnightH, JobKnight2H},
		{JobKnightB, JobKnight2B},
		{JobCrusader, JobCrusader2},
		{JobCrusaderH, JobCrusader2H},
		{JobCrusaderB, JobCrusader2B},
	}
	for _, tt := range tests {
		got, ok := MountJob(tt.job)
		if !ok || got != tt.want {
			t.Fatalf("MountJob(%d) = %d, %t; want %d, true", tt.job, got, ok, tt.want)
		}
	}
}

func TestMountJobDoesNotRemapMountedJob(t *testing.T) {
	if got, ok := MountJob(JobKnight2); ok {
		t.Fatalf("MountJob(JobKnight2) = %d, true; want no remap", got)
	}
}
