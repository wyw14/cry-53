# Bug Reproduction

Start two publishers together with the same approved bundle, selection, and idempotency key. One request succeeds while the other returns a state-transition error instead of replaying the completed publication.

The explicit start gate makes the failure deterministic; only one publication should be stored.
