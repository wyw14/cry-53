# Bug Reproduction

Import a multiline JSON bundle whose string is unterminated on line 4. The parser reports the syntax issue on line 3, so the UI navigates to the preceding line.

The failure is deterministic for the same payload. Valid JSON remains parseable.
