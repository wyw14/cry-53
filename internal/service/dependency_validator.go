package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/wyw14/cry-053/internal/domain"
	"github.com/wyw14/cry-053/internal/repository"
)

type DependencyValidator struct {
	types   repository.TypeRepository
	configs repository.ConfigurationRepository
}

func NewDependencyValidator(types repository.TypeRepository, configs repository.ConfigurationRepository) *DependencyValidator {
	return &DependencyValidator{types: types, configs: configs}
}

func (v *DependencyValidator) Validate(ctx context.Context, configs []domain.Configuration) []domain.Issue {
	issues := make([]domain.Issue, 0)
	byID := make(map[string]domain.Configuration, len(configs))
	byKey := make(map[string]domain.Configuration, len(configs))
	for index, config := range configs {
		if previous, ok := byKey[config.Key()]; ok {
			issues = append(issues, issue("DUPLICATE_NAME", fmt.Sprintf("configurations[%d].name", index), "配置包内名称与环境重复", "合并或重命名 "+previous.Name))
		}
		byKey[config.Key()] = config
		byID[config.ID] = config
		if current, err := v.configs.GetByKey(ctx, config.Key()); err == nil && current.ID != config.ID {
			if config.ReplacesID != current.ID {
				issues = append(issues, issue("NAME_CONFLICT", fmt.Sprintf("configurations[%d].name", index), "已发布配置占用该名称", "设置 replaces_id 或更换名称"))
			}
		} else if err != nil && !errors.Is(err, domain.ErrNotFound) {
			issues = append(issues, issue("LOOKUP_FAILED", fmt.Sprintf("configurations[%d]", index), "无法检查现有配置", "稍后重试"))
		}
	}
	nodes := make([]domain.DependencyNode, 0, len(configs))
	for _, config := range configs {
		nodes = append(nodes, domain.DependencyNode{ID: config.ID, Environment: config.Environment})
	}
	graph := domain.NewDependencyGraph(nodes)
	for index, config := range configs {
		definition, err := v.types.Get(ctx, config.Type)
		if err != nil {
			continue
		}
		for _, field := range definition.Fields {
			if field.Kind != domain.FieldRef {
				continue
			}
			target, ok := config.Values[field.Name].(string)
			if !ok || target == "" {
				continue
			}
			targetConfig, inBundle := byID[target]
			if !inBundle {
				targetConfig, err = v.configs.Get(ctx, target)
				if err != nil {
					issues = append(issues, issue("MISSING_REFERENCE", fmt.Sprintf("configurations[%d].values.%s", index, field.Name), "引用的配置不存在", "将依赖加入配置包或修正引用标识"))
					continue
				}
			}
			// 环境边界约束与环检测是两件独立的事：先记录跨环境冲突，再把边
			// 加入图用于环检测。若让 Connect 因环境不同而拒绝加边，跨环境引用
			// 会被静默丢弃，既漏报环境冲突，也无法在跨环境闭环中检出循环引用。
			if targetConfig.Environment != config.Environment {
				issues = append(issues, issue("ENVIRONMENT_CONFLICT", fmt.Sprintf("configurations[%d].values.%s", index, field.Name), "引用跨越了环境边界", "引用同一环境中的配置版本"))
			}
			graph.Connect(domain.DependencyEdge{From: config.ID, To: target})
		}
	}
	for _, cycle := range graph.Cycles() {
		issues = append(issues, issue("REFERENCE_CYCLE", "configurations", "检测到循环引用: "+fmt.Sprint(cycle), "解除环中的至少一条引用"))
	}
	return issues
}
