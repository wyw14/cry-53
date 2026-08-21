package application

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/wyw14/cry-053/internal/domain"
)

func TestBundleApprovalRequiresValidatedBundleAndDifferentActor(t *testing.T) {
	system := newTestSystem(t)
	payload := `{"schema_version":1,"configurations":[{"name":"primary","type":"postgres","environment":"dev","values":{"host":"db","port":5432,"database":"app","username":"u","password":"p"}}]}`
	bundle, err := system.workflow.Upload(context.Background(), UploadCommand{Filename: "bundle.json", Reader: strings.NewReader(payload), IdempotencyKey: "approval", Actor: domain.Actor{ID: "alice", Roles: []string{"config_editor"}}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = system.workflow.Approve(context.Background(), bundle.ID, domain.Actor{ID: "bob", Roles: []string{"approver"}})
	if !errors.Is(err, domain.ErrInvalidTransition) {
		t.Fatalf("unvalidated approval err=%v", err)
	}
	if _, err := system.workflow.Validate(context.Background(), bundle.ID, domain.Actor{ID: "alice", Roles: []string{"config_editor"}}); err != nil {
		t.Fatal(err)
	}
	_, err = system.workflow.Approve(context.Background(), bundle.ID, domain.Actor{ID: "alice", Roles: []string{"approver"}})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("self approval err=%v", err)
	}
	approved, err := system.workflow.Approve(context.Background(), bundle.ID, domain.Actor{ID: "bob", Roles: []string{"approver"}})
	if err != nil || approved.State != domain.BundleApproved {
		t.Fatalf("approved=%+v err=%v", approved, err)
	}
}
