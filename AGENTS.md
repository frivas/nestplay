# Nestplay working agreement

Use Wayfinder for every increment. Start with the canonical GitHub maps at https://github.com/frivas/nestplay/issues/2 and https://github.com/frivas/nestplay/issues/1; claim one ticket before work. Its Notes must explicitly allow execution for a build milestone; otherwise remain in planning. Preserve the distinction between an implemented proof and user acceptance on a physical device.

This first milestone only establishes a tvOS-to-Go health check. Do not add YouTube, media playback, profiles or service dependencies without the next scoped ticket. Perch is optional reference material, not a required dependency.

The Go server and Swift health client share the response contract documented in README.md. Validate protocol changes in both. Keep external framework dependencies out of this milestone unless necessary. Keep signing team and user credentials out of the project. Use the checked-in XcodeGen spec when changing Xcode project settings.

Prefix shell commands with rtk as required by the workspace user instructions. Run `make check` locally before pushing. Do not start GitHub-hosted Actions from a push: this private account's included minutes are exhausted by macOS tvOS jobs. GitHub CI stays disabled and is workflow_dispatch only until the allowance resets. All commits must be signed (`git commit -S`). Every commit must belong to a claimed GitHub issue and use a complete message: what changed, why, how it was verified, and a `Fixes` / `Closes` / `Resolves` link to that issue. After a branch's CI is green and the work is merged, delete the remote and local branch so the repository stays clean. Do not publish or push this project without explicit user authorization.
