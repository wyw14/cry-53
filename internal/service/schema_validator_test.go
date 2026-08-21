package service

import (
	"context"
	"testing"

	"github.com/wyw14/cry-053/internal/domain"
	"github.com/wyw14/cry-053/internal/repository/memory"
)

func TestSchemaValidatorRejectsEnvironmentAndSensitiveShape(t *testing.T) {
	store := memory.NewStore()
	types := memory.Types(store)
	if err := types.Save(context.Background(), domain.ConfigType{Name: "warehouse", Description: "warehouse", Environments: []string{"prod"}, Fields: []domain.FieldSpec{{Name: "secret", Kind: domain.FieldString, Required: true, Sensitive: true}, {Name: "enabled", Kind: domain.FieldBool, Required: true}}}); err != nil {
		t.Fatal(err)
	}
	issues := NewSchemaValidator(types).Validate(context.Background(), []domain.Configuration{{Name: "analytics", Type: "warehouse", Environment: "dev", Values: map[string]any{"secret": 42, "enabled": "yes", "extra": true}}})
	codes := map[string]int{}
	for _, issue := range issues {
		codes[issue.Code]++
	}
	if codes["ENVIRONMENT_NOT_ALLOWED"] != 1 || codes["FIELD_INVALID"] != 2 || codes["UNKNOWN_FIELD"] != 1 {
		t.Fatalf("issues=%+v", issues)
	}
}
