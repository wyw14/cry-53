package service

import (
	"context"
	"testing"

	"github.com/wyw14/cry-053/internal/domain"
	"github.com/wyw14/cry-053/internal/repository/memory"
)

func TestDependencyValidatorFindsCycleAndEnvironmentConflict(t *testing.T) {
	store := memory.NewStore()
	types := memory.Types(store)
	if err := types.Save(context.Background(), domain.ConfigType{Name: "derived", Description: "derived", Environments: []string{"dev", "prod"}, Fields: []domain.FieldSpec{{Name: "source", Kind: domain.FieldRef, Required: true, Reference: "configuration"}}}); err != nil {
		t.Fatal(err)
	}
	configs := []domain.Configuration{
		{ID: "a", Name: "a", Type: "derived", Environment: "dev", Values: map[string]any{"source": "b"}},
		{ID: "b", Name: "b", Type: "derived", Environment: "prod", Values: map[string]any{"source": "a"}},
	}
	issues := NewDependencyValidator(types, memory.Configurations(store)).Validate(context.Background(), configs)
	hasCycle, hasEnvironment := false, false
	for _, issue := range issues {
		hasCycle = hasCycle || issue.Code == "REFERENCE_CYCLE"
		hasEnvironment = hasEnvironment || issue.Code == "ENVIRONMENT_CONFLICT"
	}
	if !hasCycle || !hasEnvironment {
		t.Fatalf("issues=%+v", issues)
	}
}
