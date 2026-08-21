package repository

import (
	"context"

	"github.com/wyw14/cry-053/internal/domain"
)

type Transaction interface {
	Configurations() ConfigurationRepository
	Bundles() BundleRepository
	Versions() VersionRepository
	Audits() AuditRepository
	Commit(context.Context) error
	Rollback(context.Context) error
}

type TransactionManager interface {
	Begin(context.Context) (Transaction, error)
}

type ConfigurationRepository interface {
	Get(context.Context, string) (domain.Configuration, error)
	GetByKey(context.Context, string) (domain.Configuration, error)
	List(context.Context, domain.PageRequest) ([]domain.Configuration, int, error)
	Create(context.Context, domain.Configuration) error
	Update(context.Context, domain.Configuration, int64) error
	Delete(context.Context, string, int64) error
	References(context.Context, string) ([]domain.Configuration, error)
}

type BundleRepository interface {
	Get(context.Context, string) (domain.Bundle, error)
	Create(context.Context, domain.Bundle) error
	Update(context.Context, domain.Bundle) error
	FindByIdempotency(context.Context, string) (string, string, error)
	SaveIdempotency(context.Context, string, string, string) error
}

type VersionRepository interface {
	List(context.Context, string) ([]domain.Version, error)
	GetNumber(context.Context, string, int) (domain.Version, error)
	Create(context.Context, domain.Version) error
}

type AuditRepository interface {
	Append(context.Context, domain.AuditEvent) error
	List(context.Context, domain.PageRequest) ([]domain.AuditEvent, int, error)
}

type TypeRepository interface {
	Get(context.Context, string) (domain.ConfigType, error)
	List(context.Context) ([]domain.ConfigType, error)
	Save(context.Context, domain.ConfigType) error
}

type UsageRepository interface {
	ListByConfiguration(context.Context, string) ([]domain.Usage, error)
	Save(context.Context, domain.Usage) error
}
