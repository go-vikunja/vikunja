import {snakeCase} from 'change-case'

function trimQuotes(str: string): string {

	str = str.trim()

	if ((str.startsWith('"') && str.endsWith('"')) || 
		(str.startsWith('\'') && str.endsWith('\''))) {
		return str.slice(1, -1)
	}
	return str
}

export const DATE_FIELDS = [
	'dueDate',
	'startDate',
	'endDate',
	'doneAt',
	'reminders',
	'created',
	'updated',
]

export const ASSIGNEE_FIELDS = [
	'assignees',
]

export const LABEL_FIELDS = [
	'labels',
]

export const PROJECT_FIELDS = [
	'project',
]

export const AUTOCOMPLETE_FIELDS = [
	...LABEL_FIELDS,
	...ASSIGNEE_FIELDS,
	...PROJECT_FIELDS,
]

export const AVAILABLE_FILTER_FIELDS = [
	...DATE_FIELDS,
	...ASSIGNEE_FIELDS,
	...LABEL_FIELDS,
	...PROJECT_FIELDS,
	'done',
	'priority',
	'percentDone',
]

export const FILTER_OPERATORS = [
	'!=',
	'=',
	'>',
	'>=',
	'<',
	'<=',
	'like',
	'not in',
	'in',
	'?=',
]

export const FILTER_JOIN_OPERATOR = [
	'&&',
	'||',
	'(',
	')',
]

export const FILTER_OPERATORS_REGEX = '('+FILTER_OPERATORS.map(op => {
	// Only add word boundaries for operators that are words (like 'in', 'like', 'not in')
	const needsWordBoundary = /^[a-zA-Z]/.test(op) || /[a-zA-Z]$/.test(op)
	const escaped = op.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
	return needsWordBoundary ? '\\b' + escaped + '\\b' : escaped
}).join('|')+')'

export function hasFilterQuery(filter: string): boolean {
	return FILTER_OPERATORS.find(o => filter.includes(o)) || false
}

export function getFilterFieldRegexPattern(field: string): RegExp {
	return new RegExp('\\b(' + field + ')\\s*' + FILTER_OPERATORS_REGEX + '\\s*(?:(["\'])((?:\\\\.|(?!\\3)[^\\\\])*?)\\3|([^&|()<]+?))(?=\\s*(?:&&|\\||$))', 'g')
}

export function transformFilterStringForApi(
	filter: string,
	labelResolver: (title: string) => number | null,
	projectResolver: (title: string) => number | null,
): string {

	filter = filter.trim()

	if (filter === '') {
		return ''
	}

	AVAILABLE_FILTER_FIELDS.forEach(f => {
		const fieldPattern = new RegExp('\\b(' + f + ')\\b(?=\\s*' + FILTER_OPERATORS_REGEX + ')', 'gi')
		filter = filter.replace(fieldPattern, f)
	})

	// Transform labels and projects to ids
	function transformFieldToIds(
		fields: string[],
		resolver: (title: string) => number | null,
		filter: string,
	): string {
		fields.forEach(field => {
			const pattern = getFilterFieldRegexPattern(field)

			let match: RegExpExecArray | null
			const replacements: { start: number, length: number, replacement: string }[] = []

			while ((match = pattern.exec(filter)) !== null) {
				const [matched, fieldName, operator, quotes, quotedContent, unquotedContent] = match
				const keyword = quotedContent || unquotedContent
				if (!keyword) {
					continue
				}

				let keywords = [keyword.trim()]
				if (isMultiValueOperator(operator)) {
					keywords = keyword.trim().split(',').map(k => trimQuotes(k))
				}

				let replaced = keyword

				const transformedKeywords: string[] = []
				keywords.forEach(k => {
					let id = resolver(k)
					if (id === null && k.includes('\\')) {
						id = resolver(k.replaceAll('\\', ''))
					}
					if (id === null) {
						transformedKeywords.push(k)
						return
					}
				
					transformedKeywords.push(String(id))
				})
				
				// Join the transformed keywords back together
				if (isMultiValueOperator(operator)) {
					replaced = transformedKeywords.join(', ')
				} else {
					replaced = transformedKeywords[0] || keyword
				}

				replaced = replaced.replaceAll('"', '').replaceAll('\'', '')

				// Reconstruct the entire match with the replaced value
				let reconstructedMatch
				if (quotes && quotedContent) {
					// For quoted values, remove quotes since we converted to IDs
					reconstructedMatch = `${fieldName} ${operator} ${replaced}`
				} else if (unquotedContent) {
					// For unquoted values
					reconstructedMatch = `${fieldName} ${operator} ${replaced}`
				} else {
					continue
				}

				replacements.push({
					start: match.index!,
					length: matched.length,
					replacement: reconstructedMatch,
				})
			}

			// We're collecting the results first and then replacing the filter string in the end
			// to avoid modifying the input string as we iterate over it.
			let offset = 0
			replacements.forEach(({start, length, replacement}) => {
				filter = filter.substring(0, start + offset) +
					replacement +
					filter.substring(start + offset + length)
				offset += replacement.length - length
			})
		})
		return filter
	}

	// Transform labels to ids
	filter = transformFieldToIds(LABEL_FIELDS, labelResolver, filter)

	// Transform projects to ids
	filter = transformFieldToIds(PROJECT_FIELDS, projectResolver, filter)

	// Transform all field names (not values) to snake case
	AVAILABLE_FILTER_FIELDS.forEach(f => {
		const fieldPattern = new RegExp('\\b' + f + '\\b(?=\\s*' + FILTER_OPERATORS_REGEX + ')', 'gi')
		filter = filter.replace(fieldPattern, snakeCase(f))
	})

	return filter
}

export function transformFilterStringFromApi(
	filter: string,
	labelResolver: (id: number) => string | null | undefined,
	projectResolver: (id: number) => string | null | undefined,
): string {

	if (filter.trim() === '') {
		return ''
	}

	// Transform all attributes from snake case
	AVAILABLE_FILTER_FIELDS.forEach(f => {
		filter = filter.replaceAll(snakeCase(f), f)
	})

	// Function to transform fields to their titles
	function transformFieldsToTitles(
		fields: string[],
		resolver: (id: number) => string | null | undefined,
	) {
		fields.forEach(field => {
			const pattern = getFilterFieldRegexPattern(field)

			let match: RegExpExecArray | null
			while ((match = pattern.exec(filter)) !== null) {
				// eslint-disable-next-line @typescript-eslint/no-unused-vars
				const [matched, fieldName, operator, quotes, quotedContent, unquotedContent] = match
				const keyword = quotedContent || unquotedContent
				if (keyword) {
					let keywords = [keyword.trim()]
					if (isMultiValueOperator(operator)) {
						keywords = keyword.trim().split(',').map(k => {
							let trimmed = k.trim()
							// Strip quotes from individual values in multi-value scenarios
							trimmed = trimQuotes(trimmed)
							return trimmed
						})
					}

					keywords.forEach(k => {
						const title = resolver(parseInt(k))
						if (title) {
							filter = filter.replace(k, title)
						}
					})
				}
			}
		})
	}

	// Transform labels to their titles
	transformFieldsToTitles(LABEL_FIELDS, labelResolver)

	// Transform projects to their titles
	transformFieldsToTitles(PROJECT_FIELDS, projectResolver)

	return filter
}

export function isMultiValueOperator(operator: string): boolean {
	return ['in', '?=', 'not in', '?!='].includes(operator)
}
