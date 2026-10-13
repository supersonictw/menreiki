package utils

import (
	"reflect"
	"testing"
)

func TestSliceContent(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		maxLength int
		sep       string
		expected  []string
	}{
		{
			name:      "empty content",
			content:   "",
			maxLength: 10,
			sep:       "\n",
			expected:  []string{},
		},
		{
			name:      "zero or negative max length",
			content:   "hello",
			maxLength: 0,
			sep:       "\n",
			expected:  []string{},
		},
		{
			name:      "short content fits",
			content:   "hello world",
			maxLength: 20,
			sep:       "\n",
			expected:  []string{"hello world"},
		},
		{
			name:      "multiline content fits",
			content:   "line 1\nline 2",
			maxLength: 20,
			sep:       "\n",
			expected:  []string{"line 1\nline 2"},
		},
		{
			name:      "multiline split across limit",
			content:   "first line\nsecond line",
			maxLength: 12,
			sep:       "\n",
			expected:  []string{"first line", "second line"},
		},
		{
			name:      "single line exceeds limit",
			content:   "1234567890abcdefghij",
			maxLength: 10,
			sep:       "\n",
			expected:  []string{"1234567890", "abcdefghij"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SliceContent(tt.content, tt.maxLength, tt.sep)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("SliceContent() = %v, want %v", got, tt.expected)
			}
		})
	}
}
