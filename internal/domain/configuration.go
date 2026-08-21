package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

type ConfigStatus string

const (
	StatusDraft     ConfigStatus = "draft"
	StatusPending   ConfigStatus = "pending"
	StatusPublished ConfigStatus = "published"
	StatusWithdrawn ConfigStatus = "withdrawn"
	StatusArchived  ConfigStatus = "archived"
)

type Configuration struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Type        string         `json:"type"`
	Environment string         `json:"environment"`
	Values      map[string]any `json:"values"`
	Status      ConfigStatus   `json:"status"`
	Revision    int64          `json:"revision"`
	Version     int            `json:"version"`
	Tags        []string       `json:"tags"`
	ExpiresAt   *time.Time     `json:"expires_at,omitempty"`
	ReplacesID  string         `json:"replaces_id,omitempty"`
	CreatedBy   string         `json:"created_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

func (c Configuration) Key() string {
	return c.Environment + "/" + c.Type + "/" + c.Name
}

func (c Configuration) Clone() Configuration {
	clone := c
	clone.Values = make(map[string]any, len(c.Values))
	for key, value := range c.Values {
		clone.Values[key] = value
	}
	clone.Tags = append([]string(nil), c.Tags...)
	return clone
}

func (c Configuration) ContentHash() (string, error) {
	keys := make([]string, 0, len(c.Values))
	for key := range c.Values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	ordered := make([]struct {
		Key   string `json:"key"`
		Value any    `json:"value"`
	}, 0, len(keys))
	for _, key := range keys {
		ordered = append(ordered, struct {
			Key   string `json:"key"`
			Value any    `json:"value"`
		}{Key: key, Value: c.Values[key]})
	}
	payload, err := json.Marshal(struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		Environment string `json:"environment"`
		Values      any    `json:"values"`
	}{c.Name, c.Type, c.Environment, ordered})
	if err != nil {
		return "", fmt.Errorf("marshal content: %w", err)
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:]), nil
}

func (c Configuration) IsExpired(now time.Time) bool {
	return c.ExpiresAt != nil && !c.ExpiresAt.After(now)
}
