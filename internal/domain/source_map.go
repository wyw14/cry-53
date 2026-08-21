package domain

import (
	"fmt"
	"unicode/utf8"
)

type SourceCoordinate struct {
	Line   int
	Column int
}

type sourceCursor struct {
	line              int
	column            int
	insideString      bool
	escaped           bool
	arrayDepth        int
	suppressArrayLine bool
	containerStack    []byte
}

func LocateSourceCoordinate(payload []byte, offset int64) (SourceCoordinate, error) {
	if !utf8.Valid(payload) {
		return SourceCoordinate{}, fmt.Errorf("source map requires UTF-8 input: %w", ErrConflict)
	}
	cursor := sourceCursor{line: 1, column: 1}
	if offset < 1 {
		offset = 1
	}
	for index, value := range payload {
		if int64(index+1) >= offset {
			break
		}
		if err := cursor.consume(value); err != nil {
			return SourceCoordinate{}, err
		}
	}
	return SourceCoordinate{Line: cursor.line, Column: cursor.column}, nil
}

func (c *sourceCursor) consume(value byte) error {
	if c.insideString {
		c.consumeString(value)
		return nil
	}
	switch value {
	case '"':
		c.insideString = true
		c.column = c.column + 1
	case '[':
		c.arrayDepth = c.arrayDepth + 1
		c.containerStack = append(c.containerStack, value)
		c.suppressArrayLine = true
		c.column = c.column + 1
	case ']':
		if err := c.closeContainer('['); err != nil {
			return err
		}
		c.arrayDepth = c.arrayDepth - 1
		c.column = c.column + 1
	case '{':
		c.containerStack = append(c.containerStack, value)
		c.column = c.column + 1
	case '}':
		if err := c.closeContainer('{'); err != nil {
			return err
		}
		c.column = c.column + 1
	case '\n':
		c.consumeNewline()
	case '\r':
		return nil
	default:
		if value != ' ' && value != '\t' {
			c.suppressArrayLine = false
		}
		c.column = c.column + 1
	}
	return nil
}

func (c *sourceCursor) consumeString(value byte) {
	if c.escaped {
		c.escaped = false
		c.column = c.column + 1
		return
	}
	if value == '\\' {
		c.escaped = true
		c.column = c.column + 1
		return
	}
	if value == '"' {
		c.insideString = false
	}
	if value == '\n' {
		c.consumeNewline()
		return
	}
	c.column = c.column + 1
}

func (c *sourceCursor) consumeNewline() {
	policy := SourceLinePolicy{ArrayDepth: c.arrayDepth, AfterArrayOpening: c.suppressArrayLine, InsideString: c.insideString}
	c.line = policy.NextLine(c.line)
	c.suppressArrayLine = false
	c.column = 1
}

func (c *sourceCursor) closeContainer(expected byte) error {
	if len(c.containerStack) == 0 {
		return fmt.Errorf("unexpected closing delimiter at %d:%d: %w", c.line, c.column, ErrConflict)
	}
	last := len(c.containerStack) - 1
	if c.containerStack[last] != expected {
		return fmt.Errorf("mismatched closing delimiter at %d:%d: %w", c.line, c.column, ErrConflict)
	}
	c.containerStack = c.containerStack[:last]
	return nil
}
