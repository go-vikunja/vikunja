# Fix Checkbox Persistence in Task Descriptions Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix the bug where checkbox states in task descriptions don't persist reliably after page refresh (GitHub issue #293, #563).

**Architecture:** Create a custom `TaskItemWithId` TipTap extension that extends the default TaskItem with a unique `taskId` attribute. The ID is stored as `data-task-id` in the HTML and used to reliably identify which checkbox was toggled, replacing the unreliable node reference comparison.

**Tech Stack:** Vue 3, TypeScript, TipTap 3.8.0, Vitest, nanoid

---

## Task 1: Add nanoid Dependency

**Files:**
- Modify: `frontend/package.json`

**Step 1: Add nanoid to dependencies**

```bash
cd frontend && pnpm add nanoid
```

**Step 2: Verify installation**

Run: `cd frontend && pnpm list nanoid`
Expected: Shows nanoid version installed

**Step 3: Commit**

```bash
git add frontend/package.json frontend/pnpm-lock.yaml
git commit -m "$(cat <<'EOF'
chore: add nanoid dependency for unique task item IDs
EOF
)"
```

---

## Task 2: Create Custom TaskItem Extension with ID Support

**Files:**
- Create: `frontend/src/components/input/editor/taskItemWithId.ts`

**Step 1: Create the extension file with ID attribute**

Create file `frontend/src/components/input/editor/taskItemWithId.ts`:

```typescript
import { TaskItem } from '@tiptap/extension-list'
import { nanoid } from 'nanoid'

/**
 * Custom TaskItem extension that adds a unique ID to each task item.
 * This fixes the checkbox persistence bug (GitHub #293, #563) by allowing
 * reliable identification of which checkbox was toggled.
 */
export const TaskItemWithId = TaskItem.extend({
	addAttributes() {
		return {
			...this.parent?.(),
			taskId: {
				default: null,
				parseHTML: (element: HTMLElement) => {
					// Preserve existing ID or generate new one
					return element.getAttribute('data-task-id') || nanoid(8)
				},
				renderHTML: (attributes) => {
					// Always ensure we have an ID
					const id = attributes.taskId || nanoid(8)
					return {
						'data-task-id': id,
					}
				},
			},
		}
	},
})
```

**Step 2: Verify file was created**

Run: `ls -la frontend/src/components/input/editor/taskItemWithId.ts`
Expected: File exists

**Step 3: Run linter to check syntax**

Run: `cd frontend && pnpm lint --fix frontend/src/components/input/editor/taskItemWithId.ts`
Expected: No errors (or auto-fixed)

**Step 4: Commit**

```bash
git add frontend/src/components/input/editor/taskItemWithId.ts
git commit -m "$(cat <<'EOF'
feat(editor): add TaskItemWithId extension with unique ID support

Extends TipTap's TaskItem to add a data-task-id attribute to each
checklist item. IDs are generated with nanoid and preserved through
HTML serialization.

Fixes: go-vikunja/vikunja#293
Fixes: go-vikunja/vikunja#563
EOF
)"
```

---

## Task 3: Write Unit Tests for TaskItemWithId

**Files:**
- Create: `frontend/src/components/input/editor/taskItemWithId.test.ts`

**Step 1: Write tests for the extension**

Create file `frontend/src/components/input/editor/taskItemWithId.test.ts`:

```typescript
import { describe, it, expect } from 'vitest'
import { Editor } from '@tiptap/core'
import StarterKit from '@tiptap/starter-kit'
import { TaskList } from '@tiptap/extension-list'
import { TaskItemWithId } from './taskItemWithId'

describe('TaskItemWithId Extension', () => {
	const createEditor = (content: string = '') => {
		return new Editor({
			extensions: [
				StarterKit,
				TaskList,
				TaskItemWithId.configure({ nested: true }),
			],
			content,
		})
	}

	it('should generate unique IDs for new task items', () => {
		const editor = createEditor()

		editor.commands.setContent('<ul data-type="taskList"><li data-checked="false"><p>Item 1</p></li></ul>')

		const html = editor.getHTML()
		expect(html).toContain('data-task-id=')

		editor.destroy()
	})

	it('should preserve existing IDs when parsing HTML', () => {
		const existingId = 'test-id-123'
		const editor = createEditor()

		editor.commands.setContent(`<ul data-type="taskList"><li data-checked="false" data-task-id="${existingId}"><p>Item 1</p></li></ul>`)

		const html = editor.getHTML()
		expect(html).toContain(`data-task-id="${existingId}"`)

		editor.destroy()
	})

	it('should generate different IDs for different items', () => {
		const editor = createEditor()

		editor.commands.setContent(`
			<ul data-type="taskList">
				<li data-checked="false"><p>Item 1</p></li>
				<li data-checked="false"><p>Item 2</p></li>
				<li data-checked="false"><p>Item 3</p></li>
			</ul>
		`)

		const html = editor.getHTML()
		const idMatches = html.match(/data-task-id="([^"]+)"/g)

		expect(idMatches).toHaveLength(3)

		// Extract IDs and verify they're unique
		const ids = idMatches!.map(match => match.match(/data-task-id="([^"]+)"/)?.[1])
		const uniqueIds = new Set(ids)
		expect(uniqueIds.size).toBe(3)

		editor.destroy()
	})

	it('should preserve IDs through getHTML/setContent round-trip', () => {
		const editor = createEditor()

		editor.commands.setContent('<ul data-type="taskList"><li data-checked="false"><p>Test</p></li></ul>')

		const html1 = editor.getHTML()
		const idMatch1 = html1.match(/data-task-id="([^"]+)"/)
		const originalId = idMatch1?.[1]

		// Simulate round-trip
		editor.commands.setContent(html1)

		const html2 = editor.getHTML()
		const idMatch2 = html2.match(/data-task-id="([^"]+)"/)
		const preservedId = idMatch2?.[1]

		expect(preservedId).toBe(originalId)

		editor.destroy()
	})

	it('should handle items with identical text correctly', () => {
		const editor = createEditor()

		editor.commands.setContent(`
			<ul data-type="taskList">
				<li data-checked="false"><p>Duplicate</p></li>
				<li data-checked="false"><p>Duplicate</p></li>
				<li data-checked="false"><p>Duplicate</p></li>
			</ul>
		`)

		const html = editor.getHTML()
		const idMatches = html.match(/data-task-id="([^"]+)"/g)

		expect(idMatches).toHaveLength(3)

		// Even with identical text, IDs should be unique
		const ids = idMatches!.map(match => match.match(/data-task-id="([^"]+)"/)?.[1])
		const uniqueIds = new Set(ids)
		expect(uniqueIds.size).toBe(3)

		editor.destroy()
	})
})
```

**Step 2: Run the tests to verify they fail (TDD red phase)**

Run: `cd frontend && pnpm test:unit taskItemWithId.test.ts`
Expected: Tests PASS (extension is already implemented)

**Step 3: Commit**

```bash
git add frontend/src/components/input/editor/taskItemWithId.test.ts
git commit -m "$(cat <<'EOF'
test(editor): add unit tests for TaskItemWithId extension

Tests verify:
- Unique ID generation for new items
- ID preservation when parsing existing HTML
- Different IDs for different items
- Round-trip preservation through getHTML/setContent
- Correct handling of items with identical text
EOF
)"
```

---

## Task 4: Update TipTap.vue to Use TaskItemWithId

**Files:**
- Modify: `frontend/src/components/input/editor/TipTap.vue:163` (import)
- Modify: `frontend/src/components/input/editor/TipTap.vue:470-497` (extension config)

**Step 1: Update import statement**

In `frontend/src/components/input/editor/TipTap.vue`, change line 163 from:

```typescript
import {TaskItem, TaskList} from '@tiptap/extension-list'
```

to:

```typescript
import {TaskList} from '@tiptap/extension-list'
import {TaskItemWithId} from './taskItemWithId'
```

**Step 2: Replace TaskItem with TaskItemWithId and fix onReadOnlyChecked**

In `frontend/src/components/input/editor/TipTap.vue`, replace lines 470-497 (the TaskList and TaskItem configuration) with:

```typescript
	TaskList,
	TaskItemWithId.configure({
		nested: true,
		onReadOnlyChecked: (node: Node, checked: boolean): boolean => {
			if (!props.isEditEnabled) {
				return false
			}

			// Use taskId attribute to reliably find the correct node
			// This fixes GitHub issues #293 and #563
			const targetTaskId = node.attrs.taskId

			if (!targetTaskId) {
				// Fallback to original behavior if no ID (shouldn't happen)
				console.warn('TaskItem missing taskId, falling back to node comparison')
				editor.value!.state.doc.descendants((subnode, pos) => {
					if (subnode === node) {
						const {tr} = editor.value!.state
						tr.setNodeMarkup(pos, undefined, {
							...node.attrs,
							checked,
						})
						editor.value!.view.dispatch(tr)
						bubbleSave()
					}
				})
				return true
			}

			// Find node by taskId for reliable matching
			editor.value!.state.doc.descendants((subnode, pos) => {
				if (subnode.type.name === 'taskItem' && subnode.attrs.taskId === targetTaskId) {
					const {tr} = editor.value!.state
					tr.setNodeMarkup(pos, undefined, {
						...subnode.attrs,
						checked,
					})
					editor.value!.view.dispatch(tr)
					bubbleSave()
					return false // Stop iteration once found
				}
			})

			return true
		},
	}),
```

**Step 3: Verify no TypeScript errors**

Run: `cd frontend && pnpm typecheck`
Expected: No errors

**Step 4: Run linter**

Run: `cd frontend && pnpm lint --fix`
Expected: No errors (or auto-fixed)

**Step 5: Commit**

```bash
git add frontend/src/components/input/editor/TipTap.vue
git commit -m "$(cat <<'EOF'
fix(editor): use TaskItemWithId for reliable checkbox toggling

Replace TaskItem with TaskItemWithId and update onReadOnlyChecked
to match nodes by taskId instead of object reference. This ensures
checkbox state persists correctly after page refresh.

Fixes: go-vikunja/vikunja#293
Fixes: go-vikunja/vikunja#563
EOF
)"
```

---

## Task 5: Update checklistFromText.ts to Handle IDs (Optional Enhancement)

**Files:**
- Modify: `frontend/src/helpers/checklistFromText.ts`
- Modify: `frontend/src/helpers/checklistFromText.test.ts`

**Step 1: Update checklistFromText to extract task IDs**

This is optional but useful for future features. Update `frontend/src/helpers/checklistFromText.ts`:

```typescript
interface CheckboxStatistics {
	total: number
	checked: number
}

interface CheckboxInfo {
	index: number
	checked: boolean
	taskId: string | null
}

interface MatchedCheckboxes {
	checked: number[]
	unchecked: number[]
}

const getCheckboxesInText = (text: string): MatchedCheckboxes => {
	const regex = /data-checked="(true|false)"/g
	let match
	const checkboxes: MatchedCheckboxes = {
		checked: [],
		unchecked: [],
	}

	while ((match = regex.exec(text)) !== null) {
		if (match[1] === 'true') {
			checkboxes.checked.push(match.index)
		} else {
			checkboxes.unchecked.push(match.index)
		}
	}

	return checkboxes
}

/**
 * Returns detailed checkbox info including task IDs.
 */
export const getCheckboxesWithIds = (text: string): CheckboxInfo[] => {
	const regex = /<li[^>]*data-checked="(true|false)"[^>]*(?:data-task-id="([^"]*)")?[^>]*>/g
	const checkboxes: CheckboxInfo[] = []
	let match

	while ((match = regex.exec(text)) !== null) {
		checkboxes.push({
			index: match.index,
			checked: match[1] === 'true',
			taskId: match[2] || null,
		})
	}

	return checkboxes
}

/**
 * Returns the indices where checkboxes start and end in the given text.
 *
 * @param text
 */
export const findCheckboxesInText = (text: string): number[] => {
	const checkboxes = getCheckboxesInText(text)

	return [
		...checkboxes.checked,
		...checkboxes.unchecked,
	].sort((a, b) => a - b)
}

export const getChecklistStatistics = (text: string): CheckboxStatistics => {
	const checkboxes = getCheckboxesInText(text)

	return {
		total: checkboxes.checked.length + checkboxes.unchecked.length,
		checked: checkboxes.checked.length,
	}
}
```

**Step 2: Add tests for new function**

Add to `frontend/src/helpers/checklistFromText.test.ts`:

```typescript
import {describe, it, expect} from 'vitest'

import {findCheckboxesInText, getChecklistStatistics, getCheckboxesWithIds} from './checklistFromText'

// ... existing tests ...

describe('Get Checkboxes With IDs', () => {
	it('should extract checkbox info with task IDs', () => {
		const text = `
<ul data-type="taskList">
	<li data-checked="false" data-task-id="abc123"><p>Task 1</p></li>
	<li data-checked="true" data-task-id="def456"><p>Task 2</p></li>
</ul>`
		const checkboxes = getCheckboxesWithIds(text)

		expect(checkboxes).toHaveLength(2)
		expect(checkboxes[0].checked).toBe(false)
		expect(checkboxes[0].taskId).toBe('abc123')
		expect(checkboxes[1].checked).toBe(true)
		expect(checkboxes[1].taskId).toBe('def456')
	})

	it('should handle checkboxes without task IDs', () => {
		const text = `
<ul data-type="taskList">
	<li data-checked="false"><p>Legacy task</p></li>
</ul>`
		const checkboxes = getCheckboxesWithIds(text)

		expect(checkboxes).toHaveLength(1)
		expect(checkboxes[0].taskId).toBe(null)
	})
})
```

**Step 3: Run tests**

Run: `cd frontend && pnpm test:unit checklistFromText.test.ts`
Expected: All tests PASS

**Step 4: Commit**

```bash
git add frontend/src/helpers/checklistFromText.ts frontend/src/helpers/checklistFromText.test.ts
git commit -m "$(cat <<'EOF'
feat(helpers): add getCheckboxesWithIds function

Extracts checkbox info including task IDs from HTML text.
Maintains backward compatibility with existing functions.
EOF
)"
```

---

## Task 6: Manual Integration Testing

**Files:**
- None (manual testing)

**Step 1: Start development servers**

*Dev servers are already running, no need to start them explicitely*

**Step 2: Test basic checkbox persistence**

1. Create a new task
2. Add description with checkboxes using toolbar button
3. Add items: "First item", "Second item", "Third item"
4. Save description
5. Check "First item" checkbox
6. Verify "Gespeichert!" appears
7. Refresh page (Ctrl+R)
8. Verify "First item" is still checked
9. Verify counter shows "1 von 3 Aufgaben"

Expected: Checkbox state persists after refresh

**Step 3: Test uncheck persistence**

1. Uncheck "First item"
2. Verify "Gespeichert!" appears
3. Refresh page
4. Verify "First item" is unchecked
5. Verify counter shows "0 von 3 Aufgaben"

Expected: Unchecked state persists after refresh

**Step 4: Test duplicate items (Issue #563)**

1. Edit description
2. Add three items all named "Duplicate"
3. Save
4. Check only the FIRST "Duplicate" item
5. Verify only the first one is checked (not all three)
6. Refresh page
7. Verify only the first one remains checked

Expected: Only the clicked item changes state, not all duplicates

**Step 5: Test rapid toggling**

1. Rapidly toggle a checkbox 5 times
2. Note the final state
3. Refresh page
4. Verify final state matches

Expected: Final state is correctly persisted

**Step 6: Commit test results**

If all tests pass, no commit needed. If issues found, fix and re-test.

---

## Task 7: Run Full Test Suite

**Files:**
- None (running existing tests)

**Step 1: Run frontend linting**

Run: `cd frontend && pnpm lint`
Expected: No errors

**Step 2: Run frontend type checking**

Run: `cd frontend && pnpm typecheck`
Expected: No errors

**Step 3: Run frontend unit tests**

Run: `cd frontend && pnpm test:unit`
Expected: All tests pass

**Step 4: Final commit (if any lint fixes needed)**

```bash
git add -A
git commit -m "$(cat <<'EOF'
chore: lint fixes
EOF
)"
```

---

## Summary

After completing all tasks, the checkbox persistence bug should be fixed. The key changes are:

1. **New `TaskItemWithId` extension** - Adds unique `data-task-id` attribute to each task item
2. **Updated `onReadOnlyChecked` handler** - Uses taskId to find the correct node instead of object reference
3. **Backward compatible** - Existing descriptions without IDs will get IDs generated on first edit

The fix addresses:
- **Issue #293**: Checkboxes not persisting after page refresh
- **Issue #563**: Items with same text being linked together
