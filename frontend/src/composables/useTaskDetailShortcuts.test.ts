import {afterEach, describe, expect, it, vi} from 'vitest'
import {enableAutoUnmount, mount} from '@vue/test-utils'
import type {Task} from '@/client/generated'
import {useTaskDetailShortcuts} from './useTaskDetailShortcuts'

vi.mock('@/helpers/task', () => ({getTaskIdentifier: () => 'Task 1'}))

enableAutoUnmount(afterEach)

function mountShortcuts(blocked: boolean) {
	const onSave = vi.fn()
	const onDoneAndBack = vi.fn()
	mount({
		setup() {
			useTaskDetailShortcuts({
				task: () => ({id: 1}) as Task,
				taskTitle: () => 'Task 1',
				onSave,
				onDoneAndBack,
				isBlocked: () => blocked,
			})
			return () => null
		},
	})
	return {onSave, onDoneAndBack}
}

function pressKey(key: string, modifiers: KeyboardEventInit = {}) {
	const event = new KeyboardEvent('keydown', {
		key,
		bubbles: true,
		cancelable: true,
		...modifiers,
	})
	document.body.dispatchEvent(event)
	return event
}

describe('task detail shortcuts', () => {
	it.each(['ctrlKey', 'metaKey'])('prevents a blocked save with %s without dispatching it', modifier => {
		const {onSave, onDoneAndBack} = mountShortcuts(true)

		expect(pressKey('s', {[modifier]: true}).defaultPrevented).toBe(true)
		expect(onSave).not.toHaveBeenCalled()
		expect(onDoneAndBack).not.toHaveBeenCalled()
	})

	it.each(['ctrlKey', 'metaKey'])('saves and prevents the browser save with %s when unblocked', modifier => {
		const {onSave} = mountShortcuts(false)

		expect(pressKey('s', {[modifier]: true}).defaultPrevented).toBe(true)
		expect(onSave).toHaveBeenCalledOnce()
	})

	it.each([
		{key: 'x', modifiers: {}},
		{key: 's', modifiers: {altKey: true}},
		{key: 's', modifiers: {ctrlKey: true, shiftKey: true}},
		{key: 'T', modifiers: {shiftKey: true}},
	])('leaves blocked $key with $modifiers to the editor', ({key, modifiers}) => {
		const {onSave, onDoneAndBack} = mountShortcuts(true)

		expect(pressKey(key, modifiers).defaultPrevented).toBe(false)
		expect(onSave).not.toHaveBeenCalled()
		expect(onDoneAndBack).not.toHaveBeenCalled()
	})
})
