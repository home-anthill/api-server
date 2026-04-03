# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed
- Shadow variable declarations in `api/assign_device.go` and `api/homes.go` (`if err :=` → `if err =`)
  - `assign_device.go:59` — `ShouldBindJSON` call
  - `assign_device.go:64` — `validate.Struct` call
  - `homes.go:218` — `ShouldBindJSON` call

---

## [3.0.0] - 2026-03-31

### Added
- Configurable log folder via `LOG_FOLDER` environment variable
- Hardened Docker runtime image (`dhi.io/alpine-base:3.23`) — no shell or Go toolchain in final image
- Container runs as unprivileged user `nobody` (UID 65534); `--chown=65534:65534` on all `COPY --from=builder` instructions
- Writable `/logs` directory pre-created in builder stage and copied with correct ownership
- `-race` flag and `-count=1` flag added to `go test` in Makefile
- `test` target now depends on `proto vet lint`

### Changed
- Upgraded MongoDB driver to v2
- Upgraded to Go 1.26
- All API handler structs no longer store `context.Context` as a field — per-request context (`c.Request.Context()`) is used instead across all 10 handlers
- JWT secret read once at init and injected into `Auth` struct (was `os.Getenv` per-request)
- `logger.Sync()` moved to `main.go` for correct process-exit flush
- External HTTP call in `DeleteDevice` moved outside the MongoDB transaction (prevents broken idempotency on retry and cross-boundary side effects)
- `defer conn.Close()` / `defer cancel()` in `devices_values.go` extracted into per-call `getControllerValue` helper to prevent resource accumulation inside a for-loop
- `GetOnlineFeature` in `utils/features.go` fixed to return pointer to slice element via index (`&features[i]`), not loop variable
- Renamed `poweroutage` feature to `online` throughout
- New device values implementation supporting thermostat and mixed sensor/controller devices
- Removed unused `ctx context.Context` parameter from `RegisterRoutes`, `BuildServer`, and `Start`
- Removed dead `fmt` target from Makefile
- Cleaned up unused `"context"` imports from `server.go`, `start.go`, `profiles.go`, `fcm_token.go`, `login_github.go`, and `keepalive.go`
- Updated all 10 integration test files for new `Start()` signature

### Fixed
- API token no longer logged in plaintext in `devices_values.go`
- Unsafe type assertions in `api/auth.go` (lines 36, 57) and `utils/validator.go` guarded with comma-ok checks
- Silently ignored errors from `io.ReadAll` in `utils/http.go`, `api/online.go`, `api/devices.go`
- Silently ignored `os.Getwd` error in `initialization/environment.go` — now propagated
- Silently ignored `cur.Decode` error in `api/devices.go` — now logged and skipped
- `defer cur.Close()` in `api/homes.go` moved after error check (was before — `cur` could be nil); removed `//lint:ignore SA5001` suppression
- `gin.Context` passed to MongoDB calls in `login_github.go` replaced with `c.Request.Context()`
- CI pipeline fixes

---

## [2.0.0] - 2024-xx-xx

### Added
- Support for multiple OAuth2 apps (web and Android)
- Android/mobile login via separate GitHub OAuth2 client
- Store FCM token in `fcm_token` POST API
- Online service integration: DELETE called when removing a device
- `current_time` field added to online JSON response

### Changed
- Removed register API
- Renamed proto packages; updated Makefile proto command order

### Fixed
- Create new home with empty rooms array to prevent downstream errors
- Do not expose `apiToken` and `uuid` via online API
- FCM service passes `apiToken` when calling online service

---

## [1.3.2] and earlier

See git log for previous history.
