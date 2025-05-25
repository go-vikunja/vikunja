import {snakeCase} from 'change-case'

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

export const FILTER_OPERATORS_REGEX = '('+FILTER_OPERATORS.map(op => op.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')).join('|')+')'

export function hasFilterQuery(filter: string): boolean {
	return FILTER_OPERATORS.find(o => filter.includes(o)) || false
}

export function getFilterFieldRegexPattern(field: string): RegExp {
	return new RegExp('(' + field + ')\\s*' + FILTER_OPERATORS_REGEX + '\\s*([\'"]?)([^\'"&|()<]+?)(?=\\s*(?:&&|\\|\\||$))', 'ig')
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
		filter = filter.replace(new RegExp(f, 'ig'), f)
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
				// eslint-disable-next-line @typescript-eslint/no-unused-vars
				const [matched, fieldName, operator, quotes, keyword] = match
				if (!keyword) {
					continue
				}

				let keywords = [keyword.trim()]
				if (isMultiValueOperator(operator)) {
					keywords = keyword.trim().split(',').map(k => trimQuotes(k))
				}

				let replaced = keyword

				keywords.forEach(k => {
					const id = resolver(k)
					if (id !== null) {
						replaced = replaced.replace(k, String(id))
					}
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

				const actualKeywordStart = (match?.index || 0) + matched.length - keyword.length
				replacements.push({
					start: actualKeywordStart,
					length: keyword.length,
					replacement: replaced,
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

	// Transform all attributes to snake case
	AVAILABLE_FILTER_FIELDS.forEach(f => {
		filter = filter.replaceAll(f, snakeCase(f))
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
				const [matched, fieldName, operator, quotes, keyword] = match
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
