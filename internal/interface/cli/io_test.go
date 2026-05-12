package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadInputReadsFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "prompt.txt")
	if err := os.WriteFile(path, []byte("hello from file"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := readInput(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello from file" {
		t.Fatalf("input: got %q, want %q", got, "hello from file")
	}
}

func TestOneLine(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "trims whitespace", input: "  secret  ", want: "secret"},
		{name: "escapes newlines", input: "first\nsecond", want: `first\nsecond`},
		{name: "trims then escapes newlines", input: "\nfirst\nsecond\n", want: `first\nsecond`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := oneLine(tt.input); got != tt.want {
				t.Fatalf("oneLine(%q): got %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
