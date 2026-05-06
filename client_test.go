package main

import (
	"encoding/json"
	"testing"
)

func TestErrorMessage(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "empty input",
			raw:  "",
			want: "",
		},
		{
			name: "string form (LM Studio overflow)",
			raw:  `"Context length exceeded"`,
			want: "Context length exceeded",
		},
		{
			name: "object form with message field",
			raw:  `{"message":"model not found"}`,
			want: "model not found",
		},
		{
			name: "object with extra fields prefers message",
			raw:  `{"message":"bad request","code":400,"type":"invalid"}`,
			want: "bad request",
		},
		{
			name: "object without message field falls through to raw",
			raw:  `{"err":"oops"}`,
			want: `{"err":"oops"}`,
		},
		{
			name: "garbage non-json falls through to raw",
			raw:  `not json at all`,
			want: `not json at all`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := errorMessage(json.RawMessage(tt.raw))
			if got != tt.want {
				t.Errorf("errorMessage(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}
