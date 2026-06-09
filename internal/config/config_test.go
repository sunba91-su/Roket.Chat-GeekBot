package config

import (
	"testing"
)

func TestTrimQuotes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"'password'", "password"},
		{`"password"`, "password"},
		{`'p@ss(w0rd)!'`, "p@ss(w0rd)!"},
		{`"p@ss(w0rd)!"`, "p@ss(w0rd)!"},
		{"noquotes", "noquotes"},
		{"", ""},
		{"'", "'"},
		{`"`, `"`},
		{"''", ""},
		{`""`, ""},
		{"a", "a"},
	}
	for _, tt := range tests {
		got := trimQuotes(tt.input)
		if got != tt.expected {
			t.Errorf("trimQuotes(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
