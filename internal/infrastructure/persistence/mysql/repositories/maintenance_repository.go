package repositories

import (
	"context"
	"fmt"
	"strings"

	"math-ai.com/math-ai/internal/domain/seq"
	"math-ai.com/math-ai/internal/infrastructure/database"
)

// clearDataTables lists the user-generated tables wiped by ClearData.
// Reference / seed tables (programs, grades, semesters, schools) are
// intentionally excluded so the curriculum data seeded outside the app
// survives. Mirrors sql/clear_data.sql.
var clearDataTables = []string{
	userTable,
	aliasTable,
	deviceTable,
	loginLogTable,
	profileTable,
	otpTable,
	quizTable,
	classroomTable,
	classroomMemberTable,
	// classroomInvitationTable,
	classroomProgramTable,
	exerciseTable,
	exerciseSubmissionTable,
	notificationTable,
	bannerTable,
	chatConversationTable,
	chatParticipantTable,
	chatMessageTable,
}

// clearDataSeqs lists the external-id counters reset back to 0 for the wiped
// aggregates only. Reference sequences (program/grade/semester/school)
// are intentionally left untouched so seeded rows keep their
// ids. Mirrors sql/clear_data.sql.
var clearDataSeqs = []string{
	seq.NameUser,
	seq.NameAlias,
	seq.NameDevice,
	seq.NameLoginLog,
	seq.NameProfile,
	seq.NameOtp,
	seq.NameQuiz,
	seq.NameClassroom,
	seq.NameClassroomMember,
	seq.NameClassroomInviation,
	seq.NameClassroomProgram,
	seq.NameClassroomExercise,
	seq.NameClassroomExerciseSubmission,
	seq.NameNotification,
	seq.NameBanner,
	seq.NameChatConversation,
	seq.NameChatParticipant,
	seq.NameChatMessage,
}

// MaintenanceRepository owns destructive, cross-aggregate maintenance SQL that
// does not belong to any single aggregate repository (e.g. wiping all
// user-generated data). Keeping the raw SQL here honours the rule that no SQL
// leaks into the module / handler layer.
type MaintenanceRepository struct {
	db database.Executor
}

func NewMaintenanceRepository(db database.Executor) *MaintenanceRepository {
	return &MaintenanceRepository{db: db}
}

// ClearData TRUNCATEs every user-generated table and resets the matching
// external-id counters in ma_seqs. It returns the tables cleared and the
// sequences reset so callers can report exactly what was wiped.
//
// The schema declares no foreign keys (integrity is enforced in the
// application layer inside UnitOfWork blocks), so each TRUNCATE is independent
// and order does not matter. TRUNCATE additionally resets each table's internal
// AUTO_INCREMENT id.
func (r *MaintenanceRepository) ClearData(ctx context.Context) ([]string, []string, error) {
	for _, table := range clearDataTables {
		// Table names come from the fixed internal allow-list above, never
		// from user input, so string concatenation is safe here.
		if _, err := r.db.Exec(ctx, "TRUNCATE TABLE "+table); err != nil {
			return nil, nil, fmt.Errorf("maintenance repo clear-data: truncate %s: %w", table, err)
		}
	}

	placeholders := make([]string, len(clearDataSeqs))
	args := make([]any, len(clearDataSeqs))
	for i, name := range clearDataSeqs {
		placeholders[i] = "?"
		args[i] = name
	}
	resetQuery := `UPDATE ` + seqTable + ` SET current_value = 0 WHERE seq_name IN (` +
		strings.Join(placeholders, ", ") + `)`
	if _, err := r.db.Exec(ctx, resetQuery, args...); err != nil {
		return nil, nil, fmt.Errorf("maintenance repo clear-data: reset seqs: %w", err)
	}

	return clearDataTables, clearDataSeqs, nil
}
