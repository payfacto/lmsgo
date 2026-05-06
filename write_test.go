package main

import "testing"

func TestStripPreamble(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      string
		wantStrip bool
	}{
		{
			name:      "already clean Go file",
			input:     "package main\n\nfunc Add(a, b int) int {\n\treturn a + b\n}\n",
			want:      "package main\n\nfunc Add(a, b int) int {\n\treturn a + b\n}\n",
			wantStrip: false,
		},
		{
			name:      "trailing newlines normalized but no real strip",
			input:     "package main\n}\n\n\n",
			want:      "package main\n}\n",
			wantStrip: false,
		},
		{
			name:      "no trailing newline gets one added",
			input:     "package main\n}",
			want:      "package main\n}\n",
			wantStrip: false,
		},
		{
			name:      "planning preamble before package",
			input:     "The user wants a Go function.\n\nplan: write it.\n\npackage main\n\nfunc x() {}\n",
			want:      "package main\n\nfunc x() {}\n",
			wantStrip: true,
		},
		{
			name:      "preamble with no newline before marker",
			input:     "File content generation based on the request.package arithmetic\n\nfunc f() {}\n",
			want:      "package arithmetic\n\nfunc f() {}\n",
			wantStrip: true,
		},
		{
			name:      "markdown fence truncates duplicated body",
			input:     "package main\n\nfunc x() {}\n```\npackage main\n\nfunc x() {}\n",
			want:      "package main\n\nfunc x() {}\n",
			wantStrip: true,
		},
		{
			name:      "preamble plus fence both stripped",
			input:     "Plan: write it.\npackage main\nfunc x() {}\n```\npackage main\nfunc x() {}\n",
			want:      "package main\nfunc x() {}\n",
			wantStrip: true,
		},
		{
			name:      "no markers found returns unchanged content",
			input:     "just a plain text file with no code\nsecond line\n",
			want:      "just a plain text file with no code\nsecond line\n",
			wantStrip: false,
		},
		{
			name:      "marker preceded by word char does not match",
			input:     "mypackage main\nis the only line\n",
			want:      "mypackage main\nis the only line\n",
			wantStrip: false,
		},
		{
			name:      "earliest marker wins when several exist",
			input:     "preamble\nimport \"x\"\n\nfunc main() {}\n",
			want:      "import \"x\"\n\nfunc main() {}\n",
			wantStrip: true,
		},
		{
			name:      "Python script with shebang",
			input:     "Here is your script:\n#!/usr/bin/env python3\nprint('hi')\n",
			want:      "#!/usr/bin/env python3\nprint('hi')\n",
			wantStrip: true,
		},
		{
			name:      "HTML doctype",
			input:     "Generated HTML below.\n<!DOCTYPE html>\n<html></html>\n",
			want:      "<!DOCTYPE html>\n<html></html>\n",
			wantStrip: true,
		},
		{
			name:      "empty input",
			input:     "",
			want:      "\n",
			wantStrip: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotStrip := stripPreamble(tt.input)
			if got != tt.want {
				t.Errorf("stripPreamble() output\n got:  %q\n want: %q", got, tt.want)
			}
			if gotStrip != tt.wantStrip {
				t.Errorf("stripPreamble() didStrip = %v, want %v", gotStrip, tt.wantStrip)
			}
		})
	}
}

func TestIsWordByte(t *testing.T) {
	tests := []struct {
		b    byte
		want bool
	}{
		{'a', true}, {'z', true}, {'A', true}, {'Z', true},
		{'0', true}, {'9', true}, {'_', true},
		{' ', false}, {'\n', false}, {'.', false},
		{'-', false}, {'/', false}, {0, false},
	}
	for _, tt := range tests {
		if got := isWordByte(tt.b); got != tt.want {
			t.Errorf("isWordByte(%q) = %v, want %v", tt.b, got, tt.want)
		}
	}
}
