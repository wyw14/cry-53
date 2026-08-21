package service

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"

	"github.com/wyw14/cry-053/internal/domain"
	"github.com/wyw14/cry-053/internal/repository"
)

type ChangeKind string

const (
	ChangeCreate    ChangeKind = "create"
	ChangeUpdate    ChangeKind = "update"
	ChangeUnchanged ChangeKind = "unchanged"
)

type FieldChange struct {
	Field     string `json:"field"`
	Before    any    `json:"before,omitempty"`
	After     any    `json:"after,omitempty"`
	Sensitive bool   `json:"sensitive"`
}

type ConfigChange struct {
	Configuration domain.Configuration `json:"configuration"`
	Kind          ChangeKind           `json:"kind"`
	Fields        []FieldChange        `json:"fields"`
}

type Preview struct {
	BundleID string         `json:"bundle_id"`
	Changes  []ConfigChange `json:"changes"`
	Errors   []domain.Issue `json:"errors"`
}

type DiffService struct {
	configs repository.ConfigurationRepository
	types   repository.TypeRepository
}

func NewDiffService(configs repository.ConfigurationRepository, types repository.TypeRepository) *DiffService {
	return &DiffService{configs: configs, types: types}
}

func (s *DiffService) Preview(ctx context.Context, bundle domain.Bundle) (Preview, error) {
	preview := Preview{BundleID: bundle.ID, Errors: append([]domain.Issue(nil), bundle.Issues...)}
	for _, candidate := range bundle.Configurations {
		current, err := s.configs.GetByKey(ctx, candidate.Key())
		if errors.Is(err, domain.ErrNotFound) {
			preview.Changes = append(preview.Changes, ConfigChange{Configuration: candidate, Kind: ChangeCreate, Fields: diffFields(nil, candidate.Values, s.sensitiveFields(ctx, candidate.Type))})
			continue
		}
		if err != nil {
			return Preview{}, fmt.Errorf("load current %s: %w", candidate.Key(), err)
		}
		fields := diffFields(current.Values, candidate.Values, s.sensitiveFields(ctx, candidate.Type))
		kind := ChangeUpdate
		if len(fields) == 0 {
			kind = ChangeUnchanged
		}
		candidate.ID = current.ID
		candidate.Revision = current.Revision
		candidate.Version = current.Version
		preview.Changes = append(preview.Changes, ConfigChange{Configuration: candidate, Kind: kind, Fields: fields})
	}
	return preview, nil
}

func (s *DiffService) sensitiveFields(ctx context.Context, typeName string) map[string]bool {
	result := make(map[string]bool)
	definition, err := s.types.Get(ctx, typeName)
	if err != nil {
		return result
	}
	for _, field := range definition.Fields {
		result[field.Name] = field.Sensitive
	}
	return result
}

func diffFields(before, after map[string]any, sensitive map[string]bool) []FieldChange {
	keys := make(map[string]struct{}, len(before)+len(after))
	for key := range before {
		keys[key] = struct{}{}
	}
	for key := range after {
		keys[key] = struct{}{}
	}
	ordered := make([]string, 0, len(keys))
	for key := range keys {
		ordered = append(ordered, key)
	}
	sort.Strings(ordered)
	changes := make([]FieldChange, 0)
	for _, key := range ordered {
		if reflect.DeepEqual(before[key], after[key]) {
			continue
		}
		change := FieldChange{Field: key, Before: before[key], After: after[key], Sensitive: sensitive[key]}
		if change.Sensitive {
			if change.Before != nil {
				change.Before = "******"
			}
			if change.After != nil {
				change.After = "******"
			}
		}
		changes = append(changes, change)
	}
	return changes
}
