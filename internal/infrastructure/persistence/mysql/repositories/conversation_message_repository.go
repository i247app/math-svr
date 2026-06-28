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
)

const (
	messageTable = "ma_ai_messages"

	messageColumns = `m.id, m.message_id, m.conversation_id, m.role, m.content,
		m.seq_no, m.note, m.status,
		m.create_id, m.create_dt, m.modify_id, m.modify_dt`

	messageActiveWhere = `m.status = ? AND m.deleted_dt IS NULL`
)

func messageActiveArgs() []any {
	return []any{enum.StatusActive}
}

type ConversationMessageRepository struct {
	db database.Executor
}

func NewConversationMessageRepository(db database.Executor) conversation.IMessageRepository {
	return &ConversationMessageRepository{db: db}
}

func scanMessage(s database.RowScanner) (*models.MessageModel, error) {
	var m models.MessageModel
	if err := s.Scan(&m.Id, &m.MessageId, &m.ConversationId, &m.Role, &m.Content,
		&m.SeqNo, &m.Note, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *ConversationMessageRepository) findBareById(ctx context.Context, id int64) (*conversation.Message, error) {
	args := append(messageActiveArgs(), id)
	query := `SELECT ` + messageColumns + ` FROM ` + messageTable + ` m WHERE ` +
		messageActiveWhere + ` AND m.id = ?`

	m, err := scanMessage(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("conversation message repo find bare by id: %w", err)
	}
	return ModelToDomainMessage(m), nil
}

func (r *ConversationMessageRepository) Create(ctx context.Context, msg *conversation.Message) (*conversation.Message, error) {
	query := `
		INSERT INTO ` + messageTable + `
			(message_id, conversation_id, role, content, seq_no, note, create_id, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := mtime.Now().Time
	result, err := r.db.Exec(ctx, query,
		msg.MessageId(), msg.ConversationId(), msg.Role(), msg.Content(),
		msg.SeqNo(), msg.Note(), msg.CreateId(), now, now)
	if err != nil {
		return nil, fmt.Errorf("conversation message repo create: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("conversation message repo last insert id: %w", err)
	}
	return r.findBareById(ctx, id)
}

// ListRecentByConversationId fetches at most `limit` of the most recent
// messages (seq_no DESC) then reverses them to ascending order so the
// caller can append them to a prompt chronologically. A non-positive limit
// returns nil (history window disabled at the caller's discretion).
func (r *ConversationMessageRepository) ListRecentByConversationId(ctx context.Context, conversationId int64, limit int64) ([]*conversation.Message, error) {
	if limit <= 0 {
		return nil, nil
	}

	args := append(messageActiveArgs(), conversationId, limit)
	query := `SELECT ` + messageColumns + ` FROM ` + messageTable + ` m WHERE ` +
		messageActiveWhere + ` AND m.conversation_id = ?` +
		` ORDER BY m.seq_no DESC LIMIT ?`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("conversation message repo list recent: %w", err)
	}
	defer rows.Close()

	var messages []*conversation.Message
	for rows.Next() {
		m, err := scanMessage(rows)
		if err != nil {
			return nil, fmt.Errorf("conversation message repo scan row: %w", err)
		}
		messages = append(messages, ModelToDomainMessage(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("conversation message repo rows iteration: %w", err)
	}

	// Reverse DESC → ASC (chronological) in place.
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, nil
}

func ModelToDomainMessage(m *models.MessageModel) *conversation.Message {
	msg := conversation.NewMessage()
	msg.SetId(m.Id)
	msg.SetMessageId(m.MessageId)
	msg.SetConversationId(m.ConversationId)
	msg.SetRole(m.Role)
	msg.SetContent(m.Content)
	msg.SetSeqNo(m.SeqNo)
	msg.SetNote(m.Note)
	msg.SetStatus(m.Status)
	msg.SetCreateId(m.CreateId)
	msg.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	msg.SetModifyId(m.ModifyId)
	msg.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return msg
}
