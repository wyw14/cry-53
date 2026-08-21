package application

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wyw14/cry-053/internal/domain"
	platformcrypto "github.com/wyw14/cry-053/internal/platform/crypto"
	"github.com/wyw14/cry-053/internal/platform/notify"
	"github.com/wyw14/cry-053/internal/repository/memory"
	"github.com/wyw14/cry-053/internal/service"
)

type fixedClock struct{ value time.Time }

func (c fixedClock) Now() time.Time { return c.value }

type sequenceIDs struct{ value atomic.Int64 }

func (g *sequenceIDs) New(prefix string) string {
	return fmt.Sprintf("%s_%03d", prefix, g.value.Add(1))
}

type testSystem struct {
	store     *memory.Store
	workflow  *BundleWorkflow
	lifecycle *LifecycleService
	secrets   *SecretService
	impact    *ImpactAnalyzer
	clock     fixedClock
}

func newTestSystem(t testing.TB) *testSystem {
	t.Helper()
	store := memory.NewStore()
	types := memory.Types(store)
	if err := service.SeedTypes(context.Background(), types); err != nil {
		t.Fatal(err)
	}
	configs, bundles := memory.Configurations(store), memory.Bundles(store)
	versions, audits, usages := memory.Versions(store), memory.Audits(store), memory.Usages(store)
	clock := fixedClock{value: time.Date(2026, 8, 21, 10, 0, 0, 0, time.UTC)}
	ids := &sequenceIDs{}
	workflow := NewBundleWorkflow(service.NewBundleParser(1<<20), service.NewSchemaValidator(types), service.NewDependencyValidator(types, configs), service.NewDiffService(configs, types), bundles, memory.NewManager(store), clock, ids, &notify.MemoryNotifier{})
	cipher, err := platformcrypto.NewAESCipher("0123456789abcdef0123456789abcdef")
	if err != nil {
		t.Fatal(err)
	}
	return &testSystem{store: store, workflow: workflow, lifecycle: NewLifecycleService(configs, versions, audits, clock, ids), secrets: NewSecretService(configs, types, cipher, audits, clock, ids), impact: NewImpactAnalyzer(configs, usages), clock: clock}
}

type barrierAdapter struct {
	name    string
	gate    <-chan struct{}
	started chan<- string
	status  domain.HealthStatus
}

func (a barrierAdapter) Name() string { return a.name }
func (a barrierAdapter) Check(ctx context.Context, config domain.Configuration) domain.HealthResult {
	a.started <- a.name
	select {
	case <-a.gate:
		return domain.HealthResult{ConfigurationID: config.ID, Adapter: a.name, Status: a.status, CheckedAt: time.Now()}
	case <-ctx.Done():
		return domain.HealthResult{ConfigurationID: config.ID, Adapter: a.name, Status: domain.HealthUnhealthy, Message: ctx.Err().Error(), CheckedAt: time.Now()}
	}
}
