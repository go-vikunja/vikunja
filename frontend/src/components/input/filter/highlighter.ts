import {Plugin, PluginKey} from '@tiptap/pm/state'
import {Decoration, DecorationSet} from '@tiptap/pm/view'
import {AVAILABLE_FILTER_FIELDS, FILTER_JOIN_OPERATOR, FILTER_OPERATORS, LABEL_FIELDS, DATE_FIELDS, getFilterFieldRegexPattern} from '@/helpers/filters'
import {useLabelStore} from '@/stores/labels'
import {colorIsDark} from '@/helpers/color/colorIsDark.ts'

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

			// Get label store for color decoration
			const labelStore = useLabelStore()

			// Create a regex to match field names
			const fieldRegex = new RegExp(`\\b(${AVAILABLE_FILTER_FIELDS.join('|')})\\b`, 'g')

			// Create a regex to match operators
			const operatorRegex = new RegExp(`(${FILTER_OPERATORS.map(op => op.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')).join('|')})`, 'g')

			// Create a regex to match logical/join operators  
			const logicalRegex = new RegExp(`(${FILTER_JOIN_OPERATOR.map(op => op.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')).join('|')})`, 'g')

			// Create a regex to match field + operator + value patterns
			// This will match anything coming after an operator
			const fieldValueRegex = new RegExp(
				`(${AVAILABLE_FILTER_FIELDS.join('|')})\\s*(${FILTER_OPERATORS.map(op => op.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')).join('|')})\\s*([^&|()]+?)(?=\\s*(?:${FILTER_JOIN_OPERATOR.slice(0, 2).map(op => op.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')).join('|')})|$)`,
				'gi',
			)

			let match

			// Track ranges that are already decorated as values to avoid conflicts
			const valueRanges: Array<{ start: number, end: number }> = []

			// Handle date values with click functionality first
			DATE_FIELDS.forEach(dateField => {
				const pattern = getFilterFieldRegexPattern(dateField)
				let dateMatch
				while ((dateMatch = pattern.exec(text)) !== null) {
					if (dateMatch[4]) { // If there's a value
						const valueText = dateMatch[4].trim()
						const valueStart = dateMatch.index + dateMatch[0].indexOf(dateMatch[4])
						const valueEnd = valueStart + dateMatch[4].length

						const from = findPosForIndex(doc, valueStart)
						const to = findPosForIndex(doc, valueEnd)

						if (from !== null && to !== null) {
							decorations.push(
								Decoration.inline(from, to, {
									class: 'date-value',
									'data-date-value': valueText,
									'data-position': valueStart.toString()
								})
							)
							valueRanges.push({start: valueStart, end: valueEnd})
						}
					}
				}
			})

			// Handle label values with colors
			LABEL_FIELDS.forEach(labelField => {
				const pattern = getFilterFieldRegexPattern(labelField)
				let labelMatch
				while ((labelMatch = pattern.exec(text)) !== null) {
					if (labelMatch[4]) { // If there's a value
						const valueText = labelMatch[4].trim()
						const valueStart = labelMatch.index + labelMatch[0].indexOf(labelMatch[4])
						const valueEnd = valueStart + labelMatch[4].length

						// Find the label by its title
						const label = labelStore.getLabelByExactTitle(valueText)

						const from = findPosForIndex(doc, valueStart)
						const to = findPosForIndex(doc, valueEnd)

						if (from !== null && to !== null) {
							if (label) {
								// Use label color if found
								decorations.push(
									Decoration.inline(from, to, {
										class: 'label-value',
										style: `background-color: ${label.hexColor}; color: ${label.hexColor && colorIsDark(label.hexColor) ? 'white' : 'black'};`
									})
								)
							} else {
								// Fallback to generic value styling
								decorations.push(
									Decoration.inline(from, to, {class: 'value'})
								)
							}
							valueRanges.push({start: valueStart, end: valueEnd})
						}
					}
				}
			})

			// Match other values - anything coming after an operator (excluding labels)
			fieldValueRegex.lastIndex = 0
			while ((match = fieldValueRegex.exec(text)) !== null) {
				const [fullMatch, field, operator, value] = match

				// Skip label and date fields as they're handled above
				if (LABEL_FIELDS.includes(field) || DATE_FIELDS.includes(field)) {
					continue
				}

				if (value && value.trim()) {
					// Calculate the actual position of the value by finding where it starts after the operator
					const fieldLength = field.length
					const operatorIndex = fullMatch.indexOf(operator, fieldLength)
					const operatorEnd = operatorIndex + operator.length
					const valueIndex = fullMatch.indexOf(value.trim(), operatorEnd)

					const valueStart = match.index + valueIndex
					const valueEnd = valueStart + value.trim().length

					const from = findPosForIndex(doc, valueStart)
					const to = findPosForIndex(doc, valueEnd)

					if (from !== null && to !== null) {
						decorations.push(
							Decoration.inline(from, to, {class: 'value'}),
						)
						valueRanges.push({start: valueStart, end: valueEnd})
					}
				}
			}

			// Helper function to check if a range overlaps with any value range
			const overlapsWithValue = (start: number, end: number): boolean => {
				return valueRanges.some(range =>
					(start >= range.start && start < range.end) ||
					(end > range.start && end <= range.end) ||
					(start <= range.start && end >= range.end),
				)
			}

			// Match fields (excluding those within value ranges)
			fieldRegex.lastIndex = 0
			while ((match = fieldRegex.exec(text)) !== null) {
				const start = match.index
				const end = start + match[0].length

				// Skip if this field match is within a value range
				if (overlapsWithValue(start, end)) {
					continue
				}

				const from = findPosForIndex(doc, start)
				const to = findPosForIndex(doc, end)

				if (from !== null && to !== null) {
					decorations.push(
						Decoration.inline(from, to, {class: 'field'}),
					)
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
function findPosForIndex(doc: {
	descendants: (fn: (node: { isText: boolean, text: string }, nodePos: number) => boolean | void) => void
}, index: number): number | null {
	let pos = 0
	let found = false
	let textIndex = 0

	doc.descendants((node: { isText: boolean, text: string }, nodePos: number) => {
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
