package conversation

// Conversation purpose tags (the ma_ai_conversations.purpose column).
//
// A purpose lets a non-chat feature own a long-lived, single-thread
// conversation per subject. QuizTutoring is the per-profile "tutoring
// memory" thread the quiz module reads before generating/grading and
// appends a compact learning summary to after grading — giving the AI
// cross-quiz context without breaking the quiz JSON contract.
const (
	PurposeChat         = "CHAT"
	PurposeQuizTutoring = "QUIZ_TUTORING"
)
