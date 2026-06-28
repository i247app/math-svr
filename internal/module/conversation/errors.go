package conversation

import "errors"

// Module-scoped sentinel errors. See conventions.md §Errors.
var (
	ErrMessageRequired        = errors.New("message is required")
	ErrMessageTooLong         = errors.New("message is too long")
	ErrConversationIDInvalid  = errors.New("conversation_id must be positive")
	ErrConversationNotFound   = errors.New("conversation not found")
	ErrConversationNotOwned   = errors.New("conversation does not belong to the current user")
	ErrBotDisabled            = errors.New("conversation: bot adapter is not configured")
	ErrUidNotFoundFromSession = errors.New("uid not found from session")
)
