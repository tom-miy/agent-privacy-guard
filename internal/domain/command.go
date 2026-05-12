package domain

type CommandFinding struct {
	Command  string   `json:"command"`
	Reason   string   `json:"reason"`
	Severity Severity `json:"severity"`
}
