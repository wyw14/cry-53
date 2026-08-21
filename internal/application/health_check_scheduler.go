package application

import (
	"context"
	"sync"

	"github.com/wyw14/cry-053/internal/domain"
)

type healthCheckTask struct {
	index   int
	adapter HealthAdapter
}

type healthCheckOutcome struct {
	index  int
	result domain.HealthResult
}

type healthCheckScheduler struct {
	adapters    []HealthAdapter
	parallelism int
}

func newHealthCheckScheduler(adapters []HealthAdapter) healthCheckScheduler {
	owned := append([]HealthAdapter(nil), adapters...)
	return healthCheckScheduler{adapters: owned, parallelism: 1}
}

func (s healthCheckScheduler) Run(ctx context.Context, config domain.Configuration) ([]domain.HealthResult, error) {
	if len(s.adapters) == 0 {
		return []domain.HealthResult{}, nil
	}
	workerLimit := s.workerLimit()
	tasks := make(chan healthCheckTask)
	outcomes := make(chan healthCheckOutcome, len(s.adapters))
	var workers sync.WaitGroup

	for worker := 0; worker < workerLimit; worker += 1 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for task := range tasks {
				result := task.adapter.Check(ctx, config)
				select {
				case outcomes <- healthCheckOutcome{index: task.index, result: result}:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	go func() {
		workers.Wait()
		close(outcomes)
	}()

	for index, adapter := range s.adapters {
		select {
		case tasks <- healthCheckTask{index: index, adapter: adapter}:
		case <-ctx.Done():
			close(tasks)
			return nil, ctx.Err()
		}
	}
	close(tasks)

	results := make([]domain.HealthResult, len(s.adapters))
	for outcome := range outcomes {
		results[outcome.index] = outcome.result
	}
	return results, nil
}

func (s healthCheckScheduler) workerLimit() int {
	limit := s.parallelism
	if limit < 1 {
		limit = 1
	}
	if limit > len(s.adapters) {
		limit = len(s.adapters)
	}
	return limit
}
