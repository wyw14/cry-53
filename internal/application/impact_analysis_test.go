package application

import (
	"context"
	"testing"
	"time"

	"github.com/wyw14/cry-053/internal/domain"
	"github.com/wyw14/cry-053/internal/repository/memory"
)

func TestImpactAnalysisRunsAdaptersTogetherAndMarksBlocking(t *testing.T) {
	system := newTestSystem(t)
	config := domain.Configuration{ID: "cfg", Name: "primary", Type: "http_api", Environment: "prod", Values: map[string]any{"endpoint": "http://localhost"}, Status: domain.StatusPublished, Revision: 1, Version: 1, CreatedAt: system.clock.Now(), UpdatedAt: system.clock.Now()}
	if err := memory.Configurations(system.store).Create(context.Background(), config); err != nil {
		t.Fatal(err)
	}
	if err := memory.Usages(system.store).Save(context.Background(), domain.Usage{ID: "usage", ConfigurationID: "cfg", Consumer: "report-worker", Environment: "prod", Criticality: "critical", LastObservedAt: system.clock.Now()}); err != nil {
		t.Fatal(err)
	}
	gate := make(chan struct{})
	started := make(chan string, 2)
	analyzer := NewImpactAnalyzer(memory.Configurations(system.store), memory.Usages(system.store), barrierAdapter{name: "schema", gate: gate, started: started, status: domain.HealthHealthy}, barrierAdapter{name: "network", gate: gate, started: started, status: domain.HealthUnhealthy})
	result := make(chan domain.ImpactReport, 1)
	go func() {
		report, _ := analyzer.Analyze(context.Background(), "cfg", domain.Actor{ID: "viewer", Roles: []string{"viewer"}})
		result <- report
	}()
	timeout := time.After(time.Second)
	for count := 0; count < 2; count++ {
		select {
		case <-started:
		case <-timeout:
			t.Fatal("health adapters did not overlap")
		}
	}
	close(gate)
	report := <-result
	if !report.Blocking || len(report.Health) != 2 || len(report.Usages) != 1 {
		t.Fatalf("report=%+v", report)
	}
}
