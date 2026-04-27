# CHANGELOG_CLAUDE.md

This file summarizes significant architectural and behavioural changes made with AI assistance. Entries are grouped by change type, not by date or code location.

---

## OAuth, PKCE, And Login Flows

- **GitHub OAuth was split by client type**: Web and mobile app login now use separate handlers, routes, OAuth client configuration, session keys, and callback flows.
- **OAuth routes were simplified**: Login, callback, refresh, logout, and mobile app exchange endpoints are grouped under `/api/oauth`.
- **`golang.org/x/oauth2` was removed**: GitHub authorization URLs and token exchange are implemented explicitly while preserving standard OAuth2 authorization-code + PKCE semantics.
- **PKCE is enforced with S256**: Web login uses a server-generated GitHub PKCE verifier/challenge. Mobile login uses both a server-generated GitHub PKCE pair and an app-generated PKCE challenge for the app-code exchange.
- **OAuth state and PKCE values are stored in the session**: Callback handlers validate state with constant-time comparison and clear temporary OAuth session values after callback processing.
- **Mobile login uses a one-time app code**: The browser/OS callback returns only a short-lived app code. The app must redeem it with the original PKCE verifier before receiving local JWTs.
- **GitHub OAuth helpers were centralized**: Authorization URL building, GitHub token exchange, GitHub user fetching, profile creation, and login-result issuance were moved into shared auth utilities where appropriate.

---

## JWT And Session Security

- **JWTs use standard validation claims**: Issuer, audience, subject, issued-at, not-before, and expiry are included and validated.
- **JWT signing is fixed to HS512**: Callers cannot choose weaker or inconsistent signing algorithms.
- **Access-token type is enforced**: Protected APIs reject tokens unless the JWT explicitly has `tokenType=access`.
- **JWT and session identity must match**: Protected APIs require the session `profileID` and `githubID` to match the validated JWT identity.
- **Session storage was simplified**: The session stores primitive `profileID` and `githubID` values instead of gob-encoded structs.
- **Session cookies are signed and encrypted**: Cookie sessions use an authentication key and derived block key so session contents are tamper-protected and not readable by the client.
- **Session lifetime is bounded**: Session `MaxAge` is aligned with the web access-token TTL to avoid accepting stale profile identity longer than the access token.

---

## Refresh Tokens

- **Refresh tokens are opaque random secrets**: Local refresh tokens are no longer JWTs and are never stored in plaintext.
- **Refresh-token hashes use HMAC-SHA-256**: Stored hashes are peppered with a server secret so a database leak alone is not enough to verify guessed tokens offline.
- **Refresh-token records are persisted**: Refresh tokens are stored with profile, family, client type, creation, expiry, revocation, replacement, and last-use metadata.
- **Refresh-token rotation is enforced**: Each refresh call revokes the old token and issues a new one.
- **Refresh-token reuse triggers family revocation**: Reusing an already-revoked token revokes the full token family.
- **Refresh-token rotation is transactional**: Old-token revocation and new-token insertion happen atomically with majority write concern.
- **Web and mobile refresh flows are separated**: Web uses an HttpOnly refresh cookie; mobile uses an explicit refresh token in JSON.
- **Refresh-token cookie path was corrected**: The web refresh cookie is scoped to `/api/oauth/refresh`.
- **User-Agent binding was removed**: Refresh tokens are no longer checked against or stored with `User-Agent`, because it is not a reliable security boundary and caused brittle client behaviour.
- **Refresh/logout endpoints revoke stored tokens**: Logout revokes the refresh-token family for the presented token and clears session state.

---

## Database And Indexes

- **Refresh-token indexes were added**: Unique token hash, family lookup, and expiry TTL indexes are ensured.
- **App login code indexes were added**: App login codes have a unique code index and TTL expiry index.
- **Profile uniqueness was hardened**: GitHub profile identity is protected with a unique index.
- **Old refresh-token indexes are cleaned up**: Legacy refresh-token expiry index naming is handled during startup.
- **MongoDB startup uses explicit contexts**: Database connection validation and index setup use a startup context with timeout instead of `context.TODO`.
- **Test execution is isolated**: Tests run with `ENV=testing` and use the testing database name.

---

## API And Authorization Behaviour

- **Protected API authorization was tightened**: Requests must provide a valid bearer access token and a matching authenticated session.
- **Device value writes validate feature ownership and type**: `POST /api/devices/:id/values` now rejects feature states unless they match enabled controller features on the caller-owned device.
- **Audit logs were reduced**: `clientIP` was removed from audit logs.
- **Sensitive logs were cleaned up**: API tokens and other sensitive values are no longer logged in plaintext.
- **Request size limiting is configured**: The API enforces a maximum request body size.
- **Gin default request logging was avoided**: The router uses explicit middleware instead of Gin's default logger to avoid logging full request details.

---

## Bug Fixes And Reliability

- **Internal HTTP calls are bounded and status-aware**: Sensor/online helper calls now use a shared client timeout and treat non-2xx downstream responses as errors.
- **Internal service URL path segments are escaped**: Device UUIDs, feature UUIDs, and feature names are path-escaped before calls to sensor/online services.
- **Mobile refresh flow was fixed**: Mobile now uses `/api/oauth/app/refresh`, sends the raw refresh token in JSON, and receives rotated mobile tokens in JSON.
- **Web refresh cookie handling was fixed**: Cookie path and refresh endpoint now match.
- **OAuth callback session cleanup was made explicit**: Temporary OAuth state and PKCE values are cleared in callback handlers.
- **MongoDB transaction boundaries were improved**: External side effects are kept outside transactions where possible, and refresh-token rotation is atomic.
- **Context usage was corrected**: Per-request contexts are used for request-scoped database and network operations instead of stale stored contexts.
- **Startup error handling was made idiomatic**: Environment and database initialization now return errors through startup instead of mixing fatal logs and panics.
- **Timestamp handling was normalized**: Token and OAuth operations now capture `time.Now().UTC()` once per logical operation and derive related timestamps from it.
- **Common Go correctness issues were fixed**: Shadowed errors, unsafe type assertions, ignored errors, nil cursor cleanup, loop-variable pointer aliasing, and deferred resource accumulation were corrected.
- **Data shape issues were fixed**: New homes initialize rooms properly, internal online API fields are hidden, and FCM forwarding includes the required API token.

---

## Code Organization

- **Auth middleware was moved out of `/api`**: Authentication middleware now lives in the auth module.
- **OAuth common behaviour was separated**: Common refresh/logout code is separated from GitHub web and mobile login handlers.
- **Utilities were split by responsibility**: Cookie, JWT, PKCE, random string, and session helpers live in dedicated utility files.
- **GitHub OAuth files were renamed and split**: Web and mobile GitHub OAuth flows are in separate files with consistent naming.
- **Internal handler naming was simplified**: OAuth handler types and methods now use clearer names such as `GitHubWebHandler`, `GitHubAppHandler`, `OAuthHandler`, `RefreshMobileToken`, and `RotateAPIToken`.
- **Duplicate helper logic was removed**: Shared OAuth helpers were centralized; small session cleanup helpers were inlined where that made handlers clearer.
- **Duplicate response models were removed**: GitHub OAuth response structs are defined only in the models package.
- **Session helpers were made more idiomatic**: Session helpers accept the `sessions.Session` interface directly instead of a pointer to an interface.
- **ObjectID comparisons were simplified**: ObjectIDs are compared directly instead of converting both sides to hex strings.

---

## Testing, Tooling, And Infrastructure

- **Test target was hardened**: `make test` runs protobuf generation, vet, shadow checks, staticcheck, race-enabled tests, and coverage generation.
- **Go and MongoDB dependencies were updated**: The project was moved to newer Go and MongoDB driver versions.
- **Docker runtime was hardened**: The runtime image is minimal, runs without the Go toolchain, and uses an unprivileged user.
- **Configuration and logging were cleaned up**: Log folder configuration, logger flushing, explicit startup error propagation, unused imports, unused parameters, and obsolete Makefile targets were cleaned up.
- **Device and online features evolved**: Device assignment, device values, online status, protobuf naming, and feature naming were updated and consolidated.

---

## Known Remaining Security Work

- **Startup configuration validation should be stricter**: Secrets, OAuth client IDs/secrets, callback URLs, and cookie security mode should fail closed if missing, weak, or unsafe.
- **CORS should be environment-specific**: Credentialed CORS should use an explicit production allowlist and avoid broad localhost/internal defaults in production.
- **Web access-token delivery can be improved**: `/postlogin#token=...` avoids server logs but still exposes the access token to browser JavaScript; a BFF or secure-cookie design would reduce that exposure.
