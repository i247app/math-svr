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
	"math-ai.com/math-ai/internal/shared/pagination"
)

const (
	chatConversationTable = "ma_chat_conversations"

	chatConversationColumns = `c.id, c.conversation_id, c.conversation_type, c.classroom_id,
		c.dm_key, c.title, c.avatar_key, c.owner_profile_id, c.participant_count,
		c.last_seq_no, c.message_count, c.last_message_id, c.last_message_seq_no,
		c.last_message_type, c.last_message_preview, c.last_message_sender_profile_id,
		c.last_message_dt, c.note, c.conversation_status, c.status,
		c.create_id, c.create_dt, c.modify_id, c.modify_dt`

	chatConversationActiveWhere = `c.status = ? AND c.deleted_dt IS NULL
		AND (c.conversation_status IS NULL OR c.conversation_status != ?)`
)

func chatConversationActiveArgs() []any {
	return []any{enum.StatusActive, enum.ChatConversationStatusDeleted}
}

type ChatConversationRepository struct {
	db database.Executor
}

func NewChatConversationRepository(db database.Executor) chat.IRepository {
	return &ChatConversationRepository{db: db}
}

func scanChatConversation(s database.RowScanner) (*models.ChatConversationModel, error) {
	var m models.ChatConversationModel
	if err := s.Scan(&m.Id, &m.ConversationId, &m.ConversationType, &m.ClassroomId,
		&m.DmKey, &m.Title, &m.AvatarKey, &m.OwnerProfileId, &m.ParticipantCount,
		&m.LastSeqNo, &m.MessageCount, &m.LastMessageId, &m.LastMessageSeqNo,
		&m.LastMessageType, &m.LastMessagePreview, &m.LastMessageSenderProfileId,
		&m.LastMessageDt, &m.Note, &m.ConversationStatus, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

func ModelToDomainChatConversation(m *models.ChatConversationModel) *chat.Conversation {
	c := chat.NewConversation()
	c.SetId(m.Id)
	c.SetConversationId(m.ConversationId)
	c.SetConversationType(m.ConversationType)
	c.SetClassroomId(m.ClassroomId)
	c.SetDmKey(m.DmKey)
	c.SetTitle(m.Title)
	c.SetAvatarKey(m.AvatarKey)
	c.SetOwnerProfileId(m.OwnerProfileId)
	c.SetParticipantCount(m.ParticipantCount)
	c.SetLastSeqNo(m.LastSeqNo)
	c.SetMessageCount(m.MessageCount)
	c.SetLastMessageId(m.LastMessageId)
	c.SetLastMessageSeqNo(m.LastMessageSeqNo)
	c.SetLastMessageType(m.LastMessageType)
	c.SetLastMessagePreview(m.LastMessagePreview)
	c.SetLastMessageSenderProfileId(m.LastMessageSenderProfileId)
	if m.LastMessageDt != nil {
		c.SetLastMessageDt(mtime.MathTime{Time: *m.LastMessageDt})
	}
	c.SetNote(m.Note)
	c.SetConversationStatus(m.ConversationStatus)
	c.SetStatus(m.Status)
	c.SetCreateId(m.CreateId)
	c.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	c.SetModifyId(m.ModifyId)
	c.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return c
}

func (r *ChatConversationRepository) findOneBy(ctx context.Context, where string, args ...any) (*chat.Conversation, error) {
	fullArgs := append(chatConversationActiveArgs(), args...)
	query := `SELECT ` + chatConversationColumns + ` FROM ` + chatConversationTable + ` c WHERE ` +
		chatConversationActiveWhere + ` AND (` + where + `)`

	m, err := scanChatConversation(r.db.QueryRow(ctx, query, fullArgs...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("chat conversation repo find (%s): %w", where, err)
	}
	return ModelToDomainChatConversation(m), nil
}

func (r *ChatConversationRepository) FindByConversationId(ctx context.Context, conversationId int64) (*chat.Conversation, error) {
	return r.findOneBy(ctx, "c.conversation_id = ?", conversationId)
}

func (r *ChatConversationRepository) FindByDmKey(ctx context.Context, dmKey string) (*chat.Conversation, error) {
	return r.findOneBy(ctx, "c.dm_key = ?", dmKey)
}

func (r *ChatConversationRepository) ListByDmKeys(ctx context.Context, dmKeys []string) (map[string]*chat.Conversation, error) {
	out := make(map[string]*chat.Conversation, len(dmKeys))
	if len(dmKeys) == 0 {
		return out, nil
	}

	placeholders := strings.Repeat("?,", len(dmKeys)-1) + "?"
	args := chatConversationActiveArgs()
	for _, k := range dmKeys {
		args = append(args, k)
	}

	query := `SELECT ` + chatConversationColumns + ` FROM ` + chatConversationTable + ` c WHERE ` +
		chatConversationActiveWhere + ` AND c.dm_key IN (` + placeholders + `)`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("chat conversation repo list by dm keys: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		m, err := scanChatConversation(rows)
		if err != nil {
			return nil, fmt.Errorf("chat conversation repo scan: %w", err)
		}
		if m.DmKey != nil {
			out[*m.DmKey] = ModelToDomainChatConversation(m)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("chat conversation repo list by dm keys rows: %w", err)
	}
	return out, nil
}

func (r *ChatConversationRepository) ListByProfileId(ctx context.Context, params *chat.ListConversationsParams) ([]*chat.Conversation, *pagination.Pagination, error) {
	if params == nil {
		return nil, nil, fmt.Errorf("chat conversation repo list: params is required")
	}

	// The join is what scopes the inbox to the caller; the participant row is
	// also where the unread filter lives, so it cannot be a subquery.
	from := chatConversationTable + ` c
		INNER JOIN ma_chat_participants p
		    ON p.conversation_id = c.conversation_id
		   AND p.profile_id = ?
		   AND p.status = ?
		   AND p.deleted_dt IS NULL
		   AND p.participant_status = ?`

	joinArgs := []any{params.ProfileId, enum.StatusActive, enum.ChatParticipantStatusActive}

	where := ""
	filterArgs := []any{}
	if params.ConversationType != nil {
		where += ` AND c.conversation_type = ?`
		filterArgs = append(filterArgs, *params.ConversationType)
	}
	if params.UnreadOnly {
		where += ` AND p.unread_count > 0`
	}

	buildArgs := func() []any {
		args := append([]any{}, joinArgs...)
		args = append(args, chatConversationActiveArgs()...)
		return append(args, filterArgs...)
	}

	var total int64
	countQuery := `SELECT COUNT(*) FROM ` + from + ` WHERE ` + chatConversationActiveWhere + where
	if err := r.db.QueryRow(ctx, countQuery, buildArgs()...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("chat conversation repo count: %w", err)
	}

	// Pinned threads first, then most recent activity. Threads with no message
	// yet (last_message_dt NULL) fall back to create_dt so a freshly opened
	// conversation does not sink to the bottom of the inbox.
	query := `SELECT ` + chatConversationColumns + ` FROM ` + from + ` WHERE ` +
		chatConversationActiveWhere + where +
		` ORDER BY p.is_pinned DESC, COALESCE(c.last_message_dt, c.create_dt) DESC, c.id DESC`

	listArgs := buildArgs()
	var pg *pagination.Pagination
	if !params.TakeAll {
		pg = pagination.NewPagination(params.Page, params.Limit, total)
		query += ` LIMIT ? OFFSET ?`
		listArgs = append(listArgs, pg.Size, pg.Skip)
	}

	rows, err := r.db.Query(ctx, query, listArgs...)
	if err != nil {
		return nil, nil, fmt.Errorf("chat conversation repo list: %w", err)
	}
	defer rows.Close()

	var out []*chat.Conversation
	for rows.Next() {
		m, err := scanChatConversation(rows)
		if err != nil {
			return nil, nil, fmt.Errorf("chat conversation repo scan: %w", err)
		}
		out = append(out, ModelToDomainChatConversation(m))
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("chat conversation repo list rows: %w", err)
	}
	return out, pg, nil
}

func (r *ChatConversationRepository) Create(ctx context.Context, c *chat.Conversation) (*chat.Conversation, error) {
	query := `INSERT INTO ` + chatConversationTable + `
		  (conversation_id, conversation_type, classroom_id, dm_key, title, avatar_key,
		   owner_profile_id, participant_count, last_seq_no, message_count, note,
		   conversation_status, status, create_id, create_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, 0, ?, ?, ?, ?, ?)`

	res, err := r.db.Exec(ctx, query,
		c.ConversationId(), c.ConversationType(), c.ClassroomId(), c.DmKey(),
		c.Title(), c.AvatarKey(), c.OwnerProfileId(), c.ParticipantCount(),
		c.Note(), c.ConversationStatus(), c.Status(), c.CreateId(), mtime.Now().Time,
	)
	if err != nil {
		// Translated to a domain sentinel so the command layer can recognise
		// the lost race and re-select the winner's row without importing the
		// SQL driver.
		if isDuplicateEntry(err) {
			return nil, fmt.Errorf("chat conversation repo create: %w", chat.ErrDuplicateConversation)
		}
		return nil, fmt.Errorf("chat conversation repo create: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("chat conversation repo create last id: %w", err)
	}
	c.SetId(id)
	return r.FindByConversationId(ctx, c.ConversationId())
}

// NextSeqNo increments the conversation's allocator and reads it back. Both
// statements must run on the caller's transaction connection — see the
// interface doc for why routing them separately breaks the guarantee.
func (r *ChatConversationRepository) NextSeqNo(ctx context.Context, conversationId int64) (int64, error) {
	update := `UPDATE ` + chatConversationTable + `
		SET last_seq_no = last_seq_no + 1 WHERE conversation_id = ?`

	res, err := r.db.Exec(ctx, update, conversationId)
	if err != nil {
		return 0, fmt.Errorf("chat conversation repo next seq (update): %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("chat conversation repo next seq (affected): %w", err)
	}
	if affected == 0 {
		return 0, fmt.Errorf("chat conversation repo next seq: conversation %d not found", conversationId)
	}

	var seqNo int64
	read := `SELECT last_seq_no FROM ` + chatConversationTable + ` WHERE conversation_id = ?`
	if err := r.db.QueryRow(ctx, read, conversationId).Scan(&seqNo); err != nil {
		return 0, fmt.Errorf("chat conversation repo next seq (read): %w", err)
	}
	return seqNo, nil
}

func (r *ChatConversationRepository) UpdateLastMessage(ctx context.Context, conversationId int64, m *chat.Message, preview string) error {
	query := `UPDATE ` + chatConversationTable + ` SET
		  message_count                  = message_count + 1,
		  last_message_id                = ?,
		  last_message_seq_no            = ?,
		  last_message_type              = ?,
		  last_message_preview           = ?,
		  last_message_sender_profile_id = ?,
		  last_message_dt                = ?
		WHERE conversation_id = ?`

	if _, err := r.db.Exec(ctx, query,
		m.MessageId(), m.SeqNo(), m.MessageType(), preview,
		m.SenderProfileId(), m.SentDt().Time, conversationId,
	); err != nil {
		return fmt.Errorf("chat conversation repo update last message: %w", err)
	}
	return nil
}

func (r *ChatConversationRepository) SoftDeleteByConversationId(ctx context.Context, conversationId int64) error {
	query := `UPDATE ` + chatConversationTable + `
		SET conversation_status = ?, deleted_dt = ? WHERE conversation_id = ?`

	if _, err := r.db.Exec(ctx, query, enum.ChatConversationStatusDeleted, mtime.Now().Time, conversationId); err != nil {
		return fmt.Errorf("chat conversation repo soft delete: %w", err)
	}
	return nil
}
