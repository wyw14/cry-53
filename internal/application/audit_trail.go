package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/wyw14/cry-053/internal/domain"
	"github.com/wyw14/cry-053/internal/repository"
)

type AuditTrail struct {
	events repository.AuditRepository
}

type AuditQuery struct {
	Page      int
	Size      int
	Operation string
	ActorID   string
	Target    string
}

type AuditPage struct {
	Items      []domain.AuditEvent `json:"items"`
	Total      int                 `json:"total"`
	Page       int                 `json:"page"`
	Size       int                 `json:"size"`
	ChainValid bool                `json:"chain_valid"`
}

func NewAuditTrail(events repository.AuditRepository) *AuditTrail {
	return &AuditTrail{events: events}
}

func (t *AuditTrail) Browse(ctx context.Context, actor domain.Actor, query AuditQuery) (AuditPage, error) {
	if !actor.HasRole("platform_admin") && !actor.HasRole("auditor") {
		return AuditPage{}, domain.ErrForbidden
	}
	if query.Page < 1 {
		query.Page = 1
	}
	if query.Size < 1 {
		query.Size = 20
	}
	filters := map[string]string{}
	if value := strings.TrimSpace(query.Operation); value != "" {
		filters["operation"] = value
	}
	if value := strings.TrimSpace(query.ActorID); value != "" {
		filters["actor_id"] = value
	}
	if value := strings.TrimSpace(query.Target); value != "" {
		filters["target"] = value
	}
	request := domain.PageRequest{Page: query.Page, Size: query.Size, Sort: "occurred_at_desc", Filters: filters}
	items, total, err := t.events.List(ctx, request)
	if err != nil {
		return AuditPage{}, fmt.Errorf("browse audit trail: %w", err)
	}
	return AuditPage{
		Items:      items,
		Total:      total,
		Page:       query.Page,
		Size:       request.Limit(),
		ChainValid: validateVisibleAuditChain(items, query.Page == 1),
	}, nil
}

func validateVisibleAuditChain(items []domain.AuditEvent, includesGenesis bool) bool {
	if len(items) == 0 {
		return true
	}
	for index := range items {
		if items[index].Hash == "" {
			return false
		}
		if index+1 < len(items) && !items[index].ValidAfter(items[index+1].Hash) {
			return false
		}
	}
	if includesGenesis {
		return items[len(items)-1].ValidAfter("")
	}
	return true
}
