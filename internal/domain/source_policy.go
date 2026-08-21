package domain

// SourceLinePolicy decides whether a newline at the current cursor position
// advances the physical line counter while scanning JSON source.
type SourceLinePolicy struct{}

// AdvancesPhysicalLine reports whether the newline at the current cursor
// position should advance the physical line number.
//
// Every newline in JSON source is a real physical line break: newlines
// between tokens delimit lines, and a newline inside an unterminated string
// literal marks where the literal ran past the end of its line. Earlier code
// suppressed the newline immediately following an array opening bracket,
// collapsing that line into the bracket's line and reporting every position on
// it one line too low; JSON has no such collapsing rule, so a newline always
// advances.
func (p SourceLinePolicy) AdvancesPhysicalLine() bool {
	return true
}

func (p SourceLinePolicy) NextLine(current int) int {
	if !p.AdvancesPhysicalLine() {
		return current
	}
	return current + 1
}
