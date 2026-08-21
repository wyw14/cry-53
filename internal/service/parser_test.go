package service

import (
	"strings"
	"testing"

	"github.com/wyw14/cry-053/internal/domain"
)

func TestBundleParserReportsPreciseJSONLine(t *testing.T) {
	parser := NewBundleParser(4096)
	_, err := parser.Parse("broken.json", strings.NewReader("{\n  \"schema_version\": 1,\n  \"configurations\": [\n    {\"name\": \"primary\",}\n  ]\n}"), "editor")
	validation, ok := err.(*domain.ValidationError)
	if !ok {
		t.Fatalf("expected validation error, got %T %v", err, err)
	}
	if validation.Code != "INVALID_JSON" {
		t.Fatalf("code=%s", validation.Code)
	}
	if len(validation.Fields) != 1 || validation.Fields[0].Line != 4 {
		t.Fatalf("fields=%+v", validation.Fields)
	}
}
