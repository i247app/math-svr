package exam

import (
	"encoding/json"
	"net/http"

	"math-ai.com/math-ai/internal/application/dto"
	"math-ai.com/math-ai/internal/application/resource"
	domain "math-ai.com/math-ai/internal/domain/exam"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/metadata"
	"math-ai.com/math-ai/internal/infrastructure/session"
	"math-ai.com/math-ai/internal/module/pow"
	"math-ai.com/math-ai/internal/shared/errs"
	"math-ai.com/math-ai/internal/shared/response"
)

// ExamHandler decodes, injects the session's user id, and delegates. The
// user id is taken from the session on every route and never from the
// body: the body is the client's claim, the session is the server's.
type ExamHandler struct {
	appResource *resource.Resource
	service     *Service
	powSvc      *pow.Service
}

func NewExamHandler(appResource *resource.Resource, service *Service, powSvc *pow.Service) *ExamHandler {
	return &ExamHandler{appResource: appResource, service: service, powSvc: powSvc}
}

// POST /exams/generate
//
// The one route registered WITHOUT AuthRequiredMiddleware, because it is
// where a visitor meets the product: a request with no session and no
// profile_id opens a guest account on the spot and generates against it.
// Auth is enforced here instead, and the rules are:
//
//   - a secure session         → their own uid, exactly as before (a
//     guest's own session also fills in their single profile when the
//     request omits it);
//   - no session, no profile   → a guest is opened (or recognised) and
//     the session is initialised for them;
//   - no session, but a stated → refused. profile_id is the client's
//     profile_id                 claim about whose exams to touch, and an
//     unauthenticated caller has no standing to make it.
//
// After the guest branch the request is indistinguishable from an
// authenticated one, so validation and the service below are untouched.
// The new session reaches the client the same way /auth/login's does —
// the X-Auth-Token response header.
func (h *ExamHandler) HandleGenerateExam(w http.ResponseWriter, r *http.Request) {
	var req dto.GenerateExamReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	ctx := r.Context()
	sess, err := h.appResource.GetRequestSession(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	// Every generation may be a paid model call, and this route is open to
	// callers with no account (a guest is opened below). Spend a solved
	// proof of work first — before a guest is created — so a bot pays CPU
	// for each exam it asks for. No-op unless POW_ENABLED.
	if err := h.powSvc.ConsumePass(ctx, sess); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	if uid, ok := h.callerUID(ctx, sess); ok {
		req.UID = &uid
		if req.ProfileID <= 0 {
			profileID, isGuest, err := h.service.guest.DefaultProfileOf(ctx, uid)
			if err != nil {
				response.WriteJson(w, nil, errs.NewError(ctx, status.FAIL, nil, err))
				return
			}
			if isGuest {
				req.ProfileID = profileID
			}
		}
	} else {
		if req.ProfileID > 0 {
			response.WriteJson(w, nil, errs.NewError(ctx, domain.ExamGuestProfileNotOwned, nil, ErrGuestProfileNotOwned))
			return
		}

		identity, err := h.service.guest.EnsureGuest(ctx, metadata.GetDeviceUUID(ctx), req.ChildName)
		if err != nil {
			response.WriteJson(w, nil, err)
			return
		}

		// Deliberately NOT secure. A secure session opens every
		// auth-gated route in the product; a visitor who has not
		// registered gets only /exams/*, which gates itself on the uid
		// below. Registering is what makes the session secure — see
		// module/user.CreateUser.
		sess.Init(session.InitData{
			Source:    "guest",
			IsSecure:  false,
			UID:       identity.UID,
			LoginName: metadata.GetDeviceUUID(ctx),
		})

		req.UID = &identity.UID
		req.ProfileID = identity.ProfileID
	}

	res, err := h.service.GenerateExam(ctx, &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /exams/submit
func (h *ExamHandler) HandleSubmitExam(w http.ResponseWriter, r *http.Request) {
	var req dto.SubmitExamReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, ok := h.resolveCaller(w, r, &req.ProfileID)
	if !ok {
		return
	}
	req.UID = uid

	res, err := h.service.SubmitExam(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /exams/sessions/detail
func (h *ExamHandler) HandleGetExamSessionDetail(w http.ResponseWriter, r *http.Request) {
	var req dto.GetExamReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, ok := h.resolveCaller(w, r, &req.ProfileID)
	if !ok {
		return
	}
	req.UID = uid

	res, err := h.service.GetExamSessionDetail(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /exams/list
func (h *ExamHandler) HandleListExams(w http.ResponseWriter, r *http.Request) {
	var req dto.ListExamsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, ok := h.resolveCaller(w, r, &req.ProfileID)
	if !ok {
		return
	}
	req.UID = uid

	res, err := h.service.ListExams(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /exams/sessions/list
func (h *ExamHandler) HandleListExamSessions(w http.ResponseWriter, r *http.Request) {
	var req dto.ListExamSessionsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, ok := h.resolveCaller(w, r, &req.ProfileID)
	if !ok {
		return
	}
	req.UID = uid

	res, err := h.service.ListExamSessions(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /exams/analytics/progress
func (h *ExamHandler) HandleGetExamProgress(w http.ResponseWriter, r *http.Request) {
	var req dto.ExamProgressReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, ok := h.resolveCaller(w, r, &req.ProfileID)
	if !ok {
		return
	}
	req.UID = uid

	res, err := h.service.GetExamProgress(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /exams/journey/progress
func (h *ExamHandler) HandleGetJourneyProgress(w http.ResponseWriter, r *http.Request) {
	var req dto.JourneyProgressReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, ok := h.resolveCaller(w, r, &req.ProfileID)
	if !ok {
		return
	}
	req.UID = uid

	res, err := h.service.GetJourneyProgress(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /exams/sessions/latest
func (h *ExamHandler) HandleGetLatestJourney(w http.ResponseWriter, r *http.Request) {
	var req dto.LatestJourneyReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, ok := h.resolveCaller(w, r, &req.ProfileID)
	if !ok {
		return
	}
	req.UID = uid

	res, err := h.service.GetLatestJourney(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /exams/grade/roadmap
func (h *ExamHandler) HandleGetGradeRoadMap(w http.ResponseWriter, r *http.Request) {
	var req dto.GradeMapReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, ok := h.resolveCaller(w, r, &req.ProfileID)
	if !ok {
		return
	}
	req.UID = uid

	res, err := h.service.GetGradeRoadMap(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /exams/sessions/review
func (h *ExamHandler) HandleReviewJourney(w http.ResponseWriter, r *http.Request) {
	var req dto.JourneyReviewReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, ok := h.resolveCaller(w, r, &req.ProfileID)
	if !ok {
		return
	}
	req.UID = uid

	res, err := h.service.ReviewJourney(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /exams/grade/levels
func (h *ExamHandler) HandleGetGradeLevels(w http.ResponseWriter, r *http.Request) {
	var req dto.GradeLevelsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, ok := h.resolveCaller(w, r, &req.ProfileID)
	if !ok {
		return
	}
	req.UID = uid

	res, err := h.service.GetGradeLevels(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /exams/journeys/mark
func (h *ExamHandler) HandleMarkExamJourney(w http.ResponseWriter, r *http.Request) {
	var req dto.MarkExamJourneyReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, ok := h.resolveCaller(w, r, &req.ProfileID)
	if !ok {
		return
	}
	req.UID = uid

	res, err := h.service.MarkExamJourney(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /admin/exams/sessions/list
//
// The admin view of /exams/sessions/list: same body, same response, any
// child's journeys. profile_id is required (no guest default). The route
// sits behind adminMiddleware; the request then runs as the profile's
// owner through the ordinary service method (see Service.actAsOwner).
func (h *ExamHandler) HandleAdminListExamSessions(w http.ResponseWriter, r *http.Request) {
	var req dto.ListExamSessionsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	owner, ok := h.actAsOwner(w, r, req.ProfileID)
	if !ok {
		return
	}
	req.UID = owner

	res, err := h.service.ListExamSessions(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /admin/exams/sessions/detail
//
// The admin view of /exams/sessions/detail: same body (profile_id plus
// elink_id or esess_id), same response, any child. profile_id is required
// and must be the profile the sitting / journey belongs to — the owner
// path checks that, exactly as it does for the parent.
func (h *ExamHandler) HandleAdminGetExamSessionDetail(w http.ResponseWriter, r *http.Request) {
	var req dto.GetExamReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	owner, ok := h.actAsOwner(w, r, req.ProfileID)
	if !ok {
		return
	}
	req.UID = owner

	res, err := h.service.GetExamSessionDetail(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /exams/pools/generate — admin only (adminMiddleware): one model
// call whose question set joins the pool; nobody is handed it.
func (h *ExamHandler) HandleGenerateExamPool(w http.ResponseWriter, r *http.Request) {
	var req dto.GenerateExamPoolReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.service.GenerateExamPool(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /exams/pools/list — admin only.
func (h *ExamHandler) HandleListExamPools(w http.ResponseWriter, r *http.Request) {
	var req dto.ListExamPoolsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.service.ListExamPools(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /exams/pools/detail — admin only; carries the answer key.
func (h *ExamHandler) HandleGetExamPool(w http.ResponseWriter, r *http.Request) {
	var req dto.GetExamPoolReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.service.GetExamPool(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /exams/pools/mark-verify — admin only.
func (h *ExamHandler) HandleMarkExamPoolVerify(w http.ResponseWriter, r *http.Request) {
	var req dto.MarkExamPoolVerifyReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.service.MarkExamPoolVerify(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /exams/pools/verify — admin only.
func (h *ExamHandler) HandleVerifyExamPool(w http.ResponseWriter, r *http.Request) {
	var req dto.VerifyExamPoolReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.service.VerifyExamPool(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}
