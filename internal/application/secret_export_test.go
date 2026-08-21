package application

import (
	"context"
	"testing"
	"time"

	"github.com/wyw14/cry-053/internal/domain"
	"github.com/wyw14/cry-053/internal/repository/memory"
)

func TestSecretExportMasksByDefaultAndAuditsAuthorizedReveal(t *testing.T) {
	system := newTestSystem(t)
	plain := domain.Configuration{ID: "cfg", Name: "primary", Type: "postgres", Environment: "prod", Values: map[string]any{"host": "db", "port": 5432, "database": "app", "username": "alice", "password": "super-secret"}, Status: domain.StatusPublished, Revision: 1, Version: 1, CreatedAt: system.clock.Now(), UpdatedAt: system.clock.Now()}
	encrypted, err := system.secrets.EncryptSensitive(context.Background(), plain)
	if err != nil {
		t.Fatal(err)
	}
	if err := memory.Configurations(system.store).Create(context.Background(), encrypted); err != nil {
		t.Fatal(err)
	}
	masked, err := system.secrets.Export(context.Background(), "cfg", domain.ExportPolicy{}, domain.Actor{ID: "viewer", Roles: []string{"viewer"}}, "masked-request")
	if err != nil || !masked.Masked || masked.Values["password"] == "super-secret" {
		t.Fatalf("masked=%+v err=%v", masked, err)
	}
	revealed, err := system.secrets.Export(context.Background(), "cfg", domain.ExportPolicy{AllowSensitive: true, Reason: "incident recovery", ExpiresAt: system.clock.Now().Add(time.Minute)}, domain.Actor{ID: "security", Roles: []string{"secret_exporter"}}, "reveal-request")
	if err != nil || revealed.Masked || revealed.Values["password"] != "super-secret" || revealed.Values["username"] != "alice" {
		t.Fatalf("revealed=%+v err=%v", revealed, err)
	}
	audits, total, err := memory.Audits(system.store).List(context.Background(), domain.PageRequest{Page: 1, Size: 10})
	if err != nil || total != 2 || audits[1].Facts["reason"] != "incident recovery" || !audits[1].ValidAfter(audits[0].Hash) {
		t.Fatalf("audits=%+v total=%d err=%v", audits, total, err)
	}
}
