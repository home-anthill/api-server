# CLAUDE.md

## Project Overview

**home-anthill/api-server** is the central REST API backend for the home-anthill IoT/home automation platform. It manages homes, rooms, devices (sensors/controllers), user profiles, and real-time device status. Built with Go + Gin, backed by MongoDB, with gRPC for device communication.

## Tech Stack

- **Go 1.26** with Gin web framework
- **MongoDB** (driver v2) with replica set required for tests
- **gRPC/Protobuf** for device communication
- **GitHub OAuth2** for authentication (web + mobile variants)
- **JWT** for API session tokens
- **Zap** structured logging with lumberjack rotation
- **Ginkgo/Gomega** BDD-style integration tests

## Build & Run

```bash
make deps       # Install dev tools (staticcheck, air, shadow, go-cover-treemap)
make build      # Compile protobuf + vet + lint + build binary to ./build/api-server
make run        # Run with Air hot-reload (includes proto + vet + lint)
make test       # Run all tests with coverage (ENV=testing, needs MongoDB with replica set)
make proto      # Regenerate gRPC code from .proto files
make vet        # go vet + shadow analysis
make lint       # staticcheck
```

## Project Structure

```
main.go                    # Entry point
initialization/            # Server startup: env loading, router setup, logger config, logger setup
api/                       # HTTP handlers (one file per domain: homes, devices, profiles, etc.)
api/grpc/device/           # gRPC/protobuf definitions for device communication
models/                    # MongoDB document models
db/                        # MongoDB initialization and collection access
utils/                     # Helpers: JWT, HTTP, validation, session, gRPC, slices
customerrors/              # Custom error wrappers (RegisterError, GrpcSendError)
integration_tests/         # Ginkgo/Gomega integration tests (one per API domain)
testuutils/                # Shared test helpers
public/                    # SPA static assets (served in non-prod environments)
```

## Code Conventions

- **Handler pattern**: Each API domain is a struct with injected dependencies (logger, mongo client, validator) and a `New<Handler>()` constructor returning a pointer. Config values (e.g., JWT key, gRPC target URLs) are resolved at construction time, not per-request
- **Validation**: `go-playground/validator` with struct tags (e.g., `validate:"required,min=1,max=50"`)
- **Error handling**: Custom `ErrorWrapper` in `customerrors/` preserves HTTP status + message + original error. All errors must be checked — never use `_` to discard errors
- **Type assertions**: Always use the two-value `val, ok := ...` form; handle the `!ok` case
- **Session/Auth**: Handlers always re-fetch profile from DB (never trust stale session data)
- **Logging**: Zap `SugaredLogger` throughout; file logging disabled in test mode
- **Indentation**: Tabs for Go files (see `.editorconfig`)
- **JWT security**: Access tokens are signed with `JWT_PASSWORD`; refresh tokens with `JWT_REFRESH_PASSWORD`. All JWTs carry `iss`, `aud`, and `sub` claims (`home-anthill-api`) validated on parse. Bearer prefix is validated with `strings.HasPrefix` before slicing. Refresh token cookies use `SameSite=Lax` (required for OAuth2 cross-site redirect — see note below). Validated JWT claims are stored in Gin context under key `"jwt_claims"`.
- **Refresh token cookie — `Path` and Chrome DevTools visibility**: The `refresh_token` cookie is set with `Path=/api/token/refresh` and `HttpOnly=true`. This means:
  - The browser **only sends** the cookie to `POST /api/token/refresh` — it is never sent to `/api/homes`, `/api/devices`, or any other endpoint.
  - Chrome DevTools **Application → Cookies** only displays cookies whose `Path` is a prefix of the currently open browser URL. When on `/main` or `/postlogin`, the `refresh_token` cookie is **not visible** in that panel even though it is correctly stored. To see it, navigate the browser to `http://localhost:8082/api/token/refresh` — it will then appear.
  - This is a DevTools display-filtering behaviour, **not a bug**. The cookie is present and works correctly.
- **Refresh token cookie — `SameSite=Lax` vs `Strict`**: `SameSite=Strict` causes Chrome to silently drop the cookie during the GitHub OAuth2 callback redirect (a cross-site top-level navigation from `github.com`). `SameSite=Lax` is the correct value: it still blocks the cookie from being sent in cross-site sub-resource requests (CSRF protection), while allowing it to be stored during top-level cross-site navigations such as the OAuth2 redirect.
- **OAuth web login redirect**: `LoginCallback` redirects to `/postlogin` with the access token as a URL fragment (`#token=…`) so the token is never sent to the server in subsequent requests and does not appear in access logs. `LoginMobileAppCallback` uses a query parameter as required by the mobile deep-link scheme.
- **`LoginMobileAppCallback` — session cookie source**: `OauthAuth` middleware calls `session.Save()` which writes the updated `mysession` cookie (encoding the full profile) to the **HTTP response** headers, not back into the request. `LoginMobileAppCallback` therefore reads the response `Set-Cookie` headers first (looking for `mysession=`); only when `session.Save()` was not called (profile already in session — the repeat-login path) does it fall back to `c.Request.Cookie("mysession")`. Reading from the request directly would deliver a stale (empty-profile) cookie to the Android app on first install, causing all subsequent API calls to fail with "cannot find profile in session".

## Testing

- Integration tests in `integration_tests/` using Ginkgo/Gomega
- Tests require a running MongoDB instance with replica set
- Test database: `api-server-test` (separate from dev/prod)
- Run with: `make test` (sets `ENV=testing`)
- Tests run with `-race` (race detector) and `-count=1` (no caching)
- Coverage reports generated in `./coverage/`

## Environment

Copy `.env_template` to `.env` and fill in GitHub OAuth credentials. Key variables:
- `MONGODB_URL` - MongoDB connection string
- `JWT_PASSWORD` - JWT signing secret for access tokens (mandatory)
- `JWT_REFRESH_PASSWORD` - JWT signing secret for refresh tokens (mandatory, distinct from `JWT_PASSWORD`)
- `COOKIE_SECRET` - Session cookie signing secret
- `OAUTH2_CLIENTID/SECRETID` - GitHub OAuth for web
- `OAUTH2_APP_CLIENTID/SECRETID` - GitHub OAuth for mobile app
- `OAUTH2_CALLBACK` / `OAUTH2_APP_CALLBACK` - OAuth redirect URLs
- `GRPC_URL` - gRPC target for device service
- `GRPC_TLS` / `CERT_FOLDER_PATH` - gRPC TLS toggle and certificate path
- `HTTP_SERVER` / `HTTP_PORT` / `HTTP_CORS` - Server bind config
- `HTTP_SENSOR_*` / `HTTP_ONLINE_*` - External service endpoints
- `SINGLE_USER_LOGIN_EMAIL` - Restrict login to a single GitHub email (optional)
- `INTERNAL_CLUSTER_PATH` - Kubernetes internal service address
- `LOG_FOLDER` - Log file directory

## CI/CD

GitHub Actions (`.github/workflows/docker-image.yml`):
- **Test job**: Go 1.26, MongoDB 8.0 replica set, runs `deps` + `vet` + `lint` + `test`
- **Build job**: Multi-stage Docker build (hardened `alpine-base` runtime image, runs as UID 65534), pushes to `ks89/api-server` on Docker Hub
- Triggers on push to `master`, `develop`, `ft*` branches and semver tags, PRs to `master`/`develop`, and manual `workflow_dispatch`
