package notification

import (
	"context"
	"fmt"

	notifAdapter "math-ai.com/math-ai/internal/adapter/notification"
	command "math-ai.com/math-ai/internal/application/command/notification"
	dto "math-ai.com/math-ai/internal/application/dto/notification"
	query "math-ai.com/math-ai/internal/application/query/notification"
	appsocket "math-ai.com/math-ai/internal/application/socket"
	"math-ai.com/math-ai/internal/application/transaction"
	deviceDomain "math-ai.com/math-ai/internal/domain/device"
	domain "math-ai.com/math-ai/internal/domain/notification"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	userDomain "math-ai.com/math-ai/internal/domain/user"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/shared/enum"
)

// Service is the notification module façade. It persists notification rows
// (inbox), fans them out to a recipient's devices via FCM (out of any tx),
// and exposes the per-user inbox reads (list / unread-count / mark-read /
// soft-delete).
type Service struct {
	createCmd        *command.CreateNotificationCommandHandler
	markReadCmd      *command.MarkReadCommandHandler
	markAllReadCmd   *command.MarkAllReadCommandHandler
	softDeleteCmd    *command.SoftDeleteNotificationCommandHandler
	clearTokensCmd   *command.ClearDeadTokensCommandHandler
	listQuery        *query.ListNotificationsQueryHandler
	unreadCountQuery *query.UnreadCountQueryHandler
	userRepo         userDomain.IRepository
	deviceRepo       deviceDomain.IRepository
	push             *pushService
	socket           appsocket.Publisher
}

// NewService wires the notification module. push may be nil when the deploy
// runs with NOTIFICATION_PROVIDER=""/"disabled"; the inbox still works and
// delivery is skipped (zeroed push counts). socketPub may be nil when
// SOCKET_ENABLED=false; realtime delivery is then skipped.
func NewService(
	uow transaction.UnitOfWork,
	repo domain.IRepository,
	userRepo userDomain.IRepository,
	deviceRepo deviceDomain.IRepository,
	pushAdapter *notifAdapter.Adapter,
	socketPub appsocket.Publisher,
) *Service {
	return &Service{
		createCmd:        command.NewCreateNotificationCommandHandler(uow),
		markReadCmd:      command.NewMarkReadCommandHandler(uow),
		markAllReadCmd:   command.NewMarkAllReadCommandHandler(uow),
		softDeleteCmd:    command.NewSoftDeleteNotificationCommandHandler(uow),
		clearTokensCmd:   command.NewClearDeadTokensCommandHandler(uow),
		listQuery:        query.NewListNotificationsQueryHandler(repo),
		unreadCountQuery: query.NewUnreadCountQueryHandler(repo),
		userRepo:         userRepo,
		deviceRepo:       deviceRepo,
		push:             newPushService(pushAdapter),
		socket:           socketPub,
	}
}

func (s *Service) Ping(ctx context.Context, req *dto.PingNotificationReq) (*dto.PingNotificationRes, error) {
	log := logger.From(ctx)

	if err := ValidatePing(ctx, req); err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindById(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	notiDomain := domain.NewNotification()
	notiDomain.SetUserId(user.UserId())
	notiDomain.SetTitle("Ping")

	shortText := fmt.Sprintf("Hello %s, have a great day! ", user.UserName())
	notiDomain.SetShortText(shortText)

	cate := enum.NotificationCategoryTypeInfo.String()
	notiDomain.SetCategory(&cate)

	if s.push.enabled() {
		tokens, terr := s.recipientTokens(ctx, req.UserID)
		if terr != nil {
			log.Warnf("notification.token_lookup_failed user_id=%d err=%v", req.UserID, terr)
		} else if len(tokens) > 0 {
			sendRes, serr := s.push.send(ctx, tokens, notiDomain)
			if serr != nil {
				log.Warnf("notification.push_failed user_id=%d err=%v", req.UserID, serr)
			} else {
				if len(sendRes.InvalidTokens) > 0 {
					if cerr := s.clearTokensCmd.Handle(ctx, command.ClearDeadTokensCommand{
						Tokens: sendRes.InvalidTokens,
					}); cerr != nil {
						log.Warnf("notification.clear_dead_tokens_failed count=%d err=%v",
							len(sendRes.InvalidTokens), cerr)
					}
				}
			}
		}
	}

	res := &dto.PingNotificationRes{
		Notification: dto.DomainToResponse(notiDomain),
	}

	return res, nil
}

// SendNotification persists a notification for req.UserID then pushes it to that
// user's device tokens. Persistence is authoritative — a push failure is
// logged but never fails the request (the in-app inbox row already exists).
func (s *Service) SendNotification(ctx context.Context, req *dto.SendNotificationReq) (*dto.SendNotificationRes, error) {
	log := logger.From(ctx)

	if err := ValidateSend(ctx, req); err != nil {
		return nil, err
	}

	var actionData *string
	if len(req.ActionData) > 0 {
		raw := string(req.ActionData)
		actionData = &raw
	}

	created, err := s.createCmd.Handle(ctx, command.CreateNotificationCommand{
		UserID:     req.UserID,
		Title:      req.Title,
		ShortText:  req.ShortText,
		Category:   req.Category,
		ActionType: req.ActionType,
		ActionData: actionData,
		Priority:   req.Priority,
		Note:       req.Note,
		CreatorUID: req.CreatorUID,
	})
	if err != nil {
		return nil, err
	}

	res := &dto.SendNotificationRes{Notification: dto.DomainToResponse(created)}

	// Push delivery — out of any tx, best-effort.
	if s.push.enabled() {
		tokens, terr := s.recipientTokens(ctx, req.UserID)
		if terr != nil {
			log.Warnf("notification.token_lookup_failed user_id=%d err=%v", req.UserID, terr)
		} else if len(tokens) > 0 {
			sendRes, serr := s.push.send(ctx, tokens, created)
			if serr != nil {
				// Inbox row is already persisted; surface delivery failure in
				// logs only so the caller still sees a created notification.
				log.Warnf("notification.push_failed notification_id=%d user_id=%d err=%v",
					created.NotificationId(), req.UserID, serr)
			} else {
				res.PushSuccess = sendRes.SuccessCount
				res.PushFailure = sendRes.FailureCount
				if len(sendRes.InvalidTokens) > 0 {
					if cerr := s.clearTokensCmd.Handle(ctx, command.ClearDeadTokensCommand{
						Tokens: sendRes.InvalidTokens,
					}); cerr != nil {
						log.Warnf("notification.clear_dead_tokens_failed count=%d err=%v",
							len(sendRes.InvalidTokens), cerr)
					}
				}
			}
		}
	}

	// Realtime delivery — out of any tx, best-effort. Mirrors the FCM push: the
	// inbox row is authoritative, so a socket publish failure is logged only.
	// Reaches only the recipient's currently-connected sockets (nil when the
	// socket runtime is disabled).
	if s.socket != nil {
		topic := appsocket.NotificationsTopic(req.UserID)
		event := "notification.created"
		log.Infof("send socket event to user %d : %v", req.UserID, event)
		if perr := s.socket.Publish(ctx, topic, event, res.Notification); perr != nil {
			log.Warnf("notification.socket_publish_failed notification_id=%d user_id=%d err=%v",
				created.NotificationId(), req.UserID, perr)
		}
	}

	return res, nil
}

// recipientTokens returns the deduplicated, non-empty push tokens registered
// to the user's active devices.
func (s *Service) recipientTokens(ctx context.Context, userID int64) ([]string, error) {
	devices, err := s.deviceRepo.ListByUserId(ctx, userID)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{}, len(devices))
	tokens := make([]string, 0, len(devices))
	for _, d := range devices {
		t := d.DevicePushToken()
		if t == nil || *t == "" {
			continue
		}
		if _, ok := seen[*t]; ok {
			continue
		}
		seen[*t] = struct{}{}
		tokens = append(tokens, *t)
	}
	return tokens, nil
}

func (s *Service) ListNotifications(ctx context.Context, req *dto.ListNotificationsReq) (*dto.ListNotificationsRes, error) {
	if req.UserID == nil {
		return nil, errs.NewError(ctx, status.UNAUTHORIZED, nil, ErrUidNotFoundFromSession)
	}
	page := int64(req.Page)
	if page <= 0 {
		page = 1
	}
	size := int64(req.Size)
	if size <= 0 {
		size = 20
	}

	notifications, pg, err := s.listQuery.Handle(ctx, query.ListNotificationsQuery{
		UserID:     *req.UserID,
		OnlyUnread: req.OnlyUnread,
		Page:       page,
		Limit:      size,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	return &dto.ListNotificationsRes{
		Notifications: dto.DomainListToResponse(notifications),
		Pagination:    pg,
	}, nil
}

func (s *Service) UnreadCount(ctx context.Context, req *dto.UnreadCountReq) (*dto.UnreadCountRes, error) {
	if req.UserID == nil {
		return nil, errs.NewError(ctx, status.UNAUTHORIZED, nil, ErrUidNotFoundFromSession)
	}
	count, err := s.unreadCountQuery.Handle(ctx, query.UnreadCountQuery{UserID: *req.UserID})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	return &dto.UnreadCountRes{Count: count}, nil
}

func (s *Service) MarkRead(ctx context.Context, req *dto.MarkReadReq) error {
	if req.UserID == nil {
		return errs.NewError(ctx, status.UNAUTHORIZED, nil, ErrUidNotFoundFromSession)
	}
	if req.NotificationID <= 0 {
		return errs.NewError(ctx, status.NOTIFICATION_NOT_FOUND, nil, ErrNotificationIDInvalid)
	}
	return s.markReadCmd.Handle(ctx, command.MarkReadCommand{
		NotificationID: req.NotificationID,
		UserID:         *req.UserID,
	})
}

func (s *Service) MarkAllRead(ctx context.Context, req *dto.MarkAllReadReq) error {
	if req.UserID == nil {
		return errs.NewError(ctx, status.UNAUTHORIZED, nil, ErrUidNotFoundFromSession)
	}
	return s.markAllReadCmd.Handle(ctx, command.MarkAllReadCommand{UserID: *req.UserID})
}

func (s *Service) SoftDelete(ctx context.Context, req *dto.DeleteNotificationReq) error {
	if req.UserID == nil {
		return errs.NewError(ctx, status.UNAUTHORIZED, nil, ErrUidNotFoundFromSession)
	}
	if req.NotificationID <= 0 {
		return errs.NewError(ctx, status.NOTIFICATION_NOT_FOUND, nil, ErrNotificationIDInvalid)
	}
	return s.softDeleteCmd.Handle(ctx, command.SoftDeleteNotificationCommand{
		NotificationID: req.NotificationID,
		UserID:         *req.UserID,
	})
}
