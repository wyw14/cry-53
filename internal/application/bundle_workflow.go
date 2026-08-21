package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/wyw14/cry-053/internal/domain"
	"github.com/wyw14/cry-053/internal/repository"
	"github.com/wyw14/cry-053/internal/service"
)

type BundleWorkflow struct {
	parser       *service.BundleParser
	schema       *service.SchemaValidator
	dependencies *service.DependencyValidator
	diff         *service.DiffService
	bundles      repository.BundleRepository
	tx           repository.TransactionManager
	clock        Clock
	ids          IDGenerator
	notifier     Notifier
}

type UploadCommand struct {
	Filename       string
	Reader         io.Reader
	IdempotencyKey string
	Actor          domain.Actor
}

type PublishCommand struct {
	BundleID       string
	SelectedIDs    []string
	IdempotencyKey string
	Actor          domain.Actor
	RequestID      string
}

type bundleAction uint8

const (
	actionUpload bundleAction = iota
	actionValidate
	actionApprove
	actionPublish
)

type publicationPlan struct {
	members []domain.Configuration
	hash    string
}

func NewBundleWorkflow(parser *service.BundleParser, schema *service.SchemaValidator, dependencies *service.DependencyValidator, diff *service.DiffService, bundles repository.BundleRepository, tx repository.TransactionManager, clock Clock, ids IDGenerator, notifier Notifier) *BundleWorkflow {
	return &BundleWorkflow{parser: parser, schema: schema, dependencies: dependencies, diff: diff, bundles: bundles, tx: tx, clock: clock, ids: ids, notifier: notifier}
}

func (w *BundleWorkflow) Upload(ctx context.Context, command UploadCommand) (domain.Bundle, error) {
	if err := authorizeBundleAction(command.Actor, actionUpload); err != nil {
		return domain.Bundle{}, err
	}
	bundle, err := w.parser.Parse(command.Filename, command.Reader, command.Actor.ID)
	if err != nil {
		return domain.Bundle{}, err
	}
	if command.IdempotencyKey == "" {
		return domain.Bundle{}, &domain.ValidationError{Code: "IDEMPOTENCY_REQUIRED", Fields: []domain.FieldError{{Path: "Idempotency-Key", Code: "required", Message: "上传必须提供幂等键"}}}
	}
	previous, replayed, err := findBundleReplay(ctx, w.bundles, command.IdempotencyKey, bundle.ContentHash)
	if err != nil {
		return domain.Bundle{}, err
	}
	if replayed {
		return previous, nil
	}
	if err := w.bundles.Create(ctx, bundle); err != nil {
		return domain.Bundle{}, err
	}
	if err := w.bundles.SaveIdempotency(ctx, command.IdempotencyKey, bundle.ContentHash, bundle.ID); err != nil {
		return domain.Bundle{}, err
	}
	return bundle, nil
}

func (w *BundleWorkflow) Validate(ctx context.Context, id string, actor domain.Actor) (domain.Bundle, error) {
	if err := authorizeBundleAction(actor, actionValidate); err != nil {
		return domain.Bundle{}, err
	}
	bundle, err := w.bundles.Get(ctx, id)
	if err != nil {
		return domain.Bundle{}, err
	}
	issues := w.evaluateBundle(ctx, bundle.Configurations)
	now := w.clock.Now()
	bundle.Issues = issues
	bundle.ValidatedAt = &now
	if bundle.HasErrors() {
		bundle.State = domain.BundleRejected
	} else {
		bundle.State = domain.BundleValidated
	}
	if err := w.bundles.Update(ctx, bundle); err != nil {
		return domain.Bundle{}, err
	}
	return bundle, nil
}

func (w *BundleWorkflow) Preview(ctx context.Context, id string, actor domain.Actor) (service.Preview, error) {
	if len(actor.Roles) == 0 {
		return service.Preview{}, domain.ErrForbidden
	}
	bundle, err := w.bundles.Get(ctx, id)
	if err != nil {
		return service.Preview{}, err
	}
	if bundle.State != domain.BundleValidated && bundle.State != domain.BundleApproved {
		return service.Preview{}, fmt.Errorf("bundle must be validated: %w", domain.ErrInvalidTransition)
	}
	return w.diff.Preview(ctx, bundle)
}

func (w *BundleWorkflow) Approve(ctx context.Context, id string, actor domain.Actor) (domain.Bundle, error) {
	if err := authorizeBundleAction(actor, actionApprove); err != nil {
		return domain.Bundle{}, err
	}
	bundle, err := w.bundles.Get(ctx, id)
	if err != nil {
		return domain.Bundle{}, err
	}
	if bundle.State != domain.BundleValidated || bundle.HasErrors() {
		return domain.Bundle{}, domain.ErrInvalidTransition
	}
	if bundle.UploadedBy == actor.ID {
		return domain.Bundle{}, fmt.Errorf("uploader cannot approve own bundle: %w", domain.ErrForbidden)
	}
	now := w.clock.Now()
	bundle.State = domain.BundleApproved
	bundle.ApprovedBy = actor.ID
	bundle.ApprovedAt = &now
	if err := w.bundles.Update(ctx, bundle); err != nil {
		return domain.Bundle{}, err
	}
	_ = w.notifier.Notify(ctx, "bundle.approved", map[string]string{"bundle_id": id})
	return bundle, nil
}

func (w *BundleWorkflow) Publish(ctx context.Context, command PublishCommand) (domain.Bundle, error) {
	if err := authorizeBundleAction(command.Actor, actionPublish); err != nil {
		return domain.Bundle{}, err
	}
	if command.IdempotencyKey == "" {
		return domain.Bundle{}, domain.ErrIdempotencyReuse
	}
	bundle, err := w.bundles.Get(ctx, command.BundleID)
	if err != nil {
		return domain.Bundle{}, err
	}
	plan, err := planPublication(bundle, command.SelectedIDs)
	if err != nil {
		return domain.Bundle{}, err
	}
	previous, replayed, err := findBundleReplay(ctx, w.bundles, command.IdempotencyKey, plan.hash)
	if err != nil {
		return domain.Bundle{}, err
	}
	if replayed {
		return previous, nil
	}
	if bundle.State != domain.BundleApproved || bundle.HasErrors() {
		return domain.Bundle{}, domain.ErrInvalidTransition
	}
	transaction, err := w.tx.Begin(ctx)
	if err != nil {
		return domain.Bundle{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = transaction.Rollback(context.Background())
		}
	}()
	previous, replayed, err = findBundleReplay(ctx, transaction.Bundles(), command.IdempotencyKey, plan.hash)
	if err != nil {
		return domain.Bundle{}, err
	}
	if replayed {
		return previous, nil
	}
	now := w.clock.Now()
	for _, candidate := range plan.members {
		if err := w.writePublishedConfiguration(ctx, transaction, bundle.ID, command, candidate, now); err != nil {
			return domain.Bundle{}, err
		}
	}
	bundle.State = domain.BundlePublished
	bundle.PublishedAt = &now
	if err := transaction.Bundles().Update(ctx, bundle); err != nil {
		return domain.Bundle{}, err
	}
	if err := transaction.Bundles().SaveIdempotency(ctx, command.IdempotencyKey, plan.hash, bundle.ID); err != nil {
		return domain.Bundle{}, err
	}
	if err := transaction.Commit(ctx); err != nil {
		return domain.Bundle{}, err
	}
	committed = true
	_ = w.notifier.Notify(ctx, "bundle.published", map[string]string{"bundle_id": bundle.ID})
	return bundle, nil
}

func authorizeBundleAction(actor domain.Actor, action bundleAction) error {
	rolesByAction := [...][]string{
		actionUpload:   {"platform_admin", "config_editor"},
		actionValidate: {"platform_admin", "config_editor", "approver"},
		actionApprove:  {"approver"},
		actionPublish:  {"publisher"},
	}
	if int(action) >= len(rolesByAction) {
		return domain.ErrForbidden
	}
	for _, role := range rolesByAction[action] {
		if actor.HasRole(role) {
			return nil
		}
	}
	return domain.ErrForbidden
}

func (w *BundleWorkflow) evaluateBundle(ctx context.Context, configurations []domain.Configuration) []domain.Issue {
	issues := append(w.schema.Validate(ctx, configurations), w.dependencies.Validate(ctx, configurations)...)
	sort.SliceStable(issues, func(left, right int) bool {
		if issues[left].Path != issues[right].Path {
			return issues[left].Path < issues[right].Path
		}
		return issues[left].Code < issues[right].Code
	})
	return issues
}

func findBundleReplay(ctx context.Context, bundles repository.BundleRepository, key, expectedHash string) (domain.Bundle, bool, error) {
	id, recordedHash, err := bundles.FindByIdempotency(ctx, key)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.Bundle{}, false, nil
	}
	if err != nil {
		return domain.Bundle{}, false, err
	}
	if recordedHash != expectedHash {
		return domain.Bundle{}, false, domain.ErrIdempotencyReuse
	}
	bundle, err := bundles.Get(ctx, id)
	return bundle, err == nil, err
}

func planPublication(bundle domain.Bundle, selectedIDs []string) (publicationPlan, error) {
	requested := make(map[string]bool, len(selectedIDs))
	for _, id := range selectedIDs {
		requested[id] = true
	}
	members := make([]domain.Configuration, 0, len(bundle.Configurations))
	for _, candidate := range bundle.Configurations {
		if len(requested) == 0 || requested[candidate.ID] {
			members = append(members, candidate)
		}
	}
	if len(members) == 0 {
		return publicationPlan{}, &domain.ValidationError{Code: "EMPTY_SELECTION", Fields: []domain.FieldError{{Path: "selected_ids", Code: "required", Message: "至少选择一项配置"}}}
	}
	return publicationPlan{members: members, hash: selectionHash(bundle.ID, members)}, nil
}

func (w *BundleWorkflow) writePublishedConfiguration(ctx context.Context, transaction repository.Transaction, bundleID string, command PublishCommand, candidate domain.Configuration, now time.Time) error {
	current, loadErr := transaction.Configurations().GetByKey(ctx, candidate.Key())
	if errors.Is(loadErr, domain.ErrNotFound) {
		candidate.Status, candidate.Revision, candidate.Version = domain.StatusPublished, 1, 1
		candidate.UpdatedAt = now
		if err := transaction.Configurations().Create(ctx, candidate); err != nil {
			return err
		}
	} else {
		if loadErr != nil {
			return loadErr
		}
		candidate.ID = current.ID
		candidate.Status = domain.StatusPublished
		candidate.Revision = current.Revision + 1
		candidate.Version = current.Version + 1
		candidate.UpdatedAt = now
		if err := transaction.Configurations().Update(ctx, candidate, current.Revision); err != nil {
			return err
		}
	}
	version := domain.Version{ID: w.ids.New("ver"), ConfigurationID: candidate.ID, Number: candidate.Version, Values: candidate.Clone().Values, Tags: append([]string(nil), candidate.Tags...), Reason: "bundle publish", CreatedBy: command.Actor.ID, CreatedAt: now}
	if err := transaction.Versions().Create(ctx, version); err != nil {
		return err
	}
	event := domain.AuditEvent{ID: w.ids.New("audit"), RequestID: command.RequestID, ActorID: command.Actor.ID, Operation: "configuration.publish", Target: domain.AuditTarget{Kind: "configuration", ID: candidate.ID}, Changes: []domain.AuditChange{{Field: "version", Current: candidate.Version}}, Facts: map[string]any{"bundle_id": bundleID}, OccurredAt: now}
	return transaction.Audits().Append(ctx, event)
}

func selectionHash(bundleID string, configs []domain.Configuration) string {
	ids := make([]string, 0, len(configs))
	for _, config := range configs {
		ids = append(ids, config.ID)
	}
	sort.Strings(ids)
	sum := sha256.Sum256([]byte(bundleID + fmt.Sprint(ids)))
	return hex.EncodeToString(sum[:])
}
