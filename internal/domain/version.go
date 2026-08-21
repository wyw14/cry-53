package domain

import "time"

type Version struct {
	ID              string         `json:"id"`
	ConfigurationID string         `json:"configuration_id"`
	Number          int            `json:"number"`
	Values          map[string]any `json:"values"`
	Tags            []string       `json:"tags"`
	Reason          string         `json:"reason"`
	CreatedBy       string         `json:"created_by"`
	CreatedAt       time.Time      `json:"created_at"`
	ReplacedBy      string         `json:"replaced_by,omitempty"`
}

type RollbackRequest struct {
	ConfigurationID  string `json:"configuration_id" validate:"required"`
	TargetVersion    int    `json:"target_version" validate:"required,min=1"`
	ExpectedRevision int64  `json:"expected_revision" validate:"required,min=1"`
	Reason           string `json:"reason" validate:"required,min=3,max=300"`
}
