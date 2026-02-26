package dino_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/yuppyweb/dino"
)

func TestHelper_IsStruct(t *testing.T) {
	t.Parallel()

	type Service struct {
		Name string
	}

	type StringAlias string

	testCases := []struct {
		name     string
		input    any
		expected bool
	}{
		{
			name: "Simple struct",
			input: struct {
				Name string
				Age  int
			}{},
			expected: true,
		},
		{
			name:     "Empty struct",
			input:    struct{}{},
			expected: true,
		},
		{
			name: "Struct with nested struct",
			input: struct {
				Inner struct {
					Value string
				}
			}{},
			expected: true,
		},
		{
			name:     "String is not struct",
			input:    "test",
			expected: false,
		},
		{
			name:     "Int is not struct",
			input:    42,
			expected: false,
		},
		{
			name:     "Slice is not struct",
			input:    []string{"a", "b"},
			expected: false,
		},
		{
			name:     "Map is not struct",
			input:    map[string]int{"a": 1},
			expected: false,
		},
		{
			name:     "Pointer is not struct",
			input:    &struct{}{},
			expected: false,
		},
		{
			name:     "Named struct type",
			input:    Service{},
			expected: true,
		},
		{
			name:     "Named alias of string",
			input:    StringAlias("test"),
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := dino.MockIsStruct(reflect.TypeOf(tc.input))

			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestHelper_IsPointerToStruct(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    any
		expected bool
	}{
		{
			name: "Pointer to struct",
			input: &struct {
				Name string
			}{},
			expected: true,
		},
		{
			name:     "Pointer to empty struct",
			input:    &struct{}{},
			expected: true,
		},
		{
			name: "Pointer to struct with nested pointer",
			input: &struct {
				Inner *struct {
					Value string
				}
			}{},
			expected: true,
		},
		{
			name:     "Struct is not pointer to struct",
			input:    struct{}{},
			expected: false,
		},
		{
			name:     "Pointer to string is not pointer to struct",
			input:    new("test"),
			expected: false,
		},
		{
			name:     "Pointer to int is not pointer to struct",
			input:    new(42),
			expected: false,
		},
		{
			name:     "String is not pointer to struct",
			input:    "test",
			expected: false,
		},
		{
			name:     "Slice is not pointer to struct",
			input:    []string{"a"},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := dino.MockIsPointerToStruct(reflect.TypeOf(tc.input))

			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestHelper_IsPointerToStruct_WithDoublePointer(t *testing.T) {
	t.Parallel()

	s := struct{}{}
	ps := &s
	pps := &ps

	result := dino.MockIsPointerToStruct(reflect.TypeOf(pps))

	if result {
		t.Errorf("expected false for pointer to pointer, got %v", result)
	}
}

func TestHelper_IsFunction(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    any
		expected bool
	}{
		{
			name:     "Simple function",
			input:    func() {},
			expected: true,
		},
		{
			name:     "Function with parameters",
			input:    func(int, string) string { return "" },
			expected: true,
		},
		{
			name:     "Function with return value",
			input:    func() int { return 0 },
			expected: true,
		},
		{
			name:     "Function with multiple return values",
			input:    func() (int, error) { return 0, nil },
			expected: true,
		},
		{
			name:     "String is not function",
			input:    "test",
			expected: false,
		},
		{
			name:     "Int is not function",
			input:    42,
			expected: false,
		},
		{
			name:     "Struct is not function",
			input:    struct{}{},
			expected: false,
		},
		{
			name:     "Slice is not function",
			input:    []func(){},
			expected: false,
		},
		{
			name:     "Channel is not function",
			input:    make(chan int),
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := dino.MockIsFunction(reflect.TypeOf(tc.input))

			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestHelper_IsNil(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    any
		expected bool
	}{
		{
			name:     "Nil pointer",
			input:    (*string)(nil),
			expected: true,
		},
		{
			name:     "Nil slice",
			input:    []string(nil),
			expected: true,
		},
		{
			name:     "Nil map",
			input:    map[string]int(nil),
			expected: true,
		},
		{
			name:     "Nil channel",
			input:    (chan int)(nil),
			expected: true,
		},
		{
			name:     "Nil function",
			input:    (func())(nil),
			expected: true,
		},
		{
			name:     "Non-nil pointer",
			input:    new("test"),
			expected: false,
		},
		{
			name:     "Empty slice",
			input:    []string{},
			expected: false,
		},
		{
			name:     "Empty map",
			input:    map[string]int{},
			expected: false,
		},
		{
			name:     "String is not nil",
			input:    "test",
			expected: false,
		},
		{
			name:     "Int is not nil",
			input:    42,
			expected: false,
		},
		{
			name:     "Zero value string",
			input:    "",
			expected: false,
		},
		{
			name:     "Nil interface",
			input:    (error)(nil),
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := dino.MockIsNil(reflect.ValueOf(tc.input))
			if result != tc.expected {
				t.Errorf("expected %v, got %v for input %T", tc.expected, result, tc.input)
			}
		})
	}

	result := dino.MockIsNil(reflect.Value{})
	if !result {
		t.Errorf("expected true for zero reflect.Value, got %v", result)
	}
}

func TestHelper_AsError(t *testing.T) {
	t.Parallel()

	testError := errors.New("test error")

	testCases := []struct {
		name        string
		input       any
		shouldError bool
		errorValue  error
	}{
		{
			name:        "Error interface",
			input:       testError,
			shouldError: true,
			errorValue:  testError,
		},
		{
			name:        "Nil error",
			input:       (error)(nil),
			shouldError: false,
			errorValue:  nil,
		},
		{
			name:        "Bool is not error",
			input:       true,
			shouldError: false,
			errorValue:  nil,
		},
		{
			name:        "String is not error",
			input:       "test",
			shouldError: false,
			errorValue:  nil,
		},
		{
			name:        "Int is not error",
			input:       42,
			shouldError: false,
			errorValue:  nil,
		},
		{
			name:        "Struct is not error",
			input:       struct{}{},
			shouldError: false,
			errorValue:  nil,
		},
		{
			name:        "Slice is not error",
			input:       []string{"a", "b"},
			shouldError: false,
			errorValue:  nil,
		},
		{
			name:        "Struct with error method",
			input:       &customError{"custom error"},
			shouldError: true,
			errorValue:  &customError{"custom error"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := dino.MockAsError(reflect.ValueOf(tc.input))

			if tc.shouldError {
				if result == nil {
					t.Errorf("expected error, got nil")
				}

				if result.Error() != tc.errorValue.Error() {
					t.Errorf("expected error %v, got %v", tc.errorValue, result)
				}
			} else if result != nil {
				t.Errorf("expected nil, got %v", result)
			}
		})
	}
}

type customError struct {
	message string
}

func (c *customError) Error() string {
	return c.message
}
