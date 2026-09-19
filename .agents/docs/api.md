# API design

## Version policy

- `/api/v1` (Echo, `pkg/routes/api/v1/`) is frozen. It keeps running and is supported, but does not grow. Touch it only to fix a bug or to port a resource to v2.
- `/api/v2` (Huma, `pkg/routes/api/v2/`) gets every new route: new entities, non-CRUD actions, new actions on existing resources. "Add an endpoint for X" without a version means v2.
- Before adding a v2 route, invoke the `api-v2-routes` skill.
- New v2 operations are not exposed over MCP unless you add their operation ID to the allow-list in `pkg/modules/mcp/exposure.go`, as either a typed tool or a catalog action (`find_action`/`do_action`).
- Models in `pkg/models/` are shared by both APIs. A new entity still gets a model with `Can*` methods (`crudable` skill); only the HTTP surface differs.

## Conventions

- v2 verbs: POST creates, PUT/PATCH update. (v1: PUT creates, POST updates.)
- kebab-case URLs, snake_case JSON.
- Standard CRUD in both versions goes through the `pkg/web/handler/` `Do*` functions, which call the model's `Can*` methods.
- Permissions live in the model (`CanRead`/`CanWrite`/`CanCreate`/`CanDelete`), never in CRUD handlers. One exception: non-CRUD v2 actions have no `Do*` wrapper, so the handler must load the entity and call `Can*` itself. The `api-v2-routes` skill shows the shape.

## Frontend clients

- Always consume new API routes through the generated functions and types in `frontend/src/client/generated`. Use their snake_case fields directly.
- `frontend/src/models`, `frontend/src/modelTypes`, and `frontend/src/services` are the legacy v1 architecture. They are being migrated gradually and will be removed. Do not add models, interfaces, or service wrappers there for new routes; existing code can remain until migrated.
- After adding or changing a v2 route or schema, run `mage generate:frontend-client` and commit the generated output. Never hand-edit it. `mage check:frontend-client` verifies it is current and generation is repeatable.
- Reuse the shared client configuration in `frontend/src/client/http.ts`. Put shared query/cache behavior in `frontend/src/client/queries/` when needed; do not duplicate the generated transport layer.
- List queries that load every page use `fetchAllPages` (in `frontend/src/client/queries/`). Mutations fenced to the client request context use `contextMutationOptions`; its `optimistic` option does the cancel, snapshot and rollback. Its `toastError` option sees the mutation input, not the error, so it cannot suppress a single status.
- Failures throw the parsed problem body: `status`, `code`, `detail` are top-level and there is no Axios `.response`. When switching a consumer off the legacy services, grep it for `response?.status`, `response.status` and `response.data`; every hit is dead after the switch (`Description.vue` kept a `error.response.status === 404` guard that never matched and blocked navigation).
- Every resource gets one `XResponse` type plus `normalizeX()` applied in each `queryFn` (`ProjectResponse`, `TaskResponse`). Guaranteed fields are exactly the backend fields without `omitempty`; date strings, nullable pointers and expand-only fields stay optional. Merge write responses raw and normalize the result; normalizing the response first replaces cached expansions with `[]`.

### Query cache (TanStack Query)

`frontend/src/client/queries/labels.ts` plus `frontend/src/composables/useLabels.ts` is the reference implementation. Pattern follows the TanStack docs and TkDodo's "Effective React Query Keys" / "Mastering Mutations": one key factory, option factories, one thin hook per operation.

Layering, per feature `foo`:

- `client/queries/foo.ts`: `fooKeys` factory, `foosQuery()` via `queryOptions()`, `createFooMutationOptions()` etc. via `mutationOptions()`, thin hooks `useCreateFooMutation()` = `useMutation(createFooMutationOptions())`, imperative readers `ensureFoos()` / `refreshFoos()`, and pure lookup helpers over `Foo[]`.
- `composables/useFoos.ts`: read side only. `useQuery(foosQuery())`, `data ?? []`, `isPending`, lookup helpers bound to the reactive list. Do not put mutations in here; a create-only view must not subscribe to the list.
- Components read through `useFoos()` and write through `useCreateFooMutation()` etc. They never touch `queryClient`.
- `*MutationOptions()` are exported for tests and for the outside-component case below. Components and composables must not pass them to `useMutation` themselves; use the matching `use*Mutation()` hook.
- Outside components (Pinia setup stores, router, plain modules) `use*` hooks have no inject context. Read via `ensureFoos()` / `refreshFoos()`; write via `useMutation(createFooMutationOptions(), queryClient)` in the store setup.

Rules:

- `mutationFn` shapes input, calls the generated client, returns the narrowed entity. No separate `createFoo` wrapper unless something else calls it. The client is configured with `throwOnError: true`, so failures throw; don't add `error` checks.
- All cache writes (`setQueryData`, `invalidateQueries`, `cancelQueries`) live in mutation option callbacks. Use the `client` passed in the callback context, not the `queryClient` singleton. Never write to the cache from a plain exported function, a store action, or a socket handler.
- Updaters must bail when the cached value is `undefined` (`current ? ... : current`, or `current?.map(...)`), so a query nobody mounted is never materialized.
- Every mutation invalidates the affected list key in `onSettled`. A targeted `setQueryData` in `onSuccess` alone is not enough: a refetch started during the request would overwrite it. For expensive lists that `onSuccess` already patched (the paginated project list, task lists, boards), pass `refetchType: 'none'` so the list is only marked stale instead of reloading every page.
- Also invalidate caches of other resources that derive from the changed relation, e.g. user search after a team membership or project share change (with `refetchType: 'none'`). Don't invalidate a parent detail that doesn't depend on the change.
- `cancelQueries` only as part of a full optimistic flow: `onMutate` cancels, snapshots and writes; `onError` restores the snapshot (only if one existed); `onSettled` invalidates. Never as a standalone guard around a request.
- `setQueryData` matches keys exactly; `invalidateQueries` matches by prefix. Pass the exact key of a live query to `setQueryData`, and add `exact: true` to `invalidateQueries` when a prefix would also hit detail keys.
- Only add keys the app has queries for. A `detail(id)` key without a detail query just creates orphan cache entries; list writes go to the list key with the id in the updater, not in the key.
- Key a query only by arguments a consumer actually passes. Add a key dimension when the consumer that needs it lands. The web frontend only sends HTML, so rich-text format is never a key dimension or request option.
- Use `placeholderData: keepPreviousData` only when the previous query shares the same parent key (e.g. the same project); otherwise the previous project's results show while the new one loads.
- Data embedded in a parent response (views in projects; labels, assignees and buckets in tasks) has no query of its own. Read it from the parent query and write it into the parent's list and detail caches; its query module holds mutations only.
- Id lookups must handle pseudo projects: `-1` is Favorites and other negative ids are saved filters. They only exist in the project list and have no detail endpoint. A project move leaves their lists and boards alone (membership is server-decided); the drag source removes its own card explicitly (`removeTaskFromBoard`).
- Kanban boards are patched, never refetched by task mutations: a refetch returns the first page per bucket and collapses scrolled buckets. `invalidateTaskMembership` forces `refetchType: 'none'` on board keys whatever the caller passes. The only board refetch is bucket delete, which also refetches the project because the server clears `default_bucket_id`/`done_bucket_id`.
- Paged lists are keyed by page number, not infinite queries. Removing an item rewrites `total` and `total_pages` on every cached page of that scope (key minus page segment, group by `hashKey`), not only the page holding it.
- A composable that snapshots data on first load (so an open edit form survives a background refetch) is a draft, not a read. Name it as one, e.g. `useProjectDraft`, and never use it for tables or lists that mutations must update; those read through a live `useQuery` composable. An edit form that needs a draft is a child component mounted once the data has loaded, keyed on the entity id, that seeds the draft once in setup; not watchers plus a stored draft id.
- Toasts for the mutation outcome go in the option callbacks; UI actions like redirects stay in the component. Call `mutate(x)` when nothing follows, `mutate(x, {onSettled})` for cleanup only, and `try { await mutateAsync(x) } catch { return }` with success-only code after the block. A `mutateAsync` rejection that escapes reaches the global error handler and toasts a second time.
- Mutation input holding secrets (passwords) sets `gcTime: 0`, and the component calls the mutation's `reset()` once it settles. `gcTime` alone does nothing while a mounted `useMutation` still observes the mutation.
- Test mutation options through the real lifecycle: `queryClient.getMutationCache().build(queryClient, options).execute(vars)`. To observe an optimistic write, assert inside the mocked request before throwing. Assert on our cache writes only; don't re-test TanStack's refetch or cancellation behaviour.

## OpenAPI

- v2 generates its spec from Go types. No annotations.
- v1 uses swaggo annotations. When you touch a v1 handler, keep its annotations accurate and add them if missing.
- `pkg/swagger/` is generated by CI after commit. Never edit it, and don't run `mage generate:swagger-docs` unless asked.
