# Bug Reproduction

Run impact analysis with two health adapters that wait on a shared start barrier. Only the first adapter starts; the second never reaches the barrier and the analysis stalls.

The fixed configuration and explicit barrier make the scheduling failure deterministic.
