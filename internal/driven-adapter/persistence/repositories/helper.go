package repositories

import (
	"reflect"
	"time"
)

func PrepareForUpdate(value interface{}) interface{} {
	if value == nil {
		return nil
	}

	v := reflect.ValueOf(value)

	// If it's already a pointer, check if it's nil or points to zero value
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		// Dereference and check the underlying value
		v = v.Elem()
	}

	// Check if the value is zero for its type
	if isZeroValue(v) {
		return nil
	}

	return value
}

func PrepareForUpdateWithDefault(value interface{}, defaultValue interface{}) interface{} {
	if value == nil {
		return defaultValue
	}

	v := reflect.ValueOf(value)

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return defaultValue
		}
		v = v.Elem()
	}

	if isZeroValue(v) {
		return defaultValue
	}

	return value
}

// isZeroValue checks if a reflect.Value is a zero value for its type
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0

	case reflect.Float32, reflect.Float64:
		return v.Float() == 0.0

	case reflect.Bool:
		return v.Bool() == false

	case reflect.Struct:
		// Special handling for time.Time
		if v.Type() == reflect.TypeOf(time.Time{}) {
			t := v.Interface().(time.Time)
			return t.IsZero()
		}
		// For other structs, check if all fields are zero
		return v.IsZero()

	case reflect.Ptr, reflect.Interface:
		return v.IsNil()

	case reflect.Slice, reflect.Map, reflect.Array:
		return v.Len() == 0

	case reflect.Chan, reflect.Func:
		return v.IsNil()

	default:
		return v.IsZero()
	}
}
