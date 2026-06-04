---
name: api-v2-routes
description: Use when adding or changing a resource on the Huma-backed /api/v2 API (new endpoints, porting a v1 resource, editing pkg/routes/api/v2/). Covers per-operation Huma handlers, the shared envelopes, error/auth bridging, REST verb conventions, and what's automatic.
user-invocable: true
---

# Adding /api/v2 routes for a CRUDable resource

`/api/v2` is served by [Huma v2](https://github.com/danielgtaylor/huma) mounted on an Echo group via the vendored `pkg/modules/humaecho5` adapter. Unlike v1's generic `WebHandler`, each operation is a typed Huma handler registered explicitly. The handlers are thin: they pull auth off the context, call the same `pkg/web/handler.Do*` functions v1 uses, and translate domain errors into RFC 9457 responses.

**Reference implementation:** `pkg/routes/api/v2/labels.go` is the canonical example — copy its shape. Shared envelopes live in `pkg/routes/api/v2/types.go`; the auth/error bridge in `pkg/routes/api/v2/errors.go`; config in `pkg/routes/api/v2/huma.go`.

## Prerequisite: the model must be CRUDable

v2 handlers call `handler.DoReadAll/DoReadOne/DoCreate/DoUpdate/DoDelete`, which invoke the model's `Can*` methods. If the model isn't already a working v1 resource, do the model work first — invoke the **`crudable`** skill. Permissions are enforced at the model level; **never** re-check them in a v2 handler.

**Every exposed model field needs a `doc:` tag.** v2's schema is reflected from struct tags at runtime; Huma cannot read the Go doc comments swaggo uses for v1. A field without `doc:"..."` ships with no description in the spec. Add the tag alongside the existing comment (keep both — swaggo still reads the comment for v1, and they should stay in sync):

```go
// The title of the label. You'll see this one on tasks associated with it.
Title string `json:"title" minLength:"1" maxLength:"250" doc:"The title of the label. You'll see this one on tasks associated with it."`
```

These model edits are safe for v1 — swaggo, XORM, and govalidator all ignore the `doc` tag. (Huma *does* read validation tags like `minLength`/`maxLength`/`enum`/`format`, so those carry over without a `doc` tag.) As with operations, a `doc` tag earns its place when it says something the field name and type don't: a format hint ("hex, 6 chars"), a read-only note ("set by the server; ignored on write"), units, or allowed values. "The label description." on a `Description` field is filler. See `pkg/models/label.go` for the reference.

**Mark server-controlled fields `readOnly:"true"`.** Because the same model struct is the request body *and* the response, fields the client can never set — `id`, `created`, `updated`, `created_by`, and similar server-derived relations/IDs — should carry `readOnly:"true"`. Huma reflects this into the OpenAPI schema (`readOnly: true`), so docs and client generators present the field as response-only and drop it from request examples:

```go
ID        int64      `json:"id" readOnly:"true" doc:"The unique, numeric id of this label."`
CreatedBy *user.User `xorm:"-" json:"created_by" readOnly:"true" doc:"The user who created this label."`
Created   time.Time  `xorm:"created not null" json:"created" readOnly:"true" doc:"A timestamp when this label was created. You cannot change this value."`
```

The tag is **documentation only** — Huma does *not* reject these fields if a client sends them on create/update. Actual immutability still comes from the model layer (XORM-managed `created`/`updated`, `created_by` being `xorm:"-"` and set server-side). It's also harmless on v1 (swaggo/XORM/govalidator ignore it). Don't bother tagging fields that are already `json:"-"` (absent from the schema entirely), and skip it on response-only structs like the error model — there it's cosmetic since they never appear as a request body. See `pkg/models/label.go` and `pkg/user/user.go`.

## Steps

### 1. Create `pkg/routes/api/v2/<resource>.go`

Define the list-response body, a `Register<Resource>Routes(api huma.API)` function, and one handler per operation. Mirror `labels.go` exactly:

```go
// Element type matches what models.<Model>.ReadAll returns; extra fields
// tagged json:"-" keep the wire shape identical to the plain model.
type fooListBody struct {
	Body Paginated[*models.Foo]
}

func RegisterFooRoutes(api huma.API) {
	tags := []string{"foos"}

	Register(api, huma.Operation{
		OperationID: "foos-list",
		Summary:     "List foos",
		Description: "Returns the foos the authenticated user has access to, paginated.",
		Method:      http.MethodGet, Path: "/foos", Tags: tags,
	}, foosList)
	Register(api, huma.Operation{OperationID: "foos-read", Summary: "Get a foo", Description: "...", Method: http.MethodGet, Path: "/foos/{id}", Tags: tags}, foosRead)
	Register(api, huma.Operation{OperationID: "foos-create", Summary: "Create a foo", Description: "...", Method: http.MethodPost, Path: "/foos", Tags: tags}, foosCreate)
	Register(api, huma.Operation{OperationID: "foos-update", Summary: "Update a foo", Description: "...", Method: http.MethodPut, Path: "/foos/{id}", Tags: tags}, foosUpdate)
	Register(api, huma.Operation{OperationID: "foos-delete", Summary: "Delete a foo", Description: "...", Method: http.MethodDelete, Path: "/foos/{id}", Tags: tags}, foosDelete)
}
```

Use the package's `Register` wrapper, **not** `huma.Register` directly — it sets `DefaultStatus` from the verb (POST → 201, DELETE → 204). Don't spell out `DefaultStatus` unless you need a non-default code. Don't set `Security:` per operation — it's applied globally in `NewAPI`.

**Every operation needs a `Summary` and `Description`.** v2's OpenAPI spec is generated from these `Operation` fields at runtime — unlike v1's swaggo, Huma cannot read Go doc comments, so anything you don't put in the `Operation` (or in a `doc:` tag, see below) is simply absent from the spec and the docs UI. An operation without them ships undocumented.

**Make the description document the non-obvious — don't restate the verb+noun.** "Deletes a label" adds nothing over `DELETE /labels/{id}`. Spend the description on what a consumer *can't* infer from the method/path/schema: permission scope ("only the owner may delete it"; "returns only labels you can see, not a global list"), full-replace vs partial (PUT replaces, PATCH merges), read-only/conditional behavior (ETag → `If-None-Match` → 304), side effects (create sets ownership), non-obvious status codes. If the honest description is just the verb+noun, a short summary alone is fine — don't pad. See `labels.go` for the calibration.

### 2. Write the handlers

Every handler: pull auth with `authFromCtx(ctx)`, call the matching `handler.Do*`, wrap returned errors in `translateDomainError`. Use the shared envelopes from `types.go` (`singleBody`, `singleReadBody`, `emptyBody`, `ListParams`, `Paginated`/`NewPaginated`).

- **List** takes `*ListParams` (gives you `page`/`per_page`/`q` for free, already `doc:`-tagged in `types.go` — no need to re-document them) and returns `*fooListBody`. **You must type-assert the `DoReadAll` result to the concrete slice** — `result` is `any`, and a blind cast or a generic wrapper silently serialises `[]` (the "generic-any silent-empty trap"). Return a hard error on mismatch:
  ```go
  items, ok := result.([]*models.Foo)
  if !ok {
      return nil, fmt.Errorf("foos.ReadAll returned unexpected type %T", result)
  }
  return &fooListBody{Body: NewPaginated(items, total, in.Page, in.PerPage)}, nil
  ```
- **Extra query params go *directly* on the handler's input struct — not in a shared/embedded helper.** Beyond `ListParams`, if an operation needs its own query params (`expand`, `order_by`, `include_public`, …), declare each as a direct field with its own `query:"…"` tag on that operation's input struct, then bind it onto the model. A shared or embedded struct of query fields silently **fails to bind** under Huma when combined with other query params/embeds — the field arrives empty (hit while implementing Project's `expand`). Flatten them into the input struct.
- **Read** embeds `conditional.Params` in its input. To surface the caller's permission, define a small per-resource response struct that **embeds the model by value** and adds the permission: `type fooReadBody struct { models.Foo; MaxPermission models.Permission \`json:"max_permission" readOnly:"true" doc:"..."\` }`. Go and Huma both promote the embedded model's fields, so the wire shape is flat (model fields + `max_permission`) with no custom marshaler and nothing added to the shared model struct. Capture `DoReadOne`'s returned max permission (it is `0`/`1`/`2` on success — **never discard it as `_`**), build the body, and `return conditionalReadResponse(&in.Params, body, foo.Updated, maxPermission)`. The shared helper (in `types.go`) folds the permission into the ETag (so a share/role change invalidates the cache), applies the conditional precondition (304/412), and returns `*singleReadBody[fooReadBody]`. See `labels.go`/`project_views.go`. (A generic `struct{ T; ... }` is impossible — Go forbids embedding a type parameter — so the per-resource struct is the price of a flat shape without a marshaler.)
- **Create / Update** return `*singleBody[Model]` and set the model's `ID` from the path (URL wins over body). **Update's request body must be the same `fooReadBody` the read returns, not the bare model** — AutoPatch's GET→PUT round trip echoes the read body (max_permission included) into the PUT, and because `max_permission` is a declared `readOnly` property of `fooReadBody`'s schema, Huma accepts and ignores it on write rather than rejecting it. Take `&in.Body.Foo` (the embedded model — value-embedded, so never nil) and ignore the embedded `MaxPermission`. Create stays a bare `Body Model` (AutoPatch only round-trips into PUT).
- **Delete** returns `*emptyBody`.

### 3. Self-register the resource

Resources self-register — **you do not edit `pkg/routes/routes.go`**. In your resource file, add an `init()` that hands your registrar to `AddRouteRegistrar`:

```go
func init() { AddRouteRegistrar(RegisterFooRoutes) }

func RegisterFooRoutes(api huma.API) { ... }
```

`registerAPIRoutesV2` in `routes.go` calls `apiv2.RegisterAll(api)`, which runs every registered registrar (in init/filename order — route order is irrelevant) and then `EnableAutoPatch`. New resources touch zero shared lines, so they never conflict on `routes.go`.

Notes:

- **Give each registrar a DISTINCT name.** They share package `apiv2`, so two resources both exporting `RegisterAvatarRoutes` collide and won't compile — that actually happened and the upload one had to be renamed (`RegisterAvatarRoutes` for the binary endpoint vs `RegisterAvatarUploadRoutes` for the upload). Name yours after the specific resource.
- **Config-gated resources check the flag inside the registrar.** `RegisterAll` runs at request-router-setup time, after config is loaded, so a `RegisterFooRoutes` may early-return (or skip individual `Register` calls) based on `config.FooEnabled.GetBool()`. Don't try to gate at `init()` time — config isn't loaded yet.
- **AutoPatch is automatic.** `RegisterAll` calls `EnableAutoPatch` after all registrars — don't call it yourself, and don't register a manual PATCH (see "What's automatic").

## REST verb conventions (v2 inverts v1)

| Operation | v1 | v2 |
|---|---|---|
| create | PUT | **POST** |
| update | POST | **PUT** (and PATCH) |
| read / read-all / delete | GET / GET / DELETE | same |

## Non-CRUDable / custom routes

Not everything is plain CRUD — bulk operations, custom actions (`POST /tasks/{id}/duplicate`), sub-resource toggles, RPC-ish endpoints. These still go through Huma and reuse most of the machinery, but two responsibilities move **into your handler** because there's no `handler.Do*` doing them for you:

1. **Permission enforcement is now yours.** This is the one place the "never check permissions in the handler" rule inverts. With no generic `Do*` to call the model's `Can*`, the handler must do it explicitly — load the relevant entity and call its permission method, then refuse on denial. Mirror the v1 custom-handler shape (`pkg/routes/api/v1/task_attachment.go`):
   ```go
   func tasksDuplicate(ctx context.Context, in *struct{ ID int64 `path:"id"` }) (*singleBody[models.Task], error) {
       a, err := authFromCtx(ctx)
       if err != nil {
           return nil, err
       }
       s := db.NewSession()
       defer s.Close()

       t := &models.Task{ID: in.ID}
       can, err := t.CanUpdate(s, a) // or whichever Can* gates this action
       if err != nil {
           _ = s.Rollback()
           return nil, translateDomainError(err)
       }
       if !can {
           return nil, huma.Error403Forbidden("forbidden")
       }
       // ... do the work against s ...
       if err := s.Commit(); err != nil {
           return nil, translateDomainError(err)
       }
       return &singleBody[models.Task]{Body: t}, nil
   }
   ```
2. **Session / transaction management is now yours.** The `Do*` helpers open and commit their own `xorm.Session`; custom handlers open one with `db.NewSession()`, `defer s.Close()`, and `Commit`/`Rollback` explicitly for anything that writes.

Otherwise the same rules apply: register with the `Register` wrapper, pull auth via `authFromCtx`, route every error through `translateDomainError`, and reuse the `types.go` envelopes — or define a small body struct when none fits (don't bend a custom response into `singleBody` if it's awkward).

**Verb choice:** pick by semantics, not the CRUD table. Non-idempotent actions are `POST`. AutoPatch only synthesises PATCH for GET+PUT *pairs*, so standalone custom routes are never touched.

**Token permissions still automatic, but mind the derived name:** `collectRoutesForAPITokens` keys a route off its prefix-stripped path, so `POST /api/v2/tasks/{id}/duplicate` lands under the `tasks` group as a `duplicate` permission. Single-segment custom paths fall into the `other` group. Name the path so the derived `(group, permission)` reads sensibly — that string is what users grant tokens against.

## What's automatic — do NOT hand-roll

- **PATCH** — `EnableAutoPatch` synthesises a JSON-Merge-Patch PATCH for every GET+PUT pair. `RegisterAll` invokes it after all registrars, so it's automatic — don't call `EnableAutoPatch` and don't register PATCH yourself.
- **API token permissions** — `collectRoutesForAPITokens` walks the Echo router after registration, so your new routes land in the v2 token table automatically under the same `(group, permission)` keys as their v1 names. PATCH is intentionally not stored; `CanDoAPIRoute` accepts it as an alias for the stored PUT (see `pkg/models/api_routes.go`).
- **Security schemes** — `JWTKeyAuth` + `APITokenAuth` are declared globally in `NewAPI`. For a public endpoint, set `Security: []map[string][]string{}` on that operation and add its path to `unauthenticatedAPIPaths` in `routes.go`.
- **Error shape** — `translateDomainError` maps any `web.HTTPErrorProcessor` (e.g. `ErrFooDoesNotExist`) onto Huma's status error, producing RFC 9457 `application/problem+json`. Errors without HTTP semantics become 500.
- **OpenAPI spec / Scalar docs / `$schema` URLs** — handled in `huma.go`. Leave `Servers` alone (the relative entry must stay at index 0).

## Anti-patterns (these get flagged)

- Re-checking permissions in the handler instead of trusting `handler.Do*` → the model's `Can*`.
- Blind `result.([]*models.Foo)` without the `ok` check, or returning the `any` straight into the envelope — silent empty lists.
- `huma.Register` instead of the package `Register` wrapper (loses the verb-based status).
- Per-operation `Security:` lines (now global) or registering a manual PATCH (AutoPatch does it).
- Returning a raw model error instead of routing it through `translateDomainError` → leaks a 500 instead of the right code.
- Unquoted ETag in the response header.
- Operations without `Summary`/`Description`, or model fields without `doc:` tags — they ship undocumented because Huma can't read Go comments.
- Server-controlled fields (`id`, `created`, `updated`, `created_by`) on a shared input/output model left without `readOnly:"true"` — the docs then present them as writable request fields.

## Tests (mandatory)

Mirror the v1 webtest shape so v2 parity is readable side-by-side. Use the `webHandlerTestV2` harness in `pkg/webtests/integrations.go` — it takes the same `urlParams` map as v1's `webHandlerTest`. See `pkg/webtests/huma_label_test.go`:

- One `Test<Resource>` covering list/read/create/update/delete, positive + negative (forbidden, nonexistent), mirroring the v1 model test.
- v2-only behaviour (ETag/304, PATCH merge-patch) goes in separate top-level `Test<Resource>_*` funcs using the `humaRequest`/`humaTokenFor` helpers in `pkg/webtests/huma_helpers_test.go`.
- The RFC 9457 error-body shape is asserted **once** globally in `TestHuma_ErrorShapeIsRFC9457` — don't re-assert the full problem+json shape per resource, just the status code.

Run with `mage test:filter Test<Resource>` while iterating. **Caveat:** `mage test:filter` injects `-short`, which makes `pkg/webtests` skip entirely (the suite short-circuits in short mode), so it silently reports success without running your webtest. To actually exercise a single webtest, run it directly: `go test -run '<Name>' ./pkg/webtests/`. Save output to a file per the project test-output rule.

## Related

- `crudable` skill — the model-layer prerequisite
- `pkg/routes/api/v2/labels.go` — reference resource
- `pkg/routes/api/v2/{types,errors,huma}.go` — shared envelopes, bridge, config
- `pkg/web/handler/core.go` — the `Do*` functions handlers call
