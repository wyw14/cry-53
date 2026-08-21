package application

import (
	"context"
	"fmt"

	"github.com/wyw14/cry-053/internal/domain"
	"github.com/wyw14/cry-053/internal/repository"
)

const encryptedPrefix = "enc:v1:"

type SecretService struct {
	configs repository.ConfigurationRepository
	types   repository.TypeRepository
	cipher  Cipher
	audits  repository.AuditRepository
	clock   Clock
	ids     IDGenerator
}

func NewSecretService(configs repository.ConfigurationRepository, types repository.TypeRepository, cipher Cipher, audits repository.AuditRepository, clock Clock, ids IDGenerator) *SecretService {
	return &SecretService{configs: configs, types: types, cipher: cipher, audits: audits, clock: clock, ids: ids}
}

func (s *SecretService) EncryptSensitive(ctx context.Context, item domain.Configuration) (domain.Configuration, error) {
	definition, err := s.types.Get(ctx, item.Type)
	if err != nil {
		return domain.Configuration{}, err
	}
	result := item.Clone()
	for _, field := range definition.Fields {
		if !field.Sensitive {
			continue
		}
		value, ok := result.Values[field.Name].(string)
		if !ok || value == "" {
			continue
		}
		if len(value) >= len(encryptedPrefix) && value[:len(encryptedPrefix)] == encryptedPrefix {
			continue
		}
		encrypted, err := s.cipher.Encrypt(ctx, value)
		if err != nil {
			return domain.Configuration{}, fmt.Errorf("encrypt field %s: %w", field.Name, err)
		}
		result.Values[field.Name] = encryptedPrefix + encrypted
	}
	return result, nil
}

func (s *SecretService) Export(ctx context.Context, id string, policy domain.ExportPolicy, actor domain.Actor, requestID string) (domain.ExportedConfiguration, error) {
	item, err := s.configs.Get(ctx, id)
	if err != nil {
		return domain.ExportedConfiguration{}, err
	}
	definition, err := s.types.Get(ctx, item.Type)
	if err != nil {
		return domain.ExportedConfiguration{}, err
	}
	decision := domain.EvaluateDisclosure(policy, actor, s.clock.Now())
	values := item.Clone().Values
	masked := false
	for _, field := range definition.Fields {
		if !field.Sensitive {
			continue
		}
		value, ok := values[field.Name].(string)
		if !ok || value == "" {
			continue
		}
		if decision.ShouldMask() {
			values[field.Name] = maskSecret(value)
			masked = true
			continue
		}
		if len(value) < len(encryptedPrefix) || value[:len(encryptedPrefix)] != encryptedPrefix {
			return domain.ExportedConfiguration{}, fmt.Errorf("sensitive value is not encrypted: %w", domain.ErrConflict)
		}
		decrypted, err := s.cipher.Decrypt(ctx, value[len(encryptedPrefix):])
		if err != nil {
			return domain.ExportedConfiguration{}, err
		}
		values[field.Name] = decrypted
	}
	result := domain.ExportedConfiguration{ID: item.ID, Name: item.Name, Type: item.Type, Environment: item.Environment, Version: item.Version, Values: values, Masked: masked, ExportedAt: s.clock.Now()}
	_ = s.audits.Append(ctx, domain.AuditEvent{ID: s.ids.New("audit"), RequestID: requestID, ActorID: actor.ID, Operation: "configuration.export", Target: domain.AuditTarget{Kind: "configuration", ID: id}, Facts: decision.AuditFacts(), OccurredAt: s.clock.Now()})
	return result, nil
}

func maskSecret(value string) string {
	if len(value) <= 4 {
		return "******"
	}
	return value[:2] + "******" + value[len(value)-2:]
}
