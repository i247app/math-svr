package exam

import (
	"context"
	"net/http"

	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/session"
	"math-ai.com/math-ai/internal/shared/errs"
	"math-ai.com/math-ai/internal/shared/response"
)

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
		isGuest, err := h.service.guest.IsGuest(ctx, id)
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
		resolved, isGuest, err := h.service.guest.DefaultProfileOf(ctx, *uid)
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

// actAsOwner is the admin counterpart of resolveCaller: it answers the
// uid an /admin/exams/* request runs as — the owner of profileID (see
// Service.actAsOwner) — or writes the error and reports false.
func (h *ExamHandler) actAsOwner(w http.ResponseWriter, r *http.Request, profileID int64) (*int64, bool) {
	owner, err := h.service.actAsOwner(r.Context(), r.Method+" "+r.URL.Path, profileID)
	if err != nil {
		response.WriteJson(w, nil, err)
		return nil, false
	}
	return owner, true
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
	isGuest, err := h.service.guest.IsGuest(ctx, id)
	if err != nil || !isGuest {
		return 0, false
	}
	return id, true
}
