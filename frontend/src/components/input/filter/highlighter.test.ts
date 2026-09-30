import {afterEach, describe, expect, it} from 'vitest'
import {Schema} from '@tiptap/pm/model'
import {EditorState} from '@tiptap/pm/state'
import {EditorView} from '@tiptap/pm/view'
import type {Label} from '@/client/generated'

import {decorateDocument} from './highlighter'

const schema = new Schema({
	nodes: {
		doc: {content: 'paragraph+'},
		paragraph: {content: 'text*', toDOM: () => ['p', 0]},
		text: {inline: true},
	},
})

function filterDocument(value: string) {
	return schema.node('doc', null, [
		schema.node('paragraph', null, [schema.text(value)]),
	])
}

const views: EditorView[] = []
afterEach(() => views.splice(0).forEach(view => view.destroy()))

function renderHighlights(text: string, labels: Label[] = []) {
	const doc = filterDocument(text)
	const view = new EditorView(document.createElement('div'), {
		state: EditorState.create({doc}),
		decorations: () => decorateDocument(doc, labels),
	})
	views.push(view)
	return view
}

describe('filter highlighter', () => {
	it('uses loaded label data when recalculating decorations', () => {
		const doc = filterDocument('labels = "Work"')

		const before = decorateDocument(doc, [])
		const after = decorateDocument(doc, [{id: 1, title: 'Work', hex_color: 'ff006e'}])

		expect(before.find()).not.toEqual(after.find())
	})

	it('marks unquoted date values as clickable', () => {
		const view = renderHighlights('dueDate < now/w+1w')
		const dateValue = view.dom.querySelector('.date-value')
		expect(dateValue?.textContent).toBe('now/w+1w')
		expect(dateValue?.getAttribute('data-date-value')).toBe('now/w+1w')
	})

	it('marks unquoted label values', () => {
		const text = 'labels = Work'
		const view = renderHighlights(text, [{id: 1, title: 'Work', hex_color: 'ff006e'}])
		const labelValue = view.dom.querySelector('.label-value')
		expect(labelValue?.textContent).toBe('Work')
		expect(labelValue && view.posAtDOM(labelValue, 0)).toBe(text.lastIndexOf('Work') + 1)
	})

	it('marks the value, not the field name, when field and value share a name', () => {
		const text = 'dueDate < dueDate'
		const view = renderHighlights(text)
		const dateValue = view.dom.querySelector('.date-value')
		expect(dateValue?.textContent).toBe('dueDate')
		expect(dateValue && view.posAtDOM(dateValue, 0)).toBe(text.lastIndexOf('dueDate') + 1)
	})
})
