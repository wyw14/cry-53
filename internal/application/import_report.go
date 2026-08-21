package application

import (
	"context"
	"sort"

	"github.com/wyw14/cry-053/internal/domain"
	"github.com/wyw14/cry-053/internal/repository"
)

type ImportReportService struct {
	bundles repository.BundleRepository
}

func NewImportReportService(bundles repository.BundleRepository) *ImportReportService {
	return &ImportReportService{bundles: bundles}
}

func (s *ImportReportService) Summary(ctx context.Context, id string, actor domain.Actor) (domain.ImportSummary, error) {
	if len(actor.Roles) == 0 {
		return domain.ImportSummary{}, domain.ErrForbidden
	}
	bundle, err := s.bundles.Get(ctx, id)
	if err != nil {
		return domain.ImportSummary{}, err
	}
	result := domain.ImportSummary{BundleID: id, Total: len(bundle.Configurations), IssuesByCode: make(map[string]int)}
	fixes := make(map[string]struct{})
	catalog := domain.NewSuggestionCatalog()
	for _, issue := range bundle.Issues {
		result.IssuesByCode[issue.Code]++
		if issue.Severity == domain.SeverityError {
			result.Errors++
		} else {
			result.Warnings++
		}
		if issue.Suggestion != "" {
			fixes[issue.Suggestion] = struct{}{}
		}
		if err := catalog.Observe(issue); err != nil {
			return domain.ImportSummary{}, err
		}
	}
	for fix := range fixes {
		result.SuggestedFixes = append(result.SuggestedFixes, fix)
	}
	sort.Strings(result.SuggestedFixes)
	merged, err := catalog.Seal(result.SuggestedFixes)
	if err != nil {
		return domain.ImportSummary{}, err
	}
	result.SuggestedFixes = merged
	return result, nil
}
