package routes

import (
	"net/http"

	"github.com/i247app/gex"
	"math-ai.com/math-ai/internal/application/resource"
	"math-ai.com/math-ai/internal/bootstrap/container"
	"math-ai.com/math-ai/internal/bootstrap/middleware"
	"math-ai.com/math-ai/internal/module/auth"
	"math-ai.com/math-ai/internal/module/bot"
	"math-ai.com/math-ai/internal/module/chapter"
	"math-ai.com/math-ai/internal/module/classroom"
	"math-ai.com/math-ai/internal/module/device"
	"math-ai.com/math-ai/internal/module/exercise"
	"math-ai.com/math-ai/internal/module/grade"
	"math-ai.com/math-ai/internal/module/health"
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
	"math-ai.com/math-ai/internal/module/server"
	"math-ai.com/math-ai/internal/module/session"
	"math-ai.com/math-ai/internal/module/socket"
	"math-ai.com/math-ai/internal/module/user"
)

func SetupHttpRoutes(gexSvr *gex.Server, res *resource.Resource, services *container.ServiceContainer) {
	// reg registers a route with gex AND mirrors its spec into the route
	// classifier, so observability middleware can label metrics / name spans
	// by the matched template (keeping cardinality bounded). Use reg instead
	// of gexSvr.AddRoute for every route below.
	reg := func(spec string, handler http.HandlerFunc, mw ...gex.Middleware) {
		gexSvr.AddRoute(spec, handler, mw...)
		res.RouteClassifier.Add(spec)
	}

	// middleware
	authMiddleware := middleware.AuthRequiredMiddleware(res.SessionManager)

	// misc routes
	{
		miscHandler := misc.NewHandler(services.MiscSvc)
		reg("POST /misc/logs-time-format", miscHandler.LogsTimeFormat)
		// Destructive: wipes all user-generated data. Auth-gated (secure session).
		reg("POST /misc/clear-data", miscHandler.ClearData, authMiddleware)
	}

	// health routes
	{
		healthHandler := health.NewHealthHandler()
		reg("GET /ping", healthHandler.HandlePing)
	}

	// NOTE: /metrics is intentionally NOT registered here. It is served on a
	// separate, middleware-free listener (see bootstrap.App.setupMetricsServer)
	// to avoid the app middleware chain double-gzipping the Prometheus
	// exposition format and spamming session/request logs on every scrape.

	// server lifecycle routes
	{
		serverHandler := server.NewHandler()
		reg("POST /server/shutdown", serverHandler.HandleShutdown, authMiddleware)
	}

	// session routes
	{
		sessionHandler := session.NewHandler(res)
		reg("POST /sessions/dump", sessionHandler.HandleSessionDump)
		reg("POST /sessions/delete-unsecure", sessionHandler.HandleDeleteUnSecureSessions, authMiddleware)
		reg("POST /sessions/delete-all", sessionHandler.HandleDeleteAllSessions)
	}

	// user routes
	{
		userHandler := user.NewUserHandler(res, services.UserSvc)
		reg("GET  /users/{id}", userHandler.HandleGetUserById, authMiddleware)
		reg("POST /users/me", userHandler.HandleGetUserMe, authMiddleware)
		reg("POST /users/list", userHandler.HandleListUsers, authMiddleware)
		reg("POST /users/create", userHandler.HandleCreateUser)
		reg("POST /users/update", userHandler.HandleUpdateUser, authMiddleware)
		reg("POST /users/upload-avatar", userHandler.HandleUploadAvatar, authMiddleware)

		// admin routes
		reg("POST /users/soft-delete", userHandler.HandleSoftDeleteUser, authMiddleware)
		reg("POST /users/force-delete", userHandler.HandleForceDeleteUser, authMiddleware)
	}

	// auth routes
	{
		authHandler := auth.NewAuthHandler(res, services.AuthSvc)
		reg("POST /auth/login", authHandler.HandleLogin)
		reg("POST /auth/login-resume", authHandler.HandleLoginResume)
		reg("POST /auth/otp", authHandler.HandleLoginOTP)
		reg("POST /auth/logout", authHandler.HandleLogout, authMiddleware)
	}

	// curriculum reference routes (read-only)
	{
		programHandler := program.NewProgramHandler(services.ProgramSvc)
		reg("POST /programs/list", programHandler.HandleListPrograms)
		reg("GET  /programs/{id}", programHandler.HandleGetProgram, authMiddleware)
		reg("POST /programs/create", programHandler.HandleCreateProgram, authMiddleware)
		reg("POST /programs/update", programHandler.HandleUpdateProgram, authMiddleware)
		reg("POST /programs/soft-delete", programHandler.HandleSoftDeleteProgram, authMiddleware)
		reg("POST /programs/force-delete", programHandler.HandleForceDeleteProgram, authMiddleware)

		gradeHandler := grade.NewGradeHandler(services.GradeSvc)
		reg("POST /grades/list", gradeHandler.HandleListGrades)
		reg("GET  /grades/{id}", gradeHandler.HandleGetGrade, authMiddleware)
		reg("POST /grades/create", gradeHandler.HandleCreateGrade, authMiddleware)
		reg("POST /grades/update", gradeHandler.HandleUpdateGrade, authMiddleware)
		reg("POST /grades/soft-delete", gradeHandler.HandleSoftDeleteGrade, authMiddleware)
		reg("POST /grades/force-delete", gradeHandler.HandleForceDeleteGrade, authMiddleware)

		semesterHandler := semester.NewSemesterHandler(services.SemesterSvc)
		reg("POST /semesters/list", semesterHandler.HandleListSemesters)
		reg("GET  /semesters/{id}", semesterHandler.HandleGetSemester, authMiddleware)
		reg("POST /semesters/create", semesterHandler.HandleCreateSemester, authMiddleware)
		reg("POST /semesters/update", semesterHandler.HandleUpdateSemester, authMiddleware)
		reg("POST /semesters/soft-delete", semesterHandler.HandleSoftDeleteSemester, authMiddleware)
		reg("POST /semesters/force-delete", semesterHandler.HandleForceDeleteSemester, authMiddleware)
	}

	// profile routes
	{
		profileHandler := profile.NewProfileHandler(services.ProfileSvc)
		reg("GET  /profiles/{id}", profileHandler.HandleGetProfileById, authMiddleware)
		reg("POST /profiles/list", profileHandler.HandleListProfiles, authMiddleware)
		reg("POST /profiles/create", profileHandler.HandleCreateProfile, authMiddleware)
		reg("POST /profiles/update", profileHandler.HandleUpdateProfile, authMiddleware)
		reg("POST /profiles/soft-delete", profileHandler.HandleSoftDeleteProfile, authMiddleware)
		reg("POST /profiles/force-delete", profileHandler.HandleForceDeleteProfile, authMiddleware)
		reg("POST /profiles/upload-avatar", profileHandler.HandleUploadAvatar, authMiddleware)
		reg("POST /profiles/assign-school", profileHandler.HandleAssignSchool, authMiddleware)
		reg("POST /profiles/remove-school", profileHandler.HandleRemoveSchool, authMiddleware)

		// admin routes
		reg("POST /profiles/upload-static-file", profileHandler.HandleUploadStaticFile, authMiddleware)
		reg("POST /profiles/delete-static-file", profileHandler.HandleDeleteStaticFile, authMiddleware)
	}

	// school routes
	{
		schoolHandler := school.NewSchoolHandler(res, services.SchoolSvc)
		reg("GET  /schools/{id}", schoolHandler.HandleGetSchool, authMiddleware)
		reg("POST /schools/list", schoolHandler.HandleListSchools, authMiddleware)
		reg("POST /schools/create", schoolHandler.HandleCreateSchool, authMiddleware)
		reg("POST /schools/update", schoolHandler.HandleUpdateSchool, authMiddleware)
		reg("POST /schools/soft-delete", schoolHandler.HandleSoftDeleteSchool, authMiddleware)
		reg("POST /schools/force-delete", schoolHandler.HandleForceDeleteSchool, authMiddleware)
	}

	// device routes
	{
		deviceHandler := device.NewDeviceHandler(services.DeviceSvc)
		reg("GET  /devices/{id}", deviceHandler.HandleGetDeviceById)
		reg("POST /devices/list", deviceHandler.HandleListDevices)
		reg("POST /devices/update", deviceHandler.HandleUpdateDevice)
		reg("POST /devices/revoke", deviceHandler.HandleRevokeDevice)
		reg("POST /devices/soft-delete", deviceHandler.HandleSoftDeleteDevice)
	}

	// otp routes
	{
		otpHandler := otp.NewOtpHandler(res, services.OtpSvc)
		reg("POST /otps/send", otpHandler.HandleSend)
		reg("POST /otps/verify", otpHandler.HandleVerify)
	}

	// chapter routes
	{
		chapterHandler := chapter.NewChapterHandler(res, services.ChapterSvc)
		reg("GET  /chapters/{id}", chapterHandler.HandleGetChapter, authMiddleware)
		reg("POST /chapters/list", chapterHandler.HandleListChapters, authMiddleware)
		reg("POST /chapters/create", chapterHandler.HandleCreateChapter, authMiddleware)
		reg("POST /chapters/update", chapterHandler.HandleUpdateChapter, authMiddleware)
		reg("POST /chapters/soft-delete", chapterHandler.HandleSoftDeleteChapter, authMiddleware)
		reg("POST /chapters/force-delete", chapterHandler.HandleForceDeleteChapter, authMiddleware)
	}

	// ai routes — LLM connection warm-up (public + globally throttled) so the
	// frontend can prime the AI connection early and the first real
	// quiz/exercise generation is fast.
	{
		botHandler := bot.NewHandler(services.BotSvc)
		reg("POST /ai/shake", botHandler.HandleShake)
	}

	// quiz routes
	{
		quizHandler := quiz.NewQuizHandler(res, services.QuizSvc)
		reg("GET  /quizzes/{id}", quizHandler.HandleGetQuiz, authMiddleware)
		reg("POST /quizzes/list", quizHandler.HandleListQuizzes, authMiddleware)
		reg("POST /quizzes/generate", quizHandler.HandleGenerateQuiz, authMiddleware)
		reg("POST /quizzes/submit/cost-ai", quizHandler.HandleSubmitQuizAnswers, authMiddleware)
		reg("POST /quizzes/submit", quizHandler.HandleSubmitQuizV2, authMiddleware)
		reg("POST /quizzes/soft-delete", quizHandler.HandleSoftDeleteQuiz, authMiddleware)
	}

	// classroom routes
	{
		classroomHandler := classroom.NewClassroomHandler(res, services.ClassroomSvc)
		reg("GET  /classrooms/{id}", classroomHandler.HandleGetClassroom, authMiddleware)
		reg("POST /classrooms/list", classroomHandler.HandleListClassrooms, authMiddleware)
		reg("POST /classrooms/my-joined", classroomHandler.HandleListMyJoinedClassrooms, authMiddleware)
		reg("POST /classrooms/create", classroomHandler.HandleCreateClassroom, authMiddleware)
		reg("POST /classrooms/update", classroomHandler.HandleUpdateClassroom, authMiddleware)
		reg("POST /classrooms/archive", classroomHandler.HandleArchiveClassroom, authMiddleware)
		reg("POST /classrooms/restore", classroomHandler.HandleRestoreClassroom, authMiddleware)
		reg("POST /classrooms/soft-delete", classroomHandler.HandleSoftDeleteClassroom, authMiddleware)
		reg("POST /classrooms/force-delete", classroomHandler.HandleForceDeleteClassroom, authMiddleware)

		// membership
		reg("POST /classrooms/leave", classroomHandler.HandleLeaveClassroom, authMiddleware)
		reg("POST /classrooms/transfer-ownership", classroomHandler.HandleTransferOwnership, authMiddleware)
		reg("POST /classrooms/members/list", classroomHandler.HandleListMembers, authMiddleware)
		reg("POST /classrooms/members/remove", classroomHandler.HandleRemoveMember, authMiddleware)
		reg("POST /classrooms/members/update-role", classroomHandler.HandleUpdateMemberRole, authMiddleware)

		// invitations — teacher-initiated, member_status = PENDING_INVITATION
		reg("POST /classrooms/invitations/send", classroomHandler.HandleSendInvitation, authMiddleware)
		reg("POST /classrooms/invitations/list", classroomHandler.HandleListClassroomInvitations, authMiddleware)
		reg("POST /classrooms/invitations/my-pending", classroomHandler.HandleListMyPendingInvitations, authMiddleware)
		reg("POST /classrooms/invitations/accept", classroomHandler.HandleAcceptInvitation, authMiddleware)
		reg("POST /classrooms/invitations/reject", classroomHandler.HandleRejectInvitation, authMiddleware)
		reg("POST /classrooms/invitations/cancel", classroomHandler.HandleCancelInvitation, authMiddleware)

		// join requests — user-initiated, member_status = PENDING_REQUEST
		reg("POST /classrooms/find-by-code", classroomHandler.HandleFindClassroomByCode, authMiddleware)
		reg("POST /classrooms/join-by-code", classroomHandler.HandleJoinByCode, authMiddleware)
		reg("POST /classrooms/join-requests/list", classroomHandler.HandleListJoinRequestsByClassroom, authMiddleware)
		reg("POST /classrooms/join-requests/my-pending", classroomHandler.HandleListMyJoinRequests, authMiddleware)
		reg("POST /classrooms/join-requests/approve", classroomHandler.HandleApproveJoinRequest, authMiddleware)
		reg("POST /classrooms/join-requests/reject", classroomHandler.HandleRejectJoinRequest, authMiddleware)
		reg("POST /classrooms/join-requests/cancel", classroomHandler.HandleCancelJoinRequest, authMiddleware)

		// Profile learning-progress detail (002-profile-learning-progress).
		reg("POST /classrooms/progress/profile", classroomHandler.HandleProfileProgress, authMiddleware)
	}

	// classroom exercise routes — teacher-issued, AI-generated assignments
	{
		exerciseHandler := exercise.NewClassroomExerciseHandler(res, services.ExerciseSvc)
		reg("POST /classroom-exercises/create", exerciseHandler.HandleCreateExercise, authMiddleware)
		reg("POST /classroom-exercises/update", exerciseHandler.HandleUpdateExercise, authMiddleware)
		reg("GET  /classroom-exercises/{id}", exerciseHandler.HandleGetExercise, authMiddleware)
		reg("POST /classroom-exercises/list", exerciseHandler.HandleListExercises, authMiddleware)
		reg("POST /classroom-exercises/soft-delete", exerciseHandler.HandleSoftDeleteExercise, authMiddleware)

		// classroom exercise submissions — student submit + history,
		// teacher roster + soft-delete. Every route is auth-gated.
		submissionHandler := exercise.NewClassroomExerciseSubmissionHandler(res, services.ExerciseSvc)
		reg("POST /classroom-exercise/submissions/submit/cost-ai", submissionHandler.HandleSubmitExerciseAnswers, authMiddleware)
		reg("POST /classroom-exercise/submissions/submit", submissionHandler.HandleSubmitExerciseAnswersV2, authMiddleware)
		reg("GET  /classroom-exercise/submissions/{id}", submissionHandler.HandleGetSubmission, authMiddleware)
		reg("POST /classroom-exercise/submissions/list", submissionHandler.HandleListSubmissions, authMiddleware)
		reg("POST /classroom-exercise/submissions/by-exercise", submissionHandler.HandleListSubmissionsByExercise, authMiddleware)
		reg("POST /classroom-exercise/submissions/submitted-members", submissionHandler.HandleListSubmittedMembers, authMiddleware)
		reg("POST /classroom-exercise/submissions/non-submitted-members", submissionHandler.HandleListNonSubmittedMembers, authMiddleware)
		reg("POST /classroom-exercise/submissions/soft-delete", submissionHandler.HandleSoftDeleteSubmission, authMiddleware)
	}

	// notification routes — per-user in-app inbox + push delivery (auth-gated)
	{
		notificationHandler := notification.NewHandler(res, services.NotificationSvc)
		reg("POST /notifications/ping", notificationHandler.HandlePing, authMiddleware)
		reg("POST /notifications/list", notificationHandler.HandleList, authMiddleware)
		reg("POST /notifications/unread-count", notificationHandler.HandleUnreadCount, authMiddleware)
		reg("POST /notifications/send", notificationHandler.HandleSend, authMiddleware)
		reg("POST /notifications/mark-read", notificationHandler.HandleMarkRead, authMiddleware)
		reg("POST /notifications/mark-all-read", notificationHandler.HandleMarkAllRead, authMiddleware)
		reg("POST /notifications/soft-delete", notificationHandler.HandleSoftDelete, authMiddleware)
	}

	// home routes — role-aware dashboard composed per acting profile
	{
		homeHandler := home.NewHandler(res, services.HomeSvc)
		reg("POST /home/layout", homeHandler.HandleGetHomeLayout, authMiddleware)
	}

	// socket routes — realtime WebSocket channel (auth-gated). Registered only
	// when the socket runtime is enabled (SOCKET_ENABLED=true → SocketHub set).
	// The upgrade relies on the WS-aware middleware chain (see middleware pkg).
	if res.SocketHub != nil && services.SocketSvc != nil {
		socketHandler := socket.NewHandler(services.SocketSvc)
		reg("GET /ws/connect", socketHandler.HandleConnect)
	}

	// jobs routes
	{
		jobHandler := job.NewJobHandler(services.JobSvc)
		reg("POST /jobs/list", jobHandler.HandleListJobs, authMiddleware)
		reg("POST /jobs/trigger", jobHandler.HandleTriggerJob, authMiddleware)
		reg("POST /jobs/pause", jobHandler.HandlePauseJob, authMiddleware)
		reg("POST /jobs/resume", jobHandler.HandleResumeJob, authMiddleware)
		reg("POST /tasks/enqueue", jobHandler.HandleEnqueueTask, authMiddleware)
	}
}
