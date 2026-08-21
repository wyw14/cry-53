package memory

import (
	"context"
	"sync"

	"github.com/wyw14/cry-053/internal/domain"
	"github.com/wyw14/cry-053/internal/repository"
)

type Manager struct {
	store *Store
	gate  sync.Mutex
}

func NewManager(store *Store) *Manager {
	return &Manager{store: store}
}

func (m *Manager) Begin(_ context.Context) (repository.Transaction, error) {
	m.gate.Lock()
	return &transaction{manager: m, snapshot: m.store.Clone()}, nil
}

type transaction struct {
	manager  *Manager
	snapshot *Store
	done     bool
}

func (t *transaction) Configurations() repository.ConfigurationRepository {
	return configRepositoryAdapter{t.snapshot}
}

func (t *transaction) Bundles() repository.BundleRepository {
	return bundleRepositoryAdapter{t.snapshot}
}

func (t *transaction) Versions() repository.VersionRepository {
	return versionRepositoryAdapter{t.snapshot}
}

func (t *transaction) Audits() repository.AuditRepository {
	return auditRepositoryAdapter{t.snapshot}
}

func (t *transaction) Commit(_ context.Context) error {
	if t.done {
		return nil
	}
	t.manager.store.mu.Lock()
	t.manager.store.replace(t.snapshot)
	t.manager.store.mu.Unlock()
	t.done = true
	t.manager.gate.Unlock()
	return nil
}

func (t *transaction) Rollback(_ context.Context) error {
	if !t.done {
		t.done = true
		t.manager.gate.Unlock()
	}
	return nil
}

type configRepositoryAdapter struct{ *Store }
type bundleRepositoryAdapter struct{ *Store }
type versionRepositoryAdapter struct{ *Store }
type auditRepositoryAdapter struct{ *Store }

func Configurations(store *Store) repository.ConfigurationRepository {
	return configRepositoryAdapter{store}
}
func Bundles(store *Store) repository.BundleRepository   { return bundleRepositoryAdapter{store} }
func Versions(store *Store) repository.VersionRepository { return versionRepositoryAdapter{store} }
func Audits(store *Store) repository.AuditRepository     { return auditRepositoryAdapter{store} }

type typeRepositoryAdapter struct{ *Store }
type usageRepositoryAdapter struct{ *Store }

func Types(store *Store) repository.TypeRepository   { return typeRepositoryAdapter{store} }
func Usages(store *Store) repository.UsageRepository { return usageRepositoryAdapter{store} }

func (a bundleRepositoryAdapter) Get(ctx context.Context, id string) (domain.Bundle, error) {
	return a.GetBundle(ctx, id)
}
func (a bundleRepositoryAdapter) Create(ctx context.Context, b domain.Bundle) error {
	return a.CreateBundle(ctx, b)
}
func (a bundleRepositoryAdapter) Update(ctx context.Context, b domain.Bundle) error {
	return a.UpdateBundle(ctx, b)
}
func (a versionRepositoryAdapter) List(ctx context.Context, id string) ([]domain.Version, error) {
	return a.ListVersions(ctx, id)
}
func (a versionRepositoryAdapter) GetNumber(ctx context.Context, id string, number int) (domain.Version, error) {
	return a.GetVersion(ctx, id, number)
}
func (a versionRepositoryAdapter) Create(ctx context.Context, version domain.Version) error {
	return a.CreateVersion(ctx, version)
}
func (a auditRepositoryAdapter) List(ctx context.Context, page domain.PageRequest) ([]domain.AuditEvent, int, error) {
	return a.ListAudits(ctx, page)
}
func (a typeRepositoryAdapter) Get(ctx context.Context, name string) (domain.ConfigType, error) {
	return a.GetType(ctx, name)
}
func (a typeRepositoryAdapter) List(ctx context.Context) ([]domain.ConfigType, error) {
	return a.ListTypes(ctx)
}
func (a typeRepositoryAdapter) Save(ctx context.Context, item domain.ConfigType) error {
	return a.SaveType(ctx, item)
}
func (a usageRepositoryAdapter) Save(ctx context.Context, usage domain.Usage) error {
	return a.SaveUsage(ctx, usage)
}
