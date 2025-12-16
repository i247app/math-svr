package convert

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ============================================================
// String to Numeric Conversions
// ============================================================

// StringToInt converts string to int with default value on error
func StringToInt(s string, defaultValue int) int {
	val, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return defaultValue
	}
	return val
}

// StringToIntErr converts string to int with error
func StringToIntErr(s string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(s))
}

// StringToInt64 converts string to int64 with default value on error
func StringToInt64(s string, defaultValue int64) int64 {
	val, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return defaultValue
	}
	return val
}

// StringToInt64Err converts string to int64 with error
func StringToInt64Err(s string) (int64, error) {
	return strconv.ParseInt(strings.TrimSpace(s), 10, 64)
}

// StringToInt32 converts string to int32 with default value on error
func StringToInt32(s string, defaultValue int32) int32 {
	val, err := strconv.ParseInt(strings.TrimSpace(s), 10, 32)
	if err != nil {
		return defaultValue
	}
	return int32(val)
}

// StringToInt32Err converts string to int32 with error
func StringToInt32Err(s string) (int32, error) {
	val, err := strconv.ParseInt(strings.TrimSpace(s), 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(val), nil
}

// StringToUint converts string to uint with default value on error
func StringToUint(s string, defaultValue uint) uint {
	val, err := strconv.ParseUint(strings.TrimSpace(s), 10, 0)
	if err != nil {
		return defaultValue
	}
	return uint(val)
}

// StringToUintErr converts string to uint with error
func StringToUintErr(s string) (uint, error) {
	val, err := strconv.ParseUint(strings.TrimSpace(s), 10, 0)
	if err != nil {
		return 0, err
	}
	return uint(val), nil
}

// StringToFloat32 converts string to float32 with default value on error
func StringToFloat32(s string, defaultValue float32) float32 {
	val, err := strconv.ParseFloat(strings.TrimSpace(s), 32)
	if err != nil {
		return defaultValue
	}
	return float32(val)
}

// StringToFloat32Err converts string to float32 with error
func StringToFloat32Err(s string) (float32, error) {
	val, err := strconv.ParseFloat(strings.TrimSpace(s), 32)
	if err != nil {
		return 0, err
	}
	return float32(val), nil
}

// StringToFloat64 converts string to float64 with default value on error
func StringToFloat64(s string, defaultValue float64) float64 {
	val, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return defaultValue
	}
	return val
}

// StringToFloat64Err converts string to float64 with error
func StringToFloat64Err(s string) (float64, error) {
	return strconv.ParseFloat(strings.TrimSpace(s), 64)
}

// ============================================================
// Numeric to String Conversions
// ============================================================

// IntToString converts int to string
func IntToString(i int) string {
	return strconv.Itoa(i)
}

// Int64ToString converts int64 to string
func Int64ToString(i int64) string {
	return strconv.FormatInt(i, 10)
}

// Int32ToString converts int32 to string
func Int32ToString(i int32) string {
	return strconv.FormatInt(int64(i), 10)
}

// UintToString converts uint to string
func UintToString(i uint) string {
	return strconv.FormatUint(uint64(i), 10)
}

// Float32ToString converts float32 to string
func Float32ToString(f float32) string {
	return strconv.FormatFloat(float64(f), 'f', -1, 32)
}

// Float32ToStringPrec converts float32 to string with precision
func Float32ToStringPrec(f float32, precision int) string {
	return strconv.FormatFloat(float64(f), 'f', precision, 32)
}

// Float64ToString converts float64 to string
func Float64ToString(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// Float64ToStringPrec converts float64 to string with precision
func Float64ToStringPrec(f float64, precision int) string {
	return strconv.FormatFloat(f, 'f', precision, 64)
}

// ============================================================
// Numeric Type Conversions
// ============================================================

// IntToFloat32 converts int to float32
func IntToFloat32(i int) float32 {
	return float32(i)
}

// IntToFloat64 converts int to float64
func IntToFloat64(i int) float64 {
	return float64(i)
}

// Int64ToFloat64 converts int64 to float64
func Int64ToFloat64(i int64) float64 {
	return float64(i)
}

// Float32ToInt converts float32 to int (truncates)
func Float32ToInt(f float32) int {
	return int(f)
}

// Float64ToInt converts float64 to int (truncates)
func Float64ToInt(f float64) int {
	return int(f)
}

// Float64ToInt64 converts float64 to int64 (truncates)
func Float64ToInt64(f float64) int64 {
	return int64(f)
}

// Float32ToInt64 converts float32 to int64 (truncates)
func Float32ToInt64(f float32) int64 {
	return int64(f)
}

// Float32ToFloat64 converts float32 to float64
func Float32ToFloat64(f float32) float64 {
	return float64(f)
}

// Float64ToFloat32 converts float64 to float32
func Float64ToFloat32(f float64) float32 {
	return float32(f)
}

// Int64ToInt converts int64 to int
func Int64ToInt(i int64) int {
	return int(i)
}

// IntToInt64 converts int to int64
func IntToInt64(i int) int64 {
	return int64(i)
}

// Int32ToInt converts int32 to int
func Int32ToInt(i int32) int {
	return int(i)
}

// IntToInt32 converts int to int32
func IntToInt32(i int) int32 {
	return int32(i)
}

// UintToInt converts uint to int
func UintToInt(u uint) int {
	return int(u)
}

// IntToUint converts int to uint (returns 0 if negative)
func IntToUint(i int) uint {
	if i < 0 {
		return 0
	}
	return uint(i)
}

// RoundFloat64 rounds float64 to nearest integer
func RoundFloat64(f float64) int {
	return int(math.Round(f))
}

// RoundFloat32 rounds float32 to nearest integer
func RoundFloat32(f float32) int {
	return int(math.Round(float64(f)))
}

// CeilFloat64 rounds float64 up to nearest integer
func CeilFloat64(f float64) int {
	return int(math.Ceil(f))
}

// FloorFloat64 rounds float64 down to nearest integer
func FloorFloat64(f float64) int {
	return int(math.Floor(f))
}

// ============================================================
// Boolean Conversions
// ============================================================

// StringToBool converts string to bool with default value on error
// Accepts: "1", "t", "T", "true", "TRUE", "True" for true
// Accepts: "0", "f", "F", "false", "FALSE", "False" for false
func StringToBool(s string, defaultValue bool) bool {
	val, err := strconv.ParseBool(strings.TrimSpace(s))
	if err != nil {
		return defaultValue
	}
	return val
}

// StringToBoolErr converts string to bool with error
func StringToBoolErr(s string) (bool, error) {
	return strconv.ParseBool(strings.TrimSpace(s))
}

// BoolToString converts bool to string
func BoolToString(b bool) string {
	return strconv.FormatBool(b)
}

// BoolToInt converts bool to int (true=1, false=0)
func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// IntToBool converts int to bool (0=false, non-zero=true)
func IntToBool(i int) bool {
	return i != 0
}

// ============================================================
// Pointer Utilities
// ============================================================

// ToIntPtr returns a pointer to the int value
func ToIntPtr(i int) *int {
	return &i
}

// ToInt64Ptr returns a pointer to the int64 value
func ToInt64Ptr(i int64) *int64 {
	return &i
}

// ToInt32Ptr returns a pointer to the int32 value
func ToInt32Ptr(i int32) *int32 {
	return &i
}

// ToFloat32Ptr returns a pointer to the float32 value
func ToFloat32Ptr(f float32) *float32 {
	return &f
}

// ToFloat64Ptr returns a pointer to the float64 value
func ToFloat64Ptr(f float64) *float64 {
	return &f
}

// ToStringPtr returns a pointer to the string value
func ToStringPtr(s string) *string {
	return &s
}

// ToBoolPtr returns a pointer to the bool value
func ToBoolPtr(b bool) *bool {
	return &b
}

// FromIntPtr returns the int value or default if pointer is nil
func FromIntPtr(ptr *int, defaultValue int) int {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}

// FromInt64Ptr returns the int64 value or default if pointer is nil
func FromInt64Ptr(ptr *int64, defaultValue int64) int64 {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}

// FromInt32Ptr returns the int32 value or default if pointer is nil
func FromInt32Ptr(ptr *int32, defaultValue int32) int32 {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}

// FromFloat32Ptr returns the float32 value or default if pointer is nil
func FromFloat32Ptr(ptr *float32, defaultValue float32) float32 {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}

// FromFloat64Ptr returns the float64 value or default if pointer is nil
func FromFloat64Ptr(ptr *float64, defaultValue float64) float64 {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}

// FromStringPtr returns the string value or default if pointer is nil
func FromStringPtr(ptr *string, defaultValue string) string {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}

// FromBoolPtr returns the bool value or default if pointer is nil
func FromBoolPtr(ptr *bool, defaultValue bool) bool {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}

// ============================================================
// Interface/Any Conversions
// ============================================================

// InterfaceToString converts interface{} to string
func InterfaceToString(v interface{}, defaultValue string) string {
	if v == nil {
		return defaultValue
	}

	switch val := v.(type) {
	case string:
		return val
	case int:
		return IntToString(val)
	case int64:
		return Int64ToString(val)
	case int32:
		return Int32ToString(val)
	case float32:
		return Float32ToString(val)
	case float64:
		return Float64ToString(val)
	case bool:
		return BoolToString(val)
	case []byte:
		return string(val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// InterfaceToInt converts interface{} to int
func InterfaceToInt(v interface{}, defaultValue int) int {
	if v == nil {
		return defaultValue
	}

	switch val := v.(type) {
	case int:
		return val
	case int64:
		return Int64ToInt(val)
	case int32:
		return Int32ToInt(val)
	case float32:
		return Float32ToInt(val)
	case float64:
		return Float64ToInt(val)
	case string:
		return StringToInt(val, defaultValue)
	case bool:
		return BoolToInt(val)
	default:
		return defaultValue
	}
}

// InterfaceToInt64 converts interface{} to int64
func InterfaceToInt64(v interface{}, defaultValue int64) int64 {
	if v == nil {
		return defaultValue
	}

	switch val := v.(type) {
	case int64:
		return val
	case int:
		return IntToInt64(val)
	case int32:
		return int64(val)
	case float32:
		return Float32ToInt64(val)
	case float64:
		return Float64ToInt64(val)
	case string:
		return StringToInt64(val, defaultValue)
	case bool:
		return int64(BoolToInt(val))
	default:
		return defaultValue
	}
}

// InterfaceToFloat64 converts interface{} to float64
func InterfaceToFloat64(v interface{}, defaultValue float64) float64 {
	if v == nil {
		return defaultValue
	}

	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return Float32ToFloat64(val)
	case int:
		return IntToFloat64(val)
	case int64:
		return Int64ToFloat64(val)
	case int32:
		return float64(val)
	case string:
		return StringToFloat64(val, defaultValue)
	default:
		return defaultValue
	}
}

// InterfaceToBool converts interface{} to bool
func InterfaceToBool(v interface{}, defaultValue bool) bool {
	if v == nil {
		return defaultValue
	}

	switch val := v.(type) {
	case bool:
		return val
	case int:
		return IntToBool(val)
	case int64:
		return IntToBool(Int64ToInt(val))
	case string:
		return StringToBool(val, defaultValue)
	default:
		return defaultValue
	}
}

// ============================================================
// Slice Conversions
// ============================================================

// IntSliceToInt64Slice converts []int to []int64
func IntSliceToInt64Slice(slice []int) []int64 {
	result := make([]int64, len(slice))
	for i, v := range slice {
		result[i] = IntToInt64(v)
	}
	return result
}

// Int64SliceToIntSlice converts []int64 to []int
func Int64SliceToIntSlice(slice []int64) []int {
	result := make([]int, len(slice))
	for i, v := range slice {
		result[i] = Int64ToInt(v)
	}
	return result
}

// IntSliceToStringSlice converts []int to []string
func IntSliceToStringSlice(slice []int) []string {
	result := make([]string, len(slice))
	for i, v := range slice {
		result[i] = IntToString(v)
	}
	return result
}

// StringSliceToIntSlice converts []string to []int (skips invalid entries)
func StringSliceToIntSlice(slice []string) []int {
	result := make([]int, 0, len(slice))
	for _, v := range slice {
		if val, err := StringToIntErr(v); err == nil {
			result = append(result, val)
		}
	}
	return result
}

// ============================================================
// Math Utilities
// ============================================================

// DegreesToRadians converts degrees to radians
func DegreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

// RadiansToDegrees converts radians to degrees
func RadiansToDegrees(radians float64) float64 {
	return radians * 180 / math.Pi
}

// ============================================================
// String Utilities
// ============================================================

// TrimSpace trims whitespace from string
func TrimSpace(str string) string {
	return strings.TrimSpace(str)
}