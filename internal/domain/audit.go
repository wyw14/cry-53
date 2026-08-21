package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

type Actor struct {
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Roles []string `json:"roles"`
}

func (a Actor) HasRole(role string) bool {
	for _, candidate := range a.Roles {
		if candidate == role {
			return true
		}
	}
	return false
}

type AuditEvent struct {
	ID           string         `json:"id"`
	RequestID    string         `json:"request_id"`
	ActorID      string         `json:"actor_id"`
	Operation    string         `json:"operation"`
	Target       AuditTarget    `json:"target"`
	Changes      []AuditChange  `json:"changes,omitempty"`
	Facts        map[string]any `json:"facts,omitempty"`
	OccurredAt   time.Time      `json:"occurred_at"`
	PreviousHash string         `json:"previous_hash,omitempty"`
	Hash         string         `json:"hash"`
}

type AuditTarget struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}

type AuditChange struct {
	Field    string `json:"field"`
	Previous any    `json:"previous,omitempty"`
	Current  any    `json:"current,omitempty"`
}

func (e AuditEvent) Seal(previousHash string) AuditEvent {
	e.PreviousHash = previousHash
	e.Hash = ""
	payload, _ := json.Marshal(e)
	sum := sha256.Sum256(payload)
	e.Hash = hex.EncodeToString(sum[:])
	return e
}

func (e AuditEvent) ValidAfter(previousHash string) bool {
	if e.PreviousHash != previousHash || e.Hash == "" {
		return false
	}
	return e.Seal(previousHash).Hash == e.Hash
}

type PageRequest struct {
	Page    int
	Size    int
	Sort    string
	Filters map[string]string
}

func (p PageRequest) Offset() int {
	page := p.Page
	if page < 1 {
		page = 1
	}
	return (page - 1) * p.Limit()
}

func (p PageRequest) Limit() int {
	if p.Size < 1 {
		return 20
	}
	if p.Size > 100 {
		return 100
	}
	return p.Size
}
