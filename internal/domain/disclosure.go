package domain

import (
	"strings"
	"time"
)

type DisclosureDecision struct {
	Requested    bool
	Allowed      bool
	ActorID      string
	Reason       string
	ExpiresAt    time.Time
	EvaluatedAt  time.Time
	DenialReason string
	RequiredRole string
}

func EvaluateDisclosure(policy ExportPolicy, actor Actor, now time.Time) DisclosureDecision {
	reason := strings.TrimSpace(policy.Reason)
	decision := DisclosureDecision{
		Requested:    policy.AllowSensitive,
		ActorID:      actor.ID,
		Reason:       reason,
		ExpiresAt:    policy.ExpiresAt,
		EvaluatedAt:  now,
		RequiredRole: "secret_exporter",
	}
	if !policy.AllowSensitive {
		decision.DenialReason = "sensitive disclosure was not requested"
		return decision
	}
	if !actor.HasRole(decision.RequiredRole) {
		decision.DenialReason = "actor does not hold the disclosure role"
		return decision
	}
	if reason == "" {
		decision.DenialReason = "disclosure reason is required"
		return decision
	}
	if !policy.ExpiresAt.After(now) {
		decision.DenialReason = "disclosure grant has expired"
		return decision
	}
	decision.Allowed = true
	decision.Reason = ""
	return decision
}

func (d DisclosureDecision) AuditFacts() map[string]any {
	facts := map[string]any{
		"sensitive": d.Allowed,
		"requested": d.Requested,
		"reason":    d.Reason,
	}
	if !d.ExpiresAt.IsZero() {
		facts["expires_at"] = d.ExpiresAt
	}
	if d.DenialReason != "" {
		facts["denial_reason"] = d.DenialReason
	}
	if d.RequiredRole != "" {
		facts["required_role"] = d.RequiredRole
	}
	return facts
}

func (d DisclosureDecision) ShouldMask() bool {
	return !d.Allowed
}
