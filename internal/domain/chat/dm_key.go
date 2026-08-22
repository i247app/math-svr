// Package chat models realtime messaging: threads (conversations), their
// membership (participants) and the messages themselves.
package chat

import "fmt"

// BuildDmKey produces the deterministic identity of a 1-1 thread.
//
// Sorting the pair is what makes the key symmetric — A messaging B and B
// messaging A must resolve to the same string, or the two of them end up in
// separate threads that each only show half the conversation. The UNIQUE index
// on ma_chat_conversations.dm_key is what turns that rule into an invariant the
// database enforces, including against two clients that tap "message" at the
// same instant.
//
// The key deliberately does NOT include a classroom id: a direct thread is
// global (decision D2), so two people who share several classrooms hold one
// conversation, and the classroom is only the entry point and the permission
// basis. Changing that would mean rewriting keys for live threads.
func BuildDmKey(profileA, profileB int64) string {
	lo, hi := profileA, profileB
	if lo > hi {
		lo, hi = hi, lo
	}
	return fmt.Sprintf("p:%d:%d", lo, hi)
}
