package main

import (
	"testing"
)

func TestHtmlURLPattern(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"href=\"http://example.com\"", true},
		{"href=\"https://example.com\"", true},
		{"href=\"http://example.com/path/to/resource\"", true},
		{"href=\"https://example.com/path/to/resource\"", true},
		{"href=\"/path/to/resource\"", true},
		{"href=\"path/to/resource\"", false},
		{"href=\"../path/to/resource\"", false},
		{"href=\"resource\"", false},
		{"http://example.com", false},
		{"https://example.com", false},
		{"http://example.com/path/to/resource", false},
		{"https://example.com/path/to/resource", false},
		{"/path/to/resource", false},
		{"word http://example.com word", false},
		{"word https://example.com word", false},
		{"word /path/to/resource word", false},
		{"word", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			match := urlHtmlPattern.MatchString(tt.input)
			if match != tt.expected {
				t.Errorf("urlHtmlPattern.MatchString(%q) = %v; want %v", tt.input, match, tt.expected)
			}
		})
	}
}

func TestURLListPattern(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"href=\"http://example.com\"", false},
		{"href=\"https://example.com\"", false},
		{"href=\"http://example.com/path/to/resource\"", false},
		{"href=\"https://example.com/path/to/resource\"", false},
		{"href=\"/path/to/resource\"", false},
		{"href=\"path/to/resource\"", false},
		{"href=\"../path/to/resource\"", false},
		{"href=\"resource\"", false},
		{"http://example.com", true},
		{"https://example.com", true},
		{"http://example.com/path/to/resource", true},
		{"https://example.com/path/to/resource", true},
		{"/path/to/resource", true},
		{"word http://example.com word", false},
		{"word https://example.com word", false},
		{"word /path/to/resource word", false},
		{"word", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			match := urlListPattern.MatchString(tt.input)
			if match != tt.expected {
				t.Errorf("urlListPattern.MatchString(%q) = %v; want %v", tt.input, match, tt.expected)
			}
		})
	}
}
