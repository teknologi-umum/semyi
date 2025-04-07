package testutils

import (
	"errors"
	"testing"
	"time"
)

func TestAssertEqual(t *testing.T) {
	t.Run("EqualValues", func(t *testing.T) {
		AssertEqual(t, 42, 42, "Values should be equal")
		AssertEqual(t, "hello", "hello", "Strings should be equal")
		AssertEqual(t, []int{1, 2, 3}, []int{1, 2, 3}, "Slices should be equal")
	})

	t.Run("NotEqualValues", func(t *testing.T) {
		tt := &testing.T{}
		AssertEqual(tt, 42, 43, "Values should not be equal")
		if !tt.Failed() {
			t.Error("Expected test to fail")
		}
	})
}

func TestAssertNotEqual(t *testing.T) {
	t.Run("NotEqualValues", func(t *testing.T) {
		AssertNotEqual(t, 42, 43, "Values should not be equal")
		AssertNotEqual(t, "hello", "world", "Strings should not be equal")
	})

	t.Run("EqualValues", func(t *testing.T) {
		tt := &testing.T{}
		AssertNotEqual(tt, 42, 42, "Values should be equal")
		if !tt.Failed() {
			t.Error("Expected test to fail")
		}
	})
}

func TestAssertNil(t *testing.T) {
	t.Run("NilValue", func(t *testing.T) {
		AssertNil(t, nil, "Value should be nil")
		var s *string
		AssertNil(t, s, "Pointer should be nil")
		var i interface{}
		AssertNil(t, i, "Interface should be nil")
	})

	t.Run("NonNilValue", func(t *testing.T) {
		tt := &testing.T{}
		AssertNil(tt, "not nil", "Value should not be nil")
		if !tt.Failed() {
			t.Error("Expected test to fail")
		}
	})
}

func TestAssertNotNil(t *testing.T) {
	t.Run("NonNilValue", func(t *testing.T) {
		AssertNotNil(t, "not nil", "Value should not be nil")
		s := "hello"
		AssertNotNil(t, &s, "Pointer should not be nil")
	})

	t.Run("NilValue", func(t *testing.T) {
		tt := &testing.T{}
		AssertNotNil(tt, nil, "Value should be nil")
		if !tt.Failed() {
			t.Error("Expected test to fail")
		}
	})
}

func TestAssertTrue(t *testing.T) {
	t.Run("TrueValue", func(t *testing.T) {
		AssertTrue(t, true, "Value should be true")
	})

	t.Run("FalseValue", func(t *testing.T) {
		tt := &testing.T{}
		AssertTrue(tt, false, "Value should be false")
		if !tt.Failed() {
			t.Error("Expected test to fail")
		}
	})
}

func TestAssertFalse(t *testing.T) {
	t.Run("FalseValue", func(t *testing.T) {
		AssertFalse(t, false, "Value should be false")
	})

	t.Run("TrueValue", func(t *testing.T) {
		tt := &testing.T{}
		AssertFalse(tt, true, "Value should be true")
		if !tt.Failed() {
			t.Error("Expected test to fail")
		}
	})
}

func TestAssertError(t *testing.T) {
	t.Run("ErrorValue", func(t *testing.T) {
		err := errors.New("test error")
		AssertError(t, err, "Value should be an error")
	})

	t.Run("NoError", func(t *testing.T) {
		tt := &testing.T{}
		AssertError(tt, nil, "Value should not be an error")
		if !tt.Failed() {
			t.Error("Expected test to fail")
		}
	})
}

func TestAssertNoError(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		AssertNoError(t, nil, "Value should not be an error")
	})

	t.Run("ErrorValue", func(t *testing.T) {
		tt := &testing.T{}
		err := errors.New("test error")
		AssertNoError(tt, err, "Value should be an error")
		if !tt.Failed() {
			t.Error("Expected test to fail")
		}
	})
}

func TestAssertGreater(t *testing.T) {
	t.Run("GreaterValues", func(t *testing.T) {
		AssertGreater(t, 41, 42, "Value should be greater")
		AssertGreater(t, 41.0, 42.0, "Float should be greater")
		AssertGreater(t, time.Second, time.Minute, "Duration should be greater")
	})

	t.Run("NotGreaterValues", func(t *testing.T) {
		tt := &testing.T{}
		AssertGreater(tt, 42, 42, "Value should not be greater")
		if !tt.Failed() {
			t.Error("Expected test to fail")
		}
	})

	t.Run("UnsupportedType", func(t *testing.T) {
		tt := &testing.T{}
		AssertGreater(tt, "not a number", "not a number", "Should not support string comparison")
		if !tt.Failed() {
			t.Error("Expected test to fail")
		}
	})
}

func TestAssertLessOrEqual(t *testing.T) {
	t.Run("LessOrEqualValues", func(t *testing.T) {
		AssertLessOrEqual(t, 42, 42, "Value should be equal")
		AssertLessOrEqual(t, 43, 42, "Value should be less")
		AssertLessOrEqual(t, 42.0, 42.0, "Float should be equal")
		AssertLessOrEqual(t, time.Minute, time.Second, "Duration should be less")
	})

	t.Run("GreaterValues", func(t *testing.T) {
		tt := &testing.T{}
		AssertLessOrEqual(tt, 41, 42, "Value should be greater")
		if !tt.Failed() {
			t.Error("Expected test to fail")
		}
	})

	t.Run("UnsupportedType", func(t *testing.T) {
		tt := &testing.T{}
		AssertLessOrEqual(tt, "not a number", "not a number", "Should not support string comparison")
		if !tt.Failed() {
			t.Error("Expected test to fail")
		}
	})
}

func TestAssertContains(t *testing.T) {
	t.Run("ContainsSubstring", func(t *testing.T) {
		AssertContains(t, "hello world", "world", "String should contain substring")
	})

	t.Run("DoesNotContain", func(t *testing.T) {
		tt := &testing.T{}
		AssertContains(tt, "hello", "world", "String should not contain substring")
		if !tt.Failed() {
			t.Error("Expected test to fail")
		}
	})
}

func TestAssertNotEmpty(t *testing.T) {
	t.Run("NonEmptyString", func(t *testing.T) {
		AssertNotEmpty(t, "hello", "String should not be empty")
	})

	t.Run("EmptyString", func(t *testing.T) {
		tt := &testing.T{}
		AssertNotEmpty(tt, "", "String should be empty")
		if !tt.Failed() {
			t.Error("Expected test to fail")
		}
	})
}

func TestAssertNotZero(t *testing.T) {
	t.Run("NonZeroValues", func(t *testing.T) {
		AssertNotZero(t, 42, "Value should not be zero")
		AssertNotZero(t, "hello", "String should not be zero")
		AssertNotZero(t, []int{1, 2, 3}, "Slice should not be zero")
		AssertNotZero(t, time.Second, "Duration should not be zero")
	})

	t.Run("ZeroValues", func(t *testing.T) {
		tt := &testing.T{}
		AssertNotZero(tt, 0, "Value should be zero")
		if !tt.Failed() {
			t.Error("Expected test to fail")
		}

		tt = &testing.T{}
		AssertNotZero(tt, "", "String should be zero")
		if !tt.Failed() {
			t.Error("Expected test to fail")
		}

		tt = &testing.T{}
		AssertNotZero(tt, []int{}, "Slice should be zero")
		if !tt.Failed() {
			t.Error("Expected test to fail")
		}

		tt = &testing.T{}
		AssertNotZero(tt, time.Duration(0), "Duration should be zero")
		if !tt.Failed() {
			t.Error("Expected test to fail")
		}
	})
}
