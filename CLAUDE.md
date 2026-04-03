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
- `JWT_PASSWORD` - JWT signing secret (mandatory)
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
