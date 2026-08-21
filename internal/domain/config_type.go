package domain

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type FieldKind string

const (
	FieldString FieldKind = "string"
	FieldNumber FieldKind = "number"
	FieldBool   FieldKind = "boolean"
	FieldRef    FieldKind = "reference"
)

type FieldSpec struct {
	Name         string    `json:"name" validate:"required,lowercase"`
	Kind         FieldKind `json:"kind" validate:"required,oneof=string number boolean reference"`
	Required     bool      `json:"required"`
	Sensitive    bool      `json:"sensitive"`
	Reference    string    `json:"reference,omitempty"`
	Pattern      string    `json:"pattern,omitempty"`
	Min          *float64  `json:"min,omitempty"`
	Max          *float64  `json:"max,omitempty"`
	Environments []string  `json:"environments,omitempty"`
}

type ConfigType struct {
	Name         string      `json:"name" validate:"required,lowercase"`
	Description  string      `json:"description" validate:"required"`
	Fields       []FieldSpec `json:"fields" validate:"required,min=1,dive"`
	Environments []string    `json:"environments" validate:"required,min=1,dive,oneof=dev test staging prod"`
}

func (t ConfigType) ValidateDefinition() error {
	seen := make(map[string]struct{}, len(t.Fields))
	for _, field := range t.Fields {
		if _, ok := seen[field.Name]; ok {
			return fmt.Errorf("duplicate field %q: %w", field.Name, ErrConflict)
		}
		seen[field.Name] = struct{}{}
		if field.Kind == FieldRef && field.Reference == "" {
			return fmt.Errorf("reference field %q requires target", field.Name)
		}
		if field.Sensitive && field.Kind == FieldRef {
			return fmt.Errorf("reference field %q cannot be sensitive", field.Name)
		}
		if field.Pattern != "" {
			if _, err := regexp.Compile(field.Pattern); err != nil {
				return fmt.Errorf("field %q pattern: %w", field.Name, err)
			}
		}
	}
	if len(normalizeStrings(t.Environments)) != len(t.Environments) {
		return fmt.Errorf("duplicate environments: %w", ErrConflict)
	}
	return nil
}

func (t ConfigType) Field(name string) (FieldSpec, bool) {
	for _, field := range t.Fields {
		if field.Name == name {
			return field, true
		}
	}
	return FieldSpec{}, false
}

func (t ConfigType) SupportsEnvironment(environment string) bool {
	for _, candidate := range t.Environments {
		if candidate == environment {
			return true
		}
	}
	return false
}

func normalizeStrings(values []string) []string {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			set[value] = struct{}{}
		}
	}
	result := make([]string, 0, len(set))
	for value := range set {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
