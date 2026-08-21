package application

import (
	"context"
	"testing"
	"time"

	"github.com/wyw14/cry-053/internal/domain"
	"github.com/wyw14/cry-053/internal/repository/memory"
)

func TestRollbackCreatesTraceableSuccessorVersion(t *testing.T) {
	system := newTestSystem(t)
	configs, versions := memory.Configurations(system.store), memory.Versions(system.store)
	now := system.clock.Now()
	current := domain.Configuration{ID: "cfg", Name: "primary", Type: "postgres", Environment: "prod", Values: map[string]any{"host": "new"}, Tags: []string{"stable"}, Status: domain.StatusPublished, Revision: 6, Version: 3, CreatedAt: now.Add(-time.Hour), UpdatedAt: now}
	if err := configs.Create(context.Background(), current); err != nil {
		t.Fatal(err)
	}
	if err := versions.Create(context.Background(), domain.Version{ID: "v1", ConfigurationID: "cfg", Number: 1, Values: map[string]any{"host": "old"}, Tags: []string{"known-good"}, Reason: "initial", CreatedBy: "admin", CreatedAt: now.Add(-2 * time.Hour)}); err != nil {
		t.Fatal(err)
	}
	rolled, err := system.lifecycle.Rollback(context.Background(), domain.RollbackRequest{ConfigurationID: "cfg", TargetVersion: 1, ExpectedRevision: 6, Reason: "health regression"}, domain.Actor{ID: "publisher", Roles: []string{"publisher"}}, "rollback-request")
	if err != nil {
		t.Fatal(err)
	}
	if rolled.Version != 4 || rolled.Revision != 7 || rolled.Values["host"] != "old" || rolled.Tags[0] != "known-good" {
		t.Fatalf("rolled=%+v", rolled)
	}
	created, err := versions.GetNumber(context.Background(), "cfg", 4)
	if err != nil || created.Reason != "rollback: health regression" {
		t.Fatalf("version=%+v err=%v", created, err)
	}
}
