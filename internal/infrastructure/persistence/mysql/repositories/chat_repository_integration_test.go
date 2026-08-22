package repositories

import (
	"context"
	"errors"
	"sync"
	"testing"

	"math-ai.com/math-ai/internal/domain/chat"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/shared/enum"
)

// newTestConversation inserts a DIRECT thread and returns its external id.
func newTestConversation(t *testing.T, db database.SqlHandler, conversationID int64, dmKey *string) {
	t.Helper()

	active := string(enum.ChatConversationStatusActive)
	c := chat.NewConversation()
	c.SetConversationId(conversationID)
	c.SetConversationType(string(enum.ChatConversationTypeDirect))
	c.SetDmKey(dmKey)
	c.SetParticipantCount(2)
	c.SetConversationStatus(&active)
	c.SetStatus(string(enum.StatusActive))

	inTx(t, db, func(ctx context.Context, ex database.Executor) error {
		_, err := NewChatConversationRepository(ex).Create(ctx, c)
		return err
	})
}

// TestIntegrationNextSeqNoIsUniqueUnderConcurrency is the test this whole
// feature rests on.
//
// seq_no drives ordering, pagination, unread math and reconnect backfill. If
// two concurrent senders can obtain the same number, messages silently
// overwrite each other's position in the thread. The allocator's correctness
// depends on the increment and the read-back sharing one transaction
// connection — a property no unit test with a fake can demonstrate.
func TestIntegrationNextSeqNoIsUniqueUnderConcurrency(t *testing.T) {
	db := openTestDB(t)

	conversationID := uniqueTestID(1)
	cleanupChatRows(t, db, conversationID)
	newTestConversation(t, db, conversationID, nil)

	const senders = 20

	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		results  []int64
		failures []error
	)

	for i := 0; i < senders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var seqNo int64
			err := runInTx(db, func(ctx context.Context, ex database.Executor) error {
				var err error
				seqNo, err = NewChatConversationRepository(ex).NextSeqNo(ctx, conversationID)
				return err
			})
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				failures = append(failures, err)
				return
			}
			results = append(results, seqNo)
		}()
	}
	wg.Wait()

	if len(failures) > 0 {
		t.Fatalf("%d of %d allocations failed, first: %v", len(failures), senders, failures[0])
	}
	if len(results) != senders {
		t.Fatalf("got %d sequence numbers, want %d", len(results), senders)
	}

	seen := make(map[int64]bool, len(results))
	for _, n := range results {
		if seen[n] {
			t.Fatalf("seq_no %d was handed out twice — the allocator is not atomic", n)
		}
		seen[n] = true
	}
	// Contiguous 1..N: a gap would mean an increment was lost, which breaks
	// the "replay everything after 41" reconnect contract.
	for want := int64(1); want <= senders; want++ {
		if !seen[want] {
			t.Errorf("seq_no %d missing — allocation is not gap-free", want)
		}
	}
}

// TestIntegrationDmKeyRejectsDuplicateThread proves the constraint that stops
// two people who tap "message" at the same instant from ending up in separate
// threads, each seeing half the conversation.
func TestIntegrationDmKeyRejectsDuplicateThread(t *testing.T) {
	db := openTestDB(t)

	first := uniqueTestID(2)
	second := uniqueTestID(3)
	dmKey := chat.BuildDmKey(uniqueTestID(10), uniqueTestID(11))
	cleanupChatRows(t, db, first, second)

	newTestConversation(t, db, first, &dmKey)

	// The second insert carries the same dm_key: this is the loser of the race.
	active := string(enum.ChatConversationStatusActive)
	c := chat.NewConversation()
	c.SetConversationId(second)
	c.SetConversationType(string(enum.ChatConversationTypeDirect))
	c.SetDmKey(&dmKey)
	c.SetParticipantCount(2)
	c.SetConversationStatus(&active)
	c.SetStatus(string(enum.StatusActive))

	err := runInTx(db, func(ctx context.Context, ex database.Executor) error {
		_, err := NewChatConversationRepository(ex).Create(ctx, c)
		return err
	})
	if err == nil {
		t.Fatal("a second conversation with the same dm_key was accepted — the UNIQUE index is missing")
	}
	// The command layer branches on this sentinel to re-read the winner's row;
	// if the translation breaks, the user sees an error instead of their chat.
	if !errors.Is(err, chat.ErrDuplicateConversation) {
		t.Fatalf("error = %v, want it to wrap chat.ErrDuplicateConversation", err)
	}
}

// TestIntegrationMessagePaginationIsStable checks the read path end to end:
// the column list, the scan order, and that cursor paging returns each message
// exactly once even as the thread grows.
func TestIntegrationMessagePaginationIsStable(t *testing.T) {
	db := openTestDB(t)

	conversationID := uniqueTestID(4)
	senderProfileID := uniqueTestID(12)
	cleanupChatRows(t, db, conversationID)
	newTestConversation(t, db, conversationID, nil)

	const total = 12
	for i := 0; i < total; i++ {
		content := "tin nhắn thử nghiệm số " + string(rune('A'+i))
		runOrFail(t, db, func(ctx context.Context, ex database.Executor) error {
			convRepo := NewChatConversationRepository(ex)
			seqNo, err := convRepo.NextSeqNo(ctx, conversationID)
			if err != nil {
				return err
			}

			sent := string(enum.ChatMessageStatusSent)
			m := chat.NewMessage()
			m.SetMessageId(uniqueTestID(100 + int64(i)))
			m.SetConversationId(conversationID)
			m.SetSeqNo(seqNo)
			m.SetSenderProfileId(&senderProfileID)
			m.SetMessageType(string(enum.ChatMessageTypeText))
			m.SetContent(&content)
			m.SetSentDt(mtime.Now())
			m.SetMessageStatus(&sent)
			m.SetStatus(string(enum.StatusActive))

			created, err := NewChatMessageRepository(ex).Create(ctx, m)
			if err != nil {
				return err
			}
			// Round-trip check: a mismatch here means the column list and the
			// scan helper have drifted apart.
			if created == nil || created.Content() == nil || *created.Content() != content {
				return errors.New("message did not round-trip through the repository intact")
			}
			return convRepo.UpdateLastMessage(ctx, conversationID, created, content)
		})
	}

	msgRepo := NewChatMessageRepository(db)
	ctx := context.Background()

	// Newest page first, then walk backwards with the cursor — the exact
	// sequence the client performs when scrolling up.
	page1, err := msgRepo.ListByConversationId(ctx, &chat.ListMessagesParams{
		ConversationId: conversationID, Limit: 5,
	})
	if err != nil {
		t.Fatalf("first page: %v", err)
	}
	if len(page1) != 5 {
		t.Fatalf("first page returned %d messages, want 5", len(page1))
	}
	// Always oldest-first within a page, so the client never has to reverse.
	for i := 1; i < len(page1); i++ {
		if page1[i].SeqNo() <= page1[i-1].SeqNo() {
			t.Fatalf("page is not ascending by seq_no: %d then %d", page1[i-1].SeqNo(), page1[i].SeqNo())
		}
	}
	if page1[len(page1)-1].SeqNo() != total {
		t.Errorf("newest message seq_no = %d, want %d", page1[len(page1)-1].SeqNo(), total)
	}

	cursor := page1[0].SeqNo()
	page2, err := msgRepo.ListByConversationId(ctx, &chat.ListMessagesParams{
		ConversationId: conversationID, BeforeSeqNo: &cursor, Limit: 5,
	})
	if err != nil {
		t.Fatalf("second page: %v", err)
	}
	if len(page2) != 5 {
		t.Fatalf("second page returned %d messages, want 5", len(page2))
	}
	for _, m := range page2 {
		if m.SeqNo() >= cursor {
			t.Errorf("seq_no %d leaked past the cursor %d — pages overlap", m.SeqNo(), cursor)
		}
	}

	// Reconnect backfill: "I have up to 8, send me the rest."
	after := int64(8)
	backfill, err := msgRepo.ListByConversationId(ctx, &chat.ListMessagesParams{
		ConversationId: conversationID, AfterSeqNo: &after, Limit: 50,
	})
	if err != nil {
		t.Fatalf("backfill: %v", err)
	}
	if len(backfill) != total-int(after) {
		t.Fatalf("backfill returned %d messages, want %d", len(backfill), total-int(after))
	}
	if backfill[0].SeqNo() != after+1 {
		t.Errorf("backfill starts at %d, want %d", backfill[0].SeqNo(), after+1)
	}

	// cleared_before_seq_no is one participant's own "clear history" mark.
	cleared, err := msgRepo.ListByConversationId(ctx, &chat.ListMessagesParams{
		ConversationId: conversationID, ClearedBeforeSeqNo: 10, Limit: 50,
	})
	if err != nil {
		t.Fatalf("cleared read: %v", err)
	}
	if len(cleared) != total-10 {
		t.Errorf("cleared read returned %d messages, want %d", len(cleared), total-10)
	}
}
