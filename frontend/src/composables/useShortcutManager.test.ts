import { describe, it, expect, beforeEach, vi } from 'vitest'
import { ShortcutCategory } from '@/components/misc/keyboard-shortcuts/shortcuts'

// Mock the auth store
const mockAuthStore = {
	settings: {
		frontendSettings: {
			customShortcuts: {}
		}
	},
	saveUserSettings: vi.fn().mockResolvedValue(undefined)
}

vi.mock('@/stores/auth', () => ({
	useAuthStore: () => mockAuthStore
}))

// Mock createSharedComposable to avoid shared state issues
vi.mock('@vueuse/core', async () => {
	const actual = await vi.importActual('@vueuse/core')
	return {
		...actual,
		createSharedComposable: (fn: any) => fn
	}
})

// Import after mocking
const { useShortcutManager } = await import('./useShortcutManager')

describe('useShortcutManager', () => {
	let shortcutManager: ReturnType<typeof useShortcutManager>

	beforeEach(() => {
		// Reset mock state
		mockAuthStore.settings.frontendSettings.customShortcuts = {}
		mockAuthStore.saveUserSettings.mockClear()
		shortcutManager = useShortcutManager()
	})

	describe('getShortcut', () => {
		it('should return default shortcut when no custom shortcut exists', () => {
			const keys = shortcutManager.getShortcut('general.toggleMenu')
			expect(keys).toEqual(['ctrl', 'e']) // Adjust based on actual implementation
		})

		it('should return custom shortcut when one exists', () => {
			// Set custom shortcut in mock store
			mockAuthStore.settings.frontendSettings.customShortcuts['general.toggleMenu'] = ['alt', 'm']

			// Create new instance to pick up the change
			const newShortcutManager = useShortcutManager()
			const keys = newShortcutManager.getShortcut('general.toggleMenu')
			expect(keys).toEqual(['alt', 'm'])
		})

		it('should return null for non-existent action', () => {
			const keys = shortcutManager.getShortcut('nonexistent.action')
			expect(keys).toBeNull()
		})
	})

	describe('getHotkeyString', () => {
		it('should convert keys array to hotkey string', () => {
			const hotkeyString = shortcutManager.getHotkeyString('general.toggleMenu')
			// The actual implementation uses spaces for sequences, + for modifiers
			expect(hotkeyString).toBe('ctrl e')
		})

		it('should handle sequence shortcuts with spaces', () => {
			const hotkeyString = shortcutManager.getHotkeyString('navigation.goToOverview')
			expect(hotkeyString).toBe('g o')
		})

		it('should return empty string for non-existent action', () => {
			const hotkeyString = shortcutManager.getHotkeyString('nonexistent.action')
			expect(hotkeyString).toBe('')
		})
	})

	describe('isCustomizable', () => {
		it('should return true for customizable shortcuts', () => {
			const customizable = shortcutManager.isCustomizable('general.toggleMenu')
			expect(customizable).toBe(true)
		})

		it('should return false for non-customizable shortcuts', () => {
			const customizable = shortcutManager.isCustomizable('navigation.goToOverview')
			expect(customizable).toBe(false)
		})

		it('should return false for non-existent shortcuts', () => {
			const customizable = shortcutManager.isCustomizable('nonexistent.action')
			expect(customizable).toBe(false)
		})
	})

	describe('validateShortcut', () => {
		it('should validate a valid shortcut', () => {
			const result = shortcutManager.validateShortcut('general.toggleMenu', ['ctrl', 'x'])
			expect(result.valid).toBe(true)
		})

		it('should reject empty keys', () => {
			const result = shortcutManager.validateShortcut('general.toggleMenu', [])
			expect(result.valid).toBe(false)
			expect(result.error).toBe('keyboardShortcuts.errors.emptyShortcut')
		})

		it('should reject non-customizable shortcuts', () => {
			const result = shortcutManager.validateShortcut('navigation.goToOverview', ['ctrl', 'x'])
			expect(result.valid).toBe(false)
			expect(result.error).toBe('keyboardShortcuts.errors.notCustomizable')
		})

		it('should reject unknown actions', () => {
			const result = shortcutManager.validateShortcut('nonexistent.action', ['ctrl', 'x'])
			expect(result.valid).toBe(false)
			expect(result.error).toBe('keyboardShortcuts.errors.unknownAction')
		})
	})

	describe('findConflicts', () => {
		it('should find conflicts with existing shortcuts', () => {
			const conflicts = shortcutManager.findConflicts(['ctrl', 'e'])
			expect(conflicts).toHaveLength(1)
			expect(conflicts[0].actionId).toBe('general.toggleMenu')
		})

		it('should exclude specified action from conflict detection', () => {
			const conflicts = shortcutManager.findConflicts(['ctrl', 'e'], 'general.toggleMenu')
			expect(conflicts).toHaveLength(0)
		})

		it('should return empty array when no conflicts exist', () => {
			const conflicts = shortcutManager.findConflicts(['ctrl', 'shift', 'z'])
			expect(conflicts).toHaveLength(0)
		})
	})

	describe('setCustomShortcut', () => {
		it('should save valid custom shortcut', async () => {
			const result = await shortcutManager.setCustomShortcut('general.toggleMenu', ['ctrl', 'x'])
			expect(result.valid).toBe(true)
			expect(mockAuthStore.saveUserSettings).toHaveBeenCalledWith({
				settings: expect.objectContaining({
					frontendSettings: expect.objectContaining({
						customShortcuts: {
							'general.toggleMenu': ['ctrl', 'x']
						}
					})
				}),
				showMessage: false
			})
		})

		it('should reject invalid shortcut', async () => {
			const result = await shortcutManager.setCustomShortcut('general.toggleMenu', [])
			expect(result.valid).toBe(false)
			expect(mockAuthStore.saveUserSettings).not.toHaveBeenCalled()
		})
	})

	describe('resetShortcut', () => {
		it('should remove custom shortcut', async () => {
			mockAuthStore.settings.frontendSettings.customShortcuts = {
				'general.toggleMenu': ['ctrl', 'x']
			}
			await shortcutManager.resetShortcut('general.toggleMenu')
			expect(mockAuthStore.saveUserSettings).toHaveBeenCalledWith({
				settings: expect.objectContaining({
					frontendSettings: expect.objectContaining({
						customShortcuts: {}
					})
				}),
				showMessage: false
			})
		})
	})

	describe('resetCategory', () => {
		it('should reset all shortcuts in a category', async () => {
			mockAuthStore.settings.frontendSettings.customShortcuts = {
				'general.toggleMenu': ['ctrl', 'x'],
				'general.quickSearch': ['ctrl', 'y'],
				'task.markDone': ['ctrl', 'z']
			}
			await shortcutManager.resetCategory(ShortcutCategory.GENERAL)

			// Check that saveUserSettings was called
			expect(mockAuthStore.saveUserSettings).toHaveBeenCalled()

			// Check that the customShortcuts object was updated correctly
			const callArgs = mockAuthStore.saveUserSettings.mock.calls[0][0]
			expect(callArgs.settings.frontendSettings.customShortcuts).toEqual({
				'task.markDone': ['ctrl', 'z'] // Only non-general shortcuts remain
			})
		})
	})

	describe('resetAll', () => {
		it('should reset all custom shortcuts', async () => {
			mockAuthStore.settings.frontendSettings.customShortcuts = {
				'general.toggleMenu': ['ctrl', 'x'],
				'task.markDone': ['ctrl', 'z']
			}
			await shortcutManager.resetAll()
			expect(mockAuthStore.saveUserSettings).toHaveBeenCalledWith({
				settings: expect.objectContaining({
					frontendSettings: expect.objectContaining({
						customShortcuts: {}
					})
				}),
				showMessage: false
			})
		})
	})
})
