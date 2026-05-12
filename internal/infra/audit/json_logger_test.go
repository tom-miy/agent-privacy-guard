package audit

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestJSONLoggerWrite(t *testing.T) {
	tests := []struct {
		name   string
		events []Event
	}{
		{
			name: "writes one event",
			events: []Event{
				{Action: "sanitize", Target: "claude_api", Risk: "HIGH", Allowed: false},
			},
		},
		{
			name: "appends multiple events",
			events: []Event{
				{Action: "inspect", Target: "codex", Risk: "LOW", Allowed: true},
				{Action: "posthook", Target: "codex", Risk: "HIGH", Allowed: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "nested", "audit.jsonl")
			logger := JSONLogger{Path: path}

			for _, event := range tt.events {
				if err := logger.Write(event); err != nil {
					t.Fatal(err)
				}
			}

			got := readEvents(t, path)
			if len(got) != len(tt.events) {
				t.Fatalf("event count: got %d, want %d", len(got), len(tt.events))
			}
			for i, want := range tt.events {
				if got[i].Action != want.Action {
					t.Fatalf("event %d action: got %q, want %q", i, got[i].Action, want.Action)
				}
				if got[i].Target != want.Target {
					t.Fatalf("event %d target: got %q, want %q", i, got[i].Target, want.Target)
				}
				if got[i].Risk != want.Risk {
					t.Fatalf("event %d risk: got %q, want %q", i, got[i].Risk, want.Risk)
				}
				if got[i].Allowed != want.Allowed {
					t.Fatalf("event %d allowed: got %v, want %v", i, got[i].Allowed, want.Allowed)
				}
				if got[i].Time.IsZero() {
					t.Fatalf("event %d time was not set", i)
				}
			}
		})
	}
}

func readEvents(t *testing.T, path string) []Event {
	t.Helper()

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	var events []Event
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var event Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			t.Fatal(err)
		}
		events = append(events, event)
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	return events
}
