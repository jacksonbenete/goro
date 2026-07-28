package db

var mountJobs = map[int]int{
	JobKnight:    JobKnight2,
	JobKnightH:   JobKnight2H,
	JobKnightB:   JobKnight2B,
	JobCrusader:  JobCrusader2,
	JobCrusaderH: JobCrusader2H,
	JobCrusaderB: JobCrusader2B,
}

func MountJob(job int) (int, bool) {
	mounted, ok := mountJobs[job]
	return mounted, ok
}
