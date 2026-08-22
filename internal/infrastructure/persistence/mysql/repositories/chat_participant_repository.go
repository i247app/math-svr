package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"math-ai.com/math-ai/internal/domain/chat"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
)

const (
	chatParticipantTable = "ma_chat_participants"

	chatParticipantColumns = `p.id, p.participant_id, p.conversation_id, p.profile_id,
		p.user_id, p.participant_role, p.last_read_seq_no, p.last_read_message_id,
		p.last_read_dt, p.last_delivered_seq_no, p.unread_count, p.is_muted,
		p.muted_until_dt, p.is_pinned, p.cleared_before_seq_no, p.joined_dt, p.left_dt,
		p.invited_by_profile_id, p.note, p.participant_status, p.status,
		p.create_id, p.create_dt, p.modify_id, p.modify_dt`

	chatParticipantActiveWhere = `p.status = ? AND p.deleted_dt IS NULL
		AND (p.participant_status IS NULL OR p.participant_status != ?)`
)

func chatParticipantActiveArgs() []any {
	return []any{enum.StatusActive, enum.ChatParticipantStatusDeleted}
}

type ChatParticipantRepository struct {
	db database.Executor
}

func NewChatParticipantRepository(db database.Executor) chat.IParticipantRepository {
	return &ChatParticipantRepository{db: db}
}

func scanChatParticipant(s database.RowScanner) (*models.ChatParticipantModel, error) {
	var m models.ChatParticipantModel
	if err := s.Scan(&m.Id, &m.ParticipantId, &m.ConversationId, &m.ProfileId,
		&m.UserId, &m.ParticipantRole, &m.LastReadSeqNo, &m.LastReadMessageId,
		&m.LastReadDt, &m.LastDeliveredSeqNo, &m.UnreadCount, &m.IsMuted,
		&m.MutedUntilDt, &m.IsPinned, &m.ClearedBeforeSeqNo, &m.JoinedDt, &m.LeftDt,
		&m.InvitedByProfileId, &m.Note, &m.ParticipantStatus, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

func ModelToDomainChatParticipant(m *models.ChatParticipantModel) *chat.Participant {
	p := chat.NewParticipant()
	p.SetId(m.Id)
	p.SetParticipantId(m.ParticipantId)
	p.SetConversationId(m.ConversationId)
	p.SetProfileId(m.ProfileId)
	p.SetUserId(m.UserId)
	p.SetParticipantRole(m.ParticipantRole)
	p.SetLastReadSeqNo(m.LastReadSeqNo)
	p.SetLastReadMessageId(m.LastReadMessageId)
	if m.LastReadDt != nil {
		p.SetLastReadDt(mtime.MathTime{Time: *m.LastReadDt})
	}
	p.SetLastDeliveredSeqNo(m.LastDeliveredSeqNo)
	p.SetUnreadCount(m.UnreadCount)
	p.SetIsMuted(m.IsMuted)
	if m.MutedUntilDt != nil {
		p.SetMutedUntilDt(mtime.MathTime{Time: *m.MutedUntilDt})
	}
	p.SetIsPinned(m.IsPinned)
	p.SetClearedBeforeSeqNo(m.ClearedBeforeSeqNo)
	if m.JoinedDt != nil {
		p.SetJoinedDt(mtime.MathTime{Time: *m.JoinedDt})
	}
	if m.LeftDt != nil {
		p.SetLeftDt(mtime.MathTime{Time: *m.LeftDt})
	}
	p.SetInvitedByProfileId(m.InvitedByProfileId)
	p.SetNote(m.Note)
	p.SetParticipantStatus(m.ParticipantStatus)
	p.SetStatus(m.Status)
	p.SetCreateId(m.CreateId)
	p.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	p.SetModifyId(m.ModifyId)
	p.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return p
}

func (r *ChatParticipantRepository) findOneBy(ctx context.Context, where string, args ...any) (*chat.Participant, error) {
	fullArgs := append(chatParticipantActiveArgs(), args...)
	query := `SELECT ` + chatParticipantColumns + ` FROM ` + chatParticipantTable + ` p WHERE ` +
		chatParticipantActiveWhere + ` AND (` + where + `)`

	m, err := scanChatParticipant(r.db.QueryRow(ctx, query, fullArgs...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("chat participant repo find (%s): %w", where, err)
	}
	return ModelToDomainChatParticipant(m), nil
}

func (r *ChatParticipantRepository) FindByConversationAndProfile(ctx context.Context, conversationId, profileId int64) (*chat.Participant, error) {
	return r.findOneBy(ctx, "p.conversation_id = ? AND p.profile_id = ?", conversationId, profileId)
}

func (r *ChatParticipantRepository) ListByConversationId(ctx context.Context, params *chat.ListParticipantsParams) ([]*chat.Participant, error) {
	if params == nil {
		return nil, fmt.Errorf("chat participant repo list: params is required")
	}

	args := append(chatParticipantActiveArgs(), params.ConversationId)
	where := `p.conversation_id = ?`
	if params.Status != nil {
		where += ` AND p.participant_status = ?`
		args = append(args, *params.Status)
	}

	query := `SELECT ` + chatParticipantColumns + ` FROM ` + chatParticipantTable + ` p WHERE ` +
		chatParticipantActiveWhere + ` AND (` + where + `) ORDER BY p.id ASC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("chat participant repo list: %w", err)
	}
	defer rows.Close()

	var out []*chat.Participant
	for rows.Next() {
		m, err := scanChatParticipant(rows)
		if err != nil {
			return nil, fmt.Errorf("chat participant repo scan: %w", err)
		}
		out = append(out, ModelToDomainChatParticipant(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("chat participant repo list rows: %w", err)
	}
	return out, nil
}

func (r *ChatParticipantRepository) ListByProfileAndConversationIds(ctx context.Context, profileId int64, conversationIds []int64) (map[int64]*chat.Participant, error) {
	out := make(map[int64]*chat.Participant, len(conversationIds))
	if len(conversationIds) == 0 {
		return out, nil
	}

	placeholders := strings.Repeat("?,", len(conversationIds)-1) + "?"
	args := append(chatParticipantActiveArgs(), profileId)
	for _, id := range conversationIds {
		args = append(args, id)
	}

	query := `SELECT ` + chatParticipantColumns + ` FROM ` + chatParticipantTable + ` p WHERE ` +
		chatParticipantActiveWhere + ` AND p.profile_id = ? AND p.conversation_id IN (` + placeholders + `)`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("chat participant repo list by conversation ids: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		m, err := scanChatParticipant(rows)
		if err != nil {
			return nil, fmt.Errorf("chat participant repo scan: %w", err)
		}
		out[m.ConversationId] = ModelToDomainChatParticipant(m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("chat participant repo list rows: %w", err)
	}
	return out, nil
}

func (r *ChatParticipantRepository) Create(ctx context.Context, p *chat.Participant) (*chat.Participant, error) {
	query := `INSERT INTO ` + chatParticipantTable + `
		  (participant_id, conversation_id, profile_id, user_id, participant_role,
		   last_read_seq_no, last_delivered_seq_no, unread_count, is_muted, is_pinned,
		   cleared_before_seq_no, joined_dt, invited_by_profile_id, note,
		   participant_status, status, create_id, create_dt)
		VALUES (?, ?, ?, ?, ?, 0, 0, 0, ?, ?, 0, ?, ?, ?, ?, ?, ?, ?)`

	now := mtime.Now().Time
	joinedDt := p.JoinedDt()
	if joinedDt.Time.IsZero() {
		joinedDt = mtime.Now()
	}

	res, err := r.db.Exec(ctx, query,
		p.ParticipantId(), p.ConversationId(), p.ProfileId(), p.UserId(), p.ParticipantRole(),
		p.IsMuted(), p.IsPinned(), joinedDt.Time, p.InvitedByProfileId(), p.Note(),
		p.ParticipantStatus(), p.Status(), p.CreateId(), now,
	)
	if err != nil {
		return nil, fmt.Errorf("chat participant repo create: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("chat participant repo create last id: %w", err)
	}
	p.SetId(id)
	return r.findOneBy(ctx, "p.participant_id = ?", p.ParticipantId())
}

// IncUnreadExcept fans the count out in one statement. Doing it as a
// read-modify-write per participant would drop counts whenever two people send
// at the same moment.
//
// last_delivered_seq_no uses GREATEST so an out-of-order call cannot rewind it.
func (r *ChatParticipantRepository) IncUnreadExcept(ctx context.Context, conversationId, exceptProfileId, seqNo int64) error {
	query := `UPDATE ` + chatParticipantTable + ` SET
		  unread_count          = unread_count + 1,
		  last_delivered_seq_no = GREATEST(last_delivered_seq_no, ?)
		WHERE conversation_id = ? AND profile_id != ?
		  AND status = ? AND deleted_dt IS NULL AND participant_status = ?`

	if _, err := r.db.Exec(ctx, query, seqNo, conversationId, exceptProfileId,
		enum.StatusActive, enum.ChatParticipantStatusActive); err != nil {
		return fmt.Errorf("chat participant repo inc unread: %w", err)
	}
	return nil
}

// MarkRead is guarded by `last_read_seq_no < ?` so the watermark only ever
// moves forward. A delayed client call for an older message would otherwise
// drag it backwards and make already-read messages unread again. A no-op
// update (the guard not matching) is success, not an error.
func (r *ChatParticipantRepository) MarkRead(ctx context.Context, conversationId, profileId, seqNo int64, messageId *int64, readDt mtime.MathTime) error {
	query := `UPDATE ` + chatParticipantTable + ` SET
		  last_read_seq_no     = ?,
		  last_read_message_id = ?,
		  last_read_dt         = ?,
		  unread_count         = 0
		WHERE conversation_id = ? AND profile_id = ? AND last_read_seq_no < ?`

	if _, err := r.db.Exec(ctx, query, seqNo, messageId, readDt.Time,
		conversationId, profileId, seqNo); err != nil {
		return fmt.Errorf("chat participant repo mark read: %w", err)
	}
	return nil
}

func (r *ChatParticipantRepository) SumUnreadByProfileId(ctx context.Context, profileId int64) (int64, error) {
	// COALESCE because SUM over zero rows is NULL, and a user with no threads
	// must read as 0 rather than fail the scan.
	query := `SELECT COALESCE(SUM(p.unread_count), 0) FROM ` + chatParticipantTable + ` p
		WHERE ` + chatParticipantActiveWhere + ` AND p.profile_id = ? AND p.participant_status = ?`

	args := append(chatParticipantActiveArgs(), profileId, enum.ChatParticipantStatusActive)

	var total int64
	if err := r.db.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("chat participant repo sum unread: %w", err)
	}
	return total, nil
}

func (r *ChatParticipantRepository) MarkLeft(ctx context.Context, participantId int64, leftDt mtime.MathTime) error {
	query := `UPDATE ` + chatParticipantTable + `
		SET participant_status = ?, left_dt = ? WHERE participant_id = ?`

	if _, err := r.db.Exec(ctx, query, enum.ChatParticipantStatusLeft, leftDt.Time, participantId); err != nil {
		return fmt.Errorf("chat participant repo mark left: %w", err)
	}
	return nil
}
