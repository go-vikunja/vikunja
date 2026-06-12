import {describe, expect, it} from 'vitest'

import {formatTaskAsMarkdown} from './taskMarkdown'

describe('formatTaskAsMarkdown', () => {
	it('formats a task with title and description', () => {
		const markdown = formatTaskAsMarkdown({
			title: 'Write release notes',
			description: '<p>Summarize the changes for v1.0</p>',
		})

		expect(markdown).toBe('# Write release notes\n\nSummarize the changes for v1.0')
	})

	it('formats a task with title only when description is empty', () => {
		const markdown = formatTaskAsMarkdown({
			title: 'Write release notes',
			description: '',
		})

		expect(markdown).toBe('# Write release notes')
	})

	it('treats TipTap empty content as no description', () => {
		const markdown = formatTaskAsMarkdown({
			title: 'Write release notes',
			description: '<p></p>',
		})

		expect(markdown).toBe('# Write release notes')
	})

	it('converts links to markdown links', () => {
		const markdown = formatTaskAsMarkdown({
			title: 'Check docs',
			description: '<p>See <a href="https://vikunja.io/docs">the documentation</a> for details.</p>',
		})

		expect(markdown).toBe('# Check docs\n\nSee [the documentation](https://vikunja.io/docs) for details.')
	})

	it('converts bold and italic to markdown', () => {
		const markdown = formatTaskAsMarkdown({
			title: 'Formatting test',
			description: '<p>This is <strong>important</strong> and <em>urgent</em></p>',
		})

		expect(markdown).toBe('# Formatting test\n\nThis is **important** and *urgent*')
	})

	it('converts bullet lists to markdown', () => {
		const markdown = formatTaskAsMarkdown({
			title: 'Shopping list',
			description: '<ul><li><p>Milk</p></li><li><p>Eggs</p></li><li><p>Bread</p></li></ul>',
		})

		expect(markdown).toContain('* Milk')
		expect(markdown).toContain('* Eggs')
		expect(markdown).toContain('* Bread')
	})

	it('converts task lists with checked/unchecked state', () => {
		const markdown = formatTaskAsMarkdown({
			title: 'Checklist',
			description: '<ul data-type="taskList"><li data-type="taskItem" data-checked="true"><p>Done item</p></li><li data-type="taskItem" data-checked="false"><p>Todo item</p></li></ul>',
		})

		expect(markdown).toContain('[x] Done item')
		expect(markdown).toContain('[ ] Todo item')
	})

	it('converts tables to pipe-separated rows', () => {
		const markdown = formatTaskAsMarkdown({
			title: 'Table test',
			description: '<table><tr><th>Name</th><th>Status</th></tr><tr><td>Task A</td><td>Done</td></tr><tr><td>Task B</td><td>Pending</td></tr></table>',
		})

		expect(markdown).toContain('Name | Status')
		expect(markdown).toContain('Task A | Done')
		expect(markdown).toContain('Task B | Pending')
	})

	it('preserves line breaks across paragraphs', () => {
		const markdown = formatTaskAsMarkdown({
			title: 'Multi-paragraph',
			description: '<p>First paragraph</p><p>Second paragraph</p>',
		})

		expect(markdown).toBe('# Multi-paragraph\n\nFirst paragraph\n\nSecond paragraph')
	})

	it('trims whitespace from the title', () => {
		const markdown = formatTaskAsMarkdown({
			title: '  Spaces around  ',
			description: '',
		})

		expect(markdown).toBe('# Spaces around')
	})
})
