package repositories

import (
	"reflect"
	"testing"

	"math-ai.com/math-ai/internal/domain/device"
)

// TestBuildDeviceListFilter locks down the backward-compatible contract for
// POST /devices/list: omitting IsVerified must reproduce the exact clause/args
// this method emitted before the filter existed (just "AND d.user_id = ?"),
// and supplying it must add exactly one more bound predicate.
func TestBuildDeviceListFilter(t *testing.T) {
	trueVal := true
	falseVal := false

	tests := []struct {
		name       string
		params     *device.ListDevicesParams
		wantClause string
		wantArgs   []any
	}{
		{
			name:       "nil params produces no clause",
			params:     nil,
			wantClause: "",
			wantArgs:   nil,
		},
		{
			name:       "user id only, no verified filter — current behavior",
			params:     &device.ListDevicesParams{UserID: 42},
			wantClause: " AND d.user_id = ?",
			wantArgs:   []any{int64(42)},
		},
		{
			name:       "is_verified = true",
			params:     &device.ListDevicesParams{UserID: 42, IsVerified: &trueVal},
			wantClause: " AND d.user_id = ? AND d.is_verified = ?",
			wantArgs:   []any{int64(42), true},
		},
		{
			name:       "is_verified = false",
			params:     &device.ListDevicesParams{UserID: 42, IsVerified: &falseVal},
			wantClause: " AND d.user_id = ? AND d.is_verified = ?",
			wantArgs:   []any{int64(42), false},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			clause, args := buildDeviceListFilter(tc.params)
			if clause != tc.wantClause {
				t.Fatalf("clause = %q, want %q", clause, tc.wantClause)
			}
			if !reflect.DeepEqual(args, tc.wantArgs) {
				t.Fatalf("args = %#v, want %#v", args, tc.wantArgs)
			}
		})
	}
}
