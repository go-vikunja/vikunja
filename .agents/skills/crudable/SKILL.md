---
name: crudable
description: Use when adding or modifying a model in pkg/models/ that needs CRUD operations or permission checks. Covers Can* method placement, CRUDable interface, and required test coverage.
user-invocable: true
---

# CRUDable + Permissions

Models in `pkg/models/` that expose CRUD operations must implement the `CRUDable` interface **and** the permission methods. Permissions are enforced at the **model level** via `Can*` methods — never re-checked in route handlers.

**Reference docs:** read `pkg/web/readme.md` for the full interface definitions, DB session semantics, and call order. The interface lives at `pkg/web/web.go`. This skill is a checklist of what the review feedback surfaces on top of that.

## Before writing CRUD or route code

1. Decide which operations the model needs: Read / ReadAll / Create / Update / Delete.
2. Implement the matching permission methods on the model. Typical signatures:
   - `CanRead(s *xorm.Session, a web.Auth) (bool, int, error)`
   - `CanCreate(s *xorm.Session, a web.Auth) (bool, error)`
   - `CanUpdate(s *xorm.Session, a web.Auth) (bool, error)`
   - `CanDelete(s *xorm.Session, a web.Auth) (bool, error)`
3. If a handler or service needs to check access, call the `Can*` method. Do **not** re-implement the check inline or duplicate the logic in `pkg/routes/`.
4. Do not implement empty stub methods just to satisfy the interface, instead embed the interface in the struct. Check existing models to see how that's done.

Look at `pkg/models/project.go` or `pkg/models/task.go` for reference implementations.

The initial querying of the data should happen in the Can* function. Because we're operating on a pointer, the function that does the work should not need to re-query the model data.

## Anti-patterns (these get flagged every time)

- Permission logic inlined in `pkg/routes/` handlers instead of on the model.
- Shipping `Create` but forgetting `CanUpdate` / `CanDelete` because "only create is new right now".
- Re-querying the DB in the handler to decide access — that work belongs in `CanRead`.
- Copy-pasting permission logic across `CanUpdate` and `CanDelete` — extract a helper.
- Adding a handler that bypasses the generic CRUD handler in `pkg/web/handler/` without a clear reason (the generic handler already invokes the `Can*` methods for you).

## Tests (mandatory)

Every `Can*` method needs both positive and negative coverage. Run with `mage test:filter <TestName>` while iterating.

- User with direct permission → passes
- User without permission → denied
- Permission inherited via parent (e.g., project → task, team → project) → still passes
- Shared access edge cases (link shares, team membership) if the model supports them

## Related

- Generic CRUD handler: `pkg/web/handler/`
- Permission type definitions: `pkg/web/auth.go`, `pkg/models/permissions.go`
- After the model is stable, register the routes in `pkg/routes/api/v1/` and add Swagger annotations. Do not edit `pkg/swagger/` directly — it's generated.
