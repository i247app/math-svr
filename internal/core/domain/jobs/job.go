package domain

import (
	"context"
	"log"
	"sync"
	"time"

	di "math-ai.com/math-ai/internal/core/di/jobs"
)

type JobManager struct {
	ctx         context.Context
	jobs        []di.JobScheduler
	running     bool
	stopChan    chan struct{}
	wg          sync.WaitGroup
	mu          sync.RWMutex
	currentTick int
}

func NewJobManager(ctx context.Context) *JobManager {
	return &JobManager{
		ctx:      ctx,
		jobs:     []di.JobScheduler{},
		stopChan: make(chan struct{}),
	}
}

func (m *JobManager) RegisterJob(job di.JobScheduler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.jobs = append(m.jobs, job)
	log.Printf("Registered job: %s with schedule: %v (minutes)", job.Name(), job.TickInterval())
}

func (m *JobManager) Start() {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	m.running = true
	m.mu.Unlock()

	log.Println("Starting job manager...")

	m.wg.Add(1)
	go m.scheduler()
}

func (m *JobManager) Stop() {
	m.mu.Lock()
	if !m.running {
		m.mu.Unlock()
		return
	}
	m.running = false
	m.mu.Unlock()

	log.Println("Stopping job manager...")
	close(m.stopChan)
	m.wg.Wait()
	log.Println("Job manager stopped")
}

func (m *JobManager) scheduler() {
	defer m.wg.Done()

	ticker := time.NewTicker(1 * time.Minute) // Check every minute
	defer ticker.Stop()

	for {
		select {
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.tick()
		}
	}
}

func (m *JobManager) tick() {
	m.currentTick++

	m.mu.RLock()
	jobs := make([]di.JobScheduler, len(m.jobs))
	copy(jobs, m.jobs)
	m.mu.RUnlock()

	for _, job := range jobs {
		if m.shouldRunJob(job) {
			go m.runJob(job)
		}
	}
}

func (m *JobManager) shouldRunJob(job di.JobScheduler) bool {
	// Simple schedule matching - in production you'd use a proper scheduling strategy
	return m.currentTick%job.TickInterval() == 0
}

func (m *JobManager) runJob(job di.JobScheduler) {
	log.Printf("[%s] Starting...", job.Name())
	start := time.Now()

	ctx, cancel := context.WithTimeout(m.ctx, 1*time.Minute)
	defer cancel()

	err := job.Run(ctx)
	duration := time.Since(start)

	if err != nil {
		log.Printf("[%s] Failed after %v: %v", job.Name(), duration, err)
	} else {
		log.Printf("[%s] Completed successfully in %v", job.Name(), duration)
	}
}
