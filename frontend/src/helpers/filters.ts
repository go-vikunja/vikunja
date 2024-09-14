import {snakeCase} from 'change-case'

export const DATE_FIELDS = [
	'dueDate',
	'startDate',
	'endDate',
	'doneAt',
	'reminders',
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
	'in',
	'?=',
]

export const FILTER_JOIN_OPERATOR = [
	'&&',
	'||',
	'(',
	')',
]

export const FILTER_OPERATORS_REGEX = '(&lt;|&gt;|&lt;=|&gt;=|=|!=|in)'

export function getFilterFieldRegexPattern(field: string): RegExp {
	return new RegExp('(' + field + '\\s*' + FILTER_OPERATORS_REGEX + '\\s*)([\'"]?)([^\'"&|()<]+\\1?)?', 'ig')
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
		filter: string
	): string {
		fields.forEach(field => {
			const pattern = getFilterFieldRegexPattern(field)

			let match: RegExpExecArray | null
			while ((match = pattern.exec(filter)) !== null) {
				// eslint-disable-next-line @typescript-eslint/no-unused-vars
				const [matched, prefix, operator, space, keyword] = match
				if (!keyword) {
					continue
				}

				let keywords = [keyword.trim()]
				if (operator === 'in' || operator === '?=') {
					keywords = keyword.trim().split(',').map(k => k.trim())
				}
				
				let replaced = keyword

				keywords.forEach(k => {
					const id = resolver(k)
					if (id !== null) {
						replaced = replaced.replace(k, String(id))
					}
				})

				const actualKeywordStart = (match?.index || 0) + prefix.length
				filter = filter.substring(0, actualKeywordStart) +
					replaced +
					filter.substring(actualKeywordStart + keyword.length)
			}
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
		resolver: (id: number) => string | null | undefined
	) {
		fields.forEach(field => {
			const pattern = getFilterFieldRegexPattern(field)

			let match: RegExpExecArray | null
			while ((match = pattern.exec(filter)) !== null) {
				// eslint-disable-next-line @typescript-eslint/no-unused-vars
				const [matched, prefix, operator, space, keyword] = match
				if (keyword) {
					let keywords = [keyword.trim()]
					if (operator === 'in' || operator === '?=') {
						keywords = keyword.trim().split(',').map(k => k.trim())
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
