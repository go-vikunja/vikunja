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

	// Transform labels to ids
	LABEL_FIELDS.forEach(field => {
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
					const labelId = labelResolver(k)
					if (labelId !== null) {
						filter = filter.replace(k, String(labelId))
					}
				})
			}
		}
	})
	// Transform projects to ids
	PROJECT_FIELDS.forEach(field => {
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

				let replaced = keyword

				keywords.forEach(k => {
					const projectId = projectResolver(k)
					if (projectId !== null) {
						replaced = replaced.replace(k, String(projectId))
					}
				})

				const actualKeywordStart = (match?.index || 0) + prefix.length
				filter = filter.substring(0, actualKeywordStart) +
					replaced +
					filter.substring(actualKeywordStart + keyword.length)
			}
		}
	})

	// Transform all attributes to snake case
	AVAILABLE_FILTER_FIELDS.forEach(f => {
		filter = filter.replaceAll(f, snakeCase(f))
	})

	return filter
}

export function transformFilterStringFromApi(
	filter: string,
	labelResolver: (id: number) => string | null,
	projectResolver: (id: number) => string | null,
): string {

	if (filter.trim() === '') {
		return ''
	}

	// Transform all attributes from snake case
	AVAILABLE_FILTER_FIELDS.forEach(f => {
		filter = filter.replaceAll(snakeCase(f), f)
	})

	// Transform labels to their titles
	LABEL_FIELDS.forEach(field => {
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
					const labelTitle = labelResolver(parseInt(k))
					if (labelTitle !== null) {
						filter = filter.replace(k, labelTitle)
					}
				})
			}
		}
	})

	// Transform projects to ids
	PROJECT_FIELDS.forEach(field => {
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
					const project = projectResolver(parseInt(k))
					if (project !== null) {
						filter = filter.replace(k, project)
					}
				})
			}
		}
	})

	return filter
}
