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

// The syntax error column is independent of the line fix and must be reported
// in the message. The trailing comma sits inside an array element on line 4.
func TestBundleParserReportsSyntaxColumnUnaffected(t *testing.T) {
	parser := NewBundleParser(4096)
	_, err := parser.Parse("broken.json", strings.NewReader("{\n  \"schema_version\": 1,\n  \"configurations\": [\n    {\"name\": \"primary\",}\n  ]\n}"), "editor")
	validation, ok := err.(*domain.ValidationError)
	if !ok {
		t.Fatalf("expected validation error, got %T %v", err, err)
	}
	if !strings.Contains(validation.Fields[0].Message, "列 24") {
		t.Fatalf("column not reported in message: %q", validation.Fields[0].Message)
	}
}

// Repeatedly importing the same broken content must produce the same line,
// confirming the error is deterministic rather than order-dependent.
func TestBundleParserReportsDeterministicLineOnReimport(t *testing.T) {
	parser := NewBundleParser(4096)
	payload := "{\n  \"schema_version\": 1,\n  \"configurations\": [\n    {\"name\": \"primary\",}\n  ]\n}"
	for i := 0; i < 3; i++ {
		_, err := parser.Parse("broken.json", strings.NewReader(payload), "editor")
		validation, ok := err.(*domain.ValidationError)
		if !ok {
			t.Fatalf("pass %d: expected validation error, got %T %v", i, err, err)
		}
		if validation.Fields[0].Line != 4 {
			t.Fatalf("pass %d: line want 4, got %d", i, validation.Fields[0].Line)
		}
	}
}
