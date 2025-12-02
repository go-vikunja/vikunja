# Customizable Keyboard Shortcuts - Implementation Plan

**Date:** 2025-11-27
**Feature:** Allow users to customize keyboard shortcuts for actions in the Vikunja frontend

## Overview

This plan outlines the implementation of customizable keyboard shortcuts for Vikunja. Users will be able to customize action shortcuts (task operations, general app shortcuts) while keeping navigation shortcuts (j/k, g+key sequences) fixed. Customizations will be stored in the existing `frontendSettings` system and sync across devices.

## Requirements Summary

- **Scope:** Only action shortcuts customizable (not navigation keys like j/k or g+letter sequences)
- **Location:** Dedicated section in user settings page
- **Storage:** `frontendSettings` in auth store (syncs via backend)
- **Conflicts:** Prevent conflicts with validation and clear error messages
- **Reset:** Individual shortcut reset, category reset, and reset all
- **Display:** Show all shortcuts with non-customizable ones displayed as disabled

## Architecture Overview

### Core Component: ShortcutManager Composable

The centerpiece will be a new `useShortcutManager()` composable that becomes the single source of truth for all keyboard shortcuts. This manager will:

**Core Responsibilities:**
- Maintain the registry of all shortcuts (default + custom)
- Validate shortcut assignments and prevent conflicts
- Load/save custom shortcuts to `frontendSettings`
- Provide a reactive API for binding shortcuts to actions
- Handle the merging of defaults with user customizations

**Key Design Decisions:**

1. **Two-tier storage model:** Immutable defaults (from `shortcuts.ts`) and mutable overrides (from `frontendSettings.customShortcuts`)
2. **Semantic IDs:** Instead of hardcoded key strings, components register actions using IDs like `"general.toggleMenu"` or `"task.markDone"`
3. **Shared composable:** Uses VueUse's `createSharedComposable` for consistent state across all instances
4. **Reactive updates:** When settings change, all bound shortcuts update automatically without page reload

### Current State Analysis

**Current Implementation Uses Three Binding Approaches:**

1. **v-shortcut directive** - Element-bound shortcuts on buttons/links
   - Uses `@github/hotkey` library's `install()`/`uninstall()`
   - Example: `<BaseButton v-shortcut="'Mod+e'" @click="action()" />`

2. **Global keydown listeners** - App-wide shortcuts not tied to elements
   - Uses `eventToHotkeyString()` to normalize key events
   - Example: Ctrl+K for quick search, Ctrl+S for save

3. **Direct key checking** - View-specific navigation (j/k in lists)
   - Direct `e.key` checking in event handlers
   - Example: List navigation in `ProjectList.vue`

**All three approaches will be refactored** to use the ShortcutManager, ensuring consistent behavior and customization support.

## Data Model & Storage

### TypeScript Interfaces

**New Interface: `ICustomShortcut`**
```typescript
// frontend/src/modelTypes/ICustomShortcut.ts
export interface ICustomShortcut {
  actionId: string        // e.g., "task.markDone"
  keys: string[]          // e.g., ["t"] or ["Control", "s"]
  isCustomized: boolean   // true if user changed from default
}

export interface ICustomShortcutsMap {
  [actionId: string]: string[]  // Maps "task.markDone" -> ["t"]
}
```

**Update: `IFrontendSettings`**
```typescript
// frontend/src/modelTypes/IUserSettings.ts
export interface IFrontendSettings {
  // ... existing fields ...
  customShortcuts?: ICustomShortcutsMap  // New field
}
```

### Shortcut Action Registry

**Update: `shortcuts.ts`**

Add metadata to existing shortcut definitions:

```typescript
// frontend/src/components/misc/keyboard-shortcuts/shortcuts.ts

export interface ShortcutAction {
  actionId: string           // Unique ID like "general.toggleMenu"
  title: string             // i18n key for display
  keys: string[]            // Default keys
  customizable: boolean     // Can user customize this?
  contexts?: string[]       // Which routes/contexts apply
  category: ShortcutCategory
}

export enum ShortcutCategory {
  GENERAL = 'general',
  NAVIGATION = 'navigation',
  TASK_ACTIONS = 'taskActions',
  PROJECT_VIEWS = 'projectViews',
  LIST_VIEW = 'listView',
  GANTT_VIEW = 'ganttView',
}

export interface ShortcutGroup {
  title: string
  category: ShortcutCategory
  shortcuts: ShortcutAction[]
}

// Example updated structure:
export const KEYBOARD_SHORTCUTS: ShortcutGroup[] = [
  {
    title: 'keyboardShortcuts.general',
    category: ShortcutCategory.GENERAL,
    shortcuts: [
      {
        actionId: 'general.toggleMenu',
        title: 'keyboardShortcuts.toggleMenu',
        keys: ['Control', 'e'],
        customizable: true,
        contexts: ['*'],
        category: ShortcutCategory.GENERAL,
      },
      {
        actionId: 'general.quickSearch',
        title: 'keyboardShortcuts.quickSearch',
        keys: ['Control', 'k'],
        customizable: true,
        contexts: ['*'],
        category: ShortcutCategory.GENERAL,
      },
    ],
  },
  {
    title: 'keyboardShortcuts.navigation',
    category: ShortcutCategory.NAVIGATION,
    shortcuts: [
      {
        actionId: 'navigation.goToOverview',
        title: 'keyboardShortcuts.goToOverview',
        keys: ['g', 'o'],
        customizable: false,  // Navigation shortcuts are fixed
        contexts: ['*'],
        category: ShortcutCategory.NAVIGATION,
      },
      // ... more navigation shortcuts with customizable: false
    ],
  },
  {
    title: 'keyboardShortcuts.task',
    category: ShortcutCategory.TASK_ACTIONS,
    shortcuts: [
      {
        actionId: 'task.markDone',
        title: 'keyboardShortcuts.task.done',
        keys: ['t'],
        customizable: true,
        contexts: ['/tasks/:id'],
        category: ShortcutCategory.TASK_ACTIONS,
      },
      {
        actionId: 'task.toggleFavorite',
        title: 'keyboardShortcuts.task.favorite',
        keys: ['s'],
        customizable: true,
        contexts: ['/tasks/:id'],
        category: ShortcutCategory.TASK_ACTIONS,
      },
      // ... all task shortcuts
    ],
  },
  {
    title: 'keyboardShortcuts.listView',
    category: ShortcutCategory.LIST_VIEW,
    shortcuts: [
      {
        actionId: 'listView.nextTask',
        title: 'keyboardShortcuts.list.down',
        keys: ['j'],
        customizable: false,  // List navigation is fixed
        contexts: ['/projects/:id/list'],
        category: ShortcutCategory.LIST_VIEW,
      },
      {
        actionId: 'listView.previousTask',
        title: 'keyboardShortcuts.list.up',
        keys: ['k'],
        customizable: false,
        contexts: ['/projects/:id/list'],
        category: ShortcutCategory.LIST_VIEW,
      },
      // ...
    ],
  },
]
```

**Default Values:**

```typescript
// frontend/src/models/userSettings.ts
export default class UserSettingsModel implements IUserSettings {
  // ... existing defaults ...
  frontendSettings = {
    // ... existing frontend settings ...
    customShortcuts: {} as ICustomShortcutsMap,
  }
}
```

## ShortcutManager Composable

**File:** `frontend/src/composables/useShortcutManager.ts`

### API Design

```typescript
export interface UseShortcutManager {
  // Get effective shortcut for an action (default or custom)
  getShortcut(actionId: string): string[] | null

  // Get shortcut as hotkey string for @github/hotkey
  getHotkeyString(actionId: string): string

  // Check if action is customizable
  isCustomizable(actionId: string): boolean

  // Set custom shortcut for an action
  setCustomShortcut(actionId: string, keys: string[]): Promise<ValidationResult>

  // Reset single shortcut to default
  resetShortcut(actionId: string): Promise<void>

  // Reset all shortcuts in a category
  resetCategory(category: ShortcutCategory): Promise<void>

  // Reset all shortcuts to defaults
  resetAll(): Promise<void>

  // Get all shortcuts (for settings UI)
  getAllShortcuts(): ComputedRef<ShortcutGroup[]>

  // Get all customizable shortcuts
  getCustomizableShortcuts(): ComputedRef<ShortcutAction[]>

  // Validate a shortcut assignment
  validateShortcut(actionId: string, keys: string[]): ValidationResult

  // Find conflicts for a given key combination
  findConflicts(keys: string[]): ShortcutAction[]
}

export interface ValidationResult {
  valid: boolean
  error?: string  // i18n key
  conflicts?: ShortcutAction[]
}
```

### Implementation Structure

```typescript
import { computed, readonly } from 'vue'
import { createSharedComposable } from '@vueuse/core'
import { useAuthStore } from '@/stores/auth'
import { KEYBOARD_SHORTCUTS, ShortcutCategory } from '@/components/misc/keyboard-shortcuts/shortcuts'
import type { ShortcutAction, ShortcutGroup } from '@/components/misc/keyboard-shortcuts/shortcuts'
import type { ICustomShortcutsMap, ValidationResult } from '@/modelTypes/ICustomShortcut'

export const useShortcutManager = createSharedComposable(() => {
  const authStore = useAuthStore()

  // Build flat map of all shortcuts by actionId
  const defaultShortcuts = computed<Map<string, ShortcutAction>>(() => {
    const map = new Map()
    KEYBOARD_SHORTCUTS.forEach(group => {
      group.shortcuts.forEach(shortcut => {
        map.set(shortcut.actionId, shortcut)
      })
    })
    return map
  })

  // Get custom shortcuts from settings
  const customShortcuts = computed<ICustomShortcutsMap>(() => {
    return authStore.settings.frontendSettings.customShortcuts || {}
  })

  // Effective shortcuts (merged default + custom)
  const effectiveShortcuts = computed<Map<string, string[]>>(() => {
    const map = new Map()
    defaultShortcuts.value.forEach((action, actionId) => {
      const custom = customShortcuts.value[actionId]
      map.set(actionId, custom || action.keys)
    })
    return map
  })

  function getShortcut(actionId: string): string[] | null {
    return effectiveShortcuts.value.get(actionId) || null
  }

  function getHotkeyString(actionId: string): string {
    const keys = getShortcut(actionId)
    if (!keys) return ''

    // Convert array to hotkey string format
    // ['Control', 'k'] -> 'Control+k'
    // ['g', 'o'] -> 'g o'
    return keys.join(keys.length > 1 && !isModifier(keys[0]) ? ' ' : '+')
  }

  function isCustomizable(actionId: string): boolean {
    const action = defaultShortcuts.value.get(actionId)
    return action?.customizable ?? false
  }

  function findConflicts(keys: string[], excludeActionId?: string): ShortcutAction[] {
    const conflicts: ShortcutAction[] = []
    const keysStr = keys.join('+')

    effectiveShortcuts.value.forEach((shortcutKeys, actionId) => {
      if (actionId === excludeActionId) return
      if (shortcutKeys.join('+') === keysStr) {
        const action = defaultShortcuts.value.get(actionId)
        if (action) conflicts.push(action)
      }
    })

    return conflicts
  }

  function validateShortcut(actionId: string, keys: string[]): ValidationResult {
    // Check if action exists and is customizable
    const action = defaultShortcuts.value.get(actionId)
    if (!action) {
      return { valid: false, error: 'keyboardShortcuts.errors.unknownAction' }
    }
    if (!action.customizable) {
      return { valid: false, error: 'keyboardShortcuts.errors.notCustomizable' }
    }

    // Check if keys array is valid
    if (!keys || keys.length === 0) {
      return { valid: false, error: 'keyboardShortcuts.errors.emptyShortcut' }
    }

    // Check for conflicts
    const conflicts = findConflicts(keys, actionId)
    if (conflicts.length > 0) {
      return {
        valid: false,
        error: 'keyboardShortcuts.errors.conflict',
        conflicts
      }
    }

    return { valid: true }
  }

  async function setCustomShortcut(actionId: string, keys: string[]): Promise<ValidationResult> {
    const validation = validateShortcut(actionId, keys)
    if (!validation.valid) return validation

    // Update custom shortcuts
    const updated = {
      ...customShortcuts.value,
      [actionId]: keys,
    }

    // Save to backend via auth store
    await authStore.saveUserSettings({
      frontendSettings: {
        ...authStore.settings.frontendSettings,
        customShortcuts: updated,
      },
    })

    return { valid: true }
  }

  async function resetShortcut(actionId: string): Promise<void> {
    const updated = { ...customShortcuts.value }
    delete updated[actionId]

    await authStore.saveUserSettings({
      frontendSettings: {
        ...authStore.settings.frontendSettings,
        customShortcuts: updated,
      },
    })
  }

  async function resetCategory(category: ShortcutCategory): Promise<void> {
    const actionsInCategory = Array.from(defaultShortcuts.value.values())
      .filter(action => action.category === category)
      .map(action => action.actionId)

    const updated = { ...customShortcuts.value }
    actionsInCategory.forEach(actionId => {
      delete updated[actionId]
    })

    await authStore.saveUserSettings({
      frontendSettings: {
        ...authStore.settings.frontendSettings,
        customShortcuts: updated,
      },
    })
  }

  async function resetAll(): Promise<void> {
    await authStore.saveUserSettings({
      frontendSettings: {
        ...authStore.settings.frontendSettings,
        customShortcuts: {},
      },
    })
  }

  function getAllShortcuts(): ComputedRef<ShortcutGroup[]> {
    return computed(() => {
      // Return groups with effective shortcuts applied
      return KEYBOARD_SHORTCUTS.map(group => ({
        ...group,
        shortcuts: group.shortcuts.map(shortcut => ({
          ...shortcut,
          keys: getShortcut(shortcut.actionId) || shortcut.keys,
        })),
      }))
    })
  }

  function getCustomizableShortcuts(): ComputedRef<ShortcutAction[]> {
    return computed(() => {
      return Array.from(defaultShortcuts.value.values())
        .filter(action => action.customizable)
    })
  }

  return {
    getShortcut,
    getHotkeyString,
    isCustomizable,
    setCustomShortcut,
    resetShortcut,
    resetCategory,
    resetAll,
    getAllShortcuts,
    getCustomizableShortcuts,
    validateShortcut,
    findConflicts,
  }
})

function isModifier(key: string): boolean {
  return ['Control', 'Meta', 'Shift', 'Alt'].includes(key)
}
```

## Settings UI Components

### Main Settings Page Section

**File:** `frontend/src/views/user/settings/KeyboardShortcuts.vue`

**Structure:**
```vue
<template>
  <div class="keyboard-shortcuts-settings">
    <header>
      <h2>{{ $t('user.settings.keyboardShortcuts.title') }}</h2>
      <p class="help">{{ $t('user.settings.keyboardShortcuts.description') }}</p>
      <BaseButton
        @click="resetAll"
        variant="secondary"
      >
        {{ $t('user.settings.keyboardShortcuts.resetAll') }}
      </BaseButton>
    </header>

    <!-- Group by category -->
    <section
      v-for="group in shortcutGroups"
      :key="group.category"
      class="shortcut-group"
    >
      <div class="group-header">
        <h3>{{ $t(group.title) }}</h3>
        <BaseButton
          v-if="hasCustomizableInGroup(group)"
          @click="resetCategory(group.category)"
          variant="tertiary"
          size="small"
        >
          {{ $t('user.settings.keyboardShortcuts.resetCategory') }}
        </BaseButton>
      </div>

      <div class="shortcuts-list">
        <ShortcutEditor
          v-for="shortcut in group.shortcuts"
          :key="shortcut.actionId"
          :shortcut="shortcut"
          @update="updateShortcut"
          @reset="resetShortcut"
        />
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useShortcutManager } from '@/composables/useShortcutManager'
import ShortcutEditor from '@/components/misc/keyboard-shortcuts/ShortcutEditor.vue'

const { t } = useI18n()
const shortcutManager = useShortcutManager()

const shortcutGroups = shortcutManager.getAllShortcuts()

function hasCustomizableInGroup(group) {
  return group.shortcuts.some(s => s.customizable)
}

async function updateShortcut(actionId: string, keys: string[]) {
  const result = await shortcutManager.setCustomShortcut(actionId, keys)
  if (!result.valid) {
    // Show error notification
    console.error(result.error, result.conflicts)
  }
}

async function resetShortcut(actionId: string) {
  await shortcutManager.resetShortcut(actionId)
}

async function resetCategory(category: ShortcutCategory) {
  await shortcutManager.resetCategory(category)
}

async function resetAll() {
  if (confirm(t('user.settings.keyboardShortcuts.resetAllConfirm'))) {
    await shortcutManager.resetAll()
  }
}
</script>
```

### Shortcut Editor Component

**File:** `frontend/src/components/misc/keyboard-shortcuts/ShortcutEditor.vue`

**Features:**
- Display current shortcut with visual keys
- Edit mode with key capture
- Validation with conflict detection
- Reset button for customized shortcuts
- Disabled state for non-customizable shortcuts

**Structure:**
```vue
<template>
  <div
    class="shortcut-editor"
    :class="{ 'is-disabled': !shortcut.customizable, 'is-editing': isEditing }"
  >
    <div class="shortcut-info">
      <label>{{ $t(shortcut.title) }}</label>
      <span v-if="!shortcut.customizable" class="tag is-light">
        {{ $t('keyboardShortcuts.fixed') }}
      </span>
    </div>

    <div class="shortcut-input">
      <div v-if="!isEditing" class="shortcut-display">
        <Shortcut :keys="displayKeys" />
        <BaseButton
          v-if="shortcut.customizable"
          @click="startEditing"
          size="small"
          variant="tertiary"
        >
          {{ $t('misc.edit') }}
        </BaseButton>
      </div>

      <div v-else class="shortcut-edit">
        <input
          ref="captureInput"
          type="text"
          readonly
          :value="captureDisplay"
          :placeholder="$t('keyboardShortcuts.pressKeys')"
          @keydown.prevent="captureKey"
          @blur="cancelEditing"
          class="key-capture-input"
        />
        <BaseButton
          @click="saveShortcut"
          size="small"
          :disabled="!capturedKeys.length"
        >
          {{ $t('misc.save') }}
        </BaseButton>
        <BaseButton
          @click="cancelEditing"
          size="small"
          variant="tertiary"
        >
          {{ $t('misc.cancel') }}
        </BaseButton>
      </div>

      <BaseButton
        v-if="isCustomized && !isEditing"
        @click="resetToDefault"
        size="small"
        variant="tertiary"
        :title="$t('keyboardShortcuts.resetToDefault')"
      >
        <icon icon="undo" />
      </BaseButton>
    </div>

    <p v-if="validationError" class="help is-danger">
      {{ $t(validationError) }}
      <span v-if="conflicts.length">
        {{ conflicts.map(c => $t(c.title)).join(', ') }}
      </span>
    </p>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick } from 'vue'
import { useShortcutManager } from '@/composables/useShortcutManager'
import { eventToHotkeyString } from '@github/hotkey'
import Shortcut from './Shortcut.vue'
import type { ShortcutAction } from './shortcuts'

const props = defineProps<{
  shortcut: ShortcutAction
}>()

const emit = defineEmits<{
  update: [actionId: string, keys: string[]]
  reset: [actionId: string]
}>()

const shortcutManager = useShortcutManager()

const isEditing = ref(false)
const capturedKeys = ref<string[]>([])
const validationError = ref<string | null>(null)
const conflicts = ref<ShortcutAction[]>([])
const captureInput = ref<HTMLInputElement>()

const displayKeys = computed(() => {
  return shortcutManager.getShortcut(props.shortcut.actionId) || props.shortcut.keys
})

const isCustomized = computed(() => {
  const current = shortcutManager.getShortcut(props.shortcut.actionId)
  return JSON.stringify(current) !== JSON.stringify(props.shortcut.keys)
})

const captureDisplay = computed(() => {
  return capturedKeys.value.join(' + ')
})

async function startEditing() {
  isEditing.value = true
  capturedKeys.value = []
  validationError.value = null
  conflicts.value = []
  await nextTick()
  captureInput.value?.focus()
}

function captureKey(event: KeyboardEvent) {
  event.preventDefault()

  const hotkeyString = eventToHotkeyString(event)
  if (!hotkeyString) return

  // Parse hotkey string into keys array
  const keys = hotkeyString.includes('+')
    ? hotkeyString.split('+')
    : [hotkeyString]

  capturedKeys.value = keys

  // Validate in real-time
  const validation = shortcutManager.validateShortcut(props.shortcut.actionId, keys)
  if (!validation.valid) {
    validationError.value = validation.error || null
    conflicts.value = validation.conflicts || []
  } else {
    validationError.value = null
    conflicts.value = []
  }
}

function saveShortcut() {
  if (!capturedKeys.value.length) return

  const validation = shortcutManager.validateShortcut(props.shortcut.actionId, capturedKeys.value)
  if (!validation.valid) {
    validationError.value = validation.error || null
    conflicts.value = validation.conflicts || []
    return
  }

  emit('update', props.shortcut.actionId, capturedKeys.value)
  isEditing.value = false
  capturedKeys.value = []
}

function cancelEditing() {
  isEditing.value = false
  capturedKeys.value = []
  validationError.value = null
  conflicts.value = []
}

function resetToDefault() {
  emit('reset', props.shortcut.actionId)
}
</script>

<style scoped>
.shortcut-editor {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem;
  border-bottom: 1px solid var(--grey-200);
}

.shortcut-editor.is-disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.shortcut-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.shortcut-input {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.shortcut-display {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.shortcut-edit {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.key-capture-input {
  min-width: 200px;
  padding: 0.5rem;
  border: 2px solid var(--primary);
  border-radius: 4px;
  font-family: monospace;
  text-align: center;
}

.help.is-danger {
  color: var(--danger);
  font-size: 0.875rem;
  margin-top: 0.25rem;
}
</style>
```

### Update Settings Navigation

**File:** `frontend/src/views/user/settings/index.vue`

Add new route and navigation item for keyboard shortcuts settings.

## Migration Strategy

### Phase 1: Add Infrastructure (No Breaking Changes)

1. **Add new TypeScript interfaces**
   - `ICustomShortcut`, `ICustomShortcutsMap`
   - Update `IFrontendSettings`

2. **Update `shortcuts.ts` with metadata**
   - Add `actionId`, `customizable`, `category`, `contexts` to all shortcuts
   - Keep existing structure, only add fields

3. **Create `useShortcutManager` composable**
   - Implement all API methods
   - Test in isolation

4. **Build settings UI components**
   - `KeyboardShortcuts.vue` settings page
   - `ShortcutEditor.vue` component
   - Add to settings navigation

**Verification:** Settings UI works, can customize and persist shortcuts, but existing code still uses hardcoded shortcuts.

### Phase 2: Refactor Shortcut Bindings

Refactor components one category at a time to use the manager:

#### 2.1 Update v-shortcut Directive

**File:** `frontend/src/directives/shortcut.ts`

```typescript
import { install, uninstall } from '@github/hotkey'
import { useShortcutManager } from '@/composables/useShortcutManager'
import type { Directive } from 'vue'

const directive = <Directive<HTMLElement, string>>{
  mounted(el, { value }) {
    if (value === '') return

    // Support both old format (direct keys) and new format (actionId)
    const shortcutManager = useShortcutManager()
    const hotkeyString = value.startsWith('.')
      ? shortcutManager.getHotkeyString(value)  // New format: actionId
      : value                                    // Old format: direct keys (backwards compat)

    if (!hotkeyString) return

    install(el, hotkeyString)

    // Store for cleanup
    el.dataset.shortcutActionId = value
  },
  updated(el, { value, oldValue }) {
    if (value === oldValue) return

    // Reinstall with new shortcut
    uninstall(el)

    const shortcutManager = useShortcutManager()
    const hotkeyString = value.startsWith('.')
      ? shortcutManager.getHotkeyString(value)
      : value

    if (!hotkeyString) return
    install(el, hotkeyString)
  },
  beforeUnmount(el) {
    uninstall(el)
  },
}

export default directive
```

**Usage migration:**
```vue
<!-- Old -->
<BaseButton v-shortcut="'Mod+e'" @click="toggleMenu()" />

<!-- New -->
<BaseButton v-shortcut="'general.toggleMenu'" @click="toggleMenu()" />
```

#### 2.2 Refactor General Shortcuts

**Files to update:**
- `frontend/src/components/home/MenuButton.vue` - Menu toggle (Ctrl+E)
- `frontend/src/components/misc/OpenQuickActions.vue` - Quick search (Ctrl+K)
- `frontend/src/components/home/ContentAuth.vue` - Help modal (Shift+?)

**Pattern:**
```typescript
// Old
function handleShortcut(event) {
  const hotkeyString = eventToHotkeyString(event)
  if (hotkeyString !== 'Control+k') return
  event.preventDefault()
  action()
}

// New
import { useShortcutManager } from '@/composables/useShortcutManager'

const shortcutManager = useShortcutManager()

function handleShortcut(event) {
  const hotkeyString = eventToHotkeyString(event)
  const expectedHotkey = shortcutManager.getHotkeyString('general.quickSearch')
  if (hotkeyString !== expectedHotkey) return
  event.preventDefault()
  action()
}
```

#### 2.3 Refactor Navigation Shortcuts

**Files to update:**
- `frontend/src/components/home/Navigation.vue` - All g+key sequences

**Pattern:**
```vue
<!-- Old -->
<RouterLink v-shortcut="'g o'" :to="{ name: 'home' }">

<!-- New -->
<RouterLink v-shortcut="'navigation.goToOverview'" :to="{ name: 'home' }">
```

#### 2.4 Refactor Task Detail Shortcuts

**File:** `frontend/src/views/tasks/TaskDetailView.vue`

Update all 14 task shortcuts to use actionIds through the directive:

```vue
<!-- Old -->
<XButton v-shortcut="'t'" @click="toggleTaskDone()">

<!-- New -->
<XButton v-shortcut="'task.markDone'" @click="toggleTaskDone()">
```

For the save shortcut (global listener):
```typescript
// Old
function saveTaskViaHotkey(event) {
  const hotkeyString = eventToHotkeyString(event)
  if (hotkeyString !== 'Control+s' && hotkeyString !== 'Meta+s') return
  // ...
}

// New
const shortcutManager = useShortcutManager()

function saveTaskViaHotkey(event) {
  const hotkeyString = eventToHotkeyString(event)
  const expectedHotkey = shortcutManager.getHotkeyString('task.save')
  if (hotkeyString !== expectedHotkey) return
  // ...
}
```

#### 2.5 List View Navigation (Keep Fixed)

**File:** `frontend/src/components/project/views/ProjectList.vue`

Keep j/k/Enter as hardcoded since they're non-customizable, but add them to the registry for documentation purposes.

### Phase 3: Update Help Modal

**File:** `frontend/src/components/misc/keyboard-shortcuts/index.vue`

Update to use `shortcutManager.getAllShortcuts()` instead of the static `KEYBOARD_SHORTCUTS` constant, so the help modal always shows current effective shortcuts (including customizations).

```typescript
import { useShortcutManager } from '@/composables/useShortcutManager'

const shortcutManager = useShortcutManager()
const shortcuts = shortcutManager.getAllShortcuts()
```

Add link to settings page:
```vue
<p class="help-text">
  {{ $t('keyboardShortcuts.helpText') }}
  <RouterLink :to="{ name: 'user.settings.keyboardShortcuts' }">
    {{ $t('keyboardShortcuts.customizeShortcuts') }}
  </RouterLink>
</p>
```

### Phase 4: Testing & Cleanup

1. Remove backward compatibility from directive if all components migrated
2. Add unit tests for `useShortcutManager`
3. Add E2E tests for customization flow
4. Update documentation

## Testing Approach

### Unit Tests

**File:** `frontend/src/composables/useShortcutManager.test.ts`

Test cases:
- ✅ Returns default shortcuts when no customizations
- ✅ Returns custom shortcuts when set
- ✅ Validates conflicts correctly
- ✅ Prevents assigning shortcuts to non-customizable actions
- ✅ Reset individual/category/all works correctly
- ✅ Persists to auth store correctly

### Component Tests

**Files:**
- `ShortcutEditor.test.ts` - Test key capture, validation, save/cancel
- `KeyboardShortcuts.test.ts` - Test settings page interactions

### E2E Tests

**File:** `frontend/cypress/e2e/keyboard-shortcuts.cy.ts`

Test scenarios:
1. Navigate to settings, customize a shortcut, verify it works
2. Create conflict, verify error message prevents save
3. Reset individual shortcut, verify default restored
4. Reset all shortcuts, verify all defaults restored
5. Customize shortcut, reload page, verify persistence
6. Verify non-customizable shortcuts show as disabled

### Manual Testing Checklist

- [ ] Customize Ctrl+E (menu toggle) and verify it works
- [ ] Try to create conflict, verify error prevents save
- [ ] Customize task shortcut (t for mark done), verify in task detail
- [ ] Reset customized shortcut, verify default works again
- [ ] Reset entire category, verify all in category reset
- [ ] Reset all shortcuts, verify everything back to defaults
- [ ] Verify j/k navigation shortcuts cannot be edited
- [ ] Verify g+key navigation shortcuts cannot be edited
- [ ] Open help modal (Shift+?), verify shows customized shortcuts
- [ ] Logout/login, verify shortcuts persist
- [ ] Test on different device, verify shortcuts sync

## Translation Keys

Add to `frontend/src/i18n/lang/en.json`:

```json
{
  "user": {
    "settings": {
      "keyboardShortcuts": {
        "title": "Keyboard Shortcuts",
        "description": "Customize keyboard shortcuts for actions. Navigation shortcuts (j/k, g+keys) are fixed and cannot be changed.",
        "resetAll": "Reset All to Defaults",
        "resetAllConfirm": "Are you sure you want to reset all keyboard shortcuts to defaults?",
        "resetCategory": "Reset Category",
        "resetToDefault": "Reset to default"
      }
    }
  },
  "keyboardShortcuts": {
    "fixed": "Fixed",
    "pressKeys": "Press keys...",
    "customizeShortcuts": "Customize shortcuts",
    "helpText": "You can customize most keyboard shortcuts in settings.",
    "errors": {
      "unknownAction": "Unknown shortcut action",
      "notCustomizable": "This shortcut cannot be customized",
      "emptyShortcut": "Please press at least one key",
      "conflict": "This shortcut is already assigned to: "
    }
  }
}
```

## Implementation Checklist

### Phase 1: Infrastructure (Estimated: Core functionality)
- [ ] Create `ICustomShortcut` and `ICustomShortcutsMap` interfaces
- [ ] Update `IFrontendSettings` with `customShortcuts` field
- [ ] Update `UserSettingsModel` with default value
- [ ] Add metadata to all shortcuts in `shortcuts.ts` (`actionId`, `customizable`, `category`, `contexts`)
- [ ] Create `useShortcutManager.ts` composable with full API
- [ ] Write unit tests for `useShortcutManager`
- [ ] Create `ShortcutEditor.vue` component
- [ ] Create `KeyboardShortcuts.vue` settings page
- [ ] Add route for keyboard shortcuts settings
- [ ] Add navigation item in settings menu
- [ ] Add translation keys
- [ ] Manual test: Verify settings UI works and persists

### Phase 2: Refactor Bindings (Estimated: Progressive refactoring)
- [ ] Update `shortcut.ts` directive to support actionIds
- [ ] Refactor `MenuButton.vue` (Ctrl+E)
- [ ] Refactor `OpenQuickActions.vue` (Ctrl+K)
- [ ] Refactor `ContentAuth.vue` (Shift+?)
- [ ] Refactor `Navigation.vue` (all g+key sequences)
- [ ] Refactor `TaskDetailView.vue` (all 14 task shortcuts + Ctrl+S)
- [ ] Refactor project view switching shortcuts
- [ ] Document list navigation shortcuts (j/k) in registry (keep hardcoded)
- [ ] Manual test: Verify all refactored shortcuts work with customization

### Phase 3: Polish (Estimated: Final touches)
- [ ] Update help modal to show effective shortcuts
- [ ] Add link from help modal to settings
- [ ] Remove backward compatibility from directive (if desired)
- [ ] Write component tests for `ShortcutEditor` and `KeyboardShortcuts`
- [ ] Write E2E tests for customization flow
- [ ] Update documentation
- [ ] Full manual testing checklist

### Phase 4: Code Review & Merge
- [ ] Run frontend lints: `pnpm lint:fix && pnpm lint:styles:fix`
- [ ] Run frontend tests: `pnpm test:unit`
- [ ] Code review
- [ ] Merge to main

## Open Questions & Decisions

1. **Multi-key sequences:** Should users be able to create their own multi-key sequences (like "g p" for custom actions), or only single keys and modifier combinations?
   - **Decision:** Start with single keys + modifiers only. Can add sequences later if needed.

2. **Import/Export:** Should we add import/export functionality for sharing shortcut configurations?
   - **Decision:** Not in initial version. Can add later if users request it.

3. **Shortcut recommendations:** Should we suggest alternative shortcuts when conflicts occur?
   - **Decision:** Not in initial version. Show conflict error, user chooses different keys.

4. **Platform differences:** Mac uses Cmd while others use Ctrl. Should we allow different shortcuts per platform?
   - **Decision:** No. Use "Mod" (maps to Cmd on Mac, Ctrl elsewhere) and keep shortcuts platform-agnostic. Library already handles this.

5. **Accessibility:** Should we provide a way to disable all keyboard shortcuts for users who need screen readers?
   - **Decision:** Future enhancement. For now, shortcuts don't interfere with standard screen reader keys.

## Success Criteria

- ✅ Users can customize action shortcuts from settings page
- ✅ Navigation shortcuts (j/k, g+keys) remain fixed and clearly marked
- ✅ Conflict detection prevents duplicate shortcuts
- ✅ Individual, category, and global reset options work
- ✅ Customizations persist across sessions and devices
- ✅ Help modal reflects current effective shortcuts
- ✅ All existing shortcuts continue to work during migration
- ✅ No regressions in existing functionality
- ✅ Comprehensive test coverage (unit + E2E)
- ✅ Code follows Vikunja conventions and passes linting

## Future Enhancements

- Add import/export for shortcut configurations
- Add shortcut recommendation system for conflicts
- Allow custom multi-key sequences for advanced users
- Add keyboard shortcut recorder/tutorial for new users
- Add shortcut profiles (Vim mode, VS Code mode, etc.)
- Add analytics to track most-customized shortcuts
