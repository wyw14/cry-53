package domain

import (
	"encoding/json"
	"regexp"
)

type FieldVerdict struct {
	Accepted bool
	Reason   string
}

type FieldPolicy struct {
	spec        FieldSpec
	environment string
}

func CompileFieldPolicy(spec FieldSpec, environment string) FieldPolicy {
	return FieldPolicy{spec: spec, environment: environment}
}

func (p FieldPolicy) Evaluate(value any) FieldVerdict {
	if !p.environmentAllowed() {
		return FieldVerdict{Reason: "字段不能用于当前环境"}
	}
	if p.spec.Sensitive {
		return p.evaluateSensitive(value)
	}
	return p.evaluateShape(value)
}

func (p FieldPolicy) environmentAllowed() bool {
	if len(p.spec.Environments) == 0 {
		return true
	}
	for _, candidate := range p.spec.Environments {
		if candidate == p.environment {
			return true
		}
	}
	return false
}

func (p FieldPolicy) evaluateSensitive(value any) FieldVerdict {
	if value == nil || value == "" {
		return FieldVerdict{Reason: "敏感字段不能为空"}
	}
	// 敏感字段仍须满足字段类型规范，避免数字/布尔等错误类型绕过校验。
	return p.evaluateShape(value)
}

func (p FieldPolicy) evaluateShape(value any) FieldVerdict {
	switch p.spec.Kind {
	case FieldString, FieldRef:
		text, ok := value.(string)
		if !ok {
			return FieldVerdict{Reason: "字段必须是字符串"}
		}
		if p.spec.Pattern != "" && !regexp.MustCompile(p.spec.Pattern).MatchString(text) {
			return FieldVerdict{Reason: "字段值不符合格式约束"}
		}
	case FieldNumber:
		number, ok := policyNumber(value)
		if !ok {
			return FieldVerdict{Reason: "字段必须是数字"}
		}
		if p.spec.Min != nil && number < *p.spec.Min {
			return FieldVerdict{Reason: "字段值低于范围"}
		}
		if p.spec.Max != nil && number > *p.spec.Max {
			return FieldVerdict{Reason: "字段值高于范围"}
		}
	case FieldBool:
		if _, ok := value.(bool); !ok {
			return FieldVerdict{Reason: "字段必须是布尔值"}
		}
	default:
		return FieldVerdict{Reason: "字段类型不受支持"}
	}
	return FieldVerdict{Accepted: true}
}

func policyNumber(value any) (float64, bool) {
	switch candidate := value.(type) {
	case float64:
		return candidate, true
	case json.Number:
		parsed, err := candidate.Float64()
		return parsed, err == nil
	case int:
		return float64(candidate), true
	default:
		return 0, false
	}
}
