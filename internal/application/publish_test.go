package application

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/wyw14/cry-053/internal/domain"
	"github.com/wyw14/cry-053/internal/repository/memory"
)

func TestBundlePublishConcurrentReplayWritesOnce(t *testing.T) {
	system := newTestSystem(t)
	payload := `{"schema_version":1,"configurations":[{"name":"primary","type":"postgres","environment":"dev","values":{"host":"db","port":5432,"database":"app","username":"u","password":"p"}}]}`
	bundle, err := system.workflow.Upload(context.Background(), UploadCommand{Filename: "bundle.json", Reader: strings.NewReader(payload), IdempotencyKey: "upload-concurrent", Actor: domain.Actor{ID: "alice", Roles: []string{"config_editor"}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = system.workflow.Validate(context.Background(), bundle.ID, domain.Actor{ID: "alice", Roles: []string{"config_editor"}}); err != nil {
		t.Fatal(err)
	}
	if _, err = system.workflow.Approve(context.Background(), bundle.ID, domain.Actor{ID: "bob", Roles: []string{"approver"}}); err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	errors := make(chan error, 2)
	var wg sync.WaitGroup
	for index := 0; index < 2; index++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, publishErr := system.workflow.Publish(context.Background(), PublishCommand{BundleID: bundle.ID, IdempotencyKey: "publish-concurrent", Actor: domain.Actor{ID: "carol", Roles: []string{"publisher"}}, RequestID: "request-concurrent"})
			errors <- publishErr
		}()
	}
	close(start)
	wg.Wait()
	close(errors)
	for publishErr := range errors {
		if publishErr != nil {
			t.Fatalf("publish err=%v", publishErr)
		}
	}
	items, total, err := memory.Configurations(system.store).List(context.Background(), domain.PageRequest{Page: 1, Size: 10, Sort: "name", Filters: map[string]string{}})
	if err != nil || total != 1 || len(items) != 1 || items[0].Version != 1 {
		t.Fatalf("items=%+v total=%d err=%v", items, total, err)
	}
}
