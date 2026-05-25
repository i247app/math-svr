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
