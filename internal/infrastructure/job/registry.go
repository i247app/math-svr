package job

import "sync"

// Registry holds the static name → handler mapping for both CronJobs
// and TaskHandlers. Population happens once at boot — typically from
// internal/jobs/jobs.go via RegisterAll — and the Runtime reads it
// thereafter. Duplicate registration is a programmer bug, not a
// runtime concern; the helpers panic.
//
// The mutex exists to keep -race quiet on concurrent reads after
// startup; the post-boot access pattern is read-mostly.
type Registry struct {
	mu    sync.RWMutex
	crons map[string]CronJob
	tasks map[string]TaskHandler
}

func NewRegistry() *Registry {
	return &Registry{
		crons: make(map[string]CronJob),
		tasks: make(map[string]TaskHandler),
	}
}

// RegisterCron adds j to the registry. Panics on duplicate name.
func (r *Registry) RegisterCron(j CronJob) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.crons[j.Name()]; exists {
		panic("job: duplicate cron registration: " + j.Name())
	}
	r.crons[j.Name()] = j
}

// RegisterTask adds h to the registry. Panics on duplicate name.
func (r *Registry) RegisterTask(h TaskHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.tasks[h.Name()]; exists {
		panic("job: duplicate task registration: " + h.Name())
	}
	r.tasks[h.Name()] = h
}

func (r *Registry) Crons() []CronJob {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]CronJob, 0, len(r.crons))
	for _, j := range r.crons {
		out = append(out, j)
	}
	return out
}

func (r *Registry) Tasks() []TaskHandler {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]TaskHandler, 0, len(r.tasks))
	for _, h := range r.tasks {
		out = append(out, h)
	}
	return out
}

func (r *Registry) CronByName(name string) (CronJob, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	j, ok := r.crons[name]
	return j, ok
}

func (r *Registry) TaskByName(name string) (TaskHandler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.tasks[name]
	return h, ok
}
