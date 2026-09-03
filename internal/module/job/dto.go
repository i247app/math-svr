package job

import (
	jobruntime "math-ai.com/math-ai/internal/infrastructure/job"
)

// Wire DTOs for the job control plane. The list/snapshot path returns
// the runtime's Snapshot directly — the shape is already pure data
// with json tags — so we don't re-map. Action endpoints take a tiny
// request body and respond with a generic acknowledgment.

// JobNameRequest covers list-of-one operations: pause, resume,
// trigger. Path-style ({jobName}) is avoided to keep the admin surface
// consistent with the rest of the project (POST + JSON body).
type JobNameRequest struct {
	Name string `json:"name"`
}

// EnqueueTaskRequest schedules an arbitrary task by name. Payload is
// raw JSON the caller wants the handler to decode; the runtime never
// inspects it.
type EnqueueTaskRequest struct {
	Name    string `json:"name"`
	Payload any    `json:"payload"`

	// DelaySeconds delays the first attempt this many seconds; 0 means
	// "run as soon as a worker is free". Use this for "send the email
	// 60 seconds after the user clicks" patterns.
	DelaySeconds int `json:"delay_seconds"`

	// Attempts overrides the runtime default retry count. Zero falls
	// through to the runtime default.
	Attempts int `json:"attempts"`

	// TimeoutSeconds caps a single attempt. Zero falls through to the
	// runtime default.
	TimeoutSeconds int `json:"timeout_seconds"`
}

// SnapshotResponse is the read-only view returned by HandleListJobs.
// Aliasing keeps the JSON shape identical to what the runtime emits,
// so Tier 2 swapping to a persistent snapshot does not change the
// contract.
type SnapshotResponse = jobruntime.Snapshot

// ActionResponse is the envelope returned by every mutating handler
// (trigger, pause, resume, enqueue). Keep it skinny: the request
// already carries the identifier.
type ActionResponse struct {
	Name   string `json:"name"`
	Result string `json:"result"`
}

// ScheduleRequest is the wire form of jobruntime.Schedule. It is kept
// separate from the runtime type because the wire contract must stay
// stable and self-describing (string kind, IANA timezone) while the
// runtime type is free to change representation.
//
//	{"kind":"every","every_seconds":900}                      → every 15m
//	{"kind":"every","every_seconds":3600}                     → hourly
//	{"kind":"daily","hour":6,"minute":30,"timezone":"Asia/Ho_Chi_Minh"}
//	{"kind":"weekly","weekday":1,"hour":8,"minute":0,"timezone":"Asia/Tokyo"}
type ScheduleRequest struct {
	// Kind selects the variant: "every" | "daily" | "weekly".
	Kind string `json:"kind"`

	// EverySeconds is the interval for kind=every. Bounded by
	// [minEveryInterval, maxEveryInterval] — see validator.go.
	EverySeconds int `json:"every_seconds"`

	// Hour (0-23) and Minute (0-59) apply to kind=daily and kind=weekly.
	Hour   int `json:"hour"`
	Minute int `json:"minute"`

	// Weekday (0=Sunday … 6=Saturday) is required for kind=weekly. It is
	// a pointer so an explicit 0 (Sunday) is distinguishable from
	// "field omitted".
	Weekday *int `json:"weekday"`

	// Timezone is an IANA name resolving Hour/Minute/Weekday for
	// kind=daily and kind=weekly; empty means UTC. Ignored by
	// kind=every, which is interval-based and timezone-free.
	Timezone string `json:"timezone"`
}

// UpdateScheduleRequest retargets one CronJob's cadence at runtime.
type UpdateScheduleRequest struct {
	Name     string           `json:"name"`
	Schedule *ScheduleRequest `json:"schedule"`
}

// ScheduleResponse echoes the job's resulting state. It returns the
// runtime's own JobInfo rather than repeating the request back, so the
// caller sees the schedule as the runtime RESOLVED it — timezone
// included — plus the recomputed next_run_at. A caller that meant
// 06:30 Vietnam time but omitted the timezone sees "daily 06:30 UTC"
// and a next_run_at seven hours off, instead of finding out days later.
type ScheduleResponse struct {
	Result string             `json:"result"`
	Job    jobruntime.JobInfo `json:"job"`
}
