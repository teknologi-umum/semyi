package testutils

import (
	"reflect"
	"strings"
	"testing"
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
	if actual != nil {
		t.Errorf("%s\nExpected: nil\nActual: %v", msg, actual)
	}
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
	val1 := reflect.ValueOf(expected)
	val2 := reflect.ValueOf(actual)

	if !val1.CanConvert(val2.Type()) {
		t.Errorf("%s\nCannot compare values of different types: %T and %T", msg, expected, actual)
		return
	}

	val2 = val2.Convert(val1.Type())

	switch val1.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val1.Int() >= val2.Int() {
			t.Errorf("%s\nExpected: greater than %v\nActual: %v", msg, expected, actual)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if val1.Uint() >= val2.Uint() {
			t.Errorf("%s\nExpected: greater than %v\nActual: %v", msg, expected, actual)
		}
	case reflect.Float32, reflect.Float64:
		if val1.Float() >= val2.Float() {
			t.Errorf("%s\nExpected: greater than %v\nActual: %v", msg, expected, actual)
		}
	default:
		t.Errorf("%s\nCannot compare non-numeric types: %T and %T", msg, expected, actual)
	}
}

// AssertLessOrEqual checks if a value is less than or equal to another
func AssertLessOrEqual(t *testing.T, expected, actual interface{}, msg string) {
	t.Helper()
	val1 := reflect.ValueOf(expected)
	val2 := reflect.ValueOf(actual)

	if !val1.CanConvert(val2.Type()) {
		t.Errorf("%s\nCannot compare values of different types: %T and %T", msg, expected, actual)
		return
	}

	val2 = val2.Convert(val1.Type())

	switch val1.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val1.Int() < val2.Int() {
			t.Errorf("%s\nExpected: less than or equal to %v\nActual: %v", msg, expected, actual)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if val1.Uint() < val2.Uint() {
			t.Errorf("%s\nExpected: less than or equal to %v\nActual: %v", msg, expected, actual)
		}
	case reflect.Float32, reflect.Float64:
		if val1.Float() < val2.Float() {
			t.Errorf("%s\nExpected: less than or equal to %v\nActual: %v", msg, expected, actual)
		}
	default:
		t.Errorf("%s\nCannot compare non-numeric types: %T and %T", msg, expected, actual)
	}
}

// AssertContains checks if a string contains a substring
func AssertContains(t *testing.T, str, substr string, msg string) {
	t.Helper()
	if !strings.Contains(str, substr) {
		t.Errorf("%s\nExpected string to contain: %q\nActual: %q", msg, substr, str)
	}
}
