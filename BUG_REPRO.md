# Bug Reproduction

Upload a bundle with an idempotency key, replay the same payload, then reuse the key with different content of the same digest length. The third request incorrectly returns the first bundle instead of an idempotency conflict.

The distinct payload is silently ignored.
