package domain

// CommandFinding describes a risky command or patch fragment found in agent output.
// CommandFinding はエージェント出力から見つかった危険なコマンドやパッチ断片を表します。
type CommandFinding struct {
	Command  string   `json:"command"`
	Reason   string   `json:"reason"`
	Severity Severity `json:"severity"`
}
