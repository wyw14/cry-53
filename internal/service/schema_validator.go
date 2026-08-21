package service

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/wyw14/cry-053/internal/domain"
	"github.com/wyw14/cry-053/internal/repository"
)

type SchemaValidator struct {
	types repository.TypeRepository
}

func NewSchemaValidator(types repository.TypeRepository) *SchemaValidator {
	return &SchemaValidator{types: types}
}

func (v *SchemaValidator) Validate(ctx context.Context, configs []domain.Configuration) []domain.Issue {
	issues := make([]domain.Issue, 0)
	for index, config := range configs {
		path := fmt.Sprintf("configurations[%d]", index)
		definition, err := v.types.Get(ctx, config.Type)
		if err != nil {
			issues = append(issues, issue("UNKNOWN_TYPE", path+".type", "配置类型不存在", "先在类型管理中登记字段规范"))
			continue
		}
		if config.Name == "" {
			issues = append(issues, issue("NAME_REQUIRED", path+".name", "配置名称不能为空", "填写同一环境内唯一的名称"))
		}
		if !definition.SupportsEnvironment(config.Environment) {
			issues = append(issues, issue("ENVIRONMENT_NOT_ALLOWED", path+".environment", "配置类型不允许用于该环境", "从类型允许的环境范围中选择"))
		}
		for _, field := range definition.Fields {
			value, present := config.Values[field.Name]
			if field.Required && (!present || value == nil || value == "") {
				issues = append(issues, issue("FIELD_REQUIRED", path+".values."+field.Name, "缺少必填字段", "补充字段后重新校验"))
				continue
			}
			if !present {
				continue
			}
			if message := validateValue(field, value, config.Environment); message != "" {
				issues = append(issues, issue("FIELD_INVALID", path+".values."+field.Name, message, "按字段规范修正值或环境范围"))
			}
		}
		known := make(map[string]struct{}, len(definition.Fields))
		for _, field := range definition.Fields {
			known[field.Name] = struct{}{}
		}
		unknown := make([]string, 0)
		for key := range config.Values {
			if _, ok := known[key]; !ok {
				unknown = append(unknown, key)
			}
		}
		sort.Strings(unknown)
		for _, key := range unknown {
			issues = append(issues, issue("UNKNOWN_FIELD", path+".values."+key, "字段未在类型规范中声明", "删除该字段或更新类型规范"))
		}
	}
	return issues
}

func validateValue(spec domain.FieldSpec, value any, environment string) string {
	if len(spec.Environments) > 0 && !contains(spec.Environments, environment) {
		return "字段不能用于当前环境"
	}
	switch spec.Kind {
	case domain.FieldString, domain.FieldRef:
		text, ok := value.(string)
		if !ok {
			return "字段必须是字符串"
		}
		if spec.Pattern != "" && !regexp.MustCompile(spec.Pattern).MatchString(text) {
			return "字段值不符合格式约束"
		}
	case domain.FieldNumber:
		number, ok := numeric(value)
		if !ok {
			return "字段必须是数字"
		}
		if spec.Min != nil && number < *spec.Min {
			return fmt.Sprintf("字段值不能小于 %v", *spec.Min)
		}
		if spec.Max != nil && number > *spec.Max {
			return fmt.Sprintf("字段值不能大于 %v", *spec.Max)
		}
	case domain.FieldBool:
		if _, ok := value.(bool); !ok {
			return "字段必须是布尔值"
		}
	default:
		return "字段类型不受支持"
	}
	return ""
}

func numeric(value any) (float64, bool) {
	switch number := value.(type) {
	case float64:
		return number, true
	case json.Number:
		parsed, err := number.Float64()
		return parsed, err == nil
	case int:
		return float64(number), true
	default:
		return 0, false
	}
}

func issue(code, path, message, suggestion string) domain.Issue {
	return domain.Issue{Code: code, Severity: domain.SeverityError, Path: path, Message: message, Suggestion: suggestion}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(value, target) {
			return true
		}
	}
	return false
}
