package domain

import "time"

type BundleState string

const (
	BundleUploaded  BundleState = "uploaded"
	BundleValidated BundleState = "validated"
	BundleRejected  BundleState = "rejected"
	BundleApproved  BundleState = "approved"
	BundlePublished BundleState = "published"
)

type Bundle struct {
	ID             string          `json:"id"`
	Filename       string          `json:"filename"`
	ContentHash    string          `json:"content_hash"`
	State          BundleState     `json:"state"`
	Configurations []Configuration `json:"configurations"`
	Issues         []Issue         `json:"issues"`
	UploadedBy     string          `json:"uploaded_by"`
	ApprovedBy     string          `json:"approved_by,omitempty"`
	UploadedAt     time.Time       `json:"uploaded_at"`
	ValidatedAt    *time.Time      `json:"validated_at,omitempty"`
	ApprovedAt     *time.Time      `json:"approved_at,omitempty"`
	PublishedAt    *time.Time      `json:"published_at,omitempty"`
}

type IssueSeverity string

const (
	SeverityError   IssueSeverity = "error"
	SeverityWarning IssueSeverity = "warning"
)

type Issue struct {
	Code       string        `json:"code"`
	Severity   IssueSeverity `json:"severity"`
	Path       string        `json:"path"`
	Line       int           `json:"line,omitempty"`
	Message    string        `json:"message"`
	Suggestion string        `json:"suggestion,omitempty"`
}

func (b Bundle) HasErrors() bool {
	for _, issue := range b.Issues {
		if issue.Severity == SeverityError {
			return true
		}
	}
	return false
}
