package memory

import (
	"context"
	"sort"

	"github.com/wyw14/cry-053/internal/domain"
)

func (s *Store) ListVersions(_ context.Context, configurationID string) ([]domain.Version, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	versions := append([]domain.Version(nil), s.versions[configurationID]...)
	sort.Slice(versions, func(i, j int) bool { return versions[i].Number < versions[j].Number })
	return versions, nil
}

func (s *Store) GetVersion(_ context.Context, configurationID string, number int) (domain.Version, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, version := range s.versions[configurationID] {
		if version.Number == number {
			return version, nil
		}
	}
	return domain.Version{}, domain.ErrNotFound
}

func (s *Store) CreateVersion(_ context.Context, version domain.Version) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, current := range s.versions[version.ConfigurationID] {
		if current.Number == version.Number {
			return domain.ErrConflict
		}
	}
	s.versions[version.ConfigurationID] = append(s.versions[version.ConfigurationID], version)
	return nil
}

func (s *Store) Append(_ context.Context, event domain.AuditEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	previous := ""
	if len(s.audits) > 0 {
		previous = s.audits[len(s.audits)-1].Hash
	}
	event = event.Seal(previous)
	s.audits = append(s.audits, event)
	return nil
}

func (s *Store) ListAudits(_ context.Context, page domain.PageRequest) ([]domain.AuditEvent, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]domain.AuditEvent, 0, len(s.audits))
	for _, event := range s.audits {
		if operation := page.Filters["operation"]; operation != "" && event.Operation != operation {
			continue
		}
		if actorID := page.Filters["actor_id"]; actorID != "" && event.ActorID != actorID {
			continue
		}
		if target := page.Filters["target"]; target != "" && event.Target.Kind+":"+event.Target.ID != target {
			continue
		}
		items = append(items, event)
	}
	if page.Sort == "occurred_at_desc" {
		for left, right := 0, len(items)-1; left < right; left, right = left+1, right-1 {
			items[left], items[right] = items[right], items[left]
		}
	}
	total := len(items)
	start := page.Offset()
	if start >= total {
		return []domain.AuditEvent{}, total, nil
	}
	end := start + page.Limit()
	if end > total {
		end = total
	}
	return items[start:end], total, nil
}

func (s *Store) GetType(_ context.Context, name string) (domain.ConfigType, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.types[name]
	if !ok {
		return domain.ConfigType{}, domain.ErrNotFound
	}
	return item, nil
}

func (s *Store) ListTypes(_ context.Context) ([]domain.ConfigType, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]domain.ConfigType, 0, len(s.types))
	for _, item := range s.types {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	return items, nil
}

func (s *Store) SaveType(_ context.Context, item domain.ConfigType) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.types[item.Name] = item
	return nil
}

func (s *Store) ListByConfiguration(_ context.Context, id string) ([]domain.Usage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]domain.Usage(nil), s.usages[id]...), nil
}

func (s *Store) SaveUsage(_ context.Context, usage domain.Usage) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.usages[usage.ConfigurationID] = append(s.usages[usage.ConfigurationID], usage)
	return nil
}
