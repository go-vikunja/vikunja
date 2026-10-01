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

- Resource data goes through the generated SDK and types in `frontend/src/client/generated`, wrapped by the query modules in `frontend/src/client/queries/`. Bootstrap and session calls are the exception and call the SDK directly: `/info` in `stores/config.ts`, token refresh in `helpers/auth.ts`, login, register and logout in `stores/auth.ts`. Use the snake_case fields directly. Importing `axios` or any `models`, `modelTypes` or `services` path is a lint error (`no-restricted-imports` in `frontend/eslint.config.js`).
- After adding or changing a v2 route or schema, run `mage generate:frontend-client` and commit the generated output. Never hand-edit it. `mage check:frontend-client` verifies it is current and generation is repeatable.
- `client/http.ts` configures the shared client: base URL, bearer token, request-context fence, and a single refresh-and-retry on 401 code 11. Don't wrap the transport.
- The only other client is `client/publicClient.ts`, for calls that must skip the token and the 401 retry (`/info`, token refresh). It has no base URL; pass `baseUrl: canonicalApiBaseUrl(getApiBaseUrl())` per call. Token refresh is the one `throwOnError: false` caller: it throws a `RefreshTokenError` whose `failure.kind` (`network`, `server`, `rate-limited` with `retryAt` from `Retry-After`, `rejected`) tells an outage or rate limit from a rejection, and `refreshToken()` holds further refreshes until `retryAt`.
- Failures throw the parsed problem body (`VikunjaErrorModel`) with top-level `status` and `detail`, plus `code` when the server sets one. `client/problem.ts` stamps `status`/`detail` onto Echo-level bodies (rate limit, invalid JWT) that never reached Huma. Transport failures throw an `Error`. Read `status` and `code` directly off the caught value.
- A component that catches a request error returns early on `isRequestContextAbort(e)` (`client/requestContext.ts`): the identity, session or API base changed mid-request, so there is nothing to show.
- Blob downloads pass `parseAs: 'blob'` and narrow with `expectBlob(data, label)` (`client/queries/blobResponse.ts`). The generated types claim the JSON schema, so an unchecked cast hides a JSON body.
- Display lists paginate: key by page, clamp user-configured sizes with `pageSizeFor()`. Sweep every page with `fetchAllPages()` only when the client needs the whole set (labels, gantt tasks, shares), and pass `per_page: API_MAX_PER_PAGE` (`client/queries/pagination.ts`). The server silently caps `per_page` at its `max_items_per_page` (reported by `/info`), so a sweep may take more pages than requested; asking for more is not an error.
- Every resource gets one `XResponse` type plus `normalizeX()` applied in each `queryFn` (`ProjectResponse`, `TaskResponse`). Guaranteed fields are exactly the backend fields without `omitempty`; date strings, nullable pointers and expand-only fields stay optional. Merge write responses raw and normalize the result; normalizing the response first replaces cached expansions with `[]`.

### Query cache (TanStack Query)

`frontend/src/client/queries/projects.ts` plus `frontend/src/composables/useProjects.ts` is the reference implementation. Pattern follows the TanStack docs and TkDodo's "Effective React Query Keys" / "Mastering Mutations": one key factory, option factories, one thin hook per operation.

Layering, per feature `foo`:

- `client/queries/foo.ts`: `fooKeys` factory, `foosQuery()` via `queryOptions()`, `createFooMutationOptions()` etc. via `mutationOptions()`, thin hooks `useCreateFooMutation()` = `useMutation(createFooMutationOptions())`, imperative readers `ensureFoos()` / `refreshFoos()`, and pure lookup helpers over `Foo[]`.
- `composables/useFoos.ts`: read side only. `useQuery(foosQuery())`, `data ?? []`, `isPending`, lookup helpers bound to the reactive list. Do not put mutations in here; a create-only view must not subscribe to the list.
- Components read through `useFoos()` and write through `useCreateFooMutation()` etc. They never touch `queryClient`.
- `*MutationOptions()` are exported for tests and for the outside-component case below. Components and composables must not pass them to `useMutation` themselves; use the matching `use*Mutation()` hook.
- Outside components (Pinia setup stores, router, plain modules) `use*` hooks have no inject context. Read via `ensureFoos()` / `refreshFoos()`; write via `useMutation(createFooMutationOptions(), queryClient)` in the store setup.

Rules:

- Mutations use `contextMutationOptions` (`client/queries/contextMutation.ts`). It drops `onSuccess`/`onError`/ `onSettled` once the request context changed, toasts failures unless `toastError(input)` returns false, and its `optimistic` option does the cancel, snapshot and rollback. `toastError` sees the input, not the error, so it cannot suppress a single status.
- Failed queries toast through `QueryCache.onError` in `client/queryClient.ts`; cancellations and request-context aborts are skipped. A query whose consumer renders its own error state sets `meta: {handlesError: true}`, or the user gets both. Mutations are not covered there.
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
- `useAuthStore().info` is the `GET /user` query (`currentUserQuery`, key `accountKeys.user(id, type)`), falling back to the JWT's id and username until it loads. `is_admin` and settings come only from the query. Write the account through the mutations in `client/queries/account.ts`, not the store. `setSession()` clears the whole cache when the identity changes.
- A mutation whose input or result holds a secret (password, token) is mounted with `useSecretMutation` instead of `useMutation`. It sets `gcTime: 0` and resets the mutation once it settles, so the component needs no `reset()`; `gcTime` alone does nothing while a mounted observer still holds the mutation. Read a returned secret from `mutateAsync`, never from `data`.

## OpenAPI

- v2 generates its spec from Go types. No annotations.
- v1 uses swaggo annotations. When you touch a v1 handler, keep its annotations accurate and add them if missing.
- `pkg/swagger/` is generated by CI after commit. Never edit it, and don't run `mage generate:swagger-docs` unless asked.
