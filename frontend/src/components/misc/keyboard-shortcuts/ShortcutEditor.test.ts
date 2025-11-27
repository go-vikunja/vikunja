import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import ShortcutEditor from './ShortcutEditor.vue'
import { ShortcutCategory } from './shortcuts'
import type { ShortcutAction } from './shortcuts'

// Mock the shortcut manager
vi.mock('@/composables/useShortcutManager', () => ({
	useShortcutManager: vi.fn(() => ({
		getShortcut: vi.fn((actionId: string) => {
			if (actionId === 'general.toggleMenu') return ['⌘', 'e']
			return null
		}),
		validateShortcut: vi.fn(() => ({ valid: true }))
	}))
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
		keys: ['⌘', 'e'],
		customizable: true,
		contexts: ['*'],
		category: ShortcutCategory.GENERAL
	}

	let wrapper: any

	beforeEach(() => {
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
		expect(wrapper.find('.shortcut-mock').text()).toBe('⌘+e')
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

	it('should emit update event when shortcut is saved', async () => {
		// Enter edit mode
		await wrapper.find('button').trigger('click')
		
		// Simulate key capture (this would normally be done via keydown event)
		wrapper.vm.capturedKeys = ['ctrl', 'x']
		
		// Click save button
		const saveButton = wrapper.findAll('button').find(btn => btn.text() === 'misc.save')
		await saveButton.trigger('click')
		
		expect(wrapper.emitted('update')).toBeTruthy()
		expect(wrapper.emitted('update')[0]).toEqual(['general.toggleMenu', ['ctrl', 'x']])
	})

	it('should emit reset event when reset button is clicked', async () => {
		// Mock that this shortcut is customized
		wrapper.vm.isCustomized = true
		await wrapper.vm.$nextTick()
		
		const resetButton = wrapper.find('button[title="keyboardShortcuts.resetToDefault"]')
		await resetButton.trigger('click')
		
		expect(wrapper.emitted('reset')).toBeTruthy()
		expect(wrapper.emitted('reset')[0]).toEqual(['general.toggleMenu'])
	})

	it('should show validation error for invalid shortcuts', async () => {
		// Mock validation failure
		const mockShortcutManager = vi.mocked(await import('@/composables/useShortcutManager')).useShortcutManager()
		mockShortcutManager.validateShortcut.mockReturnValue({
			valid: false,
			error: 'keyboardShortcuts.errors.conflict',
			conflicts: []
		})
		
		// Enter edit mode and try to save invalid shortcut
		await wrapper.find('button').trigger('click')
		wrapper.vm.capturedKeys = ['ctrl', 'e'] // Conflicting shortcut
		
		// Trigger validation
		wrapper.vm.captureKey({ preventDefault: vi.fn() } as any)
		await wrapper.vm.$nextTick()
		
		expect(wrapper.find('.help.is-danger').exists()).toBe(true)
		expect(wrapper.find('.help.is-danger').text()).toContain('keyboardShortcuts.errors.conflict')
	})
})
