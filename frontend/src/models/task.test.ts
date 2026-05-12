import {describe, it, expect} from 'vitest'
import {getTaskIdentifier} from './task'
import type {ITask} from '@/modelTypes/ITask'

function makeTask(overrides: Partial<ITask> = {}): ITask {
	// Test focuses on getTaskIdentifier which only reads id, index, identifier.
	// Cast to ITask to avoid enumerating all unrelated required fields.
	return {
		id: 42,
		index: 5,
		identifier: 'TAL-5',
		title: 'Some task',
		...overrides,
	} as ITask
}

describe('getTaskIdentifier', () => {
	it('returns empty string for null task', () => {
		expect(getTaskIdentifier(null, '{identifier}')).toBe('')
	})

	it('returns empty string for undefined task', () => {
		expect(getTaskIdentifier(undefined, '{identifier}')).toBe('')
	})

	describe('with default {identifier} format', () => {
		it('returns prefix-index when project has an identifier', () => {
			const task = makeTask({identifier: 'TAL-5', index: 5})
			expect(getTaskIdentifier(task, '{identifier}')).toBe('TAL-5')
		})

		it('returns #index when project has no identifier prefix', () => {
			const task = makeTask({identifier: '', index: 5})
			expect(getTaskIdentifier(task, '{identifier}')).toBe('#5')
		})
	})

	describe('with #{id} format', () => {
		it('returns the global database id', () => {
			const task = makeTask({id: 42})
			expect(getTaskIdentifier(task, '#{id}')).toBe('#42')
		})

		it('returns the same value regardless of prefix presence', () => {
			const withPrefix = makeTask({id: 42, identifier: 'TAL-5'})
			const withoutPrefix = makeTask({id: 42, identifier: '', index: 5})
			expect(getTaskIdentifier(withPrefix, '#{id}')).toBe('#42')
			expect(getTaskIdentifier(withoutPrefix, '#{id}')).toBe('#42')
		})
	})

	describe('with combined {identifier} (#{id}) format', () => {
		it('renders both with prefix', () => {
			const task = makeTask({id: 42, identifier: 'TAL-5', index: 5})
			expect(getTaskIdentifier(task, '{identifier} (#{id})')).toBe('TAL-5 (#42)')
		})

		it('renders both without prefix', () => {
			const task = makeTask({id: 42, identifier: '', index: 5})
			expect(getTaskIdentifier(task, '{identifier} (#{id})')).toBe('#5 (#42)')
		})
	})

	describe('with raw {prefix}-{index} format (footgun)', () => {
		it('renders prefix-index when prefix is set', () => {
			const task = makeTask({identifier: 'TAL-5', index: 5})
			expect(getTaskIdentifier(task, '{prefix}-{index}')).toBe('TAL-5')
		})

		it('renders -N (broken) when no prefix is set -- documented user-responsibility caveat', () => {
			const task = makeTask({identifier: '', index: 5})
			expect(getTaskIdentifier(task, '{prefix}-{index}')).toBe('-5')
		})
	})

	describe('placeholder edge cases', () => {
		it('returns the literal string when format has no placeholders', () => {
			const task = makeTask()
			expect(getTaskIdentifier(task, 'task')).toBe('task')
		})

		it('leaves unknown placeholders intact', () => {
			const task = makeTask()
			expect(getTaskIdentifier(task, '{xyz}-{id}')).toBe('{xyz}-42')
		})

		it('substitutes all placeholders independently', () => {
			const task = makeTask({id: 42, identifier: 'TAL-5', index: 5})
			expect(getTaskIdentifier(task, '{prefix}/{index}/#{id}')).toBe('TAL/5/#42')
		})

		it('handles repeated placeholders', () => {
			const task = makeTask({id: 42})
			expect(getTaskIdentifier(task, '{id}-{id}')).toBe('42-42')
		})
	})
})
