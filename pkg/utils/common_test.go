package utils

import (
	"context"
	"testing"
	"time"
)

func TestGetStringOrDefault(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		def      string
		expected string
	}{
		{
			name:     "non-empty value",
			value:    "test",
			def:      "default",
			expected: "test",
		},
		{
			name:     "empty value",
			value:    "",
			def:      "default",
			expected: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetStringOrDefault(tt.value, tt.def)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGetBoolOrDefault(t *testing.T) {
	tests := []struct {
		name     string
		value    bool
		def      bool
		expected bool
	}{
		{
			name:     "true value",
			value:    true,
			def:      false,
			expected: true,
		},
		{
			name:     "false value",
			value:    false,
			def:      true,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetBoolOrDefault(tt.value, tt.def)
			if result != tt.expected {
				t.Errorf("Expected %t, got %t", tt.expected, result)
			}
		})
	}
}

func TestStringToIntOrDefault(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		def      int
		expected int
	}{
		{
			name:     "valid integer",
			input:    "42",
			def:      0,
			expected: 42,
		},
		{
			name:     "invalid integer",
			input:    "not-a-number",
			def:      10,
			expected: 10,
		},
		{
			name:     "empty string",
			input:    "",
			def:      5,
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringToIntOrDefault(tt.input, tt.def)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestStringToBoolOrDefault(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		def      bool
		expected bool
	}{
		{
			name:     "true string",
			input:    "true",
			def:      false,
			expected: true,
		},
		{
			name:     "false string",
			input:    "false",
			def:      true,
			expected: false,
		},
		{
			name:     "invalid string",
			input:    "maybe",
			def:      true,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringToBoolOrDefault(tt.input, tt.def)
			if result != tt.expected {
				t.Errorf("Expected %t, got %t", tt.expected, result)
			}
		})
	}
}

func TestRunWithTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// 성공하는 함수
	err := RunWithTimeout(ctx, func() error {
		time.Sleep(50 * time.Millisecond)
		return nil
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// 타임아웃되는 함수
	ctx2, cancel2 := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel2()

	err = RunWithTimeout(ctx2, func() error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})
	if err == nil {
		t.Error("Expected timeout error, got none")
	}
}
