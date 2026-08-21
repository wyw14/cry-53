# Bug Reproduction

Roll back configuration version 3 at revision 6 to historical version 1. Values and tags are restored, but the returned version remains 3 and the repository creates version 2 instead of a new version 4.

This breaks monotonic version history and traceability.
