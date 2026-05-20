package domain

// CheckResult is one named check for an account verification run.
type CheckResult struct {
	Name    string `json:"name"`    // e.g. "credentials_present"
	Passed  bool   `json:"passed"`
	Skipped bool   `json:"skipped,omitempty"`
	Detail  string `json:"detail,omitempty"` // non-empty when failed
}

// AccountVerification is the per-account verification report.
type AccountVerification struct {
	AccountNum  int           `json:"accountNum"`
	Email       string        `json:"email"`
	DisplayName string        `json:"displayName"`
	IsActive    bool          `json:"isActive"`
	Checks      []CheckResult `json:"checks"`
	SwapReady   bool          `json:"swapReady"`
}

// VerificationReport is the full result for the UI.
type VerificationReport struct {
	Results []*AccountVerification `json:"results"`
	Total   int                    `json:"total"`
	Ready   int                    `json:"ready"`
	Failed  int                    `json:"failed"`
}
