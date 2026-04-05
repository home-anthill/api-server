# CHANGELOG_CLAUDE.md

This file tracks significant architectural and behavioural changes made with AI assistance, grouped by category.

---

## Security

- **Separate JWT signing keys**: Access tokens are signed with `JWT_PASSWORD`; refresh tokens with `JWT_REFRESH_PASSWORD`. Both are mandatory env vars validated at startup. The `Auth` struct carries a dedicated `jwtRefreshKey` field.
- **JWT standard claims**: All issued tokens carry `iss`, `aud` (`home-anthill-api`), and `sub` (GitHub ID) claims, validated on every parse with `jwt.WithIssuer` / `jwt.WithAudience` to prevent cross-deployment token replay.
- **Validated JWT claims stored in Gin context**: After a successful middleware check, `*utils.JWTClaims` are stored under key `"jwt_claims"` so downstream handlers have access without re-parsing.
- **Bearer prefix validated before slicing**: `JWTMiddleware` previously sliced the `Authorization` header at a fixed offset. It now uses `strings.HasPrefix` + `strings.TrimPrefix` to guard against malformed headers.
- **Refresh token cookie `SameSite=Lax`**: `SameSite=Strict` was initially used but caused Chrome to silently discard the cookie during the GitHub OAuth2 callback (a cross-site top-level navigation from `github.com`). `SameSite=Lax` still blocks cross-site sub-resource/POST requests (CSRF protection) while allowing storage during top-level cross-site navigations.
- **Refresh token cookie scoped to `Path=/api/token/refresh`**: The browser only sends the `refresh_token` cookie to that specific endpoint. Note: Chrome DevTools Application → Cookies filters by the current browser URL, so the cookie appears invisible on `/main` or `/postlogin` — this is a DevTools display behaviour, not a storage failure.
- **OAuth web login token delivered via URL fragment**: `LoginCallback` redirects to `/postlogin#token=…` instead of a query parameter. Fragments are never forwarded to the server, so the access token no longer appears in server or proxy access logs. Mobile deep-link callbacks retain a query parameter as required by the Android scheme.

---

## Bug Fixes

- **`LoginMobileAppCallback` stale session on first install**: On first install, `OauthAuth` middleware calls `session.Save()` which writes the updated session to the HTTP *response* headers — not back into the request. Reading the request cookie at that point returned an empty-profile value, causing all subsequent API calls to fail with 401. The fix reads `Set-Cookie` response headers first and falls back to the request cookie only when `session.Save()` was not called (repeat-login path).
- **Shadow variable declarations**: In `assign_device.go` and `homes.go`, `if err :=` shadowed the outer `err` variable inside `ShouldBindJSON` and `validate.Struct` calls. Changed to `if err =` to propagate errors correctly.
- **Unsafe type assertions**: Type assertions in `api/auth.go` and `utils/validator.go` are now guarded with comma-ok checks.
- **Silently ignored errors**: `io.ReadAll` errors in `utils/http.go`, `api/online.go`, and `api/devices.go`; `os.Getwd` error in `initialization/environment.go`; and `cur.Decode` errors in `api/devices.go` — all now propagated or logged.
- **`defer cur.Close()` before nil check**: In `api/homes.go`, `defer cur.Close()` was placed before the cursor error check, risking a nil dereference. Moved after the error check and removed the associated `//lint:ignore SA5001` suppression.
- **`gin.Context` passed to MongoDB**: In `login_github.go`, `gin.Context` was passed directly to MongoDB calls. Replaced with `c.Request.Context()` as required by the driver.
- **API token logged in plaintext**: The raw API token was logged in `devices_values.go`. Removed from log output.
- **`GetOnlineFeature` returning loop variable pointer**: `utils/features.go` returned `&feature` (loop variable address) instead of `&features[i]` (slice element address), causing all returned pointers to alias the same memory. Fixed to use index-based addressing.
- **`defer` resource accumulation in device value loop**: `defer conn.Close()` and `defer cancel()` inside a for-loop in `devices_values.go` accumulated deferred calls until the function returned. Extracted into a `getControllerValue` helper so resources are released on each iteration.
- **External HTTP call inside MongoDB transaction**: In `DeleteDevice`, an external HTTP call was made inside the MongoDB transaction, breaking idempotency on retry and mixing side effects with transactional operations. Moved outside the transaction.
- **Stale `context.Context` field in handlers**: All 10 API handler structs stored a `context.Context` field set at construction time. Per-request context (`c.Request.Context()`) is now used instead, preventing context leaks and cancellation mismatches.
- **New home created without rooms array**: Creating a new home now initialises `rooms` as an empty array, preventing downstream errors when the field is absent.
- **`apiToken` and `uuid` exposed via online API**: These internal fields are no longer returned in the online API response.
- **FCM service missing `apiToken`**: The FCM service now forwards `apiToken` when calling the online service.

---

## Idiomatic & Code Quality

- **JWT secret read once at init**: `JWT_PASSWORD` is now read and stored in the `Auth` struct at construction time rather than calling `os.Getenv` on every request.
- **`logger.Sync()` placement**: Moved to `main.go` so the logger is flushed on process exit rather than being deferred inside a deeper call.
- **Unused `ctx` parameters removed**: `RegisterRoutes`, `BuildServer`, and `Start` no longer accept an unused `context.Context` parameter. All 10 integration test files updated accordingly.
- **Unused imports cleaned up**: Removed unused `"context"` imports from `server.go`, `start.go`, `profiles.go`, `fcm_token.go`, `login_github.go`, and `keepalive.go`.
- **Dead Makefile target removed**: Removed the unused `fmt` target from the Makefile.
- **`-race` and `-count=1` added to `go test`**: The `test` Makefile target now runs with the race detector enabled and caching disabled.
- **`test` target depends on `proto vet lint`**: Ensures generated code and static checks are up to date before running tests.

---

## Infrastructure & Dependencies

- **Go 1.26 upgrade**.
- **MongoDB driver upgraded to v2**.
- **Hardened Docker runtime image**: Final image is based on `dhi.io/alpine-base:3.23` with no shell or Go toolchain. The container runs as unprivileged user `nobody` (UID 65534) with correct ownership on all copied files and a pre-created writable `/logs` directory.
- **Configurable log folder**: Log output directory is now controlled via the `LOG_FOLDER` environment variable.

---

## Features

- **Refresh token rotation**: `POST /api/token/refresh` issues a new `refresh_token` cookie on each call, replacing the previous one.
- **Multiple OAuth2 app support**: Separate GitHub OAuth2 clients for web and Android/mobile login flows.
- **FCM token storage**: New `POST /api/fcm_token` endpoint stores a Firebase Cloud Messaging token for a profile.
- **Online service integration**: Calling DELETE on a device now notifies the online service; response includes a `current_time` field.
- **New device values implementation**: Supports thermostat and mixed sensor/controller devices.
- **`poweroutage` renamed to `online`**: Feature name updated throughout the codebase and API.
- **Proto packages renamed**: Protobuf package names updated; Makefile proto command order corrected.
