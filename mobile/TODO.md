# Flutter Mobile Client — TODO

Companion to [../todo-devops-mobile.md](../todo-devops-mobile.md) (DevOps backlog) and
[../todo.md](../todo.md) (compliance backlog). Scope here: the mobile client the gateway
BFF was designed for.

Own toolchain, own dependency graph — kept out of `go.work`.

**Contract source of truth:** `services/gateway/api/openapi.yaml` (JWT auth + refresh,
`Idempotency-Key` on mutations, `X-Correlation-Id`).

## Phase 1: Foundation

- [x] Scaffold Flutter app (`mobile/`).
- [x] Flavors dev/staging/prod (`--dart-define` for `API_BASE_URL`/`FLAVOR`) —
      `lib/core/config/app_config.dart`.
- [x] Lint rules (`flutter_lints`, default from `flutter create`) + `riverpod_lint`/`custom_lint`
      wired in `analysis_options.yaml`.
- [x] Architecture: feature-first folders + Riverpod (state) + go_router (navigation).
      Decision recorded in `mobile/README.md`. All six feature folders
      (`data`/`domain`/`presentation`) are now implemented, not just scaffolded — see
      Phase 2 below. Models are hand-written plain classes for now (freezed/
      json_serializable deps stay declared but unused until the generated OpenAPI
      client replaces these interim DTOs — see the item just below).
- [x] Dio interceptors:
      - JWT bearer injection + transparent refresh on 401 (single-flight refresh,
        logout on refresh failure) — `lib/core/network/auth_interceptor.dart`,
        tokens in `flutter_secure_storage` (`lib/core/storage/token_storage.dart`).
      - `Idempotency-Key: uuid` on every mutating request —
        `lib/core/network/idempotency_interceptor.dart`.
      - `X-Correlation-Id` generation — `lib/core/network/correlation_id_interceptor.dart`.
- [x] Error model — `lib/core/error/api_exception.dart` (typed failures; every repository
      catches `DioException` and rethrows `ApiException` via `core/network/error_mapper.dart`).
- [x] API client generated from `services/gateway/api/openapi.yaml` via `swagger_parser`
      (Retrofit + json_serializable) — `mobile/swagger_parser.yaml` config,
      `lib/core/api/generated/` output (47 models, 1 client, 41 requests), regenerate with
      `make mobile-generate` from the repo root (not chained into the Go `generate`/`build`
      targets — Go-only CI runners don't have Flutter/Dart). `gatewayApiClientProvider`
      (`lib/core/network/gateway_api_client_provider.dart`) wraps it around the existing
      `dioProvider`, so it inherits the auth/idempotency/correlation-id interceptors for
      free. **Not yet consumed** — migrating the six `features/*/data/*_repository.dart`
      onto it is a separate, mechanical follow-up best done one feature at a time against
      a running backend (each repository's hand-written models would be deleted in favor
      of the generated ones). Getting here required dropping `riverpod_lint`/`custom_lint`
      (pure dev tooling, unused at runtime) and bumping `freezed`/`freezed_annotation` to
      3.x and removing `riverpod_annotation`/`riverpod_generator` (never used) — all to
      resolve an `analyzer` version conflict with `retrofit_generator`; no runtime `flutter_riverpod`
      version change was needed.
- [ ] Global retry/backoff policy for idempotent GETs.

## Phase 2: Core flows (mirrors backend MVP table) — DONE

All flows are implemented end-to-end against the real gateway endpoints, with
`flutter analyze`/`flutter test` clean. Repositories are still hand-written against
plain models rather than the generated client (see Phase 1's codegen item above — that's
a deliberate, separate follow-up). Not yet exercised against a running backend or
covered by widget tests beyond the login-screen smoke test (that's Phase 3).

- [x] Auth: register, login, session restore, logout —
      `features/auth/{data,domain,presentation}`. Session status lives in
      `core/auth/session_state.dart` (shared by the router and the forced-logout path in
      `AuthInterceptor`); token refresh already covered in Phase 1's interceptor.
- [x] Onboarding/KYC: submit (`POST /v1/kyc`), status (`GET /v1/kyc/status`), gates the
      home shell on `approved` — `features/onboarding_kyc/`,
      `features/home/presentation/home_shell_screen.dart`. Rejected shows the form again
      with the reason; pending shows a manual-refresh screen (no true async KYC flow to
      poll against yet, since MVP auto-approves synchronously).
- [x] Wallet home: balance + paginated transaction history with pull-to-refresh and
      infinite scroll — `features/wallet/`. Amounts stay decimal strings end-to-end
      (never parsed to `double`), consistent with `pkg/money`.
- [x] P2P transfer: recipient (phone/email) → amount/memo → confirm → result, incl.
      fraud-declined (422) and a stable per-flow Idempotency-Key reused on retry —
      `features/transfers/`.
- [x] Cards: list, issue, detail (masked PAN via `last_four` only — full PAN is never
      returned by the gateway), freeze/unfreeze — `features/cards/`.
- [x] Authorizations list + detail, incl. capture (`POST /v1/authorizations/{id}/capture`)
      when a hold is still `authorized` — `AuthorizationsController`,
      `authorizations_list_screen.dart`, `authorization_detail_screen.dart`; entry point
      is the receipt icon in the Cards tab app bar.
- [x] Notifications inbox: list, mark read/all-read, unread badge, 30s polling —
      `features/notifications/`.

## Phase 3: Quality & mobile CI

- [ ] Unit tests (usecases/notifiers), widget tests for auth + transfer flows,
      one integration test (patrol/integration_test) against local compose stack.
- [ ] Contract safety: CI check that regenerating the Dart client from openapi.yaml is
      clean (same pattern as backend codegen check).
- [ ] `mobile.yml` GitHub workflow: analyze, test, build APK + iOS (no codesign) on PRs
      touching `mobile/`; path-filter so Go CI doesn't run for mobile-only changes.
- [ ] Security pass: certificate pinning (staging/prod), no secrets in code, screenshot
      obscuring on app switcher for balance screens, jailbreak/root detection decision.

## Phase 4: Release & beyond

- [ ] Fastlane lanes + CI: TestFlight / Play internal track from tags; build number from CI.
- [ ] Crash/analytics: Sentry (or Firebase Crashlytics) with correlation-id breadcrumbs.
- [ ] Push notifications: **requires backend work** — FCM/APNs sender in notification
      service (device token registry + send on event ingest); then deep links from push
      to transaction/card screens.
- [ ] Biometric unlock (local_auth) gating app open + transfer confirmation.
- [ ] Localization scaffold (intl), dark mode, accessibility audit.
