package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

// syntaxErrorLine decodes src and maps Go's json error offset to a source
// coordinate via LocateSourceCoordinate, mirroring the parser path.
func syntaxErrorLine(t *testing.T, src string) (line, column int) {
	t.Helper()
	decoder := json.NewDecoder(strings.NewReader(src))
	var document map[string]any
	err := decoder.Decode(&document)
	if err == nil {
		t.Fatalf("expected a JSON syntax error, got nil")
	}
	syntax, ok := err.(*json.SyntaxError)
	if !ok {
		t.Fatalf("expected *json.SyntaxError, got %T: %v", err, err)
	}
	coordinate, mapErr := LocateSourceCoordinate([]byte(src), syntax.Offset)
	if mapErr != nil {
		t.Fatalf("LocateSourceCoordinate failed: %v", mapErr)
	}
	return coordinate.Line, coordinate.Column
}

// A syntax error on line 4 (a trailing comma inside an array element) must be
// reported as line 4. Previously the cursor suppressed the newline that
// followed the array opening bracket, collapsing that line into line 3.
func TestLocateSourceCoordinateSyntaxErrorInArrayElement(t *testing.T) {
	src := "{\n  \"schema_version\": 1,\n  \"configurations\": [\n    {\"name\": \"primary\",}\n  ]\n}"
	line, _ := syntaxErrorLine(t, src)
	if line != 4 {
		t.Fatalf("line: want 4, got %d", line)
	}
}

// The reported Windows 11 scenario: a bundle whose line 4 breaks a string
// quote. The syntax error must land on line 4, not the previous line, under
// both LF and CRLF line endings.
func TestLocateSourceCoordinateBrokenQuoteOnLine4(t *testing.T) {
	cases := map[string]string{
		"LF":   "{\n  \"schema_version\": 1,\n  \"configurations\": [\n    {\"name\": \"primary}\n  ]\n}",
		"CRLF": "{\r\n  \"schema_version\": 1,\r\n  \"configurations\": [\r\n    {\"name\": \"primary}\r\n  ]\r\n}",
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			line, column := syntaxErrorLine(t, src)
			if line != 4 {
				t.Fatalf("%s: line want 4, got %d", name, line)
			}
			// The column points at the byte past the broken string literal and
			// must be reported (it is independent of the line fix).
			if column < 1 {
				t.Fatalf("%s: column must be >= 1, got %d", name, column)
			}
		})
	}
}

// Column reporting must not regress: a missing colon on a known line reports a
// column that points at the offending token, and the line matches the physical
// line of the error.
func TestLocateSourceCoordinateColumnUnaffected(t *testing.T) {
	// line 2, "a" missing its colon -> json reports the error at '1' (column 7
	// on that line, since the offset lands on the '1' after "  \"a\" ").
	src := "{\n  \"a\" 1\n}"
	line, column := syntaxErrorLine(t, src)
	if line != 2 {
		t.Fatalf("line: want 2, got %d", line)
	}
	if column != 7 {
		t.Fatalf("column: want 7, got %d", column)
	}
}

// An unterminated string literal spanning a newline reports the error on the
// line where the literal ran off the end.
func TestLocateSourceCoordinateUnterminatedString(t *testing.T) {
	src := "{\n  \"a\": 1,\n  \"b\": \"unterminated\n}"
	line, _ := syntaxErrorLine(t, src)
	if line != 3 {
		t.Fatalf("line: want 3, got %d", line)
	}
}
