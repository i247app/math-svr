package transaction

import (
	"context"

	"math-ai.com/math-ai/internal/domain/banner"
	"math-ai.com/math-ai/internal/domain/chapter"
	"math-ai.com/math-ai/internal/domain/classroom"
	"math-ai.com/math-ai/internal/domain/device"
	"math-ai.com/math-ai/internal/domain/exercise"
	"math-ai.com/math-ai/internal/domain/grade"
	"math-ai.com/math-ai/internal/domain/loginlog"
	"math-ai.com/math-ai/internal/domain/notification"
	"math-ai.com/math-ai/internal/domain/otp"
	"math-ai.com/math-ai/internal/domain/profile"
	"math-ai.com/math-ai/internal/domain/program"
	"math-ai.com/math-ai/internal/domain/quiz"
	"math-ai.com/math-ai/internal/domain/school"
	"math-ai.com/math-ai/internal/domain/semester"
	"math-ai.com/math-ai/internal/domain/seq"
	"math-ai.com/math-ai/internal/domain/user"
)

// Repositories is the bundle of aggregate repositories bound to a single
// transaction, handed to callers of UnitOfWork.Do.
//
// Seq is the central ma_seqs sequence registry — every command that
// mints a new external ID should call Seq.Next(ctx, seq.NameX) instead
// of generating a UUID, so concurrent inserts share one atomic counter.
type Repositories struct {
	User                user.IRepository
	Alias               user.IAliasRepository
	Profile             profile.IRepository
	LoginLog            loginlog.IRepository
	Device              device.IRepository
	Otp                 otp.IRepository
	Quiz                quiz.IRepository
	Chapter             chapter.IRepository
	ChapterTranslation  chapter.ITranslationRepository
	Grade               grade.IRepository
	Semester            semester.IRepository
	Program             program.IRepository
	School              school.IRepository
	Seq                 seq.IRepository
	Classroom           classroom.IRepository
	ClassroomMember     classroom.IMemberRepository
	ClassroomProgram    classroom.IClassroomProgramRepository
	Exercise            exercise.IRepository
	ExerciseSubmission  exercise.ISubmissionRepository
	Notification        notification.IRepository
	Banner              banner.IRepository
}

// UnitOfWork runs fn inside a transaction, committing on nil error and
// rolling back otherwise. Implementations live in the infrastructure layer.
type UnitOfWork interface {
	Do(ctx context.Context, fn func(ctx context.Context, repos Repositories) error) error
}
