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
	ProfileId       *int64
	OwnerProfileId  *int64
	SchoolId        *int64
	ProgramId       *int64
	ProgramIds      []int64
	GradeId         *int64
	Search          *string
	IncludeArchived bool
	Page            int64
	Limit           int64
	TakeAll         bool
}

// IRepository owns ma_classrooms persistence.
type IRepository interface {
	FindByClassroomId(ctx context.Context, classroomId int64) (*Classroom, error)
	FindByInviteCode(ctx context.Context, code string) (*Classroom, error)
	ListClassrooms(ctx context.Context, params *ListClassroomsParams) ([]*Classroom, *pagination.Pagination, error)
	ListClassroomsByIds(ctx context.Context, ids []int64) ([]*Classroom, error)
	Create(ctx context.Context, c *Classroom) (*Classroom, error)
	Update(ctx context.Context, c *Classroom) error
	// IncCounts applies signed deltas to the three classroom counters
	// (member_count, student_count, teacher_count) inside the caller's
	// transaction. Each negative delta is clamped at zero by the SQL
	// (GREATEST(0, ...)) so a drifted counter never goes negative.
	// Always called in the same UoW as the member row mutation so the
	// counters and the underlying member rows stay consistent.
	IncCounts(ctx context.Context, classroomId int64, memberDelta, studentDelta, teacherDelta int64) error
	UpdateInviteCode(ctx context.Context, classroomId int64, code *string, expiresDt mtime.MathTime) error
	SetOwnerProfileId(ctx context.Context, classroomId int64, newOwnerProfileId int64) error
	ArchiveByClassroomId(ctx context.Context, classroomId int64) error
	RestoreByClassroomId(ctx context.Context, classroomId int64) error
	SoftDeleteByClassroomId(ctx context.Context, classroomId int64) error
	ForceDeleteByClassroomId(ctx context.Context, classroomId int64) error
}

// IClassroomProgramRepository owns ma_classroom_programs — the
// many-to-many edge between a classroom and one or more programs/books.
// The replace-set update path inside UpdateClassroomCommand reads the
// current set, hard-deletes removed pairs, and inserts new ones, all
// inside the surrounding UoW.
type IClassroomProgramRepository interface {
	// ListProgramIdsByClassroomId returns the ACTIVE program ids linked
	// to a single classroom, in insertion order. Used by GetClassroom.
	ListProgramIdsByClassroomId(ctx context.Context, classroomId int64) ([]int64, error)
	// ListProgramIdsByClassroomIds batches the ListByClassroomId call
	// across a page of classrooms to avoid N+1 on ListClassrooms. The
	// returned map only contains keys for classrooms that have at least
	// one program — callers should treat a missing key as "zero
	// programs", not "not hydrated".
	ListProgramIdsByClassroomIds(ctx context.Context, classroomIds []int64) (map[int64][]int64, error)
	// Create inserts one (classroom_id, program_id) pair. The DB UNIQUE
	// (classroom_id, program_id) is the hard backstop against duplicates;
	// callers should dedupe their input slice to surface a friendly
	// CLASSROOM_PROGRAM_DUPLICATE before reaching this method.
	Create(ctx context.Context, cp *ClassroomProgram) (*ClassroomProgram, error)
	// DeleteByPair physically removes one (classroom_id, program_id)
	// row. The edge has no history worth keeping, so this is a hard
	// DELETE; soft-deletion would force a partial unique index (not
	// supported on MySQL 8) to allow re-adding the same program later.
	DeleteByPair(ctx context.Context, classroomId, programId int64) error
	// DeleteByClassroomId clears every program link for a classroom.
	// Used by the soft-delete and force-delete paths so a deleted
	// classroom doesn't leak into program-filtered listings via a stale
	// junction row.
	DeleteByClassroomId(ctx context.Context, classroomId int64) error
}

// ListMembersParams narrows ma_classroom_members reads. ClassroomId is
// required for the "members of classroom" path; ProfileId is required
// for the "my classrooms" / "my pending invitations" path; Role and
// Status further narrow. Statuses ∈ {PENDING, ACTIVE, REJECTED, LEFT,
// REMOVED} — PENDING rows are invitations, the rest are membership
// lifecycle states.
type ListMembersParams struct {
	ClassroomId *int64
	ProfileId   *int64
	Role        *string
	Status      *string
	Page        int64
	Limit       int64
	TakeAll     bool
}

// IMemberRepository owns ma_classroom_members — both the membership and
// the invitation lifecycle now ride this single table (member_status
// PENDING → ACTIVE/REJECTED → LEFT/REMOVED). The (classroom_id,
// profile_id) UNIQUE constraint means re-invitation and rejoin both
// reactivate the existing row in place rather than insert a duplicate.
type IMemberRepository interface {
	FindByMemberId(ctx context.Context, memberId int64) (*Member, error)
	FindByClassroomAndProfile(ctx context.Context, classroomId, profileId int64) (*Member, error)
	ListMembers(ctx context.Context, params *ListMembersParams) ([]*Member, *pagination.Pagination, error)
	CountActiveByClassroomId(ctx context.Context, classroomId int64) (int64, error)
	Create(ctx context.Context, m *Member) (*Member, error)
	Update(ctx context.Context, m *Member) error
	SetRole(ctx context.Context, memberId int64, role string) error
	MarkLeft(ctx context.Context, memberId int64) error
	MarkRemoved(ctx context.Context, memberId, removedByProfileId int64) error
	// Invite flips an existing row into a fresh PENDING_INVITATION
	// (member_role/invite_by/invite_dt refreshed, terminal-state
	// timestamps cleared). Used by the send-invitation path when the
	// target previously REJECTED/LEFT/REMOVED so the UNIQUE
	// (classroom_id, profile_id) constraint stays meaningful.
	Invite(ctx context.Context, memberId int64, role string, inviteBy *int64, inviteDt mtime.MathTime, note *string) error
	// RequestToJoin is Invite's user-initiated twin: flips an existing
	// terminal-state row into PENDING_REQUEST, with invite_by NULL
	// (no manager initiated it) and invite_dt = requestedDt as the
	// "requested at" timestamp.
	RequestToJoin(ctx context.Context, memberId int64, role string, requestedDt mtime.MathTime, note *string) error
	// Activate flips a PENDING_INVITATION OR PENDING_REQUEST row to
	// ACTIVE, refreshing joined_dt and clearing terminal-state
	// timestamps. Used by both accept-invitation and approve-request
	// paths; the command-layer guard decides which prior status is
	// allowed. The caller pairs this with Classroom.IncCounts inside
	// the same UoW.
	Activate(ctx context.Context, memberId int64) error
	// Reject flips a PENDING_* row to REJECTED. Reused by both the
	// invitee-rejects-invitation and owner-rejects-request paths;
	// command-layer guards enforce which source status is valid.
	// left_dt is reused as the "responded at" timestamp.
	Reject(ctx context.Context, memberId int64) error
	// Cancel flips a PENDING_* row to REMOVED with the cancelling
	// profile id captured in removed_by_profile_id. Reused by both
	// manager-cancels-invitation and user-cancels-own-request paths.
	Cancel(ctx context.Context, memberId, cancelledByProfileId int64) error
	// Reactivate flips an existing PENDING/LEFT/REMOVED/REJECTED row
	// back to ACTIVE in place. Used by the join-by-code path so the
	// UNIQUE (classroom_id, profile_id) constraint doesn't force a
	// duplicate insert when a profile rejoins a classroom they once
	// left or had a pending invitation for.
	Reactivate(ctx context.Context, memberId int64, role string, invitationId *int64) error
	SoftDeleteByClassroomId(ctx context.Context, classroomId int64) error
	ForceDeleteByClassroomId(ctx context.Context, classroomId int64) error
}
