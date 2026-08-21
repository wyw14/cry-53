# Bug Reproduction

Validate a `dev` warehouse configuration whose sensitive password field is numeric and whose boolean field is textual. The report includes the environment and boolean issues but omits the password shape issue.

The missing issue is deterministic and allows an invalid sensitive value into later processing.
