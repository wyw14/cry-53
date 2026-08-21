# Bug Reproduction

Upload and validate a clean bundle as an actor who also has the approver role. The same actor can approve the bundle, bypassing separation of duties.

An independent approver succeeds as expected, and an unvalidated bundle is still rejected.
