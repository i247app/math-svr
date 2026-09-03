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

// POST /jobs/schedule/update
//
// Body: { "name": "system.noop", "schedule": { ... } }
//
// Retargets one CronJob's cadence without a redeploy. The schedule
// object is one of:
//
//	{"kind":"every","every_seconds":60}      — every minute (the floor)
//	{"kind":"every","every_seconds":3600}    — hourly
//	{"kind":"every","every_seconds":7200}    — every 2 hours
//	{"kind":"daily","hour":6,"minute":30,"timezone":"Asia/Ho_Chi_Minh"}
//	{"kind":"weekly","weekday":1,"hour":8,"minute":0,"timezone":"Asia/Tokyo"}
//
// Takes effect immediately — the scheduler loop is woken rather than
// left to finish its pending sleep. Responds with the job's resulting
// state so the caller can confirm the resolved schedule (timezone
// included) and the recomputed next_run_at.
//
// The override is in-memory: a restart reverts the job to the schedule
// declared in internal/jobs/.
func (h *JobHandler) HandleUpdateJobSchedule(w http.ResponseWriter, r *http.Request) {
	var req UpdateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.svc.UpdateSchedule(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /jobs/schedule/reset
//
// Body: { "name": "system.noop" }
//
// Drops the runtime override so the job returns to its code-declared
// schedule. Idempotent — resetting a job that was never overridden
// succeeds and reports the declared schedule.
func (h *JobHandler) HandleResetJobSchedule(w http.ResponseWriter, r *http.Request) {
	var req JobNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.svc.ResetSchedule(r.Context(), &req)
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
