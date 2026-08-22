// Package presence is the internal façade over realtime availability. It has
// no handler and no route: the only writer is the socket runtime, and the only
// reader is the chat module's member list. Shaped like module/seq.
package presence

import (
	"context"

	command "math-ai.com/math-ai/internal/application/command/presence"
	"math-ai.com/math-ai/internal/application/transaction"
	classroomDomain "math-ai.com/math-ai/internal/domain/classroom"
	domain "math-ai.com/math-ai/internal/domain/presence"
	"math-ai.com/math-ai/internal/infrastructure/logger"
)

type Service struct {
	markOnlineCmd  *command.MarkOnlineCommandHandler
	markOfflineCmd *command.MarkOfflineCommandHandler
	resetAllCmd    *command.ResetAllCommandHandler
	repo           domain.IRepository

	// Fan-out collaborators. Both may be nil (realtime disabled), in which
	// case presence is still recorded and simply not announced.
	classroomMemberRepo classroomDomain.IMemberRepository
	publisher           Publisher
	debounce            *offlineDebouncer
}

func NewService(
	uow transaction.UnitOfWork,
	repo domain.IRepository,
	classroomMemberRepo classroomDomain.IMemberRepository,
	publisher Publisher,
) *Service {
	return &Service{
		markOnlineCmd:       command.NewMarkOnlineCommandHandler(uow),
		markOfflineCmd:      command.NewMarkOfflineCommandHandler(uow),
		resetAllCmd:         command.NewResetAllCommandHandler(uow),
		repo:                repo,
		classroomMemberRepo: classroomMemberRepo,
		publisher:           publisher,
		debounce:            newOfflineDebouncer(defaultOfflineDelay),
	}
}

// MarkOnline records one new live connection. It returns whether this was the
// offline→online transition so the caller can broadcast only on a real change
// rather than on every reconnect.
func (s *Service) MarkOnline(ctx context.Context, userId int64, deviceUuid, platform *string) (becameOnline bool, err error) {
	p, err := s.markOnlineCmd.Handle(ctx, &command.MarkOnlineCommand{
		UserId:     userId,
		DeviceUuid: deviceUuid,
		Platform:   platform,
	})
	if err != nil {
		return false, err
	}

	becameOnline = p != nil && p.ConnectionCount() == 1
	if becameOnline {
		s.announceOnline(ctx, userId)
	}
	return becameOnline, nil
}

// MarkOffline removes one live connection, reporting whether the user's last
// device just went away.
func (s *Service) MarkOffline(ctx context.Context, userId int64) (becameOffline bool, err error) {
	p, err := s.markOfflineCmd.Handle(ctx, &command.MarkOfflineCommand{UserId: userId})
	if err != nil {
		return false, err
	}

	becameOffline = p != nil && p.ConnectionCount() == 0
	if becameOffline {
		// Scheduled, not sent: a reconnect within the window cancels it.
		s.announceOffline(ctx, userId)
	}
	return becameOffline, nil
}

// ListByUserIds is the batch read behind a member list. Users absent from the
// map have never connected and must be rendered OFFLINE.
func (s *Service) ListByUserIds(ctx context.Context, userIds []int64) (map[int64]*domain.Presence, error) {
	return s.repo.ListByUserIds(ctx, userIds)
}

// ResetAll clears counters left over from the previous process. Called from
// bootstrap before the server accepts connections.
func (s *Service) ResetAll(ctx context.Context) error {
	if err := s.resetAllCmd.Handle(ctx); err != nil {
		logger.From(ctx).Warnf("presence.reset_failed err=%v", err)
		return err
	}
	return nil
}
