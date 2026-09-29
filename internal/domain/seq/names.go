package seq

// Sequence names — each is the name of the table whose ids it mints, so a
// row in ma_seqs reads as "the next id for this table". They must match the
// seed rows in migrations/seed/001_ma_seqs.sql (existing databases were
// renamed by sql/prod/squash-2026-09-29/04_rename_seq_names_to_tables.sql).
// Callers reference these constants instead of magic strings so a rename
// triggers a compile error rather than a runtime "sequence not found".
//
// One deliberate exception to "one counter per table": profile_id is minted
// from NameUser, not NameProfile. A user's default profile takes
// profile_id = uid, so every other profile must draw from the same counter
// or it could collide with a future uid. NameProfile exists only so the
// clear-data maintenance path can reset ma_profiles' row; nothing mints from it.
const (
	NameUser                        = "ma_users"
	NameAlias                       = "ma_aliases"
	NameLogin                       = "ma_logins"
	NameDevice                      = "ma_devices"
	NameLoginLog                    = "ma_login_logs"
	NameProfile                     = "ma_profiles"
	NameOtp                         = "ma_otps"
	NameProgram                     = "ma_programs"
	NameGrade                       = "ma_grades"
	NameSemester                    = "ma_semesters"
	NameSchool                      = "ma_schools"
	NameClassroom                   = "ma_classrooms"
	NameClassroomMember             = "ma_classroom_members"
	NameClassroomInviation          = "ma_classroom_invitations"
	NameClassroomProgram            = "ma_classroom_programs"
	NameClassroomExercise           = "ma_exercises"
	NameClassroomExerciseSubmission = "ma_exercise_submissions"
	NameNotification                = "ma_notifications"
	NameBanner                      = "ma_banners"
	NameChatConversation            = "ma_chat_conversations"
	NameChatParticipant             = "ma_chat_participants"
	NameChatMessage                 = "ma_chat_messages"
	NameExamPool                    = "ma_exam_pools"
	NameExamLink                    = "ma_exam_links"
	NameExamSession                 = "ma_exam_sessions"
	NameExamSessionLine             = "ma_exam_session_lines"
)
