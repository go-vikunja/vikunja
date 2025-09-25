# Fix Duplicate Project Creation Implementation Plan

## Overview

Currently, clicking the "Create" button multiple times in rapid succession creates multiple projects before the loading state prevents further clicks. This happens because there's a 100ms timeout before the loading state is applied, creating a window where duplicate submissions can occur.

## Current State Analysis

### Problem Identification:
- **File**: `frontend/src/views/project/NewProject.vue:99`
- **Issue**: No immediate loading state set when button is clicked
- **File**: `frontend/src/components/misc/CreateEdit.vue:40`
- **Current State**: Button disables when `primaryDisabled || loading` but `loading` prop is not passed from NewProject
- **File**: `frontend/src/stores/helper.ts:4`
- **Root Cause**: 100ms timeout before loading state activates allows rapid clicks

### Current Flow:
1. User clicks "Create" button (`CreateEdit.vue:42`)
2. `createNewProject()` called (`NewProject.vue:99`)
3. `projectStore.createProject(project)` called (`NewProject.vue:99`)
4. Loading state set after 100ms timeout (`stores/helper.ts:4`)
5. Multiple clicks possible during the 100ms window

### Key Discoveries:
- CreateEdit component already supports `loading` prop but NewProject doesn't use it (`CreateEdit.vue:40,63`)
- Store loading management has intentional 100ms delay to prevent flicker (`stores/helper.ts:1`)
- Button is only disabled by `primaryDisabled` (empty title), not loading state (`NewProject.vue:4`)
- Existing patterns show v-model:loading is not currently used in the codebase

## Desired End State

After implementation:
- Clicking "Create" button immediately disables it and shows loading state
- No duplicate projects can be created via rapid clicking
- CreateEdit component supports v-model:loading for two-way binding
- Existing loading behavior and 100ms timeout preserved for UI smoothness
- Solution is general and can be applied to other creation flows

### Verification Criteria:
- Rapid clicking create button only creates one project
- Loading state shows immediately on button click
- Button remains disabled until operation completes or fails
- Modal closes and redirects on successful creation

## What We're NOT Doing

- Changing the existing 100ms loading timeout behavior in stores
- Modifying the API layer or backend validation
- Adding server-side duplicate prevention 
- Changing the redirect behavior after creation
- Modifying other creation flows initially (scope limited to project creation)

## Implementation Approach

Use immediate loading state management at the component level while preserving the existing store-level loading behavior. Add v-model:loading support to CreateEdit component for two-way binding.

## Phase 1: Add v-model:loading Support to CreateEdit Component

### Overview
Enhance CreateEdit component to support two-way loading state binding, allowing parent components to control and receive loading state updates.

### Changes Required:

#### 1. CreateEdit Component Enhancement
**File**: `frontend/src/components/misc/CreateEdit.vue`
**Changes**: Add v-model:loading support with defineModel

```vue
<script setup lang="ts">
import type {IconProp} from '@fortawesome/fontawesome-svg-core'

// Add defineModel for loading state
const loading = defineModel<boolean>('loading', { default: false })

withDefaults(defineProps<{
	title: string,
	primaryLabel?: string,
	primaryIcon?: IconProp,
	primaryDisabled?: boolean,
	hasPrimaryAction?: boolean,
	tertiary?: string,
	wide?: boolean,
	// Remove loading from props since it's now a model
}>(), {
	primaryLabel: '',
	primaryIcon: 'plus',
	primaryDisabled: false,
	hasPrimaryAction: true,
	tertiary: '',
	wide: false,
})

const emit = defineEmits<{
	'create': [event: MouseEvent],
	'primary': [event: MouseEvent],
	'tertiary': [event: MouseEvent]
}>()

function primary(event: MouseEvent) {
	emit('create', event)
	emit('primary', event)
}
</script>
```

**Template Changes**: No changes needed - `loading` model will work with existing template usage

### Success Criteria:

#### Automated Verification:
- [ ] Frontend type checking passes: `cd frontend && pnpm typecheck`
- [ ] Frontend linting passes: `cd frontend && pnpm lint`

#### Manual Verification:
- [ ] CreateEdit component accepts v-model:loading prop
- [ ] Loading state properly disables primary button
- [ ] Two-way binding works (parent can read/write loading state)

---

## Phase 2: Implement Immediate Loading State in NewProject

### Overview
Modify NewProject component to manage immediate loading state and use v-model:loading with CreateEdit component.

### Changes Required:

#### 1. NewProject Component Loading Management
**File**: `frontend/src/views/project/NewProject.vue`
**Changes**: Add immediate loading state management

```vue
<template>
	<CreateEdit
		:title="$t('project.create.header')"
		:primary-disabled="project.title === ''"
		v-model:loading="isLoading"
		@create="createNewProject()"
	>
		<!-- Form fields remain unchanged -->
	</CreateEdit>
</template>

<script setup lang="ts">
import {ref, reactive, shallowReactive, watch} from 'vue'
import {useI18n} from 'vue-i18n'

// ... existing imports ...

// Add immediate loading state
const isLoading = ref(false)

// ... existing code ...

async function createNewProject() {
	if (project.title === '') {
		showError.value = true
		return
	}
	
	// Prevent multiple submissions
	if (isLoading.value) {
		return
	}
	
	showError.value = false
	
	// Set loading immediately to prevent duplicate clicks
	isLoading.value = true
	
	try {
		if (parentProject.value) {
			project.parentProjectId = parentProject.value.id
		}

		await projectStore.createProject(project)
		success({message: t('project.create.createdSuccess')})
		// Loading will be cleared when component unmounts due to redirect
	} catch (error) {
		// Clear loading on error so user can retry
		isLoading.value = false
		throw error
	}
}
</script>
```

### Success Criteria:

#### Automated Verification:
- [ ] Frontend type checking passes: `cd frontend && pnpm typecheck`
- [ ] Frontend linting passes: `cd frontend && pnpm lint`
- [ ] Frontend unit tests pass: `cd frontend && pnpm test:unit`

#### Manual Verification:
- [ ] Button disables immediately when clicked
- [ ] Loading spinner appears immediately
- [ ] Rapid clicking only creates one project
- [ ] Error states properly reset loading
- [ ] Successful creation redirects as before

---

## Phase 3: Test Edge Cases and Error Handling

### Overview
Ensure the loading state management works correctly in all scenarios including errors, network failures, and edge cases.

### Changes Required:

#### 1. Error Handling Verification
**Files**: No code changes required
**Testing**: Manual verification of error scenarios

#### 2. Integration Testing
**Files**: No additional changes
**Testing**: End-to-end testing of creation flow

### Success Criteria:

#### Automated Verification:
- [ ] Frontend E2E tests pass: `cd frontend && pnpm test:e2e`
- [ ] All linting and type checking passes

#### Manual Verification:
- [ ] Network timeout scenarios handled correctly
- [ ] API error responses clear loading state
- [ ] Keyboard navigation (Enter/Escape) works correctly
- [ ] Loading state clears on modal close/cancel
- [ ] Parent project selection works with loading states
- [ ] Form validation errors don't affect loading state
- [ ] No loading state stuck in true after successful creation

---

## Testing Strategy

### Unit Tests:
- Test CreateEdit component with v-model:loading
- Test immediate loading state setting in NewProject
- Test loading state clearing on errors

### Integration Tests:
- Test complete project creation flow
- Test rapid clicking prevention
- Test error handling with loading states

### Manual Testing Steps:
1. Open project creation modal
2. Fill in valid project title
3. Click "Create" button rapidly (10+ clicks within 1 second)
4. Verify only one project is created
5. Test with invalid/empty title
6. Test with network errors (disable network briefly)
7. Test keyboard shortcuts (Enter to submit, Escape to cancel)
8. Test parent project selection with loading states

## Performance Considerations

- Immediate loading state has no performance impact
- Existing 100ms timeout in store preserved for UI smoothness
- Two-way binding adds minimal overhead
- No additional API calls or network requests

## Migration Notes

No migration required - this is a pure frontend enhancement that maintains backward compatibility with existing CreateEdit usage patterns.

## References

- Current CreateEdit component: `frontend/src/components/misc/CreateEdit.vue:40`
- NewProject implementation: `frontend/src/views/project/NewProject.vue:99`
- Store loading helper: `frontend/src/stores/helper.ts:3`
- Similar patterns in codebase: `frontend/src/components/tasks/AddTask.vue:36` (loading state with early return)