package service

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/wyw14/cry-053/internal/domain"
)

type BundleParser struct {
	maxBytes int64
	now      func() time.Time
}

type bundleDocument struct {
	SchemaVersion  int                  `json:"schema_version"`
	Configurations []configurationInput `json:"configurations"`
}

type configurationInput struct {
	Name        string         `json:"name"`
	Type        string         `json:"type"`
	Environment string         `json:"environment"`
	Values      map[string]any `json:"values"`
	Tags        []string       `json:"tags,omitempty"`
	ExpiresAt   *time.Time     `json:"expires_at,omitempty"`
	ReplacesID  string         `json:"replaces_id,omitempty"`
}

func NewBundleParser(maxBytes int64) *BundleParser {
	return &BundleParser{maxBytes: maxBytes, now: time.Now}
}

func (p *BundleParser) Parse(filename string, reader io.Reader, actor string) (domain.Bundle, error) {
	if !strings.HasSuffix(strings.ToLower(filename), ".json") {
		return domain.Bundle{}, &domain.ValidationError{Code: "UNSUPPORTED_FILE", Fields: []domain.FieldError{{Path: "file", Code: "json_required", Message: "只支持 JSON 配置包"}}}
	}
	limited := io.LimitReader(reader, p.maxBytes+1)
	payload, err := io.ReadAll(limited)
	if err != nil {
		return domain.Bundle{}, fmt.Errorf("read bundle: %w", err)
	}
	if int64(len(payload)) > p.maxBytes {
		return domain.Bundle{}, &domain.ValidationError{Code: "FILE_TOO_LARGE", Fields: []domain.FieldError{{Path: "file", Code: "max_bytes", Message: fmt.Sprintf("配置包不能超过 %d 字节", p.maxBytes)}}}
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	var document bundleDocument
	if err := decoder.Decode(&document); err != nil {
		line, column := locateJSONError(payload, err)
		return domain.Bundle{}, &domain.ValidationError{Code: "INVALID_JSON", Fields: []domain.FieldError{{Path: "$", Code: "syntax", Message: fmt.Sprintf("JSON 格式错误（列 %d）", column), Line: line}}}
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return domain.Bundle{}, &domain.ValidationError{Code: "INVALID_JSON", Fields: []domain.FieldError{{Path: "$", Code: "trailing_data", Message: "JSON 根对象后存在多余内容"}}}
	}
	if document.SchemaVersion != 1 {
		return domain.Bundle{}, &domain.ValidationError{Code: "SCHEMA_VERSION", Fields: []domain.FieldError{{Path: "schema_version", Code: "unsupported", Message: "仅支持 schema_version=1"}}}
	}
	if len(document.Configurations) == 0 {
		return domain.Bundle{}, &domain.ValidationError{Code: "EMPTY_BUNDLE", Fields: []domain.FieldError{{Path: "configurations", Code: "required", Message: "配置包至少包含一项配置"}}}
	}
	configs := make([]domain.Configuration, 0, len(document.Configurations))
	now := p.now().UTC()
	for index, input := range document.Configurations {
		configs = append(configs, domain.Configuration{
			ID:   newID("cfg", input.Environment+input.Type+input.Name),
			Name: strings.TrimSpace(input.Name), Type: strings.TrimSpace(input.Type),
			Environment: strings.TrimSpace(input.Environment), Values: input.Values,
			Status: domain.StatusDraft, Revision: 1, Version: 1,
			Tags: append([]string(nil), input.Tags...), ExpiresAt: input.ExpiresAt,
			ReplacesID: input.ReplacesID, CreatedBy: actor,
			CreatedAt: now.Add(time.Duration(index) * time.Nanosecond), UpdatedAt: now,
		})
	}
	sum := sha256.Sum256(payload)
	return domain.Bundle{
		ID: newID("bundle", hex.EncodeToString(sum[:8])), Filename: filename,
		ContentHash: hex.EncodeToString(sum[:]), State: domain.BundleUploaded,
		Configurations: configs, UploadedBy: actor, UploadedAt: now,
	}, nil
}

func locateJSONError(payload []byte, err error) (int, int) {
	offset := int64(1)
	if syntax, ok := err.(*json.SyntaxError); ok {
		offset = syntax.Offset
	}
	if typed, ok := err.(*json.UnmarshalTypeError); ok {
		offset = typed.Offset
	}
	if offset < 1 {
		offset = 1
	}
	line, column := 1, 1
	for index, value := range payload {
		if int64(index+1) >= offset {
			break
		}
		if value == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}
	return line, column
}

func newID(prefix, seed string) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%d", prefix, seed, time.Now().UnixNano())))
	return prefix + "_" + hex.EncodeToString(sum[:8])
}
