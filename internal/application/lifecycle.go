package application

import (
	"context"
	"fmt"

	"github.com/wyw14/cry-053/internal/domain"
	"github.com/wyw14/cry-053/internal/repository"
)

type LifecycleService struct {
	configs  repository.ConfigurationRepository
	versions repository.VersionRepository
	audits   repository.AuditRepository
	clock    Clock
	ids      IDGenerator
}

func NewLifecycleService(configs repository.ConfigurationRepository, versions repository.VersionRepository, audits repository.AuditRepository, clock Clock, ids IDGenerator) *LifecycleService {
	return &LifecycleService{configs: configs, versions: versions, audits: audits, clock: clock, ids: ids}
}

func decideLifecycleApproval(bundle domain.Bundle, actor domain.Actor) domain.ApprovalDecision {
	return domain.DecideBundleApproval(domain.ApprovalRequest{
		State:      bundle.State,
		Issues:     bundle.Issues,
		UploadedBy: bundle.UploadedBy,
		Actor:      actor,
	})
}

func (s *LifecycleService) Transition(ctx context.Context, id string, target domain.ConfigStatus, expected int64, actor domain.Actor, requestID string) (domain.Configuration, error) {
	if !actor.HasRole("publisher") && !actor.HasRole("platform_admin") {
		return domain.Configuration{}, domain.ErrForbidden
	}
	current, err := s.configs.Get(ctx, id)
	if err != nil {
		return domain.Configuration{}, err
	}
	allowed := map[domain.ConfigStatus]map[domain.ConfigStatus]bool{
		domain.StatusDraft:     {domain.StatusPending: true, domain.StatusArchived: true},
		domain.StatusPending:   {domain.StatusPublished: true, domain.StatusDraft: true},
		domain.StatusPublished: {domain.StatusWithdrawn: true},
		domain.StatusWithdrawn: {domain.StatusPublished: true, domain.StatusArchived: true},
		domain.StatusArchived:  {},
	}
	if !allowed[current.Status][target] {
		return domain.Configuration{}, domain.ErrInvalidTransition
	}
	if current.Revision != expected {
		return domain.Configuration{}, domain.ErrVersionConflict
	}
	before := current.Status
	current.Status = target
	current.Revision++
	current.UpdatedAt = s.clock.Now()
	if err := s.configs.Update(ctx, current, expected); err != nil {
		return domain.Configuration{}, err
	}
	_ = s.audits.Append(ctx, domain.AuditEvent{ID: s.ids.New("audit"), RequestID: requestID, ActorID: actor.ID, Operation: "configuration.transition", Target: domain.AuditTarget{Kind: "configuration", ID: id}, Changes: []domain.AuditChange{{Field: "status", Previous: before, Current: target}}, OccurredAt: s.clock.Now()})
	return current, nil
}

func (s *LifecycleService) Rollback(ctx context.Context, request domain.RollbackRequest, actor domain.Actor, requestID string) (domain.Configuration, error) {
	if !actor.HasRole("publisher") {
		return domain.Configuration{}, domain.ErrForbidden
	}
	current, err := s.configs.Get(ctx, request.ConfigurationID)
	if err != nil {
		return domain.Configuration{}, err
	}
	if current.Revision != request.ExpectedRevision {
		return domain.Configuration{}, domain.ErrVersionConflict
	}
	target, err := s.versions.GetNumber(ctx, request.ConfigurationID, request.TargetVersion)
	if err != nil {
		return domain.Configuration{}, fmt.Errorf("target version: %w", err)
	}
	current.Values = target.Values
	current.Tags = append([]string(nil), target.Tags...)
	current.Version++
	current.Revision++
	current.Status = domain.StatusPublished
	current.UpdatedAt = s.clock.Now()
	if err := s.configs.Update(ctx, current, request.ExpectedRevision); err != nil {
		return domain.Configuration{}, err
	}
	version := domain.Version{ID: s.ids.New("ver"), ConfigurationID: current.ID, Number: current.Version, Values: current.Clone().Values, Tags: append([]string(nil), current.Tags...), Reason: "rollback: " + request.Reason, CreatedBy: actor.ID, CreatedAt: s.clock.Now()}
	if err := s.versions.Create(ctx, version); err != nil {
		return domain.Configuration{}, err
	}
	_ = s.audits.Append(ctx, domain.AuditEvent{ID: s.ids.New("audit"), RequestID: requestID, ActorID: actor.ID, Operation: "configuration.rollback", Target: domain.AuditTarget{Kind: "configuration", ID: current.ID}, Changes: []domain.AuditChange{{Field: "version", Previous: request.TargetVersion, Current: current.Version}}, Facts: map[string]any{"reason": request.Reason}, OccurredAt: s.clock.Now()})
	return current, nil
}

func (s *LifecycleService) Delete(ctx context.Context, id string, expected int64, actor domain.Actor) error {
	if !actor.HasRole("platform_admin") {
		return domain.ErrForbidden
	}
	references, err := s.configs.References(ctx, id)
	if err != nil {
		return err
	}
	if len(references) > 0 {
		return domain.ErrReferenced
	}
	return s.configs.Delete(ctx, id, expected)
}
