package application

import (
	"context"
	"fmt"

	"github.com/wyw14/cry-053/internal/domain"
	"github.com/wyw14/cry-053/internal/repository"
)

type ImpactAnalyzer struct {
	configs  repository.ConfigurationRepository
	usages   repository.UsageRepository
	adapters []HealthAdapter
}

func NewImpactAnalyzer(configs repository.ConfigurationRepository, usages repository.UsageRepository, adapters ...HealthAdapter) *ImpactAnalyzer {
	return &ImpactAnalyzer{configs: configs, usages: usages, adapters: adapters}
}

func (a *ImpactAnalyzer) Analyze(ctx context.Context, id string, actor domain.Actor) (domain.ImpactReport, error) {
	if len(actor.Roles) == 0 {
		return domain.ImpactReport{}, domain.ErrForbidden
	}
	config, err := a.configs.Get(ctx, id)
	if err != nil {
		return domain.ImpactReport{}, err
	}
	usages, err := a.usages.ListByConfiguration(ctx, id)
	if err != nil {
		return domain.ImpactReport{}, err
	}
	scheduler := newHealthCheckScheduler(a.adapters)
	results, err := scheduler.Run(ctx, config)
	if err != nil {
		return domain.ImpactReport{}, err
	}
	blocking := false
	for _, usage := range usages {
		if usage.Criticality == "critical" {
			blocking = true
		}
	}
	for _, result := range results {
		if result.Status == domain.HealthUnhealthy {
			blocking = true
		}
	}
	summary := fmt.Sprintf("%d 个使用方，%d 项本地健康检查", len(usages), len(results))
	return domain.ImpactReport{ConfigurationID: id, Usages: usages, Health: results, Blocking: blocking, Summary: summary}, nil
}
