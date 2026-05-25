package job

import (
	"encoding/json"
	"net/http"

	"math-ai.com/math-ai/internal/shared/response"
)

type JobHandler struct {
	svc *Service
}

func NewJobHandler(svc *Service) *JobHandler {
	return &JobHandler{svc: svc}
}

// POST /jobs/list
//
// Returns the consolidated runtime snapshot — every CronJob with its
// schedule, status (running|paused), last run, next run, last error,
// in_flight flag; every registered task by name; and the in-memory
// queue gauges. Read-only.
func (h *JobHandler) HandleListJobs(w http.ResponseWriter, r *http.Request) {
	snap, err := h.svc.Snapshot(r.Context())
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, snap, nil)
}

// POST /jobs/trigger
//
// Body: { "name": "system.session_cleanup" }
//
// Fires the named CronJob immediately, out of schedule. Returns
// JOB_IN_FLIGHT if the job is already executing.
func (h *JobHandler) HandleTriggerJob(w http.ResponseWriter, r *http.Request) {
	var req JobNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.svc.TriggerJob(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /jobs/pause
//
// Body: { "name": "leaderboard.recalc" }
//
// Suspends the scheduler loop. In-flight executions are NOT cancelled;
// pause only stops new scheduling.
func (h *JobHandler) HandlePauseJob(w http.ResponseWriter, r *http.Request) {
	var req JobNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.svc.PauseJob(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /jobs/resume
//
// Body: { "name": "leaderboard.recalc" }
//
// Reverses Pause. The schedule resumes computing from time.Now().
func (h *JobHandler) HandleResumeJob(w http.ResponseWriter, r *http.Request) {
	var req JobNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.svc.ResumeJob(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /tasks/enqueue
//
// Body: { "name": "digest.send", "payload": {...}, "delay_seconds": 0,
//
//	"attempts": 3, "timeout_seconds": 60 }
//
// Submits one task for execution. Mostly an ops tool — application
// code that wants to enqueue a task should call s.runtime.Enqueue
// directly from inside its UnitOfWork (in Tier 2) for atomicity.
func (h *JobHandler) HandleEnqueueTask(w http.ResponseWriter, r *http.Request) {
	var req EnqueueTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.svc.EnqueueTask(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}
