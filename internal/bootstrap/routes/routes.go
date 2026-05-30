package routes

import (
	"github.com/i247app/gex"
	"math-ai.com/math-ai/internal/application/resource"
	"math-ai.com/math-ai/internal/bootstrap/container"
	"math-ai.com/math-ai/internal/bootstrap/middleware"
	"math-ai.com/math-ai/internal/module/auth"
	"math-ai.com/math-ai/internal/module/chapter"
	"math-ai.com/math-ai/internal/module/classroom"
	"math-ai.com/math-ai/internal/module/device"
	"math-ai.com/math-ai/internal/module/grade"
	"math-ai.com/math-ai/internal/module/health"
	"math-ai.com/math-ai/internal/module/job"
	"math-ai.com/math-ai/internal/module/otp"
	"math-ai.com/math-ai/internal/module/profile"
	"math-ai.com/math-ai/internal/module/program"
	"math-ai.com/math-ai/internal/module/quiz"
	"math-ai.com/math-ai/internal/module/school"
	"math-ai.com/math-ai/internal/module/semester"
	"math-ai.com/math-ai/internal/module/session"
	"math-ai.com/math-ai/internal/module/user"
)

func SetupHttpRoutes(gexSvr *gex.Server, res *resource.Resource, services *container.ServiceContainer) {
	// middleware
	authMiddleware := middleware.AuthRequiredMiddleware(res.SessionManager)

	// session routes
	{
		sessionHandler := session.NewHandler(res)
		gexSvr.AddRoute("POST /sessions/dump", sessionHandler.HandleSessionDump)
		gexSvr.AddRoute("POST /sessions/delete-unsecure", sessionHandler.HandleDeleteUnSecureSessions, authMiddleware)
		gexSvr.AddRoute("POST /sessions/delete-all", sessionHandler.HandleDeleteAllSessions, authMiddleware)
	}

	// user routes
	{
		userHandler := user.NewUserHandler(res, services.UserSvc)
		gexSvr.AddRoute("GET  /users/{id}", userHandler.HandleGetUserById, authMiddleware)
		gexSvr.AddRoute("POST /users/me", userHandler.HandleGetUserMe, authMiddleware)
		gexSvr.AddRoute("POST /users/list", userHandler.HandleListUsers, authMiddleware)
		gexSvr.AddRoute("POST /users/create", userHandler.HandleCreateUser)
		gexSvr.AddRoute("POST /users/update", userHandler.HandleUpdateUser, authMiddleware)
		gexSvr.AddRoute("POST /users/upload-avatar", userHandler.HandleUploadAvatar, authMiddleware)

		// admin routes
		gexSvr.AddRoute("POST /users/soft-delete", userHandler.HandleSoftDeleteUser, authMiddleware)
		gexSvr.AddRoute("POST /users/force-delete", userHandler.HandleForceDeleteUser, authMiddleware)
	}

	// auth routes
	{
		authHandler := auth.NewAuthHandler(res, services.AuthSvc)
		gexSvr.AddRoute("POST /auth/login", authHandler.HandleLogin)
		gexSvr.AddRoute("POST /auth/login-resume", authHandler.HandleLoginResume, authMiddleware)
		gexSvr.AddRoute("POST /auth/otp", authHandler.HandleLoginOTP)
		gexSvr.AddRoute("POST /auth/logout", authHandler.HandleLogout, authMiddleware)
	}

	// curriculum reference routes (read-only)
	{
		programHandler := program.NewProgramHandler(services.ProgramSvc)
		gexSvr.AddRoute("POST /programs/list", programHandler.HandleListPrograms)
		gexSvr.AddRoute("GET  /programs/{id}", programHandler.HandleGetProgram, authMiddleware)
		gexSvr.AddRoute("POST /programs/create", programHandler.HandleCreateProgram, authMiddleware)
		gexSvr.AddRoute("POST /programs/update", programHandler.HandleUpdateProgram, authMiddleware)
		gexSvr.AddRoute("POST /programs/soft-delete", programHandler.HandleSoftDeleteProgram, authMiddleware)
		gexSvr.AddRoute("POST /programs/force-delete", programHandler.HandleForceDeleteProgram, authMiddleware)

		gradeHandler := grade.NewGradeHandler(services.GradeSvc)
		gexSvr.AddRoute("POST /grades/list", gradeHandler.HandleListGrades)
		gexSvr.AddRoute("GET  /grades/{id}", gradeHandler.HandleGetGrade, authMiddleware)
		gexSvr.AddRoute("POST /grades/create", gradeHandler.HandleCreateGrade, authMiddleware)
		gexSvr.AddRoute("POST /grades/update", gradeHandler.HandleUpdateGrade, authMiddleware)
		gexSvr.AddRoute("POST /grades/soft-delete", gradeHandler.HandleSoftDeleteGrade, authMiddleware)
		gexSvr.AddRoute("POST /grades/force-delete", gradeHandler.HandleForceDeleteGrade, authMiddleware)

		semesterHandler := semester.NewSemesterHandler(services.SemesterSvc)
		gexSvr.AddRoute("POST /semesters/list", semesterHandler.HandleListSemesters)
		gexSvr.AddRoute("GET  /semesters/{id}", semesterHandler.HandleGetSemester, authMiddleware)
		gexSvr.AddRoute("POST /semesters/create", semesterHandler.HandleCreateSemester, authMiddleware)
		gexSvr.AddRoute("POST /semesters/update", semesterHandler.HandleUpdateSemester, authMiddleware)
		gexSvr.AddRoute("POST /semesters/soft-delete", semesterHandler.HandleSoftDeleteSemester, authMiddleware)
		gexSvr.AddRoute("POST /semesters/force-delete", semesterHandler.HandleForceDeleteSemester, authMiddleware)
	}

	// profile routes
	{
		profileHandler := profile.NewProfileHandler(services.ProfileSvc)
		gexSvr.AddRoute("GET  /profiles/{id}", profileHandler.HandleGetProfileById, authMiddleware)
		gexSvr.AddRoute("POST /profiles/list", profileHandler.HandleListProfiles, authMiddleware)
		gexSvr.AddRoute("POST /profiles/create", profileHandler.HandleCreateProfile, authMiddleware)
		gexSvr.AddRoute("POST /profiles/update", profileHandler.HandleUpdateProfile, authMiddleware)
		gexSvr.AddRoute("POST /profiles/soft-delete", profileHandler.HandleSoftDeleteProfile, authMiddleware)
		gexSvr.AddRoute("POST /profiles/force-delete", profileHandler.HandleForceDeleteProfile, authMiddleware)
		gexSvr.AddRoute("POST /profiles/upload-avatar", profileHandler.HandleUploadAvatar, authMiddleware)
		gexSvr.AddRoute("POST /profiles/assign-school", profileHandler.HandleAssignSchool, authMiddleware)
		gexSvr.AddRoute("POST /profiles/remove-school", profileHandler.HandleRemoveSchool, authMiddleware)
	}

	// school routes
	{
		schoolHandler := school.NewSchoolHandler(res, services.SchoolSvc)
		gexSvr.AddRoute("GET  /schools/{id}", schoolHandler.HandleGetSchool, authMiddleware)
		gexSvr.AddRoute("POST /schools/list", schoolHandler.HandleListSchools, authMiddleware)
		gexSvr.AddRoute("POST /schools/create", schoolHandler.HandleCreateSchool, authMiddleware)
		gexSvr.AddRoute("POST /schools/update", schoolHandler.HandleUpdateSchool, authMiddleware)
		gexSvr.AddRoute("POST /schools/soft-delete", schoolHandler.HandleSoftDeleteSchool, authMiddleware)
		gexSvr.AddRoute("POST /schools/force-delete", schoolHandler.HandleForceDeleteSchool, authMiddleware)
	}

	// device routes
	{
		deviceHandler := device.NewDeviceHandler(services.DeviceSvc)
		gexSvr.AddRoute("GET  /devices/{id}", deviceHandler.HandleGetDeviceById)
		gexSvr.AddRoute("POST /devices/list", deviceHandler.HandleListDevices)
		gexSvr.AddRoute("POST /devices/update", deviceHandler.HandleUpdateDevice)
		gexSvr.AddRoute("POST /devices/revoke", deviceHandler.HandleRevokeDevice)
		gexSvr.AddRoute("POST /devices/soft-delete", deviceHandler.HandleSoftDeleteDevice)
	}

	// otp routes
	{
		otpHandler := otp.NewOtpHandler(res, services.OtpSvc)
		gexSvr.AddRoute("POST /otps/send", otpHandler.HandleSend)
		gexSvr.AddRoute("POST /otps/verify", otpHandler.HandleVerify)
	}

	// chapter routes
	{
		chapterHandler := chapter.NewChapterHandler(res, services.ChapterSvc)
		gexSvr.AddRoute("GET  /chapters/{id}", chapterHandler.HandleGetChapter, authMiddleware)
		gexSvr.AddRoute("POST /chapters/list", chapterHandler.HandleListChapters, authMiddleware)
		gexSvr.AddRoute("POST /chapters/create", chapterHandler.HandleCreateChapter, authMiddleware)
		gexSvr.AddRoute("POST /chapters/update", chapterHandler.HandleUpdateChapter, authMiddleware)
		gexSvr.AddRoute("POST /chapters/soft-delete", chapterHandler.HandleSoftDeleteChapter, authMiddleware)
		gexSvr.AddRoute("POST /chapters/force-delete", chapterHandler.HandleForceDeleteChapter, authMiddleware)
	}

	// quiz routes
	{
		quizHandler := quiz.NewQuizHandler(res, services.QuizSvc)
		gexSvr.AddRoute("GET  /quizzes/{id}", quizHandler.HandleGetQuiz, authMiddleware)
		gexSvr.AddRoute("POST /quizzes/list", quizHandler.HandleListQuizzes, authMiddleware)
		gexSvr.AddRoute("POST /quizzes/generate", quizHandler.HandleGenerateQuiz, authMiddleware)
		gexSvr.AddRoute("POST /quizzes/submit", quizHandler.HandleSubmitQuizAnswers, authMiddleware)
		gexSvr.AddRoute("POST /quizzes/soft-delete", quizHandler.HandleSoftDeleteQuiz, authMiddleware)
	}

	// classroom routes
	{
		classroomHandler := classroom.NewClassroomHandler(res, services.ClassroomSvc)
		gexSvr.AddRoute("GET  /classrooms/{id}", classroomHandler.HandleGetClassroom, authMiddleware)
		gexSvr.AddRoute("POST /classrooms/list", classroomHandler.HandleListClassrooms, authMiddleware)
		gexSvr.AddRoute("POST /classrooms/create", classroomHandler.HandleCreateClassroom, authMiddleware)
		gexSvr.AddRoute("POST /classrooms/update", classroomHandler.HandleUpdateClassroom, authMiddleware)
		gexSvr.AddRoute("POST /classrooms/archive", classroomHandler.HandleArchiveClassroom, authMiddleware)
		gexSvr.AddRoute("POST /classrooms/restore", classroomHandler.HandleRestoreClassroom, authMiddleware)
		gexSvr.AddRoute("POST /classrooms/soft-delete", classroomHandler.HandleSoftDeleteClassroom, authMiddleware)
		gexSvr.AddRoute("POST /classrooms/force-delete", classroomHandler.HandleForceDeleteClassroom, authMiddleware)

		// membership
		gexSvr.AddRoute("POST /classrooms/join-by-code", classroomHandler.HandleJoinByCode, authMiddleware)
		gexSvr.AddRoute("POST /classrooms/leave", classroomHandler.HandleLeaveClassroom, authMiddleware)
		gexSvr.AddRoute("POST /classrooms/transfer-ownership", classroomHandler.HandleTransferOwnership, authMiddleware)
		gexSvr.AddRoute("POST /classrooms/members/list", classroomHandler.HandleListMembers, authMiddleware)
		gexSvr.AddRoute("POST /classrooms/members/remove", classroomHandler.HandleRemoveMember, authMiddleware)
		gexSvr.AddRoute("POST /classrooms/members/update-role", classroomHandler.HandleUpdateMemberRole, authMiddleware)
	}

	// health routes
	{
		healthHandler := health.NewHealthHandler()
		gexSvr.AddRoute("GET /ping", healthHandler.HandlePing)
	}

	// jobs routes
	{
		jobHandler := job.NewJobHandler(services.JobSvc)
		gexSvr.AddRoute("POST /jobs/list", jobHandler.HandleListJobs, authMiddleware)
		gexSvr.AddRoute("POST /jobs/trigger", jobHandler.HandleTriggerJob, authMiddleware)
		gexSvr.AddRoute("POST /jobs/pause", jobHandler.HandlePauseJob, authMiddleware)
		gexSvr.AddRoute("POST /jobs/resume", jobHandler.HandleResumeJob, authMiddleware)
		gexSvr.AddRoute("POST /tasks/enqueue", jobHandler.HandleEnqueueTask, authMiddleware)
	}
}
