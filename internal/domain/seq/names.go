package seq

// Sequence names — must match the seed rows in
// migrations/000_ma_seqs_table.sql. Callers reference these constants
// instead of magic strings so a rename triggers a compile error rather
// than a runtime "sequence not found".
const (
	NameUser                        = "user"
	NameAlias                       = "alias"
	NameDevice                      = "device"
	NameLoginLog                    = "login_log"
	NameProfile                     = "profile"
	NameOtp                         = "otp"
	NameProgram                     = "program"
	NameGrade                       = "grade"
	NameSemester                    = "semester"
	NameChapter                     = "chapter"
	NameChapterTranslation          = "chapter_translation"
	NameQuiz                        = "quiz"
	NameSchool                      = "school"
	NameClassroom                   = "classroom"
	NameClassroomMember             = "classroom_member"
	NameClassroomProgram            = "classroom_program"
	NameClassroomExercise           = "classroom_exercise"
	NameClassroomExerciseSubmission = "classroom_exercise_submission"
	NameNotification                = "notification"
	NameBanner                      = "banner"
	NameChatConversation            = "chat_conversation"
	NameChatParticipant             = "chat_participant"
	NameChatMessage                 = "chat_message"
)
