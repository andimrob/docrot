package freshness

type Status string

const (
	StatusFresh              Status = "fresh"
	StatusStale              Status = "stale"
	StatusMissingFrontmatter Status = "missing_frontmatter"
)

type Result struct {
	Path         string   `json:"path"`
	Status       Status   `json:"status"`
	Strategy     string   `json:"strategy,omitempty"`
	LastReviewed string   `json:"last_reviewed,omitempty"`
	StaleSince   string   `json:"stale_since,omitempty"`
	Reason       string   `json:"reason,omitempty"`
	Expires      string   `json:"expires,omitempty"`
	ChangedFiles []string `json:"changed_files,omitempty"`
}
