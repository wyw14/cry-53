package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/wyw14/cry-053/internal/domain"
	"github.com/wyw14/cry-053/internal/repository"
)

type TypeCatalog struct {
	repository repository.TypeRepository
	validate   *validator.Validate
}

func NewTypeCatalog(repository repository.TypeRepository) *TypeCatalog {
	return &TypeCatalog{repository: repository, validate: validator.New()}
}

func (c *TypeCatalog) Save(ctx context.Context, actor domain.Actor, item domain.ConfigType) error {
	if !actor.HasRole("platform_admin") {
		return domain.ErrForbidden
	}
	if err := c.validate.Struct(item); err != nil {
		return fmt.Errorf("validate type: %w", err)
	}
	if err := item.ValidateDefinition(); err != nil {
		return err
	}
	current, err := c.repository.Get(ctx, item.Name)
	if err == nil {
		if err := allowDefinitionEvolution(current, item, actor); err != nil {
			return err
		}
	} else if !errors.Is(err, domain.ErrNotFound) {
		return fmt.Errorf("load existing type: %w", err)
	}
	return c.repository.Save(ctx, item)
}

func (c *TypeCatalog) List(ctx context.Context) ([]domain.ConfigType, error) {
	return c.repository.List(ctx)
}

func allowDefinitionEvolution(before, after domain.ConfigType, actor domain.Actor) error {
	previous := make(map[string]domain.FieldSpec, len(before.Fields))
	for _, field := range before.Fields {
		previous[field.Name] = field
	}
	for _, field := range after.Fields {
		old, existed := previous[field.Name]
		if !existed {
			continue
		}
		if old.Sensitive && !field.Sensitive && !actor.HasRole("security_admin") {
			return fmt.Errorf("sensitive marker removal for %s: %w", field.Name, domain.ErrForbidden)
		}
		if old.Kind != field.Kind {
			return fmt.Errorf("field kind change for %s: %w", field.Name, domain.ErrConflict)
		}
		delete(previous, field.Name)
	}
	for name, removed := range previous {
		if removed.Required {
			return fmt.Errorf("required field removal for %s: %w", name, domain.ErrConflict)
		}
	}
	return nil
}
