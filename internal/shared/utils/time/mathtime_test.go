package time

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNow(t *testing.T) {
	mt := Now()
	if mt.IsZero() {
		t.Error("Now() should not return zero time")
	}
}

func TestNewMathTime(t *testing.T) {
	now := time.Now()
	mt := NewMathTime(now)

	if !mt.Equal(now) {
		t.Errorf("NewMathTime() = %v, want %v", mt, now)
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid RFC3339",
			input:   "2023-12-19T10:30:00Z",
			wantErr: false,
		},
		{
			name:    "valid RFC3339 with timezone",
			input:   "2023-12-19T10:30:00+07:00",
			wantErr: false,
		},
		{
			name:    "invalid format",
			input:   "2023-12-19",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && mt.IsZero() {
				t.Error("Parse() returned zero time for valid input")
			}
		})
	}
}

func TestParseDate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid date",
			input:   "2023-12-19",
			wantErr: false,
		},
		{
			name:    "invalid date",
			input:   "2023-12-19T10:30:00Z",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt, err := ParseDate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && mt.IsZero() {
				t.Error("ParseDate() returned zero time for valid input")
			}
		})
	}
}

func TestParseWithFormat(t *testing.T) {
	mt, err := ParseWithFormat(time.DateOnly, "2023-12-19")
	if err != nil {
		t.Errorf("ParseWithFormat() error = %v", err)
	}
	if mt.IsZero() {
		t.Error("ParseWithFormat() returned zero time")
	}
}

func TestMarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		mathTime MathTime
		want     string
	}{
		{
			name:     "zero time",
			mathTime: MathTime{},
			want:     "null",
		},
		{
			name:     "valid time",
			mathTime: MathTime{Time: time.Date(2023, 12, 19, 10, 30, 0, 0, time.UTC)},
			want:     `"2023-12-19T10:30:00Z"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.mathTime)
			if err != nil {
				t.Errorf("MarshalJSON() error = %v", err)
				return
			}
			if string(got) != tt.want {
				t.Errorf("MarshalJSON() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		isZero  bool
	}{
		{
			name:    "null",
			input:   "null",
			wantErr: false,
			isZero:  true,
		},
		{
			name:    "empty string",
			input:   `""`,
			wantErr: false,
			isZero:  true,
		},
		{
			name:    "valid RFC3339",
			input:   `"2023-12-19T10:30:00Z"`,
			wantErr: false,
			isZero:  false,
		},
		{
			name:    "invalid format",
			input:   `"2023-12-19"`,
			wantErr: true,
			isZero:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mt MathTime
			err := json.Unmarshal([]byte(tt.input), &mt)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && mt.IsZero() != tt.isZero {
				t.Errorf("UnmarshalJSON() isZero = %v, want %v", mt.IsZero(), tt.isZero)
			}
		})
	}
}

func TestJSONRoundTrip(t *testing.T) {
	original := MathTime{Time: time.Date(2023, 12, 19, 10, 30, 0, 0, time.UTC)}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded MathTime
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if !original.Equal(decoded.Time) {
		t.Errorf("Round trip failed: got %v, want %v", decoded, original)
	}
}

func TestValue(t *testing.T) {
	tests := []struct {
		name     string
		mathTime MathTime
		wantNil  bool
	}{
		{
			name:     "zero time",
			mathTime: MathTime{},
			wantNil:  true,
		},
		{
			name:     "valid time",
			mathTime: MathTime{Time: time.Date(2023, 12, 19, 10, 30, 0, 0, time.UTC)},
			wantNil:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := tt.mathTime.Value()
			if err != nil {
				t.Errorf("Value() error = %v", err)
				return
			}
			if (val == nil) != tt.wantNil {
				t.Errorf("Value() nil = %v, want nil %v", val == nil, tt.wantNil)
			}
		})
	}
}

func TestScan(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
		isZero  bool
	}{
		{
			name:    "nil",
			input:   nil,
			wantErr: false,
			isZero:  true,
		},
		{
			name:    "time.Time",
			input:   now,
			wantErr: false,
			isZero:  false,
		},
		{
			name:    "[]byte RFC3339",
			input:   []byte("2023-12-19T10:30:00Z"),
			wantErr: false,
			isZero:  false,
		},
		{
			name:    "string RFC3339",
			input:   "2023-12-19T10:30:00Z",
			wantErr: false,
			isZero:  false,
		},
		{
			name:    "invalid []byte",
			input:   []byte("invalid"),
			wantErr: true,
			isZero:  false,
		},
		{
			name:    "invalid type",
			input:   123,
			wantErr: true,
			isZero:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mt MathTime
			err := mt.Scan(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && mt.IsZero() != tt.isZero {
				t.Errorf("Scan() isZero = %v, want %v", mt.IsZero(), tt.isZero)
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		name     string
		mathTime MathTime
		wantStr  string
	}{
		{
			name:     "zero time",
			mathTime: MathTime{},
			wantStr:  "",
		},
		{
			name:     "valid time",
			mathTime: MathTime{Time: time.Date(2023, 12, 19, 10, 30, 0, 0, time.UTC)},
			wantStr:  "2023-12-19T10:30:00Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.mathTime.String()
			if got != tt.wantStr {
				t.Errorf("String() = %v, want %v", got, tt.wantStr)
			}
		})
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		name     string
		mathTime MathTime
		want     bool
	}{
		{
			name:     "zero time is invalid",
			mathTime: MathTime{},
			want:     false,
		},
		{
			name:     "non-zero time is valid",
			mathTime: MathTime{Time: time.Date(2023, 12, 19, 10, 30, 0, 0, time.UTC)},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.mathTime.IsValid()
			if got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPtr(t *testing.T) {
	mt := MathTime{Time: time.Date(2023, 12, 19, 10, 30, 0, 0, time.UTC)}
	ptr := mt.Ptr()

	if ptr == nil {
		t.Error("Ptr() returned nil")
	}
	if !ptr.Equal(mt.Time) {
		t.Errorf("Ptr() time mismatch: got %v, want %v", ptr.Time, mt.Time)
	}
}

func TestNewMathTimePtr(t *testing.T) {
	tests := []struct {
		name    string
		input   time.Time
		wantNil bool
	}{
		{
			name:    "zero time",
			input:   time.Time{},
			wantNil: true,
		},
		{
			name:    "valid time",
			input:   time.Date(2023, 12, 19, 10, 30, 0, 0, time.UTC),
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewMathTimePtr(tt.input)
			if (got == nil) != tt.wantNil {
				t.Errorf("NewMathTimePtr() nil = %v, want nil %v", got == nil, tt.wantNil)
			}
		})
	}
}

func TestMathTimeFromPtr(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		input   *time.Time
		isZero  bool
	}{
		{
			name:   "nil pointer",
			input:  nil,
			isZero: true,
		},
		{
			name:   "valid pointer",
			input:  &now,
			isZero: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MathTimeFromPtr(tt.input)
			if got.IsZero() != tt.isZero {
				t.Errorf("MathTimeFromPtr() isZero = %v, want %v", got.IsZero(), tt.isZero)
			}
		})
	}
}

func TestTimeToMathTimePtr(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		input   *time.Time
		wantNil bool
	}{
		{
			name:    "nil pointer",
			input:   nil,
			wantNil: true,
		},
		{
			name:    "valid pointer",
			input:   &now,
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TimeToMathTimePtr(tt.input)
			if (got == nil) != tt.wantNil {
				t.Errorf("TimeToMathTimePtr() nil = %v, want nil %v", got == nil, tt.wantNil)
			}
		})
	}
}

func TestMathTimePtrToTime(t *testing.T) {
	mt := MathTime{Time: time.Date(2023, 12, 19, 10, 30, 0, 0, time.UTC)}

	tests := []struct {
		name    string
		input   *MathTime
		wantNil bool
	}{
		{
			name:    "nil pointer",
			input:   nil,
			wantNil: true,
		},
		{
			name:    "valid pointer",
			input:   &mt,
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MathTimePtrToTime(tt.input)
			if (got == nil) != tt.wantNil {
				t.Errorf("MathTimePtrToTime() nil = %v, want nil %v", got == nil, tt.wantNil)
			}
		})
	}
}

func TestToTime(t *testing.T) {
	now := time.Now()
	mt := MathTime{Time: now}

	got := mt.ToTime()
	if !got.Equal(now) {
		t.Errorf("ToTime() = %v, want %v", got, now)
	}
}
