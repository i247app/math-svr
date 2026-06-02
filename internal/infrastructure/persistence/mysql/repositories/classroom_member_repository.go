package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"math-ai.com/math-ai/internal/domain/classroom"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

const (
	classroomMemberTable = "ma_classroom_members"

	classroomMemberColumns = `m.id, m.member_id, m.classroom_id, m.profile_id, m.member_role,
		m.invitation_id, m.joined_dt, m.left_dt, m.removed_by_profile_id, m.removed_dt,
		m.last_seen_dt, m.note, m.invite_by, m.invite_dt,
		m.member_status, m.status,
		m.create_id, m.create_dt, m.modify_id, m.modify_dt`

	// classroomMemberActiveWhere excludes only system-inactive and the
	// DELETED business state. PENDING/ACTIVE/REJECTED/LEFT/REMOVED are
	// all visible — callers narrow further via params.Status.
	classroomMemberActiveWhere = `m.status = ? AND m.deleted_dt IS NULL
		AND (m.member_status IS NULL OR m.member_status != ?)`
)

func classroomMemberActiveArgs() []any {
	return []any{enum.StatusActive, enum.ClassroomMemberStatusTypeDeleted}
}

type ClassroomMemberRepository struct {
	db database.Executor
}

func NewClassroomMemberRepository(db database.Executor) classroom.IMemberRepository {
	return &ClassroomMemberRepository{db: db}
}

func scanClassroomMember(s database.RowScanner) (*models.ClassroomMemberModel, error) {
	var m models.ClassroomMemberModel
	if err := s.Scan(&m.Id, &m.MemberId, &m.ClassroomId, &m.ProfileId, &m.MemberRole,
		&m.InvitationId, &m.JoinedDt, &m.LeftDt, &m.RemovedByProfileId, &m.RemovedDt,
		&m.LastSeenDt, &m.Note, &m.InviteBy, &m.InviteDt,
		&m.MemberStatus, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *ClassroomMemberRepository) findOneBy(ctx context.Context, where string, args ...any) (*classroom.Member, error) {
	fullArgs := append(classroomMemberActiveArgs(), args...)
	query := `SELECT ` + classroomMemberColumns + ` FROM ` + classroomMemberTable + ` m WHERE ` +
		classroomMemberActiveWhere + ` AND (` + where + `)`

	m, err := scanClassroomMember(r.db.QueryRow(ctx, query, fullArgs...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("classroom_member repo find (%s): %w", where, err)
	}
	return ModelToDomainClassroomMember(m), nil
}

func (r *ClassroomMemberRepository) findBareById(ctx context.Context, id int64) (*classroom.Member, error) {
	args := append(classroomMemberActiveArgs(), id)
	query := `SELECT ` + classroomMemberColumns + ` FROM ` + classroomMemberTable + ` m WHERE ` +
		classroomMemberActiveWhere + ` AND m.id = ?`

	m, err := scanClassroomMember(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("classroom_member repo find bare by id: %w", err)
	}
	return ModelToDomainClassroomMember(m), nil
}

func (r *ClassroomMemberRepository) FindByMemberId(ctx context.Context, memberId int64) (*classroom.Member, error) {
	return r.findOneBy(ctx, "m.member_id = ?", memberId)
}

func (r *ClassroomMemberRepository) FindByClassroomAndProfile(ctx context.Context, classroomId, profileId int64) (*classroom.Member, error) {
	return r.findOneBy(ctx, "m.classroom_id = ? AND m.profile_id = ?", classroomId, profileId)
}

func (r *ClassroomMemberRepository) FindByClassroomAndProfileAndInvitedBy(ctx context.Context, classroomId, profileId, inviterProfileId int64) (*classroom.Member, error) {
	return r.findOneBy(ctx, "m.classroom_id = ? AND m.profile_id = ? AND m.invite_by = ?", classroomId, profileId, inviterProfileId)
}

func (r *ClassroomMemberRepository) ListMembers(ctx context.Context, params *classroom.ListMembersParams) ([]*classroom.Member, *pagination.Pagination, error) {
	filterWhere, filterArgs := buildClassroomMemberFilter(params)

	countArgs := append(classroomMemberActiveArgs(), filterArgs...)
	countQuery := `SELECT COUNT(*) FROM ` + classroomMemberTable + ` m WHERE ` +
		classroomMemberActiveWhere + filterWhere

	var total int64
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("classroom_member repo count: %w", err)
	}

	listArgs := append(classroomMemberActiveArgs(), filterArgs...)
	query := `SELECT ` + classroomMemberColumns + ` FROM ` + classroomMemberTable + ` m WHERE ` +
		classroomMemberActiveWhere + filterWhere +
		` ORDER BY m.joined_dt DESC, m.id DESC`

	var pg *pagination.Pagination
	if params == nil || !params.TakeAll {
		page := int64(1)
		limit := int64(20)
		if params != nil {
			page = params.Page
			limit = params.Limit
		}
		pg = pagination.NewPagination(page, limit, total)
		query += ` LIMIT ? OFFSET ?`
		listArgs = append(listArgs, pg.Size, pg.Skip)
	} else {
		pg = pagination.NewPagination(1, total, total)
	}

	rows, err := r.db.Query(ctx, query, listArgs...)
	if err != nil {
		return nil, nil, fmt.Errorf("classroom_member repo list: %w", err)
	}
	defer rows.Close()

	var out []*classroom.Member
	for rows.Next() {
		m, err := scanClassroomMember(rows)
		if err != nil {
			return nil, nil, fmt.Errorf("classroom_member repo scan row: %w", err)
		}
		out = append(out, ModelToDomainClassroomMember(m))
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("classroom_member repo rows iteration: %w", err)
	}
	return out, pg, nil
}

func buildClassroomMemberFilter(params *classroom.ListMembersParams) (string, []any) {
	if params == nil {
		return "", nil
	}
	var (
		clause string
		args   []any
	)
	if params.ClassroomId != nil && *params.ClassroomId != 0 {
		clause += ` AND m.classroom_id = ?`
		args = append(args, *params.ClassroomId)
	}
	if params.ProfileId != nil && *params.ProfileId != 0 {
		clause += ` AND m.profile_id = ?`
		args = append(args, *params.ProfileId)
	}
	if params.Role != nil && *params.Role != "" {
		clause += ` AND m.member_role = ?`
		args = append(args, *params.Role)
	}
	if params.Status != nil && strings.TrimSpace(*params.Status) != "" {
		clause += ` AND m.member_status = ?`
		args = append(args, strings.TrimSpace(*params.Status))
	}
	return clause, args
}

// ListByProfileAndClassroomIds returns every active (non-DELETED) row
// for the given profile across the requested classroom ids. Pulled in
// one round trip via IN(...). Terminal-state rows
// (REJECTED/LEFT/REMOVED) are included — the caller is responsible for
// mapping them to "not participating" when composing relationship.
func (r *ClassroomMemberRepository) ListByProfileAndClassroomIds(ctx context.Context, profileId int64, classroomIds []int64) ([]*classroom.Member, error) {
	if profileId == 0 || len(classroomIds) == 0 {
		return nil, nil
	}
	placeholders := make([]string, len(classroomIds))
	args := classroomMemberActiveArgs()
	args = append(args, profileId)
	for i, id := range classroomIds {
		placeholders[i] = "?"
		args = append(args, id)
	}
	query := `SELECT ` + classroomMemberColumns + ` FROM ` + classroomMemberTable + ` m WHERE ` +
		classroomMemberActiveWhere +
		` AND m.profile_id = ? AND m.classroom_id IN (` + strings.Join(placeholders, ", ") + `)`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("classroom_member repo list by profile and classroom ids: %w", err)
	}
	defer rows.Close()

	var out []*classroom.Member
	for rows.Next() {
		m, err := scanClassroomMember(rows)
		if err != nil {
			return nil, fmt.Errorf("classroom_member repo scan row: %w", err)
		}
		out = append(out, ModelToDomainClassroomMember(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("classroom_member repo rows iteration: %w", err)
	}
	return out, nil
}

func (r *ClassroomMemberRepository) CountActiveByClassroomId(ctx context.Context, classroomId int64) (int64, error) {
	query := `SELECT COUNT(*) FROM ` + classroomMemberTable + ` m WHERE ` +
		classroomMemberActiveWhere + ` AND m.classroom_id = ? AND m.member_status = ?`
	args := append(classroomMemberActiveArgs(),
		classroomId, enum.ClassroomMemberStatusTypeActive)

	var n int64
	if err := r.db.QueryRow(ctx, query, args...).Scan(&n); err != nil {
		return 0, fmt.Errorf("classroom_member repo count active: %w", err)
	}
	return n, nil
}

// ListMembersByExerciseSubmission powers the teacher-side roster split
// for a given exercise. ACTIVE members are filtered through a JOIN
// against ma_classroom_exercise_submissions:
//
//	Submitted=true   → INNER JOIN  (only members with a non-DELETED
//	                                 submission for the exercise)
//	Submitted=false  → LEFT JOIN s ON … WHERE s.id IS NULL  (members
//	                                 with no submission row)
//
// Search joins ma_profiles for a case-insensitive LIKE on
// ma_profiles.name (utf8mb4_0900_ai_ci collation). One SQL query per
// page; pagination is applied AFTER the JOIN so the page total
// reflects the filtered cohort.
func (r *ClassroomMemberRepository) ListMembersByExerciseSubmission(
	ctx context.Context,
	params *classroom.ListMembersByExerciseSubmissionParams,
) ([]*classroom.Member, *pagination.Pagination, error) {
	if params == nil || params.ClassroomId == 0 || params.ClassroomExerciseId == 0 {
		return nil, nil, fmt.Errorf("classroom_member repo list-by-submission: classroom_id and classroom_exercise_id are required")
	}
	if params.Limit <= 0 {
		params.Limit = pagination.DefaultPageSize
	}
	if params.Page < 1 {
		params.Page = 1
	}

	joinKind := "LEFT JOIN"
	if params.Submitted {
		joinKind = "INNER JOIN"
	}

	subOn := `s.classroom_exercise_id = ?
		AND s.profile_id = m.profile_id
		AND s.status = ?
		AND (s.submission_status IS NULL OR s.submission_status != ?)
		AND s.deleted_dt IS NULL`

	// Member is always ACTIVE — both endpoints are roster-scoped to
	// currently participating students.
	memberWhere := `m.classroom_id = ?
		AND m.status = ?
		AND m.deleted_dt IS NULL
		AND m.member_role IN (?, ?)
		AND m.member_status = ?`

	args := []any{
		// JOIN args
		params.ClassroomExerciseId,
		enum.StatusActive,
		enum.ClassroomExerciseSubmissionStatusDeleted,
		// Member where args
		params.ClassroomId,
		enum.StatusActive,
		enum.ClassroomMemberRoleTypeTeacher,
		enum.ClassroomMemberRoleTypeStudent,
		enum.ClassroomMemberStatusTypeActive,
	}

	searchClause := ""
	if params.Search != nil {
		if v := strings.TrimSpace(*params.Search); v != "" {
			pattern := "%" + escapeLikePattern(v) + "%"
			searchClause = ` AND p.name LIKE ? ESCAPE '\\'`
			args = append(args, pattern)
		}
	}

	// LEFT JOIN ma_profiles for the search predicate. The join stays in
	// place even when no search is supplied so the WHERE shape doesn't
	// branch — MySQL elides the unused join cheaply via index lookups
	// on profile_id.
	whereClause := ""
	if params.Submitted {
		whereClause = memberWhere + searchClause
	} else {
		whereClause = memberWhere + ` AND s.id IS NULL` + searchClause
	}

	from := classroomMemberTable + ` m
		LEFT JOIN ma_profiles p ON p.profile_id = m.profile_id
		` + joinKind + ` ` + classroomExerciseSubmissionTable + ` s ON ` + subOn

	countQuery := `SELECT COUNT(*) FROM ` + from + ` WHERE ` + whereClause
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("classroom_member repo list-by-submission count: %w", err)
	}

	pg := pagination.NewPagination(params.Page, params.Limit, total)

	orderBy := buildClassroomMemberByExerciseOrderBy(params)
	listArgs := append([]any{}, args...)
	listArgs = append(listArgs, pg.Size, pg.Skip)
	listQuery := `SELECT ` + classroomMemberColumns + ` FROM ` + from + ` WHERE ` + whereClause +
		` ORDER BY ` + orderBy + ` LIMIT ? OFFSET ?`

	rows, err := r.db.Query(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, nil, fmt.Errorf("classroom_member repo list-by-submission: %w", err)
	}
	defer rows.Close()

	var out []*classroom.Member
	for rows.Next() {
		m, err := scanClassroomMember(rows)
		if err != nil {
			return nil, nil, fmt.Errorf("classroom_member repo list-by-submission scan: %w", err)
		}
		out = append(out, ModelToDomainClassroomMember(m))
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("classroom_member repo list-by-submission iteration: %w", err)
	}
	return out, pg, nil
}

func buildClassroomMemberByExerciseOrderBy(params *classroom.ListMembersByExerciseSubmissionParams) string {
	column := "m.joined_dt"
	if params.SortBy != nil {
		switch *params.SortBy {
		case "joined":
			column = "m.joined_dt"
		case "name":
			column = "p.name"
		}
	}
	direction := "DESC"
	if params.SortOrder != nil && *params.SortOrder == "asc" {
		direction = "ASC"
	}
	return column + " " + direction + ", m.id " + direction
}

// CountPendingRequestsByClassroomIds groups the PENDING_REQUEST rows
// for the given classroom ids in one round trip. Zero-count classrooms
// are absent from the map — the service layer treats a missing key as 0.
func (r *ClassroomMemberRepository) CountPendingRequestsByClassroomIds(ctx context.Context, classroomIds []int64) (map[int64]int64, error) {
	if len(classroomIds) == 0 {
		return map[int64]int64{}, nil
	}
	placeholders := strings.Repeat("?,", len(classroomIds))
	placeholders = placeholders[:len(placeholders)-1]
	query := `SELECT m.classroom_id, COUNT(*) FROM ` + classroomMemberTable + ` m WHERE ` +
		classroomMemberActiveWhere + ` AND m.member_status = ? AND m.classroom_id IN (` + placeholders + `) GROUP BY m.classroom_id`
	args := append(classroomMemberActiveArgs(), enum.ClassroomMemberStatusTypePendingRequest)
	for _, id := range classroomIds {
		args = append(args, id)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("classroom_member repo count pending requests: %w", err)
	}
	defer rows.Close()

	out := make(map[int64]int64, len(classroomIds))
	for rows.Next() {
		var cid, n int64
		if err := rows.Scan(&cid, &n); err != nil {
			return nil, fmt.Errorf("classroom_member repo count pending requests scan: %w", err)
		}
		out[cid] = n
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("classroom_member repo count pending requests iteration: %w", err)
	}
	return out, nil
}

func (r *ClassroomMemberRepository) Create(ctx context.Context, m *classroom.Member) (*classroom.Member, error) {
	var joinedArg, leftArg, removedArg, lastSeenArg, inviteDtArg any
	if m.JoinedDt().IsValid() {
		joinedArg = m.JoinedDt().Time
	}
	if m.LeftDt().IsValid() {
		leftArg = m.LeftDt().Time
	}
	if m.RemovedDt().IsValid() {
		removedArg = m.RemovedDt().Time
	}
	if m.LastSeenDt().IsValid() {
		lastSeenArg = m.LastSeenDt().Time
	}
	if m.InviteDt().IsValid() {
		inviteDtArg = m.InviteDt().Time
	}

	query := `
		INSERT INTO ` + classroomMemberTable + `
			(member_id, classroom_id, profile_id, member_role,
			 invitation_id, joined_dt, left_dt, removed_by_profile_id, removed_dt,
			 last_seen_dt, note, invite_by, invite_dt, member_status, create_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(ctx, query,
		m.MemberId(), m.ClassroomId(), m.ProfileId(), m.MemberRole(),
		m.InvitationId(), joinedArg, leftArg, m.RemovedByProfileId(), removedArg,
		lastSeenArg, m.Note(), m.InviteBy(), inviteDtArg, m.MemberStatus(), m.CreateId())
	if err != nil {
		return nil, fmt.Errorf("classroom_member repo create: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("classroom_member repo last insert id: %w", err)
	}
	return r.findBareById(ctx, id)
}

// Update applies a partial patch — Role / Status / Note / timestamp
// transitions go through COALESCE so callers can leave fields untouched.
func (r *ClassroomMemberRepository) Update(ctx context.Context, m *classroom.Member) error {
	var roleArg any
	if m.MemberRole() != "" {
		roleArg = m.MemberRole()
	}
	var joinedArg, leftArg, removedArg, lastSeenArg, inviteDtArg any
	if m.JoinedDt().IsValid() {
		joinedArg = m.JoinedDt().Time
	}
	if m.LeftDt().IsValid() {
		leftArg = m.LeftDt().Time
	}
	if m.RemovedDt().IsValid() {
		removedArg = m.RemovedDt().Time
	}
	if m.LastSeenDt().IsValid() {
		lastSeenArg = m.LastSeenDt().Time
	}
	if m.InviteDt().IsValid() {
		inviteDtArg = m.InviteDt().Time
	}

	query := `
		UPDATE ` + classroomMemberTable + `
		SET member_role           = COALESCE(?, member_role),
			invitation_id         = COALESCE(?, invitation_id),
			joined_dt             = COALESCE(?, joined_dt),
			left_dt               = COALESCE(?, left_dt),
			removed_by_profile_id = COALESCE(?, removed_by_profile_id),
			removed_dt            = COALESCE(?, removed_dt),
			last_seen_dt          = COALESCE(?, last_seen_dt),
			note                  = COALESCE(?, note),
			invite_by             = COALESCE(?, invite_by),
			invite_dt             = COALESCE(?, invite_dt),
			member_status         = COALESCE(?, member_status),
			modify_id             = COALESCE(?, modify_id),
			modify_dt             = ?
		WHERE member_id = ?
	`
	if _, err := r.db.Exec(ctx, query,
		roleArg, m.InvitationId(), joinedArg, leftArg,
		m.RemovedByProfileId(), removedArg, lastSeenArg,
		m.Note(), m.InviteBy(), inviteDtArg, m.MemberStatus(), m.ModifyId(),
		mtime.Now().Time, m.MemberId()); err != nil {
		return fmt.Errorf("classroom_member repo update: %w", err)
	}
	return nil
}

func (r *ClassroomMemberRepository) SetRole(ctx context.Context, memberId int64, role string) error {
	query := `
		UPDATE ` + classroomMemberTable + `
		SET member_role = ?,
			modify_dt   = ?
		WHERE member_id = ?
	`
	if _, err := r.db.Exec(ctx, query, role, mtime.Now().Time, memberId); err != nil {
		return fmt.Errorf("classroom_member repo set role: %w", err)
	}
	return nil
}

func (r *ClassroomMemberRepository) MarkLeft(ctx context.Context, memberId int64) error {
	query := `
		UPDATE ` + classroomMemberTable + `
		SET member_status = ?,
			left_dt       = ?,
			modify_dt     = ?
		WHERE member_id = ?
	`
	now := mtime.Now().Time
	if _, err := r.db.Exec(ctx, query, enum.ClassroomMemberStatusTypeLeft, now, now, memberId); err != nil {
		return fmt.Errorf("classroom_member repo mark left: %w", err)
	}
	return nil
}

func (r *ClassroomMemberRepository) MarkRemoved(ctx context.Context, memberId, removedByProfileId int64) error {
	query := `
		UPDATE ` + classroomMemberTable + `
		SET member_status         = ?,
			removed_by_profile_id = ?,
			removed_dt            = ?,
			modify_dt             = ?
		WHERE member_id = ?
	`
	now := mtime.Now().Time
	if _, err := r.db.Exec(ctx, query,
		enum.ClassroomMemberStatusTypeRemoved, removedByProfileId, now, now, memberId); err != nil {
		return fmt.Errorf("classroom_member repo mark removed: %w", err)
	}
	return nil
}

// Reactivate flips an existing row's lifecycle fields back to a fresh
// ACTIVE membership. invitationId is kept on the signature so historical
// callers compile, but is no longer written by the new flow (the
// legacy ma_classroom_invitations link is unused).
func (r *ClassroomMemberRepository) Reactivate(ctx context.Context, memberId int64, role string, invitationId *int64) error {
	query := `
		UPDATE ` + classroomMemberTable + `
		SET member_role           = ?,
			member_status         = ?,
			invitation_id         = ?,
			joined_dt             = ?,
			left_dt               = NULL,
			removed_by_profile_id = NULL,
			removed_dt            = NULL,
			modify_dt             = ?
		WHERE member_id = ?
	`
	now := mtime.Now().Time
	if _, err := r.db.Exec(ctx, query,
		role, enum.ClassroomMemberStatusTypeActive, invitationId, now, now, memberId); err != nil {
		return fmt.Errorf("classroom_member repo reactivate: %w", err)
	}
	return nil
}

// Invite refreshes a row into a PENDING_INVITATION in place — used
// when the send-invitation path finds an existing row in a terminal
// state (REJECTED/LEFT/REMOVED) for the (classroom, profile) pair.
// The member_role and inviter metadata are rewritten so the new
// invitation is indistinguishable from one inserted fresh.
func (r *ClassroomMemberRepository) Invite(ctx context.Context, memberId int64, role string, inviteBy *int64, inviteDt mtime.MathTime, note *string) error {
	var inviteDtArg any
	if inviteDt.IsValid() {
		inviteDtArg = inviteDt.Time
	} else {
		inviteDtArg = mtime.Now().Time
	}
	query := `
		UPDATE ` + classroomMemberTable + `
		SET member_role           = ?,
			member_status         = ?,
			invite_by             = ?,
			invite_dt             = ?,
			note                  = COALESCE(?, note),
			joined_dt             = NULL,
			left_dt               = NULL,
			removed_by_profile_id = NULL,
			removed_dt            = NULL,
			modify_dt             = ?
		WHERE member_id = ?
	`
	if _, err := r.db.Exec(ctx, query,
		role, enum.ClassroomMemberStatusTypePendingInvitation,
		inviteBy, inviteDtArg, note,
		mtime.Now().Time, memberId); err != nil {
		return fmt.Errorf("classroom_member repo invite: %w", err)
	}
	return nil
}

// RequestToJoin refreshes a row into a PENDING_REQUEST in place — the
// user-initiated counterpart to Invite. invite_by is explicitly
// nulled (no manager is sponsoring) and invite_dt carries the
// "requested at" timestamp. Used by the join-by-code path when an
// existing terminal-state row is reactivated.
func (r *ClassroomMemberRepository) RequestToJoin(ctx context.Context, memberId int64, role string, requestedDt mtime.MathTime, note *string) error {
	var requestedDtArg any
	if requestedDt.IsValid() {
		requestedDtArg = requestedDt.Time
	} else {
		requestedDtArg = mtime.Now().Time
	}
	query := `
		UPDATE ` + classroomMemberTable + `
		SET member_role           = ?,
			member_status         = ?,
			invite_by             = NULL,
			invite_dt             = ?,
			note                  = COALESCE(?, note),
			joined_dt             = NULL,
			left_dt               = NULL,
			removed_by_profile_id = NULL,
			removed_dt            = NULL,
			modify_dt             = ?
		WHERE member_id = ?
	`
	if _, err := r.db.Exec(ctx, query,
		role, enum.ClassroomMemberStatusTypePendingRequest,
		requestedDtArg, note,
		mtime.Now().Time, memberId); err != nil {
		return fmt.Errorf("classroom_member repo request to join: %w", err)
	}
	return nil
}

// Activate flips a PENDING_* row to ACTIVE, refreshes joined_dt, and
// clears terminal-state markers. Shared by accept-invitation and
// approve-request paths; the command-layer guard decides which prior
// status is valid. Caller pairs with Classroom.IncCounts inside the
// same UoW so the classroom counters stay aligned.
func (r *ClassroomMemberRepository) Activate(ctx context.Context, memberId int64) error {
	query := `
		UPDATE ` + classroomMemberTable + `
		SET member_status         = ?,
			joined_dt             = ?,
			left_dt               = NULL,
			removed_by_profile_id = NULL,
			removed_dt            = NULL,
			modify_dt             = ?
		WHERE member_id = ?
	`
	now := mtime.Now().Time
	if _, err := r.db.Exec(ctx, query,
		enum.ClassroomMemberStatusTypeActive, now, now, memberId); err != nil {
		return fmt.Errorf("classroom_member repo activate: %w", err)
	}
	return nil
}

// Reject flips a PENDING_* row to REJECTED. left_dt is reused as the
// "responded at" timestamp so the schema doesn't need a new column.
func (r *ClassroomMemberRepository) Reject(ctx context.Context, memberId int64) error {
	query := `
		UPDATE ` + classroomMemberTable + `
		SET member_status = ?,
			left_dt       = ?,
			modify_dt     = ?
		WHERE member_id = ?
	`
	now := mtime.Now().Time
	if _, err := r.db.Exec(ctx, query,
		enum.ClassroomMemberStatusTypeRejected, now, now, memberId); err != nil {
		return fmt.Errorf("classroom_member repo reject: %w", err)
	}
	return nil
}

// Cancel flips PENDING → REMOVED with the manager's profile id in
// removed_by_profile_id. Distinguishable from a regular member-removal
// only by the prior state (PENDING vs ACTIVE) — UI labels it
// "invitation cancelled" when it sees a removed PENDING row.
func (r *ClassroomMemberRepository) Cancel(ctx context.Context, memberId, cancelledByProfileId int64) error {
	query := `
		UPDATE ` + classroomMemberTable + `
		SET member_status         = ?,
			removed_by_profile_id = ?,
			removed_dt            = ?,
			modify_dt             = ?
		WHERE member_id = ?
	`
	now := mtime.Now().Time
	if _, err := r.db.Exec(ctx, query,
		enum.ClassroomMemberStatusTypeRemoved, cancelledByProfileId, now, now, memberId); err != nil {
		return fmt.Errorf("classroom_member repo cancel: %w", err)
	}
	return nil
}

func (r *ClassroomMemberRepository) SoftDeleteByClassroomId(ctx context.Context, classroomId int64) error {
	query := `
		UPDATE ` + classroomMemberTable + `
		SET member_status = ?,
			status        = ?,
			deleted_dt    = ?,
			modify_dt     = ?
		WHERE classroom_id = ?
	`
	now := mtime.Now().Time
	if _, err := r.db.Exec(ctx, query,
		enum.ClassroomMemberStatusTypeDeleted, enum.StatusInactive, now, now, classroomId); err != nil {
		return fmt.Errorf("classroom_member repo soft delete by classroom: %w", err)
	}
	return nil
}

func (r *ClassroomMemberRepository) ForceDeleteByClassroomId(ctx context.Context, classroomId int64) error {
	query := `DELETE FROM ` + classroomMemberTable + ` WHERE classroom_id = ?`
	if _, err := r.db.Exec(ctx, query, classroomId); err != nil {
		return fmt.Errorf("classroom_member repo force delete by classroom: %w", err)
	}
	return nil
}

func ModelToDomainClassroomMember(m *models.ClassroomMemberModel) *classroom.Member {
	d := classroom.NewMember()
	d.SetId(m.Id)
	d.SetMemberId(m.MemberId)
	d.SetClassroomId(m.ClassroomId)
	d.SetProfileId(m.ProfileId)
	d.SetMemberRole(m.MemberRole)
	d.SetInvitationId(m.InvitationId)
	if m.JoinedDt != nil {
		d.SetJoinedDt(mtime.MathTime{Time: *m.JoinedDt})
	}
	if m.LeftDt != nil {
		d.SetLeftDt(mtime.MathTime{Time: *m.LeftDt})
	}
	d.SetRemovedByProfileId(m.RemovedByProfileId)
	if m.RemovedDt != nil {
		d.SetRemovedDt(mtime.MathTime{Time: *m.RemovedDt})
	}
	if m.LastSeenDt != nil {
		d.SetLastSeenDt(mtime.MathTime{Time: *m.LastSeenDt})
	}
	d.SetNote(m.Note)
	d.SetInviteBy(m.InviteBy)
	if m.InviteDt != nil {
		d.SetInviteDt(mtime.MathTime{Time: *m.InviteDt})
	}
	d.SetMemberStatus(m.MemberStatus)
	d.SetStatus(m.Status)
	d.SetCreateId(m.CreateId)
	d.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	d.SetModifyId(m.ModifyId)
	d.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return d
}
