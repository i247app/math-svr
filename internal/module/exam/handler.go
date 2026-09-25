package exam

import (
	"context"
	"encoding/json"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/exam"
	"math-ai.com/math-ai/internal/application/resource"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/metadata"
	"math-ai.com/math-ai/internal/infrastructure/session"
	"math-ai.com/math-ai/internal/module/pow"
	"math-ai.com/math-ai/internal/shared/response"
)

// ExamHandler decodes, injects the session's user id, and delegates. The
// user id is taken from the session on every route and never from the
// body: the body is the client's claim, the session is the server's.
type ExamHandler struct {
	appResource *resource.Resource
	examSvc     *Service
	powSvc      *pow.Service
}

func NewExamHandler(appResource *resource.Resource, examSvc *Service, powSvc *pow.Service) *ExamHandler {
	return &ExamHandler{appResource: appResource, examSvc: examSvc, powSvc: powSvc}
}

// uid pulls the caller's user id out of the request's session, applying
// the same gate AuthRequiredMiddleware applies elsewhere — with one
// deliberate exception for guests.
//
// A guest session is NEVER marked secure. Secure is the key to every
// auth-gated route in the product (classrooms, banners, notifications),
// and someone who has not registered must not hold it. So /exams/* is
// registered without AuthRequiredMiddleware and gates itself here:
//
//   - secure session          → allowed, as before;
//   - unsecure, uid is a guest → allowed, this is their only surface;
//   - unsecure, uid is a real  → refused, exactly as the middleware
//     account (mid-login)        would have refused them.
//
// The uid is server-written session state, so a client cannot claim to
// be a guest they are not.
func (h *ExamHandler) uid(w http.ResponseWriter, r *http.Request) (*int64, bool) {
	ctx := r.Context()
	ss, err := h.appResource.GetRequestSession(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return nil, false
	}
	id, ok := ss.UID()
	if !ok || !ss.IsValid() {
		response.WriteJson(w, nil, errs.NewError(ctx, status.UNAUTHORIZED, nil, session.ErrUidNotFoundFromSession))
		return nil, false
	}
	if !ss.IsSecure() {
		isGuest, err := h.examSvc.guest.IsGuest(ctx, id)
		if err != nil {
			response.WriteJson(w, nil, errs.NewError(ctx, status.FAIL, nil, err))
			return nil, false
		}
		if !isGuest {
			response.WriteJson(w, nil, errs.NewError(ctx, status.UNAUTHORIZED, nil, ErrSessionNotSecure))
			return nil, false
		}
	}
	return &id, true
}

// resolveCaller pins a request to its user and, when that user is a
// guest, fills in the child they must have meant.
//
// Every /exams/* route names a profile, because a parent has several
// children. A guest has exactly one — the server opened it for them —
// so requiring them to repeat its id is asking a question with a single
// answer, and a client that did not keep the id from the hand-out has no
// way to answer it. profileID is only written when it was left empty and
// the caller really is a guest; a registered parent still has to say
// which child, and a stated id is never overridden.
func (h *ExamHandler) resolveCaller(w http.ResponseWriter, r *http.Request, profileID *int64) (*int64, bool) {
	uid, ok := h.uid(w, r)
	if !ok {
		return nil, false
	}
	if profileID != nil && *profileID <= 0 {
		ctx := r.Context()
		resolved, isGuest, err := h.examSvc.guest.DefaultProfileOf(ctx, *uid)
		if err != nil {
			response.WriteJson(w, nil, errs.NewError(ctx, status.FAIL, nil, err))
			return nil, false
		}
		if isGuest {
			*profileID = resolved
		}
	}
	return uid, true
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
		req.UserID = &uid
		if req.ProfileID <= 0 {
			profileID, isGuest, err := h.examSvc.guest.DefaultProfileOf(ctx, uid)
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
			response.WriteJson(w, nil, errs.NewError(ctx, status.EXAM_GUEST_PROFILE_NOT_OWNED, nil, ErrGuestProfileNotOwned))
			return
		}

		identity, err := h.examSvc.guest.EnsureGuest(ctx, metadata.GetDeviceUUID(ctx), req.ChildName)
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
			UID:       identity.UserID,
			LoginName: metadata.GetDeviceUUID(ctx),
		})

		req.UserID = &identity.UserID
		req.ProfileID = identity.ProfileID
	}

	res, err := h.examSvc.GenerateExam(ctx, &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// callerUID reports the caller's uid when the session already belongs to
// someone — a secure session, or an unsecure one carrying a guest the
// server opened earlier. Anything short of that is treated as "no
// session at all" rather than as an error, because on this route that is
// a legitimate way to arrive: it is where a guest is born.
func (h *ExamHandler) callerUID(ctx context.Context, sess *session.AppSession) (int64, bool) {
	if sess == nil || !sess.IsValid() {
		return 0, false
	}
	id, ok := sess.UID()
	if !ok {
		return 0, false
	}
	if sess.IsSecure() {
		return id, true
	}
	isGuest, err := h.examSvc.guest.IsGuest(ctx, id)
	if err != nil || !isGuest {
		return 0, false
	}
	return id, true
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
	req.UserID = uid

	res, err := h.examSvc.SubmitExam(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /exams/detail
func (h *ExamHandler) HandleGetExam(w http.ResponseWriter, r *http.Request) {
	var req dto.GetExamReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, ok := h.resolveCaller(w, r, &req.ProfileID)
	if !ok {
		return
	}
	req.UserID = uid

	res, err := h.examSvc.GetExam(r.Context(), &req)
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
	req.UserID = uid

	res, err := h.examSvc.ListExams(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /exams/stats
func (h *ExamHandler) HandleGetExamStats(w http.ResponseWriter, r *http.Request) {
	var req dto.GetExamStatsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, ok := h.resolveCaller(w, r, &req.ProfileID)
	if !ok {
		return
	}
	req.UserID = uid

	res, err := h.examSvc.GetExamStats(r.Context(), &req)
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
	req.UserID = uid

	res, err := h.examSvc.GetExamProgress(r.Context(), &req)
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
	req.UserID = uid

	res, err := h.examSvc.GetJourneyProgress(r.Context(), &req)
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
	req.UserID = uid

	res, err := h.examSvc.MarkExamJourney(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}
