package command_test

import (
	"context"
	"testing"

	command "math-ai.com/math-ai/internal/application/command/exam"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/exam"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// fakeUoW runs fn directly against the embedded repos — no real
// transaction, mirroring the hand-rolled fake style used across the
// command packages (stdlib only).
type fakeUoW struct {
	repos transaction.Repositories
}

func (f fakeUoW) Do(ctx context.Context, fn func(ctx context.Context, repos transaction.Repositories) error) error {
	return fn(ctx, f.repos)
}

// fakeUserExamRepo embeds the real interface (nil) so only the methods
// this path exercises need bodies; an accidental call to anything else
// panics loudly instead of silently no-oping.
//
// It models the one thing the real repository guarantees and the command
// leans on: MarkStatus only succeeds against an ACTIVE row.
type fakeUserExamRepo struct {
	exam.IUserExamRepository
	rows       map[int64]*exam.UserExam
	markCalls  int
	raceEnding bool // simulate another mark landing between the read and the write
}

func (f *fakeUserExamRepo) FindByUserExamId(ctx context.Context, id int64) (*exam.UserExam, error) {
	return f.rows[id], nil
}

func (f *fakeUserExamRepo) MarkStatus(ctx context.Context, id int64, newStatus string, endedDt mtime.MathTime) error {
	f.markCalls++
	row, ok := f.rows[id]
	if !ok {
		return exam.ErrJourneyNotActive
	}
	if f.raceEnding {
		// Someone else ended it after the command read it as ACTIVE.
		return exam.ErrJourneyNotActive
	}
	if s := row.UserExamStatus(); s == nil || *s != string(enum.UserExamStatusActive) {
		return exam.ErrJourneyNotActive
	}
	ns := newStatus
	row.SetUserExamStatus(&ns)
	row.SetEndedDt(endedDt)
	return nil
}

func journey(id, userID, profileID int64, st enum.UserExamStatusType) *exam.UserExam {
	j := exam.NewUserExam()
	j.SetUserExamId(id)
	j.SetUserId(userID)
	j.SetProfileId(profileID)
	j.SetReqExamType(string(enum.ExamTypePractice))
	s := string(st)
	j.SetUserExamStatus(&s)
	return j
}

func codeOf(t *testing.T, err error) status.StatusCode {
	t.Helper()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	mErr, ok := errs.IsMathError(err)
	if !ok {
		t.Fatalf("expected a MathError, got %T: %v", err, err)
	}
	return mErr.GetStatusCode()
}

func TestMarkUserExamCommand(t *testing.T) {
	const (
		journeyID = int64(9001)
		userID    = int64(1)
		profileID = int64(11)
	)

	tests := []struct {
		name      string
		seed      *exam.UserExam
		race      bool
		cmd       command.MarkUserExamCommand
		wantCode  status.StatusCode
		wantMarks int
		wantEnded enum.UserExamStatusType
	}{
		{
			name: "completes an active journey",
			seed: journey(journeyID, userID, profileID, enum.UserExamStatusActive),
			cmd: command.MarkUserExamCommand{
				UserExamID: journeyID, UserID: userID, ProfileID: profileID,
				Status: enum.UserExamStatusComplete,
			},
			wantMarks: 1,
			wantEnded: enum.UserExamStatusComplete,
		},
		{
			name: "cancels an active journey",
			seed: journey(journeyID, userID, profileID, enum.UserExamStatusActive),
			cmd: command.MarkUserExamCommand{
				UserExamID: journeyID, UserID: userID, ProfileID: profileID,
				Status: enum.UserExamStatusCancel,
			},
			wantMarks: 1,
			wantEnded: enum.UserExamStatusCancel,
		},
		{
			name: "refuses a status that is not an ending",
			seed: journey(journeyID, userID, profileID, enum.UserExamStatusActive),
			cmd: command.MarkUserExamCommand{
				UserExamID: journeyID, UserID: userID, ProfileID: profileID,
				Status: enum.UserExamStatusActive,
			},
			wantCode: status.EXAM_INVALID_JOURNEY_STATUS,
		},
		{
			name: "refuses DELETED as an ending — that is a soft delete, not a finish",
			seed: journey(journeyID, userID, profileID, enum.UserExamStatusActive),
			cmd: command.MarkUserExamCommand{
				UserExamID: journeyID, UserID: userID, ProfileID: profileID,
				Status: enum.UserExamStatusDeleted,
			},
			wantCode: status.EXAM_INVALID_JOURNEY_STATUS,
		},
		{
			name: "unknown journey",
			seed: nil,
			cmd: command.MarkUserExamCommand{
				UserExamID: journeyID, UserID: userID, ProfileID: profileID,
				Status: enum.UserExamStatusComplete,
			},
			wantCode: status.EXAM_JOURNEY_NOT_FOUND,
		},
		{
			name: "another family's journey",
			seed: journey(journeyID, userID, profileID, enum.UserExamStatusActive),
			cmd: command.MarkUserExamCommand{
				UserExamID: journeyID, UserID: userID + 1, ProfileID: profileID,
				Status: enum.UserExamStatusComplete,
			},
			wantCode: status.EXAM_JOURNEY_NOT_OWNED,
		},
		{
			name: "same user, different child",
			seed: journey(journeyID, userID, profileID, enum.UserExamStatusActive),
			cmd: command.MarkUserExamCommand{
				UserExamID: journeyID, UserID: userID, ProfileID: profileID + 1,
				Status: enum.UserExamStatusComplete,
			},
			wantCode: status.EXAM_JOURNEY_NOT_OWNED,
		},
		{
			name: "an ended journey cannot be ended again",
			seed: journey(journeyID, userID, profileID, enum.UserExamStatusComplete),
			cmd: command.MarkUserExamCommand{
				UserExamID: journeyID, UserID: userID, ProfileID: profileID,
				Status: enum.UserExamStatusCancel,
			},
			wantCode: status.EXAM_JOURNEY_ALREADY_ENDED,
		},
		{
			// The read said ACTIVE, the write found otherwise — the guard in
			// the repository's WHERE clause is what turns that into a clean
			// error instead of a second, conflicting ending.
			name: "loses a race with another mark",
			seed: journey(journeyID, userID, profileID, enum.UserExamStatusActive),
			race: true,
			cmd: command.MarkUserExamCommand{
				UserExamID: journeyID, UserID: userID, ProfileID: profileID,
				Status: enum.UserExamStatusComplete,
			},
			wantCode:  status.EXAM_JOURNEY_ALREADY_ENDED,
			wantMarks: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeUserExamRepo{rows: map[int64]*exam.UserExam{}, raceEnding: tc.race}
			if tc.seed != nil {
				repo.rows[tc.seed.UserExamId()] = tc.seed
			}
			handler := command.NewMarkUserExamCommandHandler(fakeUoW{repos: transaction.Repositories{UserExam: repo}})

			got, err := handler.Handle(context.Background(), tc.cmd)

			if tc.wantCode != 0 {
				if code := codeOf(t, err); code != tc.wantCode {
					t.Fatalf("code = %d, want %d", code, tc.wantCode)
				}
				if repo.markCalls != tc.wantMarks {
					t.Errorf("MarkStatus called %d times, want %d", repo.markCalls, tc.wantMarks)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil || got.UserExamStatus() == nil || *got.UserExamStatus() != string(tc.wantEnded) {
				t.Fatalf("journey status = %v, want %s", got.UserExamStatus(), tc.wantEnded)
			}
			if got.EndedDt().IsZero() {
				t.Error("ended_dt was not stamped")
			}
			if repo.markCalls != tc.wantMarks {
				t.Errorf("MarkStatus called %d times, want %d", repo.markCalls, tc.wantMarks)
			}
		})
	}
}
