# Build and validate the first LAN health check
<!-- Exported Wayfinder snapshot. -->
Type: task
Label: wayfinder:task
Status: resolved
Assignee: root
Parent: ../map.md
Blocked by: none

## Question

Can a minimal native tvOS client contact a Go GET /health endpoint, validate the Nestplay response and recover when the server is unavailable?

## Answer

Yes at simulator level. The completed prototype is in this repository. Go endpoint tests and the arm64 build passed. Five Swift test functions passed. An unsigned generic tvOS build succeeded. Two tvOS 26.5 simulator UI checks passed against the real Go process: connected and unreachable-with-retry.

[Validation record](../../../../VALIDATION.md) and [milestone handoff](../../../../HANDOFF.md) preserve the evidence and continuation steps. The physical Apple TV HD A1625 on tvOS 26.6 has not been exercised.
