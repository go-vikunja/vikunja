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

		editor.commands.setContent('<ul data-type="taskList"><li data-type="taskItem" data-checked="false"><p>Item 1</p></li></ul>')

		const html = editor.getHTML()
		expect(html).toContain('data-task-id=')

		editor.destroy()
	})

	it('should preserve existing IDs when parsing HTML', () => {
		const existingId = 'test-id-123'
		const editor = createEditor()

		editor.commands.setContent(`<ul data-type="taskList"><li data-type="taskItem" data-checked="false" data-task-id="${existingId}"><p>Item 1</p></li></ul>`)

		const html = editor.getHTML()
		expect(html).toContain(`data-task-id="${existingId}"`)

		editor.destroy()
	})

	it('should generate different IDs for different items', () => {
		const editor = createEditor()

		editor.commands.setContent('<ul data-type="taskList"><li data-type="taskItem" data-checked="false"><p>Item 1</p></li><li data-type="taskItem" data-checked="false"><p>Item 2</p></li><li data-type="taskItem" data-checked="false"><p>Item 3</p></li></ul>')

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

		editor.commands.setContent('<ul data-type="taskList"><li data-type="taskItem" data-checked="false"><p>Test</p></li></ul>')

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

		editor.commands.setContent('<ul data-type="taskList"><li data-type="taskItem" data-checked="false"><p>Duplicate</p></li><li data-type="taskItem" data-checked="false"><p>Duplicate</p></li><li data-type="taskItem" data-checked="false"><p>Duplicate</p></li></ul>')

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
