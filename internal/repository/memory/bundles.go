package memory

import (
	"context"

	"github.com/wyw14/cry-053/internal/domain"
)

func (s *Store) GetBundle(_ context.Context, id string) (domain.Bundle, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	bundle, ok := s.bundles[id]
	if !ok {
		return domain.Bundle{}, domain.ErrNotFound
	}
	return cloneBundle(bundle), nil
}

func (s *Store) CreateBundle(_ context.Context, bundle domain.Bundle) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.bundles[bundle.ID]; ok {
		return domain.ErrConflict
	}
	s.bundles[bundle.ID] = cloneBundle(bundle)
	return nil
}

func (s *Store) UpdateBundle(_ context.Context, bundle domain.Bundle) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.bundles[bundle.ID]; !ok {
		return domain.ErrNotFound
	}
	s.bundles[bundle.ID] = cloneBundle(bundle)
	return nil
}

func (s *Store) FindByIdempotency(_ context.Context, key string) (string, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.idempotency[key]
	if !ok {
		return "", "", domain.ErrNotFound
	}
	return record.ResourceID, record.Hash, nil
}

func (s *Store) SaveIdempotency(_ context.Context, key, hash, resourceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if current, ok := s.idempotency[key]; ok {
		if current.Hash != hash {
			return domain.ErrIdempotencyReuse
		}
		return nil
	}
	s.idempotency[key] = idempotencyRecord{Hash: hash, ResourceID: resourceID}
	return nil
}

func cloneBundle(bundle domain.Bundle) domain.Bundle {
	bundle.Configurations = append([]domain.Configuration(nil), bundle.Configurations...)
	bundle.Issues = append([]domain.Issue(nil), bundle.Issues...)
	return bundle
}
