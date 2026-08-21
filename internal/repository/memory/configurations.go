package memory

import (
	"context"
	"sort"
	"strings"

	"github.com/wyw14/cry-053/internal/domain"
)

func (s *Store) Get(_ context.Context, id string) (domain.Configuration, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.configs[id]
	if !ok {
		return domain.Configuration{}, domain.ErrNotFound
	}
	return item.Clone(), nil
}

func (s *Store) GetByKey(_ context.Context, key string) (domain.Configuration, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.configKeys[key]
	if !ok {
		return domain.Configuration{}, domain.ErrNotFound
	}
	return s.configs[id].Clone(), nil
}

func (s *Store) List(_ context.Context, page domain.PageRequest) ([]domain.Configuration, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]domain.Configuration, 0, len(s.configs))
	for _, item := range s.configs {
		if value := page.Filters["environment"]; value != "" && item.Environment != value {
			continue
		}
		if value := page.Filters["status"]; value != "" && string(item.Status) != value {
			continue
		}
		if value := page.Filters["type"]; value != "" && item.Type != value {
			continue
		}
		items = append(items, item.Clone())
	}
	sort.Slice(items, func(i, j int) bool {
		if page.Sort == "-updated_at" {
			return items[i].UpdatedAt.After(items[j].UpdatedAt)
		}
		return strings.Compare(items[i].Key(), items[j].Key()) < 0
	})
	total := len(items)
	start := page.Offset()
	if start >= total {
		return []domain.Configuration{}, total, nil
	}
	end := start + page.Limit()
	if end > total {
		end = total
	}
	return items[start:end], total, nil
}

func (s *Store) Create(_ context.Context, item domain.Configuration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.configs[item.ID]; ok {
		return domain.ErrConflict
	}
	if _, ok := s.configKeys[item.Key()]; ok {
		return domain.ErrConflict
	}
	s.configs[item.ID] = item.Clone()
	s.configKeys[item.Key()] = item.ID
	return nil
}

func (s *Store) Update(_ context.Context, item domain.Configuration, expected int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.configs[item.ID]
	if !ok {
		return domain.ErrNotFound
	}
	if current.Revision != expected {
		return domain.ErrVersionConflict
	}
	if existing, ok := s.configKeys[item.Key()]; ok && existing != item.ID {
		return domain.ErrConflict
	}
	delete(s.configKeys, current.Key())
	s.configs[item.ID] = item.Clone()
	s.configKeys[item.Key()] = item.ID
	return nil
}

func (s *Store) Delete(_ context.Context, id string, expected int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.configs[id]
	if !ok {
		return domain.ErrNotFound
	}
	if current.Revision != expected {
		return domain.ErrVersionConflict
	}
	delete(s.configs, id)
	delete(s.configKeys, current.Key())
	return nil
}

func (s *Store) References(_ context.Context, id string) ([]domain.Configuration, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]domain.Configuration, 0)
	for _, item := range s.configs {
		for _, value := range item.Values {
			if value == id {
				result = append(result, item.Clone())
				break
			}
		}
	}
	return result, nil
}
