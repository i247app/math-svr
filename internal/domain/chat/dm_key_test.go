package chat

import "testing"

// Symmetry is the whole point of the key: if A→B and B→A produced different
// strings, the UNIQUE index would happily accept both and the two of them
// would sit in separate threads each showing half the conversation.
func TestBuildDmKeyIsSymmetric(t *testing.T) {
	tests := []struct {
		name string
		a, b int64
		want string
	}{
		{name: "ascending order", a: 7, b: 42, want: "p:7:42"},
		{name: "descending order yields the same key", a: 42, b: 7, want: "p:7:42"},
		{name: "large ids", a: 900000000002, b: 900000000001, want: "p:900000000001:900000000002"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := BuildDmKey(tc.a, tc.b); got != tc.want {
				t.Errorf("BuildDmKey(%d, %d) = %q, want %q", tc.a, tc.b, got, tc.want)
			}
			if forward, reverse := BuildDmKey(tc.a, tc.b), BuildDmKey(tc.b, tc.a); forward != reverse {
				t.Errorf("not symmetric: %q vs %q", forward, reverse)
			}
		})
	}
}

// Distinct pairs must not collide. A naive key such as concatenating the two
// ids without a separator maps (1, 23) and (12, 3) onto the same string.
func TestBuildDmKeyDistinguishesPairs(t *testing.T) {
	seen := map[string][2]int64{}
	pairs := [][2]int64{{1, 23}, {12, 3}, {1, 2}, {11, 2}, {2, 3}}

	for _, p := range pairs {
		key := BuildDmKey(p[0], p[1])
		if prev, ok := seen[key]; ok {
			t.Fatalf("key %q collides: %v and %v", key, prev, p)
		}
		seen[key] = p
	}
}
