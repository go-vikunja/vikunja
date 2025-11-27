import { computed, type ComputedRef } from 'vue'
import { createSharedComposable } from '@vueuse/core'
import { useAuthStore } from '@/stores/auth'
import { KEYBOARD_SHORTCUTS, ShortcutCategory } from '@/components/misc/keyboard-shortcuts/shortcuts'
import type { ShortcutAction, ShortcutGroup } from '@/components/misc/keyboard-shortcuts/shortcuts'
import type { ICustomShortcutsMap, ValidationResult } from '@/modelTypes/ICustomShortcut'

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
				conflicts,
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
			settings: {
				...authStore.settings,
				frontendSettings: {
					...authStore.settings.frontendSettings,
					customShortcuts: updated,
				},
			},
			showMessage: false,
		})

		return { valid: true }
	}

	async function resetShortcut(actionId: string): Promise<void> {
		const updated = { ...customShortcuts.value }
		delete updated[actionId]

		await authStore.saveUserSettings({
			settings: {
				...authStore.settings,
				frontendSettings: {
					...authStore.settings.frontendSettings,
					customShortcuts: updated,
				},
			},
			showMessage: false,
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
			settings: {
				...authStore.settings,
				frontendSettings: {
					...authStore.settings.frontendSettings,
					customShortcuts: updated,
				},
			},
			showMessage: false,
		})
	}

	async function resetAll(): Promise<void> {
		await authStore.saveUserSettings({
			settings: {
				...authStore.settings,
				frontendSettings: {
					...authStore.settings.frontendSettings,
					customShortcuts: {},
				},
			},
			showMessage: false,
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

	function isModifier(key: string): boolean {
		return ['Control', 'Meta', 'Shift', 'Alt'].includes(key)
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
