package orchestrator

import "github.com/robfig/cron"

type Scheduler struct {
	cronJob *cron.Cron
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		cronJob: cron.New(),
	}
}


func (s *Scheduler) Start() {
	s.cronJob.Start()
}

func (s *Scheduler) AddTask(spec string, cmd func()) {
	s.cronJob.AddFunc(spec, cmd)
}