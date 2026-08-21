package domain

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound          = newRuleViolation("RESOURCE_NOT_FOUND", "资源不存在", false)
	ErrConflict          = newRuleViolation("RESOURCE_CONFLICT", "资源与当前状态冲突", false)
	ErrInvalidTransition = newRuleViolation("STATE_TRANSITION_DENIED", "当前状态不允许该转换", false)
	ErrForbidden         = newRuleViolation("ACTOR_FORBIDDEN", "当前身份无权执行该操作", false)
	ErrReferenced        = newRuleViolation("CONFIGURATION_REFERENCED", "配置仍有活动引用", false)
	ErrVersionConflict   = newRuleViolation("REVISION_STALE", "配置版本已经变化", true)
	ErrIdempotencyReuse  = newRuleViolation("IDEMPOTENCY_PAYLOAD_MISMATCH", "幂等键已绑定其他请求内容", false)
)

type RuleViolation struct {
	Code       string
	PublicText string
	Retryable  bool
}

func newRuleViolation(code, publicText string, retryable bool) *RuleViolation {
	return &RuleViolation{Code: code, PublicText: publicText, Retryable: retryable}
}

func (v *RuleViolation) Error() string {
	return v.Code + ": " + v.PublicText
}

func Violation(err error) (*RuleViolation, bool) {
	var violation *RuleViolation
	if !errors.As(err, &violation) {
		return nil, false
	}
	return violation, true
}

type FieldError struct {
	Path    string `json:"path"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Line    int    `json:"line,omitempty"`
}

type ValidationError struct {
	Code   string       `json:"code"`
	Fields []FieldError `json:"fields"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %d field errors", e.Code, len(e.Fields))
}

func IsValidation(err error) bool {
	var target *ValidationError
	return errors.As(err, &target)
}
