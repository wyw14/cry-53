package memory

import (
	"sync"

	"github.com/wyw14/cry-053/internal/domain"
)

type Store struct {
	mu          sync.RWMutex
	configs     map[string]domain.Configuration
	configKeys  map[string]string
	bundles     map[string]domain.Bundle
	versions    map[string][]domain.Version
	audits      []domain.AuditEvent
	types       map[string]domain.ConfigType
	usages      map[string][]domain.Usage
	idempotency map[string]idempotencyRecord
}

type idempotencyRecord struct {
	Hash       string
	ResourceID string
}

func NewStore() *Store {
	return &Store{
		configs:     make(map[string]domain.Configuration),
		configKeys:  make(map[string]string),
		bundles:     make(map[string]domain.Bundle),
		versions:    make(map[string][]domain.Version),
		types:       make(map[string]domain.ConfigType),
		usages:      make(map[string][]domain.Usage),
		idempotency: make(map[string]idempotencyRecord),
	}
}

func (s *Store) Clone() *Store {
	s.mu.RLock()
	defer s.mu.RUnlock()
	clone := NewStore()
	for id, config := range s.configs {
		clone.configs[id] = config.Clone()
	}
	for key, id := range s.configKeys {
		clone.configKeys[key] = id
	}
	for id, bundle := range s.bundles {
		bundle.Configurations = append([]domain.Configuration(nil), bundle.Configurations...)
		bundle.Issues = append([]domain.Issue(nil), bundle.Issues...)
		clone.bundles[id] = bundle
	}
	for id, versions := range s.versions {
		clone.versions[id] = append([]domain.Version(nil), versions...)
	}
	clone.audits = append([]domain.AuditEvent(nil), s.audits...)
	for name, item := range s.types {
		clone.types[name] = item
	}
	for id, usages := range s.usages {
		clone.usages[id] = append([]domain.Usage(nil), usages...)
	}
	for key, record := range s.idempotency {
		clone.idempotency[key] = record
	}
	return clone
}

func (s *Store) replace(snapshot *Store) {
	s.configs = snapshot.configs
	s.configKeys = snapshot.configKeys
	s.bundles = snapshot.bundles
	s.versions = snapshot.versions
	s.audits = snapshot.audits
	s.types = snapshot.types
	s.usages = snapshot.usages
	s.idempotency = snapshot.idempotency
}
