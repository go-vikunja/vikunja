import {Plugin, PluginKey} from '@tiptap/pm/state'
import {Decoration, DecorationSet} from '@tiptap/pm/view'

// Define the available fields for filtering
const FIELDS = [
	'assignees', 
	'created', 
	'done',
	'doneAt', 
	'dueDate', 
	'endDate', 
	'labels', 
	'percentDone', 
	'priority', 
	'project',
	'reminders', 
	'startDate',
	'updated',
]

// Operators
const OPERATORS = ['!=', '=', '>=', '<=', '>', '<', 'like', 'in', 'not in']

// Logical operators
const LOGICAL_OPERATORS = ['&&', '||']

// Create a plugin key for our plugin
const filterHighlighterKey = new PluginKey('filterHighlighter')

export const filterHighlighter = new Plugin({
	key: filterHighlighterKey,
	state: {
		init() {
			return DecorationSet.empty
		},
		apply(tr, oldState) {
			// Only recompute decorations if the document changed
			if (!tr.docChanged) return oldState

			const decorations: Decoration[] = []
			const doc = tr.doc

			// Get the text content of the document
			const text = doc.textContent

			// Create a regex to match field names
			const fieldRegex = new RegExp(`\\b(${FIELDS.join('|')})\\b`, 'g')

			// Create a regex to match operators
			const operatorRegex = new RegExp(`(${OPERATORS.map(op => op.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')).join('|')})`, 'g')

			// Create a regex to match logical operators
			const logicalRegex = new RegExp(`(${LOGICAL_OPERATORS.map(op => op.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')).join('|')})`, 'g')

			// Create a regex for parentheses
			const groupingRegex = /[()]/g

			// Create a regex for values
			const valueRegex = /(?:true|false|\d+|"[^"]*"|'[^']*'|now(?:[+-]\d+[smhdwMy])?(?:\/[dwMy])?)/g

			// Match fields
			let match
			while ((match = fieldRegex.exec(text)) !== null) {
				const start = match.index
				const end = start + match[0].length

				// Get the position in the document
				const from = findPosForIndex(doc, start)
				const to = findPosForIndex(doc, end)

				if (from !== null && to !== null) {
					decorations.push(
						Decoration.inline(from, to, {class: 'field'}),
					)

					// If this is an assignees field, look for the next value
					if (match[0] === 'assignees') {
						// Look for the next value after this field
						const afterField = text.slice(end)
						const operatorMatch = /\s*(?:in|=)\s*/.exec(afterField)
						if (operatorMatch) {
							const valueStart = end + operatorMatch[0].length
							const valueMatch = /([a-zA-Z0-9_]+)(?:,\s*[a-zA-Z0-9_]+)*/.exec(text.slice(valueStart))
							if (valueMatch) {
								const users = valueMatch[0].split(/,\s*/)
								users.forEach(user => {
									const userStart = text.indexOf(user, valueStart)
									const userEnd = userStart + user.length
									const userFrom = findPosForIndex(doc, userStart)
									const userTo = findPosForIndex(doc, userEnd)

									if (userFrom !== null && userTo !== null) {
										decorations.push(
											Decoration.inline(userFrom, userTo, {
												class: 'value user-value',
												'data-user': user,
											}),
										)
									}
								})
								continue
							}
						}
					}
				}
			}

			// Match operators
			operatorRegex.lastIndex = 0
			while ((match = operatorRegex.exec(text)) !== null) {
				const start = match.index
				const end = start + match[0].length

				const from = findPosForIndex(doc, start)
				const to = findPosForIndex(doc, end)

				if (from !== null && to !== null) {
					decorations.push(
						Decoration.inline(from, to, {class: 'operator'}),
					)
				}
			}

			// Match logical operators
			logicalRegex.lastIndex = 0
			while ((match = logicalRegex.exec(text)) !== null) {
				const start = match.index
				const end = start + match[0].length

				const from = findPosForIndex(doc, start)
				const to = findPosForIndex(doc, end)

				if (from !== null && to !== null) {
					decorations.push(
						Decoration.inline(from, to, {class: 'logical'}),
					)
				}
			}

			// Match grouping symbols
			groupingRegex.lastIndex = 0
			while ((match = groupingRegex.exec(text)) !== null) {
				const start = match.index
				const end = start + match[0].length

				const from = findPosForIndex(doc, start)
				const to = findPosForIndex(doc, end)

				if (from !== null && to !== null) {
					decorations.push(
						Decoration.inline(from, to, {class: 'grouping'}),
					)
				}
			}

			// Match values that aren't already matched as user values
			valueRegex.lastIndex = 0
			while ((match = valueRegex.exec(text)) !== null) {
				const start = match.index
				const end = start + match[0].length

				// Skip if this is a field name
				const value = match[0]
				if (FIELDS.includes(value)) continue

				const from = findPosForIndex(doc, start)
				const to = findPosForIndex(doc, end)

				if (from !== null && to !== null) {
					// Check if this position is already decorated as a user value
					const hasUserValue = decorations.some(d =>
						d.from === from && d.to === to && d.type.attrs.class?.includes('user-value'),
					)

					if (!hasUserValue) {
						decorations.push(
							Decoration.inline(from, to, {class: 'value'}),
						)
					}
				}
			}

			return DecorationSet.create(doc, decorations)
		},
	},
	props: {
		decorations(state) {
			return this.getState(state)
		},
	},
})

// Helper function to find the position in the document for a given text index
function findPosForIndex(doc: any, index: number): number | null {
	let pos = 0
	let found = false
	let textIndex = 0

	doc.descendants((node: any, nodePos: number) => {
		if (found) return false

		if (node.isText) {
			const endIndex = textIndex + node.text.length

			if (textIndex <= index && index <= endIndex) {
				pos = nodePos + (index - textIndex)
				found = true
				return false
			}

			textIndex = endIndex
		}
	})

	return found ? pos : null
}
