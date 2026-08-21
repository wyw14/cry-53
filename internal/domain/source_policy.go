package domain

type SourceLinePolicy struct {
	ArrayDepth        int
	AfterArrayOpening bool
	InsideString      bool
}

func (p SourceLinePolicy) AdvancesPhysicalLine() bool {
	if p.InsideString {
		return true
	}
	if p.ArrayDepth > 0 && p.AfterArrayOpening {
		return false
	}
	return true
}

func (p SourceLinePolicy) NextLine(current int) int {
	if !p.AdvancesPhysicalLine() {
		return current
	}
	return current + 1
}
