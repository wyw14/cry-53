package domain

import "time"

type ExportPolicy struct {
	AllowSensitive bool      `json:"allow_sensitive"`
	Reason         string    `json:"reason"`
	ExpiresAt      time.Time `json:"expires_at"`
}

type ExportedConfiguration struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Type        string         `json:"type"`
	Environment string         `json:"environment"`
	Version     int            `json:"version"`
	Values      map[string]any `json:"values"`
	Masked      bool           `json:"masked"`
	ExportedAt  time.Time      `json:"exported_at"`
}

type ImportSummary struct {
	BundleID       string         `json:"bundle_id"`
	Total          int            `json:"total"`
	Errors         int            `json:"errors"`
	Warnings       int            `json:"warnings"`
	IssuesByCode   map[string]int `json:"issues_by_code"`
	SuggestedFixes []string       `json:"suggested_fixes"`
}
