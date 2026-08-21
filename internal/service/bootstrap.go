package service

import (
	"context"

	"github.com/wyw14/cry-053/internal/domain"
	"github.com/wyw14/cry-053/internal/repository"
)

func SeedTypes(ctx context.Context, repository repository.TypeRepository) error {
	minimum, maximum := 1.0, 65535.0
	types := []domain.ConfigType{
		{
			Name: "postgres", Description: "PostgreSQL 数据源连接",
			Environments: []string{"dev", "test", "staging", "prod"},
			Fields: []domain.FieldSpec{
				{Name: "host", Kind: domain.FieldString, Required: true, Pattern: `^[a-zA-Z0-9.-]+$`},
				{Name: "port", Kind: domain.FieldNumber, Required: true, Min: &minimum, Max: &maximum},
				{Name: "database", Kind: domain.FieldString, Required: true},
				{Name: "username", Kind: domain.FieldString, Required: true, Sensitive: true},
				{Name: "password", Kind: domain.FieldString, Required: true, Sensitive: true},
				{Name: "endpoint", Kind: domain.FieldString, Required: false},
			},
		},
		{
			Name: "http_api", Description: "离线环境内的 HTTP 数据源",
			Environments: []string{"dev", "test", "staging", "prod"},
			Fields: []domain.FieldSpec{
				{Name: "endpoint", Kind: domain.FieldString, Required: true, Pattern: `^https?://`},
				{Name: "token", Kind: domain.FieldString, Required: false, Sensitive: true},
				{Name: "timeout_ms", Kind: domain.FieldNumber, Required: true, Min: &minimum},
			},
		},
		{
			Name: "derived", Description: "引用另一数据源生成的派生配置",
			Environments: []string{"dev", "test", "staging", "prod"},
			Fields: []domain.FieldSpec{
				{Name: "source", Kind: domain.FieldRef, Required: true, Reference: "configuration"},
				{Name: "query", Kind: domain.FieldString, Required: true},
			},
		},
	}
	for _, item := range types {
		if err := item.ValidateDefinition(); err != nil {
			return err
		}
		if err := repository.Save(ctx, item); err != nil {
			return err
		}
	}
	return nil
}
