package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Event is one audit record emitted by inspect, sanitize, or posthook actions.
// Event は inspect、sanitize、posthook が出力する監査レコードです。
type Event struct {
	Time    time.Time   `json:"time"`
	Action  string      `json:"action"`
	Target  string      `json:"target,omitempty"`
	Risk    string      `json:"risk,omitempty"`
	Allowed bool        `json:"allowed"`
	Details interface{} `json:"details,omitempty"`
}

// JSONLogger appends audit events as JSON Lines to Path.
// JSONLogger は監査イベントを JSON Lines 形式で Path に追記します。
type JSONLogger struct {
	Path string
}

// Write appends event to the audit log, creating parent directories when needed.
// Write は必要に応じて親ディレクトリを作成し、イベントを監査ログへ追記します。
func (l JSONLogger) Write(event Event) error {
	if l.Path == "" {
		l.Path = "audit/agent-privacy-guard.jsonl"
	}
	event.Time = time.Now().UTC()
	if err := os.MkdirAll(filepath.Dir(l.Path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(l.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	return enc.Encode(event)
}
