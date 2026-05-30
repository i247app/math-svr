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
	classroomInvitationTable = "ma_classroom_invitations"

	classroomInvitationColumns = `i.id, i.invitation_id, i.classroom_id, i.inviter_profile_id,
		i.invited_profile_id, i.invitee_identifier, i.invitee_identifier_type, i.proposed_role,
		i.token, i.message, i.sent_dt, i.expires_dt,
		i.responded_dt, i.response_profile_id, i.cancelled_by_profile_id, i.note,
		i.invitation_status, i.status,
		i.create_id, i.create_dt, i.modify_id, i.modify_dt`

	// classroomInvitationActiveWhere excludes only system-inactive rows.
	// Lifecycle states (PENDING/ACCEPTED/REJECTED/EXPIRED/...) are
	// surfaced to callers; they filter via params.Status.
	classroomInvitationActiveWhere = `i.status = ? AND i.deleted_dt IS NULL`
)

func classroomInvitationActiveArgs() []any {
	return []any{enum.StatusActive}
}

type ClassroomInvitationRepository struct {
	db database.Executor
}

func NewClassroomInvitationRepository(db database.Executor) classroom.IInvitationRepository {
	return &ClassroomInvitationRepository{db: db}
}

func scanClassroomInvitation(s database.RowScanner) (*models.ClassroomInvitationModel, error) {
	var m models.ClassroomInvitationModel
	if err := s.Scan(&m.Id, &m.InvitationId, &m.ClassroomId, &m.InviterProfileId,
		&m.InvitedProfileId, &m.InviteeIdentifier, &m.InviteeIdentifierType, &m.ProposedRole,
		&m.Token, &m.Message, &m.SentDt, &m.ExpiresDt,
		&m.RespondedDt, &m.ResponseProfileId, &m.CancelledByProfileId, &m.Note,
		&m.InvitationStatus, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *ClassroomInvitationRepository) findOneBy(ctx context.Context, where string, args ...any) (*classroom.Invitation, error) {
	fullArgs := append(classroomInvitationActiveArgs(), args...)
	query := `SELECT ` + classroomInvitationColumns + ` FROM ` + classroomInvitationTable + ` i WHERE ` +
		classroomInvitationActiveWhere + ` AND (` + where + `)`

	m, err := scanClassroomInvitation(r.db.QueryRow(ctx, query, fullArgs...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("classroom_invitation repo find (%s): %w", where, err)
	}
	return ModelToDomainClassroomInvitation(m), nil
}

func (r *ClassroomInvitationRepository) findBareById(ctx context.Context, id int64) (*classroom.Invitation, error) {
	args := append(classroomInvitationActiveArgs(), id)
	query := `SELECT ` + classroomInvitationColumns + ` FROM ` + classroomInvitationTable + ` i WHERE ` +
		classroomInvitationActiveWhere + ` AND i.id = ?`

	m, err := scanClassroomInvitation(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("classroom_invitation repo find bare by id: %w", err)
	}
	return ModelToDomainClassroomInvitation(m), nil
}

func (r *ClassroomInvitationRepository) FindByInvitationId(ctx context.Context, invitationId string) (*classroom.Invitation, error) {
	return r.findOneBy(ctx, "i.invitation_id = ?", invitationId)
}

func (r *ClassroomInvitationRepository) FindByToken(ctx context.Context, token string) (*classroom.Invitation, error) {
	return r.findOneBy(ctx, "i.token = ?", token)
}

func (r *ClassroomInvitationRepository) FindPendingByClassroomAndProfile(ctx context.Context, classroomId, profileId string) (*classroom.Invitation, error) {
	return r.findOneBy(ctx,
		"i.classroom_id = ? AND i.invited_profile_id = ? AND i.invitation_status = ?",
		classroomId, profileId, enum.ClassroomInvitationStatusTypePending)
}

func (r *ClassroomInvitationRepository) FindPendingByClassroomAndIdentifier(ctx context.Context, classroomId, identifier string) (*classroom.Invitation, error) {
	return r.findOneBy(ctx,
		"i.classroom_id = ? AND i.invitee_identifier = ? AND i.invitation_status = ?",
		classroomId, identifier, enum.ClassroomInvitationStatusTypePending)
}

func (r *ClassroomInvitationRepository) ListInvitations(ctx context.Context, params *classroom.ListInvitationsParams) ([]*classroom.Invitation, *pagination.Pagination, error) {
	filterWhere, filterArgs := buildClassroomInvitationFilter(params)

	countArgs := append(classroomInvitationActiveArgs(), filterArgs...)
	countQuery := `SELECT COUNT(*) FROM ` + classroomInvitationTable + ` i WHERE ` +
		classroomInvitationActiveWhere + filterWhere

	var total int64
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("classroom_invitation repo count: %w", err)
	}

	listArgs := append(classroomInvitationActiveArgs(), filterArgs...)
	query := `SELECT ` + classroomInvitationColumns + ` FROM ` + classroomInvitationTable + ` i WHERE ` +
		classroomInvitationActiveWhere + filterWhere +
		` ORDER BY i.sent_dt DESC, i.id DESC`

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
		return nil, nil, fmt.Errorf("classroom_invitation repo list: %w", err)
	}
	defer rows.Close()

	var out []*classroom.Invitation
	for rows.Next() {
		m, err := scanClassroomInvitation(rows)
		if err != nil {
			return nil, nil, fmt.Errorf("classroom_invitation repo scan row: %w", err)
		}
		out = append(out, ModelToDomainClassroomInvitation(m))
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("classroom_invitation repo rows iteration: %w", err)
	}
	return out, pg, nil
}

func buildClassroomInvitationFilter(params *classroom.ListInvitationsParams) (string, []any) {
	if params == nil {
		return "", nil
	}
	var (
		clause string
		args   []any
	)
	if params.ClassroomId != nil && strings.TrimSpace(*params.ClassroomId) != "" {
		clause += ` AND i.classroom_id = ?`
		args = append(args, strings.TrimSpace(*params.ClassroomId))
	}
	if params.InviterProfileId != nil && strings.TrimSpace(*params.InviterProfileId) != "" {
		clause += ` AND i.inviter_profile_id = ?`
		args = append(args, strings.TrimSpace(*params.InviterProfileId))
	}
	if params.InvitedProfileId != nil && strings.TrimSpace(*params.InvitedProfileId) != "" {
		clause += ` AND i.invited_profile_id = ?`
		args = append(args, strings.TrimSpace(*params.InvitedProfileId))
	}
	if params.InviteeIdentifier != nil && strings.TrimSpace(*params.InviteeIdentifier) != "" {
		clause += ` AND i.invitee_identifier = ?`
		args = append(args, strings.TrimSpace(*params.InviteeIdentifier))
	}
	if params.Status != nil && strings.TrimSpace(*params.Status) != "" {
		clause += ` AND i.invitation_status = ?`
		args = append(args, strings.TrimSpace(*params.Status))
	}
	return clause, args
}

func (r *ClassroomInvitationRepository) Create(ctx context.Context, inv *classroom.Invitation) (*classroom.Invitation, error) {
	var sentArg, expiresArg, respondedArg any
	if inv.SentDt().IsValid() {
		sentArg = inv.SentDt().Time
	}
	if inv.ExpiresDt().IsValid() {
		expiresArg = inv.ExpiresDt().Time
	}
	if inv.RespondedDt().IsValid() {
		respondedArg = inv.RespondedDt().Time
	}

	query := `
		INSERT INTO ` + classroomInvitationTable + `
			(invitation_id, classroom_id, inviter_profile_id,
			 invited_profile_id, invitee_identifier, invitee_identifier_type, proposed_role,
			 token, message, sent_dt, expires_dt,
			 responded_dt, response_profile_id, cancelled_by_profile_id, note,
			 invitation_status, create_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(ctx, query,
		inv.InvitationId(), inv.ClassroomId(), inv.InviterProfileId(),
		inv.InvitedProfileId(), inv.InviteeIdentifier(), inv.InviteeIdentifierType(), inv.ProposedRole(),
		inv.Token(), inv.Message(), sentArg, expiresArg,
		respondedArg, inv.ResponseProfileId(), inv.CancelledByProfileId(), inv.Note(),
		inv.InvitationStatus(), inv.CreateId())
	if err != nil {
		return nil, fmt.Errorf("classroom_invitation repo create: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("classroom_invitation repo last insert id: %w", err)
	}
	return r.findBareById(ctx, id)
}

func (r *ClassroomInvitationRepository) Update(ctx context.Context, inv *classroom.Invitation) error {
	var expiresArg, respondedArg any
	if inv.ExpiresDt().IsValid() {
		expiresArg = inv.ExpiresDt().Time
	}
	if inv.RespondedDt().IsValid() {
		respondedArg = inv.RespondedDt().Time
	}

	query := `
		UPDATE ` + classroomInvitationTable + `
		SET proposed_role           = COALESCE(?, proposed_role),
			message                 = COALESCE(?, message),
			expires_dt              = COALESCE(?, expires_dt),
			responded_dt            = COALESCE(?, responded_dt),
			response_profile_id     = COALESCE(?, response_profile_id),
			cancelled_by_profile_id = COALESCE(?, cancelled_by_profile_id),
			note                    = COALESCE(?, note),
			invitation_status       = COALESCE(?, invitation_status),
			modify_id               = COALESCE(?, modify_id),
			modify_dt               = ?
		WHERE invitation_id = ?
	`
	var roleArg any
	if inv.ProposedRole() != "" {
		roleArg = inv.ProposedRole()
	}
	if _, err := r.db.Exec(ctx, query,
		roleArg, inv.Message(), expiresArg, respondedArg,
		inv.ResponseProfileId(), inv.CancelledByProfileId(),
		inv.Note(), inv.InvitationStatus(), inv.ModifyId(),
		mtime.Now().Time, inv.InvitationId()); err != nil {
		return fmt.Errorf("classroom_invitation repo update: %w", err)
	}
	return nil
}

func (r *ClassroomInvitationRepository) SetStatus(ctx context.Context, invitationId, newStatus string, respondedByProfileId *string) error {
	query := `
		UPDATE ` + classroomInvitationTable + `
		SET invitation_status   = ?,
			responded_dt        = ?,
			response_profile_id = COALESCE(?, response_profile_id),
			modify_dt           = ?
		WHERE invitation_id = ?
	`
	now := mtime.Now().Time
	if _, err := r.db.Exec(ctx, query, newStatus, now, respondedByProfileId, now, invitationId); err != nil {
		return fmt.Errorf("classroom_invitation repo set status: %w", err)
	}
	return nil
}

// ExpirePending is called by the reaper job. It only touches rows whose
// expires_dt is set and already past — null expiries are intentionally
// long-lived invitations (e.g. share-with-class link).
func (r *ClassroomInvitationRepository) ExpirePending(ctx context.Context) (int64, error) {
	query := `
		UPDATE ` + classroomInvitationTable + `
		SET invitation_status = ?,
			modify_dt         = ?
		WHERE invitation_status = ?
		  AND expires_dt IS NOT NULL
		  AND expires_dt < ?
	`
	now := mtime.Now().Time
	res, err := r.db.Exec(ctx, query,
		enum.ClassroomInvitationStatusTypeExpired, now,
		enum.ClassroomInvitationStatusTypePending, now)
	if err != nil {
		return 0, fmt.Errorf("classroom_invitation repo expire pending: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("classroom_invitation repo rows affected: %w", err)
	}
	return n, nil
}

func (r *ClassroomInvitationRepository) SoftDeleteByClassroomId(ctx context.Context, classroomId string) error {
	query := `
		UPDATE ` + classroomInvitationTable + `
		SET status     = ?,
			deleted_dt = ?,
			modify_dt  = ?
		WHERE classroom_id = ?
	`
	now := mtime.Now().Time
	if _, err := r.db.Exec(ctx, query, enum.StatusInactive, now, now, classroomId); err != nil {
		return fmt.Errorf("classroom_invitation repo soft delete by classroom: %w", err)
	}
	return nil
}

func (r *ClassroomInvitationRepository) ForceDeleteByClassroomId(ctx context.Context, classroomId string) error {
	query := `DELETE FROM ` + classroomInvitationTable + ` WHERE classroom_id = ?`
	if _, err := r.db.Exec(ctx, query, classroomId); err != nil {
		return fmt.Errorf("classroom_invitation repo force delete by classroom: %w", err)
	}
	return nil
}

func ModelToDomainClassroomInvitation(m *models.ClassroomInvitationModel) *classroom.Invitation {
	d := classroom.NewInvitation()
	d.SetId(m.Id)
	d.SetInvitationId(m.InvitationId)
	d.SetClassroomId(m.ClassroomId)
	d.SetInviterProfileId(m.InviterProfileId)
	d.SetInvitedProfileId(m.InvitedProfileId)
	d.SetInviteeIdentifier(m.InviteeIdentifier)
	d.SetInviteeIdentifierType(m.InviteeIdentifierType)
	d.SetProposedRole(m.ProposedRole)
	d.SetToken(m.Token)
	d.SetMessage(m.Message)
	d.SetSentDt(mtime.MathTime{Time: m.SentDt})
	if m.ExpiresDt != nil {
		d.SetExpiresDt(mtime.MathTime{Time: *m.ExpiresDt})
	}
	if m.RespondedDt != nil {
		d.SetRespondedDt(mtime.MathTime{Time: *m.RespondedDt})
	}
	d.SetResponseProfileId(m.ResponseProfileId)
	d.SetCancelledByProfileId(m.CancelledByProfileId)
	d.SetNote(m.Note)
	d.SetInvitationStatus(m.InvitationStatus)
	d.SetStatus(m.Status)
	d.SetCreateId(m.CreateId)
	d.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	d.SetModifyId(m.ModifyId)
	d.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return d
}
