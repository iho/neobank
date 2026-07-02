# neobank_mobile

Flutter client for the neobank gateway BFF. See [TODO.md](TODO.md) for the phased backlog
and [../services/gateway/api/openapi.yaml](../services/gateway/api/openapi.yaml) for the
API contract this app is built against.

## Architecture decisions

- **State management: Riverpod** (`flutter_riverpod`, plain `AsyncNotifier`/`Notifier`
  classes — no code generation). Chosen over Bloc/Provider for compile-time-safe DI,
  testability without `BuildContext`, and first-class async matching the gateway's
  request/response shape.
- **Navigation: go_router.** Declarative routes, typed path params, and a single place
  (`core/router/app_router.dart`) to gate routes on auth/KYC state via `redirect`.
- **Models: hand-written plain classes for now.** Each `features/*/domain/*_models.dart`
  has a small immutable class with a `fromJson` factory, matching the OpenAPI schema field
  for field. A generated client already exists (`lib/core/api/generated/`, via
  `swagger_parser` — see below) but the six feature repositories haven't been migrated
  onto it yet; that's tracked in `TODO.md` as a deliberate, separate follow-up. `freezed`/
  `json_serializable` stay as dev dependencies for that migration and for the generated
  client's own models.
- **Folder structure: feature-first**, not layer-first:

  ```
  lib/
    core/            # cross-cutting: config, network, router, storage, theme, error
    features/
      auth/
        data/        # repositories, DTOs, API calls
        domain/      # entities, use cases
        presentation/# screens, widgets, Riverpod notifiers
      onboarding_kyc/
      wallet/
      transfers/
      cards/
      notifications/
  ```

  Each feature owns its full stack; `core/` holds only what's genuinely shared (the Dio
  client, token storage, the router shell, theme, typed API errors).

## Networking

- `core/network/dio_client.dart` exposes two Riverpod providers:
  - `authDioProvider` — unauthenticated, used only for `/v1/auth/register|login|refresh`.
  - `dioProvider` — authenticated; attaches the bearer token and transparently refreshes
    it on 401 via `AuthInterceptor` (single-flight refresh, retries the original request once).
- `IdempotencyInterceptor` stamps every mutating request with a fresh `Idempotency-Key`
  unless the caller already set one (retries of the same logical operation, e.g. a transfer
  after a timeout, must reuse the original key — set it explicitly on `RequestOptions`).
- `CorrelationIdInterceptor` stamps `X-Correlation-Id` on every request for support/tracing.
- `core/error/api_exception.dart` is the typed failure surface — repositories should catch
  `DioException` and rethrow `ApiException`; UI code never handles raw Dio errors.

## Generated API client

`lib/core/api/generated/` is a [`swagger_parser`](https://pub.dev/packages/swagger_parser)
Retrofit + json_serializable client generated from
[`services/gateway/api/openapi.yaml`](../services/gateway/api/openapi.yaml) — config in
`swagger_parser.yaml`. It's checked in (same convention as the backend's `oapi-codegen`/
`sqlc` output) and regenerated with:

```bash
make mobile-generate   # from the repo root; not part of `make generate`/`make build`
```

`gatewayApiClientProvider` (`core/network/gateway_api_client_provider.dart`) wraps it
around the same `dioProvider` the hand-written repositories use, so it already gets the
auth/idempotency/correlation-id interceptors. It isn't consumed by any feature yet.

## Flavors

Configured via `--dart-define`, read in `core/config/app_config.dart`:

```bash
flutter run --dart-define=FLAVOR=dev --dart-define=API_BASE_URL=http://localhost:8080
```

Default (no defines) points at `http://localhost:8080` for local backend development.

## Getting started

```bash
flutter pub get
flutter analyze
flutter test
flutter run --dart-define=API_BASE_URL=http://localhost:8080
```
