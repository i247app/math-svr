package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"math-ai.com/math-ai/internal/domain/chat"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
)

const (
	chatMessageTable = "ma_chat_messages"

	chatMessageColumns = `m.id, m.message_id, m.conversation_id, m.seq_no,
		m.sender_profile_id, m.sender_user_id, m.message_type, m.content,
		m.attachment_count, m.reply_to_message_id, m.system_event, m.system_payload,
		m.metadata, m.client_msg_id, m.sent_dt, m.edited_dt, m.revoked_dt,
		m.note, m.message_status, m.status, m.create_id, m.create_dt,
		m.modify_id, m.modify_dt`

	// REVOKED rows deliberately survive this filter: the client renders them
	// in place as "tin nhắn đã được thu hồi", and hiding them would reopen the
	// gap in seq_no that the thread view relies on being contiguous.
	chatMessageActiveWhere = `m.status = ? AND m.deleted_dt IS NULL
		AND (m.message_status IS NULL OR m.message_status != ?)`

	// defaultMessagePageSize caps a thread page when the caller passes none.
	defaultMessagePageSize int64 = 30
	maxMessagePageSize     int64 = 100
)

func chatMessageActiveArgs() []any {
	return []any{enum.StatusActive, enum.ChatMessageStatusDeleted}
}

type ChatMessageRepository struct {
	db database.Executor
}

func NewChatMessageRepository(db database.Executor) chat.IMessageRepository {
	return &ChatMessageRepository{db: db}
}

func scanChatMessage(s database.RowScanner) (*models.ChatMessageModel, error) {
	var m models.ChatMessageModel
	if err := s.Scan(&m.Id, &m.MessageId, &m.ConversationId, &m.SeqNo,
		&m.SenderProfileId, &m.SenderUserId, &m.MessageType, &m.Content,
		&m.AttachmentCount, &m.ReplyToMessageId, &m.SystemEvent, &m.SystemPayload,
		&m.Metadata, &m.ClientMsgId, &m.SentDt, &m.EditedDt, &m.RevokedDt,
		&m.Note, &m.MessageStatus, &m.Status, &m.CreateId, &m.CreateDt,
		&m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

func ModelToDomainChatMessage(m *models.ChatMessageModel) *chat.Message {
	msg := chat.NewMessage()
	msg.SetId(m.Id)
	msg.SetMessageId(m.MessageId)
	msg.SetConversationId(m.ConversationId)
	msg.SetSeqNo(m.SeqNo)
	msg.SetSenderProfileId(m.SenderProfileId)
	msg.SetSenderUserId(m.SenderUserId)
	msg.SetMessageType(m.MessageType)
	msg.SetContent(m.Content)
	msg.SetAttachmentCount(m.AttachmentCount)
	msg.SetReplyToMessageId(m.ReplyToMessageId)
	msg.SetSystemEvent(m.SystemEvent)
	msg.SetSystemPayload(m.SystemPayload)
	msg.SetMetadata(m.Metadata)
	msg.SetClientMsgId(m.ClientMsgId)
	msg.SetSentDt(mtime.MathTime{Time: m.SentDt})
	if m.EditedDt != nil {
		msg.SetEditedDt(mtime.MathTime{Time: *m.EditedDt})
	}
	if m.RevokedDt != nil {
		msg.SetRevokedDt(mtime.MathTime{Time: *m.RevokedDt})
	}
	msg.SetNote(m.Note)
	msg.SetMessageStatus(m.MessageStatus)
	msg.SetStatus(m.Status)
	msg.SetCreateId(m.CreateId)
	msg.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	msg.SetModifyId(m.ModifyId)
	msg.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return msg
}

func (r *ChatMessageRepository) findOneBy(ctx context.Context, where string, args ...any) (*chat.Message, error) {
	fullArgs := append(chatMessageActiveArgs(), args...)
	query := `SELECT ` + chatMessageColumns + ` FROM ` + chatMessageTable + ` m WHERE ` +
		chatMessageActiveWhere + ` AND (` + where + `)`

	m, err := scanChatMessage(r.db.QueryRow(ctx, query, fullArgs...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("chat message repo find (%s): %w", where, err)
	}
	return ModelToDomainChatMessage(m), nil
}

func (r *ChatMessageRepository) FindByMessageId(ctx context.Context, messageId int64) (*chat.Message, error) {
	return r.findOneBy(ctx, "m.message_id = ?", messageId)
}

func (r *ChatMessageRepository) FindByClientMsgId(ctx context.Context, conversationId, senderProfileId int64, clientMsgId string) (*chat.Message, error) {
	return r.findOneBy(ctx,
		"m.conversation_id = ? AND m.sender_profile_id = ? AND m.client_msg_id = ?",
		conversationId, senderProfileId, clientMsgId)
}

// ListByConversationId pages on seq_no in both directions.
//
// BeforeSeqNo walks backwards for scroll-up history; AfterSeqNo replays
// forward for reconnect backfill. Neither uses OFFSET, which would shift under
// the reader as new messages arrive. Both orderings are served by the existing
// UNIQUE (conversation_id, seq_no) index — MySQL 8 walks a B-tree backwards as
// cheaply as forwards, so no DESC index is needed.
//
// Results are always returned oldest-first so the caller never has to reverse.
func (r *ChatMessageRepository) ListByConversationId(ctx context.Context, params *chat.ListMessagesParams) ([]*chat.Message, error) {
	if params == nil {
		return nil, fmt.Errorf("chat message repo list: params is required")
	}

	limit := params.Limit
	if limit <= 0 {
		limit = defaultMessagePageSize
	}
	if limit > maxMessagePageSize {
		limit = maxMessagePageSize
	}

	// cleared_before_seq_no is the caller's own "clear history" watermark, so
	// one side clearing their copy hides nothing from the other participants.
	args := append(chatMessageActiveArgs(), params.ConversationId, params.ClearedBeforeSeqNo)
	where := `m.conversation_id = ? AND m.seq_no > ?`
	order := ` ORDER BY m.seq_no ASC`
	needsReverse := false

	switch {
	case params.AfterSeqNo != nil:
		where += ` AND m.seq_no > ?`
		args = append(args, *params.AfterSeqNo)
	case params.BeforeSeqNo != nil:
		where += ` AND m.seq_no < ?`
		args = append(args, *params.BeforeSeqNo)
		// Take the newest rows below the cursor, then flip to oldest-first.
		order = ` ORDER BY m.seq_no DESC`
		needsReverse = true
	default:
		// No cursor: the thread's most recent page.
		order = ` ORDER BY m.seq_no DESC`
		needsReverse = true
	}

	query := `SELECT ` + chatMessageColumns + ` FROM ` + chatMessageTable + ` m WHERE ` +
		chatMessageActiveWhere + ` AND (` + where + `)` + order + ` LIMIT ?`
	args = append(args, limit)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("chat message repo list: %w", err)
	}
	defer rows.Close()

	var out []*chat.Message
	for rows.Next() {
		m, err := scanChatMessage(rows)
		if err != nil {
			return nil, fmt.Errorf("chat message repo scan: %w", err)
		}
		out = append(out, ModelToDomainChatMessage(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("chat message repo list rows: %w", err)
	}

	if needsReverse {
		for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
			out[i], out[j] = out[j], out[i]
		}
	}
	return out, nil
}

func (r *ChatMessageRepository) Create(ctx context.Context, m *chat.Message) (*chat.Message, error) {
	query := `INSERT INTO ` + chatMessageTable + `
		  (message_id, conversation_id, seq_no, sender_profile_id, sender_user_id,
		   message_type, content, attachment_count, reply_to_message_id,
		   system_event, system_payload, metadata, client_msg_id, sent_dt,
		   note, message_status, status, create_id, create_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	res, err := r.db.Exec(ctx, query,
		m.MessageId(), m.ConversationId(), m.SeqNo(), m.SenderProfileId(), m.SenderUserId(),
		m.MessageType(), m.Content(), m.AttachmentCount(), m.ReplyToMessageId(),
		m.SystemEvent(), m.SystemPayload(), m.Metadata(), m.ClientMsgId(), m.SentDt().Time,
		m.Note(), m.MessageStatus(), m.Status(), m.CreateId(), mtime.Now().Time,
	)
	if err != nil {
		// Translated to a domain sentinel so the command layer can recognise a
		// retried send and return the message it already stored.
		if isDuplicateEntry(err) {
			return nil, fmt.Errorf("chat message repo create: %w", chat.ErrDuplicateMessage)
		}
		return nil, fmt.Errorf("chat message repo create: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("chat message repo create last id: %w", err)
	}
	m.SetId(id)
	return r.FindByMessageId(ctx, m.MessageId())
}

func (r *ChatMessageRepository) SoftDeleteByMessageId(ctx context.Context, messageId int64) error {
	query := `UPDATE ` + chatMessageTable + `
		SET message_status = ?, deleted_dt = ? WHERE message_id = ?`

	if _, err := r.db.Exec(ctx, query, enum.ChatMessageStatusDeleted, mtime.Now().Time, messageId); err != nil {
		return fmt.Errorf("chat message repo soft delete: %w", err)
	}
	return nil
}
