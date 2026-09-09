# Project Guidelines

## Project Intent

This is an educational SaaS proof of concept built with Go, Chi, Templ, HTMX,
pgx, PostgreSQL, and Goose. Keep changes explicit, readable, and narrowly
scoped. Prefer the standard library and the project's existing dependencies
over adding frameworks or abstractions without a concrete need.

See `README.md` for setup and feature documentation and `Makefile` for the
supported development commands.

## Architecture

Follow the dependency direction used by the TODO feature:

```text
HTTP handler -> service/domain logic -> repository -> PostgreSQL
```

- `cmd/server/main.go` is the composition root. Create infrastructure,
  repositories, and services there, then inject them into route registration.
- `internal/delivery/<area>` owns HTTP concerns: routes, request parsing,
  status codes, headers, locale selection, and Templ rendering.
- `internal/<domain>` owns domain types, validation, business rules, service
  contracts, and persistence contracts.
- Repository implementations own SQL and persistence error translation. They
  must not contain HTTP, rendering, or business workflow logic.
- Keep handlers thin: parse and validate transport syntax, call a service or
  domain function, map domain outcomes to HTTP, and render a response.
- Pass `context.Context` from the request through services and repositories.
- Inject dependencies through constructors or handler closures. Do not add
  package globals for repositories, services, or database connections.
- Add interfaces at meaningful boundaries where multiple implementations or
  isolated tests are useful. Do not create pass-through abstractions solely to
  satisfy a layer diagram.

## Go Conventions

- Format changed Go files with `gofmt`.
- Keep packages focused and use existing naming patterns: `RegisterRoutes`,
  `handle<Feature><Action>`, `New<Type>`, and `Err<Condition>`.
- Return errors to the layer that can handle them. Define stable domain errors
  in the domain package and inspect them with `errors.Is`.
- Log unexpected errors with useful operation context, but return generic
  server-error messages to clients.
- Avoid unrelated refactors and preserve the project's small, direct style.

## HTTP, HTMX, And Templ

- Render Templ components with `.Render(r.Context(), w)` and set the response
  content type before writing HTML.
- Use semantic status codes. Existing conventions include `201` for creation,
  `422` for validation failures, `404` for missing resources, `429` for rate
  limiting, and `500` for unexpected failures.
- Full-page handlers render page components; HTMX mutation handlers render the
  smallest region needed to update the UI.
- Preserve the dashboard `lang` query parameter in HTMX URLs by using the
  existing `localizedPath` helper.
- Pass typed view models to templates rather than generic maps.
- Edit `*.templ`, never generated `*_templ.go` or `*.templ.go` files. Run
  `make generate` after template changes; generated files are ignored by Git.

## Localization

- Dashboard handlers load locale from `lang` with `views.LoadLocale`.
- User-visible dashboard text belongs in both
  `internal/delivery/dashboard/views/locales/en.json` and `it.json`.
- Use dotted, domain-oriented translation keys and access them through
  `Locale.Text`; do not duplicate translated strings in handlers or templates.
- When business validation needs localized messages, depend on a small message
  lookup contract rather than importing the delivery package into the domain.

## Persistence And Migrations

- Define repository contracts in the owning domain package and PostgreSQL
  implementations in `postgres_repository.go`-style files.
- Use pgx with parameterized `$1`, `$2`, ... placeholders. Never concatenate
  request data into SQL.
- Translate storage-specific absence errors such as `pgx.ErrNoRows` into
  domain errors before returning from the repository.
- Add schema changes as reversible Goose migrations under `migrations/` using
  the next zero-padded sequence number and `-- +goose Up`/`Down` sections.
- Keep database constraints for data integrity even when the service also
  validates the same business input.

## Testing And Validation

- Add focused unit tests beside domain code. Use small hand-written stubs for
  service dependencies, as in `internal/todos/service_test.go`.
- Use dashboard integration tests for complete HTTP-to-database behavior.
  Keep them isolated through `TEST_DATABASE_URL` and reset changed tables.
- After Go-only changes, run `gofmt` on touched files and `go test ./...`.
- After Templ changes, run `make generate` and then `make test`.
- After migration or persistence changes, also run `make test-integration`.
- Use `make build` when changing startup, dependency wiring, Docker-related
  build behavior, or generated templates.
- Do not weaken or delete tests to make a change pass. Report unrelated test
  failures separately.

## Security And Runtime Boundaries

- Treat all request values and forwarding headers as untrusted input.
- Rely on Templ's escaped expressions for user-controlled strings; do not emit
  unescaped HTML without an explicit review of the data source.
- Preserve parameterized SQL, bounded server timeouts, and the unprivileged
  scratch-container runtime.
- Do not assume authentication, authorization, CSRF protection, or distributed
  rate limiting exists. These are not currently general platform guarantees;
  flag them when a change depends on them.
- Keep secrets and environment-specific connection strings out of source.
  Runtime database configuration uses `DATABASE_URL`; integration tests use
  `TEST_DATABASE_URL`.

## Change Discipline

- Read the owning handler, service/domain code, repository, and nearest tests
  before changing behavior. Make the smallest coherent change across layers.
- Update documentation when commands, environment variables, routes, or
  architecture change.
- Do not edit generated output, vendored dependencies, or unrelated files.
- Before finishing, check the diff for accidental generated files, credentials,
  broad formatting churn, and stale documentation claims.
