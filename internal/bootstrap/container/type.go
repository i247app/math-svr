package container

import (
	chapterDomain "math-ai.com/math-ai/internal/domain/chapter"
	classroomDomain "math-ai.com/math-ai/internal/domain/classroom"
	conversationDomain "math-ai.com/math-ai/internal/domain/conversation"
	deviceDomain "math-ai.com/math-ai/internal/domain/device"
	exerciseDomain "math-ai.com/math-ai/internal/domain/exercise"
	gradeDomain "math-ai.com/math-ai/internal/domain/grade"
	loginLogDomain "math-ai.com/math-ai/internal/domain/loginlog"
	notificationDomain "math-ai.com/math-ai/internal/domain/notification"
	otpDomain "math-ai.com/math-ai/internal/domain/otp"
	profileDomain "math-ai.com/math-ai/internal/domain/profile"
	programDomain "math-ai.com/math-ai/internal/domain/program"
	quizDomain "math-ai.com/math-ai/internal/domain/quiz"
	schoolDomain "math-ai.com/math-ai/internal/domain/school"
	semesterDomain "math-ai.com/math-ai/internal/domain/semester"
	seqDomain "math-ai.com/math-ai/internal/domain/seq"
	userDomain "math-ai.com/math-ai/internal/domain/user"
	"math-ai.com/math-ai/internal/module/auth"
	"math-ai.com/math-ai/internal/module/bot"
	"math-ai.com/math-ai/internal/module/chapter"
	"math-ai.com/math-ai/internal/module/classroom"
	"math-ai.com/math-ai/internal/module/conversation"
	"math-ai.com/math-ai/internal/module/device"
	"math-ai.com/math-ai/internal/module/exercise"
	"math-ai.com/math-ai/internal/module/grade"
	"math-ai.com/math-ai/internal/module/home"
	"math-ai.com/math-ai/internal/module/job"
	"math-ai.com/math-ai/internal/module/misc"
	"math-ai.com/math-ai/internal/module/notification"
	"math-ai.com/math-ai/internal/module/otp"
	"math-ai.com/math-ai/internal/module/profile"
	"math-ai.com/math-ai/internal/module/program"
	"math-ai.com/math-ai/internal/module/quiz"
	"math-ai.com/math-ai/internal/module/school"
	"math-ai.com/math-ai/internal/module/semester"
	"math-ai.com/math-ai/internal/module/seq"
	"math-ai.com/math-ai/internal/module/user"
)

type ServiceContainer struct {
	MiscSvc         *misc.Service
	UserSvc         *user.Service
	AuthSvc         *auth.Service
	ProgramSvc      *program.Service
	GradeSvc        *grade.Service
	SemesterSvc     *semester.Service
	ProfileSvc      *profile.Service
	DeviceSvc       *device.Service
	OtpSvc          *otp.Service
	QuizSvc         *quiz.Service
	ChapterSvc      *chapter.Service
	SchoolSvc       *school.Service
	JobSvc          *job.Service
	SeqSvc          *seq.Service
	ClassroomSvc    *classroom.Service
	ExerciseSvc     *exercise.Service
	HomeSvc         *home.Service
	BotSvc          *bot.Service
	ConversationSvc *conversation.Service
	NotificationSvc *notification.Service
}

type RepositoryContainer struct {
	UserRepository                userDomain.IRepository
	ProgramRepository             programDomain.IRepository
	GradeRepository               gradeDomain.IRepository
	SemesterRepository            semesterDomain.IRepository
	ProfileRepository             profileDomain.IRepository
	LoginLogRepository            loginLogDomain.IRepository
	DeviceRepository              deviceDomain.IRepository
	OtpRepository                 otpDomain.IRepository
	QuizRepository                quizDomain.IRepository
	ChapterRepository             chapterDomain.IRepository
	ChapterTranslationRepository  chapterDomain.ITranslationRepository
	GradeTranslationRepository    gradeDomain.ITranslationRepository
	SemesterTranslationRepository semesterDomain.ITranslationRepository
	ProgramTranslationRepository  programDomain.ITranslationRepository
	SchoolRepository              schoolDomain.IRepository
	SeqRepository                 seqDomain.IRepository
	ClassroomRepository           classroomDomain.IRepository
	ClassroomMemberRepository     classroomDomain.IMemberRepository
	ClassroomProgramRepository    classroomDomain.IClassroomProgramRepository
	ExerciseRepository            exerciseDomain.IRepository
	ExerciseSubmissionRepository  exerciseDomain.ISubmissionRepository
	ConversationRepository        conversationDomain.IRepository
	ConversationMessageRepository conversationDomain.IMessageRepository
	NotificationRepository        notificationDomain.IRepository
}
