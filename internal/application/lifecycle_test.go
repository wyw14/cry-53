package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/wyw14/cry-053/internal/domain"
	"github.com/wyw14/cry-053/internal/repository/memory"
)

func TestLifecycleRejectsStaleRevisionAndReferencedDelete(t *testing.T) {
	system := newTestSystem(t)
	repository := memory.Configurations(system.store)
	now := system.clock.Now()
	base := domain.Configuration{ID: "base", Name: "base", Type: "postgres", Environment: "dev", Values: map[string]any{"host": "db"}, Status: domain.StatusPublished, Revision: 4, Version: 2, CreatedAt: now, UpdatedAt: now}
	dependent := domain.Configuration{ID: "dependent", Name: "derived", Type: "derived", Environment: "dev", Values: map[string]any{"source": "base"}, Status: domain.StatusPublished, Revision: 1, Version: 1, CreatedAt: now.Add(time.Second), UpdatedAt: now}
	if err := repository.Create(context.Background(), base); err != nil {
		t.Fatal(err)
	}
	if err := repository.Create(context.Background(), dependent); err != nil {
		t.Fatal(err)
	}
	_, err := system.lifecycle.Transition(context.Background(), "base", domain.StatusWithdrawn, 3, domain.Actor{ID: "publisher", Roles: []string{"publisher"}}, "req")
	if !errors.Is(err, domain.ErrVersionConflict) {
		t.Fatalf("stale transition err=%v", err)
	}
	err = system.lifecycle.Delete(context.Background(), "base", 4, domain.Actor{ID: "admin", Roles: []string{"platform_admin"}})
	if !errors.Is(err, domain.ErrReferenced) {
		t.Fatalf("delete err=%v", err)
	}
	current, err := repository.Get(context.Background(), "base")
	if err != nil || current.Status != domain.StatusPublished || current.Revision != 4 {
		t.Fatalf("current=%+v err=%v", current, err)
	}
}
