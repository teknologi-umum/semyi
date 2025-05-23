package testutils

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

// AssertEqual checks if two values are equal
func AssertEqual(t *testing.T, expected, actual interface{}, msg string) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("%s\nExpected: %v\nActual:   %v", msg, expected, actual)
	}
}

// AssertNotEqual checks if two values are not equal
func AssertNotEqual(t *testing.T, expected, actual interface{}, msg string) {
	t.Helper()
	if reflect.DeepEqual(expected, actual) {
		t.Errorf("%s\nExpected not equal to: %v\nActual: %v", msg, expected, actual)
	}
}

// AssertNil checks if a value is nil
func AssertNil(t *testing.T, actual interface{}, msg string) {
	t.Helper()
	if actual == nil {
		return
	}
	val := reflect.ValueOf(actual)
	if val.Kind() == reflect.Ptr && val.IsNil() {
		return
	}
	t.Errorf("%s\nExpected: nil\nActual: %v", msg, actual)
}

// AssertNotNil checks if a value is not nil
func AssertNotNil(t *testing.T, actual interface{}, msg string) {
	t.Helper()
	if actual == nil {
		t.Errorf("%s\nExpected: not nil\nActual: nil", msg)
	}
}

// AssertTrue checks if a boolean value is true
func AssertTrue(t *testing.T, actual bool, msg string) {
	t.Helper()
	if !actual {
		t.Errorf("%s\nExpected: true\nActual: false", msg)
	}
}

// AssertFalse checks if a boolean value is false
func AssertFalse(t *testing.T, actual bool, msg string) {
	t.Helper()
	if actual {
		t.Errorf("%s\nExpected: false\nActual: true", msg)
	}
}

// AssertError checks if an error is not nil
func AssertError(t *testing.T, err error, msg string) {
	t.Helper()
	if err == nil {
		t.Errorf("%s\nExpected: error\nActual: nil", msg)
	}
}

// AssertNoError checks if an error is nil
func AssertNoError(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Errorf("%s\nExpected: no error\nActual: %v", msg, err)
	}
}

// AssertGreater checks if a value is greater than another
func AssertGreater(t *testing.T, expected, actual interface{}, msg string) {
	t.Helper()
	switch v := actual.(type) {
	case int:
		if v <= expected.(int) {
			t.Errorf("%s: expected greater than %v, got %v", msg, expected, actual)
		}
	case int64:
		if v <= expected.(int64) {
			t.Errorf("%s: expected greater than %v, got %v", msg, expected, actual)
		}
	case float64:
		if v <= expected.(float64) {
			t.Errorf("%s: expected greater than %v, got %v", msg, expected, actual)
		}
	case time.Duration:
		if v <= expected.(time.Duration) {
			t.Errorf("%s: expected greater than %v, got %v", msg, expected, actual)
		}
	default:
		t.Errorf("%s: unsupported type for comparison: %T", msg, actual)
	}
}

// AssertLessOrEqual checks if a value is less than or equal to another
func AssertLessOrEqual(t *testing.T, expected, actual interface{}, msg string) {
	t.Helper()
	switch v := actual.(type) {
	case int:
		if v > expected.(int) {
			t.Errorf("%s: expected less than or equal to %v, got %v", msg, expected, actual)
		}
	case int64:
		if v > expected.(int64) {
			t.Errorf("%s: expected less than or equal to %v, got %v", msg, expected, actual)
		}
	case float64:
		if v > expected.(float64) {
			t.Errorf("%s: expected less than or equal to %v, got %v", msg, expected, actual)
		}
	case time.Duration:
		if v > expected.(time.Duration) {
			t.Errorf("%s: expected less than or equal to %v, got %v", msg, expected, actual)
		}
	default:
		t.Errorf("%s: unsupported type for comparison: %T", msg, actual)
	}
}

// AssertContains checks if a string contains a substring
func AssertContains(t *testing.T, str, substr string, msg string) {
	t.Helper()
	if !strings.Contains(str, substr) {
		t.Errorf("%s\nExpected string to contain: %q\nActual: %q", msg, substr, str)
	}
}

// AssertNotEmpty checks if a string is not empty
func AssertNotEmpty(t *testing.T, actual string, msg string) {
	t.Helper()
	if actual == "" {
		t.Errorf("%s: expected non-empty string, got empty", msg)
	}
}

// AssertNotZero checks if a value is not zero
func AssertNotZero(t *testing.T, actual interface{}, msg string) {
	t.Helper()
	if actual == nil {
		t.Errorf("%s: expected non-zero value, got nil", msg)
		return
	}

	val := reflect.ValueOf(actual)
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val.Int() == 0 {
			t.Errorf("%s: expected non-zero value, got zero", msg)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if val.Uint() == 0 {
			t.Errorf("%s: expected non-zero value, got zero", msg)
		}
	case reflect.Float32, reflect.Float64:
		if val.Float() == 0 {
			t.Errorf("%s: expected non-zero value, got zero", msg)
		}
	case reflect.String:
		if val.String() == "" {
			t.Errorf("%s: expected non-zero value, got empty string", msg)
		}
	case reflect.Slice, reflect.Map, reflect.Chan:
		if val.Len() == 0 {
			t.Errorf("%s: expected non-zero value, got empty %v", msg, val.Kind())
		}
	case reflect.Ptr:
		if val.IsNil() {
			t.Errorf("%s: expected non-zero value, got nil pointer", msg)
		}
	case reflect.Interface:
		if val.IsNil() {
			t.Errorf("%s: expected non-zero value, got nil interface", msg)
		}
	default:
		zero := reflect.Zero(val.Type()).Interface()
		if reflect.DeepEqual(actual, zero) {
			t.Errorf("%s: expected non-zero value, got zero", msg)
		}
	}
}
