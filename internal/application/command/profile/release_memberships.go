package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/classroom"
)

// releaseMemberships takes the profile's ACTIVE memberships out of their
// classrooms' member / student / teacher counters. It must run before
// ProfileOwnedData deletes the member rows, which are then gone from the
// active list. Pending and terminal rows were never counted.
func releaseMemberships(ctx context.Context, repos transaction.Repositories, profileId int64) error {
	members, err := repos.ClassroomMember.ListActiveByProfileIds(ctx, []int64{profileId})
	if err != nil {
		return err
	}
	for _, m := range members {
		studentDelta, teacherDelta := classroom.RoleCountDeltas(m.MemberRole(), -1)
		if err := repos.Classroom.IncCounts(ctx, m.ClassroomId(), -1, studentDelta, teacherDelta); err != nil {
			return err
		}
	}
	return nil
}
