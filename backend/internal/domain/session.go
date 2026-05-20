package domain

// ClaudeSession is a live Claude Code process from ~/.claude/sessions/{pid}.json.
type ClaudeSession struct {
	PID        int    `json:"pid"`
	SessionID  string `json:"sessionId"`
	CWD        string `json:"cwd"`
	StartedAt  int64  `json:"startedAt"`
	Kind       string `json:"kind"`       // interactive | bg | daemon | daemon-worker
	Entrypoint string `json:"entrypoint"` // cli | claude-vscode | claude-desktop | sdk-cli | mcp
	Status     string `json:"status"`     // busy | idle | waiting
}

// IsBusy returns true if the session is actively generating or running a tool.
func (s ClaudeSession) IsBusy() bool {
	return s.Status == "busy" || s.Status == "waiting"
}

// IsInteractive returns true if this is a CLI session a user is driving.
func (s ClaudeSession) IsInteractive() bool {
	return s.Kind == "interactive"
}

// SessionReport summarises liveness for the auto-swap state machine.
type SessionReport struct {
	Total           int  `json:"total"`
	BusyOrWaiting   int  `json:"busyOrWaiting"`
	InteractiveOnly int  `json:"interactiveOnly"`
	SafeToSwap      bool `json:"safeToSwap"`
}
