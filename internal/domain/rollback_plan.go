package domain

import (
	"fmt"
	"strings"
	"time"
)

type RollbackPlan struct {
	ConfigurationID string
	TargetVersion   int
	SuccessorNumber int
	Values          map[string]any
	Tags            []string
	Reason          string
	CreatedBy       string
	CreatedAt       time.Time
}

func PlanRollback(current Configuration, target Version, request RollbackRequest, actor Actor, now time.Time) (RollbackPlan, error) {
	if current.ID == "" || current.ID != request.ConfigurationID {
		return RollbackPlan{}, fmt.Errorf("rollback configuration identity: %w", ErrConflict)
	}
	if target.ConfigurationID != current.ID || target.Number != request.TargetVersion {
		return RollbackPlan{}, fmt.Errorf("rollback target identity: %w", ErrConflict)
	}
	if current.Status == StatusArchived {
		return RollbackPlan{}, ErrInvalidTransition
	}
	reason := strings.TrimSpace(request.Reason)
	if reason == "" {
		return RollbackPlan{}, fmt.Errorf("rollback reason: %w", ErrConflict)
	}
	if actor.ID == "" {
		return RollbackPlan{}, ErrForbidden
	}

	values := make(map[string]any, len(target.Values))
	for key, value := range target.Values {
		values[key] = value
	}
	tags := append([]string(nil), target.Tags...)
	successor := target.Number + 1
	if successor <= target.Number {
		return RollbackPlan{}, fmt.Errorf("rollback successor: %w", ErrVersionConflict)
	}

	return RollbackPlan{
		ConfigurationID: current.ID,
		TargetVersion:   target.Number,
		SuccessorNumber: successor,
		Values:          values,
		Tags:            tags,
		Reason:          reason,
		CreatedBy:       actor.ID,
		CreatedAt:       now,
	}, nil
}

func (p RollbackPlan) Apply(current Configuration) Configuration {
	updated := current.Clone()
	updated.Values = cloneRollbackValues(p.Values)
	updated.Tags = append([]string(nil), p.Tags...)
	updated.Version = p.SuccessorNumber
	updated.Revision = current.Revision
	updated.Status = StatusPublished
	updated.UpdatedAt = p.CreatedAt
	return updated
}

func (p RollbackPlan) Successor(id string) Version {
	return Version{
		ID:              id,
		ConfigurationID: p.ConfigurationID,
		Number:          p.SuccessorNumber,
		Values:          cloneRollbackValues(p.Values),
		Tags:            append([]string(nil), p.Tags...),
		Reason:          "rollback: " + p.Reason,
		CreatedBy:       p.CreatedBy,
		CreatedAt:       p.CreatedAt,
	}
}

func cloneRollbackValues(source map[string]any) map[string]any {
	result := make(map[string]any, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}
