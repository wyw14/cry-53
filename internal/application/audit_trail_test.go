package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/wyw14/cry-053/internal/application"
	"github.com/wyw14/cry-053/internal/domain"
	"github.com/wyw14/cry-053/internal/repository/memory"
)

func TestAuditTrailFiltersNewestFirstAndValidatesChain(t *testing.T) {
	store := memory.NewStore()
	audits := memory.Audits(store)
	first := domain.AuditEvent{ID: "audit-1", ActorID: "publisher", Operation: "configuration.publish", Target: domain.AuditTarget{Kind: "configuration", ID: "source"}, OccurredAt: time.Unix(1, 0)}
	second := domain.AuditEvent{ID: "audit-2", ActorID: "publisher", Operation: "configuration.publish", Target: domain.AuditTarget{Kind: "configuration", ID: "warehouse"}, OccurredAt: time.Unix(2, 0)}
	third := domain.AuditEvent{ID: "audit-3", ActorID: "auditor", Operation: "configuration.export", Target: domain.AuditTarget{Kind: "configuration", ID: "warehouse"}, OccurredAt: time.Unix(3, 0)}
	for _, event := range []domain.AuditEvent{first, second, third} {
		if err := audits.Append(context.Background(), event); err != nil {
			t.Fatal(err)
		}
	}

	trail := application.NewAuditTrail(audits)
	page, err := trail.Browse(context.Background(), domain.Actor{ID: "reviewer", Roles: []string{"auditor"}}, application.AuditQuery{Operation: "configuration.publish"})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 2 || len(page.Items) != 2 || page.Items[0].ID != "audit-2" || page.Items[1].ID != "audit-1" {
		t.Fatalf("unexpected page: %+v", page)
	}
	if !page.ChainValid {
		t.Fatal("expected filtered audit page to preserve a verifiable chain")
	}
	if _, err := trail.Browse(context.Background(), domain.Actor{ID: "editor", Roles: []string{"config_editor"}}, application.AuditQuery{}); err != domain.ErrForbidden {
		t.Fatalf("expected forbidden, got %v", err)
	}
}
