# Build and validate the first LAN health check
<!-- Canonical local Wayfinder ticket. -->
Type: task
Label: wayfinder:task
Status: resolved
Assignee: root
Parent: [Prove Apple TV can reach the Nestplay Go server](https://github.com/frivas/nestplay/issues/1)
Blocked by: none

## Question

Can a minimal native tvOS client contact a Go GET /health endpoint, validate the Nestplay response and recover when the server is unavailable? Build the explicitly requested PoC foundation, exercise the protocol and build the tvOS target, then record exactly which runtime/device checks are possible. Leave user/device acceptance open if not observed.

## Answer

Yes at simulator level. The completed prototype is in this repository and contains a configurable Go server, a shared Swift health client, and a native SwiftUI tvOS application. The contract is `GET /health` returning service identity, status and API version. The client exposes ready, checking, connected and actionable failure states, remembers the server address, validates the response, times out, and permits retry.

Evidence recorded on 11 September 2026:

- Go endpoint tests passed and the arm64 server binary built.
- Five Swift Testing cases passed, covering the valid endpoint, invalid addresses, wrong service/version, malformed or failing HTTP responses, and an unreachable server.
- An unsigned generic tvOS build succeeded against the tvOS 26.5 SDK.
- Two tvOS 26.5 simulator UI checks passed against the real Go process: connected and unreachable-with-retry.

[Validation record](../../VALIDATION.md) and [milestone handoff](../../HANDOFF.md) preserve exact commands, scope, limitations and continuation steps.

The user's physical Apple TV HD A1625 on tvOS 26.6 was not paired, signed or exercised. Simulator success does not close physical-device acceptance; that is the next ticket.
