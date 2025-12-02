import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import ShortcutEditor from './ShortcutEditor.vue'
import { ShortcutCategory } from './shortcuts'
import type { ShortcutAction } from './shortcuts'

// Mock the shortcut manager
const mockShortcutManager = {
	getShortcut: vi.fn((actionId: string) => {
		if (actionId === 'general.toggleMenu') return ['ctrl', 'e']
		return null
	}),
	validateShortcut: vi.fn(() => ({ valid: true })),
	isCustomized: vi.fn(() => false)
}

vi.mock('@/composables/useShortcutManager', () => ({
	useShortcutManager: () => mockShortcutManager
}))

// Mock the Shortcut component
vi.mock('@/components/misc/Shortcut.vue', () => ({
	default: {
		name: 'Shortcut',
		template: '<div class="shortcut-mock">{{ keys.join("+") }}</div>',
		props: ['keys']
	}
}))

// Mock BaseButton component
vi.mock('@/components/base/BaseButton.vue', () => ({
	default: {
		name: 'BaseButton',
		template: '<button @click="$emit(\'click\')"><slot /></button>',
		emits: ['click']
	}
}))

describe('ShortcutEditor', () => {
	const mockShortcut: ShortcutAction = {
		actionId: 'general.toggleMenu',
		title: 'keyboardShortcuts.toggleMenu',
		keys: ['ctrl', 'e'],
		customizable: true,
		contexts: ['*'],
		category: ShortcutCategory.GENERAL
	}

	let wrapper: any

	beforeEach(() => {
		// Reset mocks
		mockShortcutManager.getShortcut.mockReturnValue(['ctrl', 'e'])
		mockShortcutManager.validateShortcut.mockReturnValue({ valid: true })
		mockShortcutManager.isCustomized.mockReturnValue(false)

		wrapper = mount(ShortcutEditor, {
			props: {
				shortcut: mockShortcut
			},
			global: {
				mocks: {
					$t: (key: string) => key // Simple mock for i18n
				}
			}
		})
	})

	it('should render shortcut information', () => {
		expect(wrapper.find('.shortcut-info label').text()).toBe('keyboardShortcuts.toggleMenu')
		expect(wrapper.find('.shortcut-mock').text()).toBe('ctrl+e')
	})

	it('should show edit button for customizable shortcuts', () => {
		expect(wrapper.find('button').exists()).toBe(true)
		expect(wrapper.find('button').text()).toBe('misc.edit')
	})

	it('should not show edit button for non-customizable shortcuts', async () => {
		const nonCustomizableShortcut = {
			...mockShortcut,
			customizable: false
		}
		await wrapper.setProps({ shortcut: nonCustomizableShortcut })
		expect(wrapper.find('button').exists()).toBe(false)
		expect(wrapper.find('.tag').text()).toBe('keyboardShortcuts.fixed')
	})

	it('should enter edit mode when edit button is clicked', async () => {
		const editButton = wrapper.find('button')
		await editButton.trigger('click')

		expect(wrapper.find('.key-capture-input').exists()).toBe(true)
		expect(wrapper.find('input[placeholder="keyboardShortcuts.pressKeys"]').exists()).toBe(true)
	})

	// Simplified tests that don't rely on complex DOM manipulation
	it('should have correct initial state', () => {
		expect(wrapper.vm.isEditing).toBe(false)
		expect(wrapper.vm.capturedKeys).toEqual([])
		// validationError might be null initially
		expect(wrapper.vm.validationError).toBeFalsy()
	})

	it('should call shortcut manager methods', () => {
		// Test that the component calls the shortcut manager
		expect(mockShortcutManager.getShortcut).toHaveBeenCalledWith('general.toggleMenu')
	})
})
