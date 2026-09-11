# Accept the health check on the Apple TV HD
<!-- Canonical local Wayfinder ticket. -->
Type: task
Label: wayfinder:task
Status: open
Assignee: none
Parent: [Prove Apple TV can reach the Nestplay Go server](https://github.com/frivas/nestplay/issues/1)
Blocked by: [Build and validate the first LAN health check](https://github.com/frivas/nestplay/issues/11)

## Question

Can the signed Nestplay PoC installed on the user's Apple TV HD A1625 running tvOS 26.6 reach the Go server through the real LAN, display a useful failure after the server stops, and recover after restart? Pair the device in Xcode, select the user's signing team, run the documented five-step acceptance check, and record the result. Decide any manual-address UX changes only after observing this proof.

## Comments

At the 11 September 2026 handoff, Xcode listed only generic tvOS destinations and installed simulators. The physical A1625 was not paired or visible, so this ticket correctly remains open.
