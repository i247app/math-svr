package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"math-ai.com/math-ai/internal/domain/conversation"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

const (
	conversationTable = "ma_ai_conversations"

	conversationColumns = `c.id, c.conversation_id, c.user_id, c.profile_id,
		c.title, c.purpose, c.message_count, c.note,
		c.conversation_status, c.status,
		c.create_id, c.create_dt, c.modify_id, c.modify_dt`

	// Reads exclude both system-inactive rows and the DELETED business
	// status used by the soft-delete path.
	conversationActiveWhere = `c.status = ? AND c.deleted_dt IS NULL
		AND (c.conversation_status IS NULL OR c.conversation_status != ?)`
)

func conversationActiveArgs() []any {
	return []any{enum.StatusActive, enum.ConversationStatusTypeDeleted}
}

type ConversationRepository struct {
	db database.Executor
}

func NewConversationRepository(db database.Executor) conversation.IRepository {
	return &ConversationRepository{db: db}
}

func scanConversation(s database.RowScanner) (*models.ConversationModel, error) {
	var m models.ConversationModel
	if err := s.Scan(&m.Id, &m.ConversationId, &m.UserId, &m.ProfileId,
		&m.Title, &m.Purpose, &m.MessageCount, &m.Note,
		&m.ConversationStatus, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

// findOneBy is the single-row read helper. conversationActiveWhere is
// prepended so every read excludes soft-deleted and inactive rows.
func (r *ConversationRepository) findOneBy(ctx context.Context, where string, args ...any) (*conversation.Conversation, error) {
	fullArgs := append(conversationActiveArgs(), args...)
	query := `SELECT ` + conversationColumns + ` FROM ` + conversationTable + ` c WHERE ` +
		conversationActiveWhere + ` AND (` + where + `)`

	m, err := scanConversation(r.db.QueryRow(ctx, query, fullArgs...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("conversation repo find (%s): %w", where, err)
	}
	return ModelToDomainConversation(m), nil
}

func (r *ConversationRepository) findBareById(ctx context.Context, id int64) (*conversation.Conversation, error) {
	args := append(conversationActiveArgs(), id)
	query := `SELECT ` + conversationColumns + ` FROM ` + conversationTable + ` c WHERE ` +
		conversationActiveWhere + ` AND c.id = ?`

	m, err := scanConversation(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("conversation repo find bare by id: %w", err)
	}
	return ModelToDomainConversation(m), nil
}

func (r *ConversationRepository) FindByConversationId(ctx context.Context, conversationId int64) (*conversation.Conversation, error) {
	return r.findOneBy(ctx, "c.conversation_id = ?", conversationId)
}

func (r *ConversationRepository) ListByUserId(ctx context.Context, params *conversation.ListConversationsParams) ([]*conversation.Conversation, *pagination.Pagination, error) {
	if params == nil {
		return nil, nil, fmt.Errorf("conversation repo list: params is required")
	}

	countArgs := append(conversationActiveArgs(), params.UserID)
	countQuery := `SELECT COUNT(*) FROM ` + conversationTable + ` c WHERE ` +
		conversationActiveWhere + ` AND c.user_id = ?`

	var total int64
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("conversation repo count: %w", err)
	}

	pg := pagination.NewPagination(params.Page, params.Limit, total)

	listArgs := append(conversationActiveArgs(), params.UserID, pg.Size, pg.Skip)
	query := `SELECT ` + conversationColumns + ` FROM ` + conversationTable + ` c WHERE ` +
		conversationActiveWhere + ` AND c.user_id = ?` +
		` ORDER BY c.modify_dt DESC, c.id DESC LIMIT ? OFFSET ?`

	rows, err := r.db.Query(ctx, query, listArgs...)
	if err != nil {
		return nil, nil, fmt.Errorf("conversation repo list: %w", err)
	}
	defer rows.Close()

	var conversations []*conversation.Conversation
	for rows.Next() {
		m, err := scanConversation(rows)
		if err != nil {
			return nil, nil, fmt.Errorf("conversation repo scan row: %w", err)
		}
		conversations = append(conversations, ModelToDomainConversation(m))
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("conversation repo rows iteration: %w", err)
	}
	return conversations, pg, nil
}

func (r *ConversationRepository) Create(ctx context.Context, c *conversation.Conversation) (*conversation.Conversation, error) {
	query := `
		INSERT INTO ` + conversationTable + `
			(conversation_id, user_id, profile_id, title, purpose, message_count, note, conversation_status, create_id, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := mtime.Now().Time
	result, err := r.db.Exec(ctx, query,
		c.ConversationId(), c.UserId(), c.ProfileId(), c.Title(), c.Purpose(),
		c.MessageCount(), c.Note(), c.ConversationStatus(), c.CreateId(), now, now)
	if err != nil {
		return nil, fmt.Errorf("conversation repo create: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("conversation repo last insert id: %w", err)
	}
	return r.findBareById(ctx, id)
}

// IncMessageCount advances message_count atomically (row-locked UPDATE). It
// MUST run inside the same UoW tx as the message insert that consumes the
// new seq_no, otherwise concurrent turns on one conversation could derive
// the same seq_no.
func (r *ConversationRepository) IncMessageCount(ctx context.Context, conversationId int64, delta int64) error {
	query := `
		UPDATE ` + conversationTable + `
		SET message_count = message_count + ?,
			modify_dt     = ?
		WHERE conversation_id = ?
	`
	if _, err := r.db.Exec(ctx, query, delta, mtime.Now().Time, conversationId); err != nil {
		return fmt.Errorf("conversation repo inc message count: %w", err)
	}
	return nil
}

func (r *ConversationRepository) SoftDeleteByConversationId(ctx context.Context, conversationId int64) error {
	query := `
		UPDATE ` + conversationTable + `
		SET conversation_status = ?,
			status              = ?,
			deleted_dt          = ?,
			modify_dt           = ?
		WHERE conversation_id = ?
	`
	now := mtime.Now().Time
	if _, err := r.db.Exec(ctx, query,
		enum.ConversationStatusTypeDeleted, enum.StatusInactive, now, now, conversationId); err != nil {
		return fmt.Errorf("conversation repo soft delete: %w", err)
	}
	return nil
}

func ModelToDomainConversation(m *models.ConversationModel) *conversation.Conversation {
	c := conversation.NewConversation()
	c.SetId(m.Id)
	c.SetConversationId(m.ConversationId)
	c.SetUserId(m.UserId)
	c.SetProfileId(m.ProfileId)
	c.SetTitle(m.Title)
	c.SetPurpose(m.Purpose)
	c.SetMessageCount(m.MessageCount)
	c.SetNote(m.Note)
	c.SetConversationStatus(m.ConversationStatus)
	c.SetStatus(m.Status)
	c.SetCreateId(m.CreateId)
	c.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	c.SetModifyId(m.ModifyId)
	c.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return c
}
