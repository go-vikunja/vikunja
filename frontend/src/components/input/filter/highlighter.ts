import {EditorState, Plugin, PluginKey, Transaction} from '@tiptap/pm/state'
import {Decoration, DecorationSet} from '@tiptap/pm/view'
import {
	AVAILABLE_FILTER_FIELDS,
	DATE_FIELDS,
	FILTER_JOIN_OPERATOR,
	FILTER_OPERATORS,
	FILTER_OPERATORS_REGEX,
	getFilterFieldRegexPattern,
	LABEL_FIELDS,
} from '@/helpers/filters'
import {useLabelStore} from '@/stores/labels'
import {colorIsDark} from '@/helpers/color/colorIsDark.ts'
import {Node} from '@tiptap/pm/model'

export const filterHighlighter = new Plugin({
	key: new PluginKey('filterHighlighter'),
	state: {
		init(_, state: EditorState) {
			return decorateDocument(state.doc)
		},
		apply(tr: Transaction, oldState) {
			if (!tr.docChanged) return oldState

			return decorateDocument(tr.doc)
		},
	},
	props: {
		decorations(state) {
			return this.getState(state)
		},
	},
})

function decorateDocument(doc: Node) {
	const decorations: Decoration[] = []

	const text = doc.textContent

	const labelStore = useLabelStore()

	const fieldRegex = new RegExp(`\\b(${AVAILABLE_FILTER_FIELDS.join('|')})\\b`, 'g')
	const operatorRegex = new RegExp(FILTER_OPERATORS_REGEX, 'g')
	const logicalRegex = new RegExp(`(${FILTER_JOIN_OPERATOR.map(op => op.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')).join('|')})`, 'g')
	const fieldValueRegex = new RegExp(
		`(${AVAILABLE_FILTER_FIELDS.join('|')})\\s*(${FILTER_OPERATORS.map(op => op.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')).join('|')})\\s*([^\\s&|()]+)`,
		'gi',
	)

	let match

	const valueRanges: Array<{ start: number, end: number }> = []

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
							'data-position': valueStart.toString(),
						}),
					)
					valueRanges.push({start: valueStart, end: valueEnd})
				}
			}
		}
	})

	LABEL_FIELDS.forEach(labelField => {
		const pattern = getFilterFieldRegexPattern(labelField)
		let labelMatch
		while ((labelMatch = pattern.exec(text)) !== null) {
			const labelValue = labelMatch[4]?.trim()
			const operator = labelMatch[2]?.trim()
			if (labelValue) { // If there's a value
				const valueStart = labelMatch.index + labelMatch[0].lastIndexOf(labelValue)
				const valueEnd = valueStart + labelValue.length

				const addLabelDecoration = (labelValue: string, start: number, end: number) => {
					const label = labelStore.getLabelByExactTitle(labelValue)

					const from = findPosForIndex(doc, start)
					const to = findPosForIndex(doc, end)

					if (from === null || to === null) {
						return
					}
					
					valueRanges.push({start, end})

					if (label) {
						// Use label color if found
						decorations.push(
							Decoration.inline(from, to, {
								class: 'label-value',
								style: `background-color: ${label.hexColor}; color: ${label.hexColor && colorIsDark(label.hexColor) ? 'white' : 'black'};`,
							}),
						)
						
						return
					}

					// Fallback to generic value styling
					decorations.push(
						Decoration.inline(from, to, {class: 'value'}),
					)
				}

				// Check if this is a multi-value operator and the value contains commas
				const isMultiValueOperator = ['in', '?=', 'not in', '?!='].includes(operator)
				if (isMultiValueOperator && labelValue.includes(',')) {
					// Split by commas and create decorations for each individual label
					const labels = labelValue.split(',').map(l => l.trim()).filter(l => l.length > 0)
					let currentOffset = 0
					
					labels.forEach(individualLabel => {
						// Find the position of this individual label within the full value
						const labelIndex = labelValue.indexOf(individualLabel, currentOffset)
						if (labelIndex !== -1) {
							const individualStart = valueStart + labelIndex
							const individualEnd = individualStart + individualLabel.length

							addLabelDecoration(individualLabel, individualStart, individualEnd)
							
							currentOffset = labelIndex + individualLabel.length
						}
					})
					
					continue
				}
				
				addLabelDecoration(labelValue, valueStart, valueEnd)
			}

			const valueStart = labelMatch.index + labelMatch[0].lastIndexOf(labelValue)
			const valueEnd = valueStart + labelValue.length

			const addLabelDecoration = (labelValue: string, start: number, end: number) => {
				const label = labelStore.getLabelByExactTitle(labelValue)

				const from = findPosForIndex(doc, start)
				const to = findPosForIndex(doc, end)

				if (from === null || to === null) {
					return
				}
				
				valueRanges.push({start, end})

				if (label) {
					// Use label color if found
					decorations.push(
						Decoration.inline(from, to, {
							class: 'label-value',
							style: `background-color: ${label.hexColor}; color: ${label.hexColor && colorIsDark(label.hexColor) ? 'white' : 'black'};`,
						}),
					)
					
					return
				}

				// Fallback to generic value styling
				decorations.push(
					Decoration.inline(from, to, {class: 'value'}),
				)
			}

			// Check if this is a multi-value operator and the value contains commas
			const isMultiValueOperator = ['in', '?=', 'not in', '?!='].includes(operator)
			if (isMultiValueOperator && labelValue.includes(',')) {
				// Split by commas and create decorations for each individual label
				const labels = labelValue.split(',').map(l => l.trim()).filter(l => l.length > 0)
				let currentOffset = 0
				
				labels.forEach(individualLabel => {
					// Find the position of this individual label within the full value
					const labelIndex = labelValue.indexOf(individualLabel, currentOffset)
					if (labelIndex !== -1) {
						const individualStart = valueStart + labelIndex
						const individualEnd = individualStart + individualLabel.length

						addLabelDecoration(individualLabel, individualStart, individualEnd)
						
						currentOffset = labelIndex + individualLabel.length
					}
				})
				
				continue
			}
			
			addLabelDecoration(labelValue, valueStart, valueEnd)
		}
	})

	// Match other values - anything coming after an operator (excluding labels and dates)
	fieldValueRegex.lastIndex = 0
	while ((match = fieldValueRegex.exec(text)) !== null) {
		const [fullMatch, field, operator, value] = match

		if (LABEL_FIELDS.includes(field) || DATE_FIELDS.includes(field)) {
			continue
		}

		if (value && value.trim()) {
			// Calculate the actual position of the value by finding where it starts after the operator
			const fieldLength = field.length
			const operatorIndex = fullMatch.indexOf(operator, fieldLength)
			const operatorEnd = operatorIndex + operator.length
			const valueIndex = fullMatch.indexOf(value, operatorEnd)

			const valueStart = match.index + valueIndex
			const valueEnd = valueStart + value.length

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
}

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
