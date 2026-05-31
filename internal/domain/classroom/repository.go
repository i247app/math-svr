package classroom

import (
	"context"

	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// ListClassroomsParams narrows the listing query. ProfileId filters to
// classrooms the profile is an active member of (any role). OwnerProfileId
// filters to classrooms the profile owns. Either or both may be set.
// Search matches case-insensitively against name. IncludeArchived
// controls whether ARCHIVED rows are returned alongside ACTIVE ones.
//
// Program filters operate via the ma_classroom_programs junction table:
// ProgramId matches a single program; ProgramIds matches the OR-set
// ("classroom contains at least one of these programs"). When both are
// set the singleton is appended to the slice and the union is matched.
type ListClassroomsParams struct {
	ProfileId        *string
	OwnerProfileId   *string
	SchoolId         *string
	ProgramId        *string
	ProgramIds       []string
	GradeId          *string
	Search           *string
	IncludeArchived  bool
	Page             int64
	Limit            int64
	TakeAll          bool
}

// IRepository owns ma_classrooms persistence.
type IRepository interface {
	FindByClassroomId(ctx context.Context, classroomId string) (*Classroom, error)
	FindByInviteCode(ctx context.Context, code string) (*Classroom, error)
	ListClassrooms(ctx context.Context, params *ListClassroomsParams) ([]*Classroom, *pagination.Pagination, error)
	ListClassroomsByIds(ctx context.Context, ids []string) ([]*Classroom, error)
	Create(ctx context.Context, c *Classroom) (*Classroom, error)
	Update(ctx context.Context, c *Classroom) error
	// IncCounts applies signed deltas to the three classroom counters
	// (member_count, student_count, teacher_count) inside the caller's
	// transaction. Each negative delta is clamped at zero by the SQL
	// (GREATEST(0, ...)) so a drifted counter never goes negative.
	// Always called in the same UoW as the member row mutation so the
	// counters and the underlying member rows stay consistent.
	IncCounts(ctx context.Context, classroomId string, memberDelta, studentDelta, teacherDelta int64) error
	UpdateInviteCode(ctx context.Context, classroomId string, code *string, expiresDt mtime.MathTime) error
	SetOwnerProfileId(ctx context.Context, classroomId, newOwnerProfileId string) error
	ArchiveByClassroomId(ctx context.Context, classroomId string) error
	RestoreByClassroomId(ctx context.Context, classroomId string) error
	SoftDeleteByClassroomId(ctx context.Context, classroomId string) error
	ForceDeleteByClassroomId(ctx context.Context, classroomId string) error
}

// IClassroomProgramRepository owns ma_classroom_programs — the
// many-to-many edge between a classroom and one or more programs/books.
// The replace-set update path inside UpdateClassroomCommand reads the
// current set, hard-deletes removed pairs, and inserts new ones, all
// inside the surrounding UoW.
type IClassroomProgramRepository interface {
	// ListProgramIdsByClassroomId returns the ACTIVE program ids linked
	// to a single classroom, in insertion order. Used by GetClassroom.
	ListProgramIdsByClassroomId(ctx context.Context, classroomId string) ([]string, error)
	// ListProgramIdsByClassroomIds batches the ListByClassroomId call
	// across a page of classrooms to avoid N+1 on ListClassrooms. The
	// returned map only contains keys for classrooms that have at least
	// one program — callers should treat a missing key as "zero
	// programs", not "not hydrated".
	ListProgramIdsByClassroomIds(ctx context.Context, classroomIds []string) (map[string][]string, error)
	// Create inserts one (classroom_id, program_id) pair. The DB UNIQUE
	// (classroom_id, program_id) is the hard backstop against duplicates;
	// callers should dedupe their input slice to surface a friendly
	// CLASSROOM_PROGRAM_DUPLICATE before reaching this method.
	Create(ctx context.Context, cp *ClassroomProgram) (*ClassroomProgram, error)
	// DeleteByPair physically removes one (classroom_id, program_id)
	// row. The edge has no history worth keeping, so this is a hard
	// DELETE; soft-deletion would force a partial unique index (not
	// supported on MySQL 8) to allow re-adding the same program later.
	DeleteByPair(ctx context.Context, classroomId, programId string) error
	// DeleteByClassroomId clears every program link for a classroom.
	// Used by the soft-delete and force-delete paths so a deleted
	// classroom doesn't leak into program-filtered listings via a stale
	// junction row.
	DeleteByClassroomId(ctx context.Context, classroomId string) error
}

// ListMembersParams narrows ma_classroom_members reads. ClassroomId is
// required for the "members of classroom" path; ProfileId is required
// for the "my classrooms" path; Role / Status filter both.
type ListMembersParams struct {
	ClassroomId *string
	ProfileId   *string
	Role        *string
	Status      *string
	Page        int64
	Limit       int64
	TakeAll     bool
}

// IMemberRepository owns ma_classroom_members. Note: the same (classroom,
// profile) pair has at most one row — Upsert flips an existing inactive
// row back to ACTIVE rather than inserting a duplicate, so the unique
// constraint stays meaningful.
type IMemberRepository interface {
	FindByMemberId(ctx context.Context, memberId string) (*Member, error)
	FindByClassroomAndProfile(ctx context.Context, classroomId, profileId string) (*Member, error)
	ListMembers(ctx context.Context, params *ListMembersParams) ([]*Member, *pagination.Pagination, error)
	CountActiveByClassroomId(ctx context.Context, classroomId string) (int64, error)
	Create(ctx context.Context, m *Member) (*Member, error)
	Update(ctx context.Context, m *Member) error
	SetRole(ctx context.Context, memberId, role string) error
	MarkLeft(ctx context.Context, memberId string) error
	MarkRemoved(ctx context.Context, memberId, removedByProfileId string) error
	// Reactivate flips an existing INVITED/LEFT/REMOVED member back to
	// ACTIVE. Resets joined_dt, clears left_dt / removed_dt /
	// removed_by_profile_id, and assigns the given role. Used by the
	// join-by-code path (and the future accept-invitation path) so the
	// UNIQUE (classroom_id, profile_id) constraint doesn't force a
	// duplicate insert when a profile rejoins a classroom they once left.
	Reactivate(ctx context.Context, memberId, role string, invitationId *string) error
	SoftDeleteByClassroomId(ctx context.Context, classroomId string) error
	ForceDeleteByClassroomId(ctx context.Context, classroomId string) error
}

// ListInvitationsParams narrows ma_classroom_invitations reads. At least
// one of ClassroomId / InvitedProfileId / InviteeIdentifier should be
// set in practice; an empty params returns every active invitation.
type ListInvitationsParams struct {
	ClassroomId        *string
	InviterProfileId   *string
	InvitedProfileId   *string
	InviteeIdentifier  *string
	Status             *string
	Page               int64
	Limit              int64
	TakeAll            bool
}

// IInvitationRepository owns ma_classroom_invitations.
type IInvitationRepository interface {
	FindByInvitationId(ctx context.Context, invitationId string) (*Invitation, error)
	FindByToken(ctx context.Context, token string) (*Invitation, error)
	// FindPendingByClassroomAndProfile guards against duplicate active
	// invitations for the same (classroom, profile) pair.
	FindPendingByClassroomAndProfile(ctx context.Context, classroomId, profileId string) (*Invitation, error)
	// FindPendingByClassroomAndIdentifier is the equivalent guard for
	// identifier-based invitations (email / phone / explicit profile id).
	FindPendingByClassroomAndIdentifier(ctx context.Context, classroomId, identifier string) (*Invitation, error)
	ListInvitations(ctx context.Context, params *ListInvitationsParams) ([]*Invitation, *pagination.Pagination, error)
	Create(ctx context.Context, inv *Invitation) (*Invitation, error)
	Update(ctx context.Context, inv *Invitation) error
	SetStatus(ctx context.Context, invitationId, newStatus string, respondedByProfileId *string) error
	// ExpirePending sweeps PENDING rows whose expires_dt < now into the
	// EXPIRED state. Returns the number of rows affected for the job's
	// telemetry. Used by the reaper job in Phase 6.
	ExpirePending(ctx context.Context) (int64, error)
	SoftDeleteByClassroomId(ctx context.Context, classroomId string) error
	ForceDeleteByClassroomId(ctx context.Context, classroomId string) error
}
