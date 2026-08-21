package application

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/wyw14/cry-053/internal/domain"
)

func TestBundleUploadIdempotencyRejectsPayloadReuse(t *testing.T) {
	system := newTestSystem(t)
	actor := domain.Actor{ID: "editor", Roles: []string{"config_editor"}}
	first := `{"schema_version":1,"configurations":[{"name":"primary","type":"postgres","environment":"dev","values":{"host":"db","port":5432,"database":"app","username":"u","password":"p"}}]}`
	second := strings.Replace(first, "primary", "replica", 1)
	bundle, err := system.workflow.Upload(context.Background(), UploadCommand{Filename: "one.json", Reader: strings.NewReader(first), IdempotencyKey: "same-key", Actor: actor})
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := system.workflow.Upload(context.Background(), UploadCommand{Filename: "one.json", Reader: strings.NewReader(first), IdempotencyKey: "same-key", Actor: actor})
	if err != nil || replayed.ID != bundle.ID {
		t.Fatalf("replay=%+v err=%v", replayed, err)
	}
	_, err = system.workflow.Upload(context.Background(), UploadCommand{Filename: "two.json", Reader: strings.NewReader(second), IdempotencyKey: "same-key", Actor: actor})
	if !errors.Is(err, domain.ErrIdempotencyReuse) {
		t.Fatalf("expected reuse error, got %v", err)
	}
}
