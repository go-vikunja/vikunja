# OpenAPI 3 Migration Analysis

Investigation of options for producing an OpenAPI 3 spec for the Vikunja API,
currently generated as Swagger 2.0 via `swaggo/swag`.

## Current state

- Generator: `github.com/swaggo/swag` v1.16.6 (annotation-driven)
- Invocation: `mage generate:swagger-docs` →
  `swag init -g ./pkg/routes/routes.go --parseDependency -d . -o ./pkg/swagger`
- Output: `pkg/swagger/swagger.json`, `"swagger": "2.0"`
- Web framework: `labstack/echo/v5` (community fork, not mainline v4)
- API surface: **119 paths, 161 operations, 87 model definitions**
- Spec-to-code consistency is **not enforced** — drift is possible and silent

## Option A — Upgrade to swaggo/swag v2

swaggo has a `v2` branch that adds OpenAPI 3.1 generation via a `-v3.1` CLI flag.

**Readiness verdict: experimental, not production-ready.**

- Latest tag: `v2.0.0-rc5` (2026-01-08). No stable v2.0.0 — RC line started in 2023
  (~3 years in RC)
- Tracking issue #548 "Proposal swag 2.0 using OpenAPI 3.0" still open, no GA
  timeline (178 reactions, 44 comments, last activity Dec 2025)
- PR #2156 (Mar 2026) explicitly states *"v2 is out of sync with master by a
  large margin"*
- ~11 open PRs against `base:v2`, several sitting 6–12+ months
- Known unfixed correctness bugs in v3 output: formData (#1554, since 2023),
  echo-swagger integration (#1588), oneOf/embedded-struct/binary attachment
  fixes unmerged
- Stable production line is still v1.16.x (Swagger 2.0 only)

**Recommendation:** Track, revisit at GA. Not viable today.

## Option B — Convert (post-process Swagger 2.0 → OpenAPI 3)

Keep the existing swag-driven pipeline, add a conversion step.

| Tool | Output | Notes |
|---|---|---|
| `swagger2openapi` (Mike Ralphson, Node.js) | OAS 3.0.x | De facto standard, battle-tested |
| `api-spec-converter` (LucyBot) | OAS 3.0 | Older, less maintained |
| `gnostic` (Google) | OAS 3.0 protobuf model | Go-native, heavier |

**Trade-offs**

- Delivers OAS **3.0** only, not 3.1
- Some 2.0 patterns translate awkwardly (`additionalProperties: true`,
  `nullable`, file uploads, security schemes)
- Documentation-quality output is fine; client/server *generation* from
  converted specs sometimes hits oddities
- Does **not** solve the spec-to-code drift problem

**Effort:** ~half a day. Add Node toolchain (or Docker step) to CI.

**When it wins:** If downstream consumers just need OAS 3.0 for docs/SDK
generation and spec fidelity isn't a correctness requirement.

## Option C — Switch the code-first generator

| Tool | Approach | Echo v5 fit | Notes |
|---|---|---|---|
| **Huma** (`danielgtaylor/huma` v2) | Typed handler funcs, spec derived from input/output structs | Adapter is v4; ~1-day fork for v5 | OAS **3.1** + JSON Schema 2020-12. Most active Go-native generator. |
| Fuego | Web framework + auto-spec | Own router | Would replace Echo entirely |
| go-swagger | Annotation-driven | n/a | Still 2.0, no real OAS 3 support |

**Huma details** (source: `github.com/danielgtaylor/huma/v2`, v2.37.3 Mar 2026,
~4k stars, 1,371 commits, production-ready per README):

- Runs on top of the existing router — existing Echo middleware (JWT, CORS,
  rate-limit, Sentry, logger) keeps working unchanged
- Handler signature:
  ```go
  huma.Register(api, huma.Operation{
      OperationID: "get-task", Method: "GET", Path: "/tasks/{id}",
  }, func(ctx context.Context, in *struct{
      ID int `path:"id"`
  }) (*struct{ Body Task }, error) { ... })
  ```
- Validation automatic from struct tags (`minLength`, `pattern`, `enum`,
  `format`, `required`...); business-rule validation stays in models/services
- Security schemes declared in `huma.Config`, referenced per-operation; Huma
  does **not** verify tokens (existing JWT middleware stays authoritative)
- Errors use RFC 9457 `application/problem+json`; customizable via
  `huma.NewError` hook to bridge to existing `models.Err*` types
- Incremental migration supported: same `*echo.Echo`, two specs served in
  parallel, migrate endpoint-by-endpoint

## Option D — Spec-first / inverted

Hand-write `openapi.yaml`, generate handlers from it.

| Tool | What it generates | Server frameworks |
|---|---|---|
| `oapi-codegen` | Types + server interface + clients | Echo v4, Chi, Gin, Fiber, net/http (no v5) |
| `ogen` (ogen-go/ogen) | Strict types + own radix router + clients | None — replaces router |
| OpenAPI Generator (openapitools) | Stubs + clients (many languages) | Go template, not idiomatic |

**Cost for vikunja**

1. Author ~3000+ lines of hand-curated YAML for 161 ops × 87 models
  (weeks of work to reach parity)
2. Wire generated interfaces into existing services (touches every route)
3. Lose "code is source of truth"; gain "spec is source of truth" *only if*
  enforced by CI
4. Echo v5 friction for `oapi-codegen`; ogen drops Echo entirely

**When it wins:** If you also want multi-language SDKs, server-side validation
as a primary feature, and can absorb several weeks of upfront investment.

## Recommendation: Huma (Option C)

Best balance of:
- OAS 3.1 output (not just 3.0)
- Compile-time guarantee that handlers match spec (solves drift)
- Preserves existing Echo + middleware + permissions investment
- Supports incremental migration with zero big-bang risk

ogen is technically excellent but the wrong shape — it requires abandoning
Echo and writing the spec first, which conflicts with vikunja's existing
architecture.

## Huma migration plan

### Vikunja-side ground truth

- **31 generic `WebHandler` instances** in `pkg/routes/routes.go` drive the
  bulk of the 161 endpoints via a shared `Create/Read/ReadAll/Update/Delete`
  codepath (`pkg/web/handler/*.go`) with a struct-factory and central
  permissions/validation. Re-implementing this generic layer once in Huma
  auto-covers most CRUD endpoints.
- **~30 custom handler files** in `pkg/routes/api/v1/` (login, OAuth, OIDC,
  TOTP, attachments, exports, webhooks, CalDAV) — per-endpoint rewrite work.
- **Validation:** `govalidator` via `c.Validate(i)`
  (`pkg/routes/validation.go:52`).
- **Errors:** centralized `CreateHTTPErrorHandler`
  (`pkg/routes/error_handler.go`) mapping `models.Err*` and `httpCodeGetter`
  to HTTP codes.
- **Web framework:** `labstack/echo/v5` with `*echo.Context` (pointer).

### What the migration touches

**1. Echo v5 adapter (~1 day)**
Huma ships `humaecho` for `labstack/echo/v4` only (~150 LOC wrapping
`echo.Context`). Fork it for v5: swap the import, adjust for the pointer
`Context` shape, write tests. Community precedents exist (`superstas/huma`,
`eugenepentland/huma`).

**2. Generic `WebHandler` rewrite (~2–3 days)**
The leverage point. Re-express `pkg/web/handler/{create,read,readall,update,delete}.go`
using Huma generics so all 31 existing generic handlers migrate in one pass.

**3. Custom handlers (~3–6 days)**
Each becomes one `huma.Register` call. Mechanical but per-file.

**4. Validation bridge (~3–5 days)**
Two paths:
- **Migrate tags:** convert `valid:"required,email"` →
  `validate:"required" format:"email"`. Tedious but a clean win — constraints
  become visible in the spec.
- **Keep govalidator:** Huma middleware calls the existing validator after
  tag validation. Cheaper, but constraints stay invisible in the spec.

**5. Error mapping (~1 day)**
Wrap existing `models.Err*` types via `huma.NewError` so
`httpCodeGetter` + RFC 9457 `application/problem+json` produce matching
responses.

**6. Auth declaration (~0.5 day)**
JWT middleware stays on Echo. Declare bearer scheme once in
`huma.Config.Components.SecuritySchemes`; reference per-operation via
`Operation.Security`. Permissions logic in models stays exactly as-is.

### Incremental rollout

Same `*echo.Echo`, two specs served in parallel during transition:

```
/api/v1/<legacy>   ← swag-annotated handlers, /swagger.json (OAS 2.0)
/api/v1/<migrated> ← Huma handlers,          /openapi.json (OAS 3.1)
```

Migrate one resource at a time, ship continuously. `check:got-swag` stays
valid for unmigrated routes; add `check:openapi` for the Huma portion. When
the last endpoint is moved, retire swag.

### Budget

| Phase | Effort |
|---|---|
| Adapter fork + generic CRUD shell | ~1 week |
| Custom handlers | ~1 week (parallelizable) |
| Validation tag migration | ~3–5 days |
| Error/auth bridge + spec wiring + tests | ~3 days |
| Frontend SDK regeneration + smoke testing | ~1 week |
| **Total** | **3–4 weeks focused work, shippable incrementally** |

### Risks

- **Echo v5 fork divergence** — if labstack/echo/v5 changes its `Context`
  shape, the adapter needs updating. Surface is small, low risk.
- **Frontend impact** — JSON shapes stay compatible (same tags), but error
  JSON changes to RFC 9457. Plan one round of frontend service updates.
- **Two-spec window** — needs documenting for API consumers. The only
  realistic alternative is a multi-month freeze.

## Suggested first step: spike

Before committing to the full migration, run a 2–3 day spike:

1. Fork `humaecho` for `echo/v5`
2. Port one self-contained resource (e.g. `Label` — small surface, full CRUD)
3. Stand up `/openapi.json` alongside the existing `/swagger.json`
4. Verify adapter, generic-CRUD bridge, auth, and error wiring end-to-end

If the spike lands cleanly, the rest is repetition at known cost.
