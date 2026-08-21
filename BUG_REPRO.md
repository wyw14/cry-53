# Bug Reproduction

Export an encrypted configuration once with masking and once with an authorized, unexpired sensitive disclosure reason. The plaintext export succeeds, but its audit event stores an empty reason.

The audit hash chain remains valid while the authorization rationale is lost.
