# Fix Frontend S3 Upload Error Handling Implementation Plan

**Goal:** Show meaningful error notifications when file attachment uploads fail, instead of silently swallowing errors.

**Architecture:** The backend already returns `{errors: [...], success: [...]}` in the response body (HTTP 200) for batch upload semantics. The frontend helper `uploadFiles()` detects errors and throws, but the component caller never awaits or catches. Fix the caller to handle errors, and fix the error formatting to produce readable messages.

**Tech Stack:** Vue 3 + TypeScript, vue-i18n

---

### Task 1: Fix error message formatting in `uploadFiles()`

The current code `throw Error(response.errors)` passes an array of `{code, message}` objects to `Error()`, which coerces them to `"[object Object]"`. Instead, extract the message strings and join them.

**Files:**
- Modify: `frontend/src/helpers/attachments.ts:32-34`

**Step 1: Fix the error throw to produce a readable message**

Replace lines 32-34 in `frontend/src/helpers/attachments.ts`:

```ts
// Before:
if (response.errors !== null) {
	throw Error(response.errors)
}

// After:
if (response.errors !== null) {
	const messages = response.errors.map((e: {message: string}) => e.message)
	throw new Error(messages.join('\n'))
}
```

**Step 2: Run frontend lint to verify**

Run: `cd frontend && pnpm lint`
Expected: PASS (no new lint errors)

**Step 3: Commit**

```bash
git add frontend/src/helpers/attachments.ts
git commit -m "fix: format attachment upload error messages as readable strings"
```

---

### Task 2: Add error handling in `Attachments.vue` upload caller

The function `uploadFilesToTask()` at line 336-338 calls `uploadFiles()` without `await` or error handling. The promise rejection goes unhandled. Add `try/catch` with the existing `error()` notification function (already imported at line 192).

**Files:**
- Modify: `frontend/src/components/tasks/partials/Attachments.vue:336-338`

**Step 1: Add try/catch to `uploadFilesToTask`**

Replace lines 336-338:

```ts
// Before:
function uploadFilesToTask(files: File[] | FileList) {
	uploadFiles(attachmentService, props.task.id, files)
}

// After:
async function uploadFilesToTask(files: File[] | FileList) {
	try {
		await uploadFiles(attachmentService, props.task.id, files)
	} catch (e) {
		error(e)
	}
}
```

This mirrors the existing pattern used in `deleteAttachment()` at line 346-358 in the same file. The `error` import from `@/message` is already present at line 192.

**Step 2: Run frontend lint to verify**

Run: `cd frontend && pnpm lint`
Expected: PASS

**Step 3: Run frontend type check**

Run: `cd frontend && pnpm typecheck`
Expected: PASS

**Step 4: Commit**

```bash
git add frontend/src/components/tasks/partials/Attachments.vue
git commit -m "fix: handle attachment upload errors with user-visible notifications"
```

---

### Task 3: Manual verification

There are no existing unit tests for `uploadFilesToTask` or `uploadFiles`. Verify manually:

**Step 1: Verify error display with backend returning errors**

1. Start the dev server: `cd frontend && pnpm dev`
2. Upload a file to a task
3. If using S3 with a misconfigured endpoint, the error should now show as a toast notification at the bottom left
4. If using local storage, simulate by temporarily making the files directory read-only, then uploading

**Step 2: Verify partial success still works**

1. Upload multiple files where one exceeds the size limit
2. The successful files should still appear as attachments
3. The failed file(s) should show an error toast with the message from the backend

---

### Summary of changes

| File | Change |
|------|--------|
| `frontend/src/helpers/attachments.ts:32-34` | Format error array into readable message string |
| `frontend/src/components/tasks/partials/Attachments.vue:336-338` | Add `async`, `await`, `try/catch` with `error()` notification |
