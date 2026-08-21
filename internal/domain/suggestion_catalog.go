package domain

import (
	"fmt"
	"sort"
	"strings"
)

type SuggestionState string

const (
	SuggestionCollecting SuggestionState = "collecting"
	SuggestionSealed     SuggestionState = "sealed"
)

type suggestionEntry struct {
	key   string
	value string
}

type SuggestionCatalog struct {
	state   SuggestionState
	entries []suggestionEntry
	seen    map[string]struct{}
}

func NewSuggestionCatalog() *SuggestionCatalog {
	return &SuggestionCatalog{
		state:   SuggestionCollecting,
		entries: make([]suggestionEntry, 0),
		seen:    make(map[string]struct{}),
	}
}

func (c *SuggestionCatalog) Observe(issue Issue) error {
	if c.state != SuggestionCollecting {
		return fmt.Errorf("suggestion catalog is %s: %w", c.state, ErrConflict)
	}
	value := strings.TrimSpace(issue.Suggestion)
	if value == "" {
		return nil
	}
	key := strings.Join([]string{strings.TrimSpace(issue.Code), strings.TrimSpace(issue.Path), value}, "\x00")
	if _, exists := c.seen[key]; exists {
		return nil
	}
	c.seen[key] = struct{}{}
	c.entries = append(c.entries, suggestionEntry{key: key, value: value})
	return nil
}

func (c *SuggestionCatalog) Seal(existing []string) ([]string, error) {
	if c.state != SuggestionCollecting {
		return nil, fmt.Errorf("suggestion catalog cannot be sealed from %s: %w", c.state, ErrConflict)
	}
	c.state = SuggestionSealed
	entries := append([]suggestionEntry(nil), c.entries...)
	sort.SliceStable(entries, func(left, right int) bool {
		if entries[left].value == entries[right].value {
			return entries[left].key < entries[right].key
		}
		return entries[left].value < entries[right].value
	})
	merged := append([]string(nil), existing...)
	for _, entry := range entries {
		merged = append(merged, entry.value)
	}
	sort.Strings(merged)
	return merged, nil
}

func (c *SuggestionCatalog) State() SuggestionState {
	return c.state
}
