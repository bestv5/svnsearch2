package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Job struct {
	ID       string
	Interval time.Duration
	Handler  func()
	ticker   *time.Ticker
	cancel   context.CancelFunc
}

type Scheduler struct {
	jobs map[string]*Job
	mu   sync.RWMutex
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		jobs: make(map[string]*Job),
	}
}

func (s *Scheduler) AddJob(id string, interval time.Duration, handler func()) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[id]; exists {
		return fmt.Errorf("任务已存在: %s", id)
	}

	ctx, cancel := context.WithCancel(context.Background())
	ticker := time.NewTicker(interval)

	job := &Job{
		ID:       id,
		Interval: interval,
		Handler:  handler,
		ticker:   ticker,
		cancel:   cancel,
	}

	s.jobs[id] = job

	go func() {
		for {
			select {
			case <-ticker.C:
				handler()
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (s *Scheduler) RemoveJob(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[id]
	if !exists {
		return fmt.Errorf("任务不存在: %s", id)
	}

	job.ticker.Stop()
	job.cancel()
	delete(s.jobs, id)

	return nil
}

func (s *Scheduler) UpdateJob(id string, interval time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[id]
	if !exists {
		return fmt.Errorf("任务不存在: %s", id)
	}

	job.ticker.Stop()
	job.Interval = interval
	job.ticker = time.NewTicker(interval)

	return nil
}

func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, job := range s.jobs {
		job.ticker.Stop()
		job.cancel()
		delete(s.jobs, id)
	}
}

func (s *Scheduler) GetJobIDs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := make([]string, 0, len(s.jobs))
	for id := range s.jobs {
		ids = append(ids, id)
	}
	return ids
}
