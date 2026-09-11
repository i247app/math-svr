package enum

import "testing"

// TestUserExamStatusIsEnding pins which statuses a client may end a
// journey with. DELETED is the trap: it is a valid lifecycle value, so a
// check written as IsValid() would let a client "finish" a journey by
// soft-deleting it.
func TestUserExamStatusIsEnding(t *testing.T) {
	tests := []struct {
		status UserExamStatusType
		want   bool
	}{
		{UserExamStatusComplete, true},
		{UserExamStatusCancel, true},
		{UserExamStatusActive, false},
		{UserExamStatusDeleted, false},
		{UserExamStatusType("DONE"), false},
		{UserExamStatusType(""), false},
	}
	for _, tc := range tests {
		if got := tc.status.IsEnding(); got != tc.want {
			t.Errorf("%q.IsEnding() = %v, want %v", tc.status, got, tc.want)
		}
	}
}
