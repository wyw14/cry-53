package application

import (
	"context"
	"testing"
	"time"

	"github.com/wyw14/cry-053/internal/domain"
	"github.com/wyw14/cry-053/internal/repository/memory"
)

func TestImportReportAggregatesIssuesAndUniqueSuggestions(t *testing.T) {
	store := memory.NewStore()
	bundles := memory.Bundles(store)
	bundle := domain.Bundle{
		ID: "bundle", Filename: "invalid.json", State: domain.BundleRejected,
		Configurations: []domain.Configuration{{ID: "a"}, {ID: "b"}}, UploadedAt: time.Now(),
		Issues: []domain.Issue{
			{Code: "MISSING_REFERENCE", Severity: domain.SeverityError, Suggestion: "补充依赖配置"},
			{Code: "MISSING_REFERENCE", Severity: domain.SeverityError, Suggestion: "补充依赖配置"},
			{Code: "EXPIRING_SOON", Severity: domain.SeverityWarning, Suggestion: "延长过期时间"},
		},
	}
	if err := bundles.Create(context.Background(), bundle); err != nil {
		t.Fatal(err)
	}
	report, err := NewImportReportService(bundles).Summary(context.Background(), "bundle", domain.Actor{ID: "auditor", Roles: []string{"auditor"}})
	if err != nil {
		t.Fatal(err)
	}
	if report.Total != 2 || report.Errors != 2 || report.Warnings != 1 {
		t.Fatalf("report=%+v", report)
	}
	if report.IssuesByCode["MISSING_REFERENCE"] != 2 || len(report.SuggestedFixes) != 2 {
		t.Fatalf("report=%+v", report)
	}
	if report.SuggestedFixes[0] != "延长过期时间" || report.SuggestedFixes[1] != "补充依赖配置" {
		t.Fatalf("suggestions=%v", report.SuggestedFixes)
	}
}
