package classroom

import (
	"context"

	"math-ai.com/math-ai/internal/adapter/storage"
	dto "math-ai.com/math-ai/internal/application/dto/classroom"
	classroomDomain "math-ai.com/math-ai/internal/domain/classroom"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/shared/enum"
)

// hydrateOwnersAndRelationships fills the Owner and Relationship fields
// for a page of classroom responses using batched cross-aggregate
// queries — one ListByProfileIds against ma_profiles and one
// ListByProfileAndClassroomIds against ma_classroom_members. Total
// cross-aggregate cost is +2 round trips per list call, independent of
// page size, so N+1 is structurally impossible.
//
// callerProfileID is the acting profile id from the request; it drives
// the relationship lookup. When it is zero (caller did not supply it)
// every classroom defaults to Relationship=NONE so the field is still
// stable for the frontend.
func (s *Service) hydrateOwnersAndRelationships(
	ctx context.Context,
	classrooms []*classroomDomain.Classroom,
	responses []*dto.ClassroomResponse,
	callerProfileID int64,
) error {
	if len(classrooms) == 0 {
		return nil
	}

	ownerIDSet := make(map[int64]struct{}, len(classrooms))
	classroomIDs := make([]int64, 0, len(classrooms))
	for _, c := range classrooms {
		ownerIDSet[c.OwnerProfileId()] = struct{}{}
		classroomIDs = append(classroomIDs, c.ClassroomId())
	}
	ownerIDs := make([]int64, 0, len(ownerIDSet))
	for id := range ownerIDSet {
		ownerIDs = append(ownerIDs, id)
	}

	ownerMap := make(map[int64]*dto.ClassroomOwnerSummary, len(ownerIDs))
	if len(ownerIDs) > 0 && s.profileRepo != nil {
		owners, err := s.profileRepo.ListByProfileIds(ctx, ownerIDs)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		for _, p := range owners {
			summary := &dto.ClassroomOwnerSummary{
				ProfileID: p.ProfileId(),
				Name:      p.Name(),
				Role:      p.Role(),
				AvatarKey: p.AvatarKey(),
			}
			s.signOwnerAvatarURL(ctx, summary)
			ownerMap[p.ProfileId()] = summary
		}
	}

	memberMap := make(map[int64]*classroomDomain.Member, len(classroomIDs))
	if callerProfileID != 0 && s.classroomMemberRepo != nil {
		rows, err := s.classroomMemberRepo.ListByProfileAndClassroomIds(ctx, callerProfileID, classroomIDs)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		for _, m := range rows {
			memberMap[m.ClassroomId()] = m
		}
	}

	for i, c := range classrooms {
		resp := responses[i]
		if resp == nil {
			continue
		}
		if owner, ok := ownerMap[c.OwnerProfileId()]; ok {
			resp.Owner = owner
		}
		rel, role := relationshipFromMember(memberMap[c.ClassroomId()])
		resp.Relationship = string(rel)
		resp.MyRole = role
	}
	return nil
}

// relationshipFromMember maps a member row's status into the public
// relationship enum. Terminal states (REJECTED/LEFT/REMOVED) and a
// missing row both flatten to NONE — the UI treats them identically as
// "not currently participating". Returns the caller's member_role only
// when the relationship is MEMBER; nil otherwise so the JSON omits it.
func relationshipFromMember(m *classroomDomain.Member) (enum.ClassroomRelationshipType, *string) {
	if m == nil || m.MemberStatus() == nil {
		return enum.ClassroomRelationshipTypeNone, nil
	}
	switch enum.ClassroomMemberStatusType(*m.MemberStatus()) {
	case enum.ClassroomMemberStatusTypeActive:
		role := m.MemberRole()
		return enum.ClassroomRelationshipTypeMember, &role
	case enum.ClassroomMemberStatusTypePendingInvitation:
		return enum.ClassroomRelationshipTypePendingInvitation, nil
	case enum.ClassroomMemberStatusTypePendingRequest:
		return enum.ClassroomRelationshipTypePendingRequest, nil
	default:
		return enum.ClassroomRelationshipTypeNone, nil
	}
}

// hydratePendingRequestCounts groups the PENDING_REQUEST member rows
// for the given classroom ids in one round trip and writes the result
// onto each response's PendingRequestCount field. Classrooms with zero
// pending requests are absent from the repo map and default to 0.
func (s *Service) hydratePendingRequestCounts(
	ctx context.Context,
	classrooms []*classroomDomain.Classroom,
	responses []*dto.ClassroomResponse,
) error {
	if len(classrooms) == 0 || s.classroomMemberRepo == nil {
		return nil
	}
	ids := make([]int64, 0, len(classrooms))
	for _, c := range classrooms {
		ids = append(ids, c.ClassroomId())
	}
	counts, err := s.classroomMemberRepo.CountPendingRequestsByClassroomIds(ctx, ids)
	if err != nil {
		return errs.NewError(ctx, status.FAIL, nil, err)
	}
	for i, c := range classrooms {
		if responses[i] == nil {
			continue
		}
		responses[i].PendingRequestCount = counts[c.ClassroomId()]
	}
	return nil
}

// signOwnerAvatarURL mutates the owner summary in place to add a
// short-lived presigned URL for the avatar. No-op when storage is
// disabled or the owner has no avatar_key.
func (s *Service) signOwnerAvatarURL(ctx context.Context, summary *dto.ClassroomOwnerSummary) {
	if summary == nil || s.storageProvider == nil || summary.AvatarKey == nil || *summary.AvatarKey == "" {
		return
	}
	url, err := s.storageProvider.CreatePresignedUrl(ctx, &storage.CreatePresignedUrlRequest{
		Key:        *summary.AvatarKey,
		Expiration: coverUrlTTL,
	})
	if err != nil {
		logger.From(ctx).Warnf("classroom.owner_avatar presign failed profile_id=%d err=%v", summary.ProfileID, err)
		return
	}
	summary.AvatarURL = &url
}

// populateCoverUrl mutates resp in place to add a short-lived presigned
// URL when the classroom carries a cover_key. No-op if storage is
// disabled or the classroom has no cover.
func (s *Service) populateCoverUrl(ctx context.Context, resp *dto.ClassroomResponse) {
	if resp == nil || s.storageProvider == nil || resp.CoverKey == nil || *resp.CoverKey == "" {
		return
	}
	url, err := s.storageProvider.CreatePresignedUrl(ctx, &storage.CreatePresignedUrlRequest{
		Key:        *resp.CoverKey,
		Expiration: coverUrlTTL,
	})
	if err != nil {
		logger.From(ctx).Warnf("classroom.cover presign failed classroom_id=%s err=%v", resp.ClassroomID, err)
		return
	}
	resp.CoverURL = &url
}
