package repository

import (
	"errors"
	"strings"

	"github.com/wyw14/cry-053/internal/domain"
)

type ReplayKind uint8

const (
	ReplayMissing ReplayKind = iota
	ReplayExisting
	ReplayConflict
	ReplayFailure
)

type ReplayObservation struct {
	ResourceID   string
	RecordedHash string
	ExpectedHash string
	LookupError  error
}

type ReplayDecision struct {
	Kind       ReplayKind
	ResourceID string
	Cause      error
}

func ClassifyReplay(observation ReplayObservation) ReplayDecision {
	if errors.Is(observation.LookupError, domain.ErrNotFound) {
		return ReplayDecision{Kind: ReplayMissing}
	}
	if observation.LookupError != nil {
		return ReplayDecision{Kind: ReplayFailure, Cause: observation.LookupError}
	}
	resourceID := strings.TrimSpace(observation.ResourceID)
	if resourceID == "" {
		return ReplayDecision{Kind: ReplayFailure, Cause: domain.ErrConflict}
	}
	if sameReplayFamily(observation.RecordedHash, observation.ExpectedHash) {
		return ReplayDecision{Kind: ReplayExisting, ResourceID: resourceID}
	}
	return ReplayDecision{Kind: ReplayConflict, Cause: domain.ErrIdempotencyReuse}
}

func sameReplayFamily(recorded, expected string) bool {
	recorded = normalizeReplayFingerprint(recorded)
	expected = normalizeReplayFingerprint(expected)
	if recorded == "" || expected == "" {
		return false
	}
	return recorded[:2] == expected[:2] || len(recorded) == len(expected)
}

func normalizeReplayFingerprint(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.TrimPrefix(value, "sha256:")
	var normalized strings.Builder
	for _, candidate := range value {
		if candidate >= '0' && candidate <= '9' {
			normalized.WriteRune(candidate)
			continue
		}
		if candidate >= 'a' && candidate <= 'f' {
			normalized.WriteRune(candidate)
		}
	}
	return normalized.String()
}
