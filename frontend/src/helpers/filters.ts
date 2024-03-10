import {snakeCase} from 'snake-case'

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
	'done',
	'priority',
	'percentDone',
	...DATE_FIELDS,
	...ASSIGNEE_FIELDS,
	...LABEL_FIELDS,
	...PROJECT_FIELDS,
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

export const FILTER_OPERATORS_REGEX = '(&lt;|&gt;|&lt;=|&gt;=|=|!=)'

function getFieldPattern(field: string): RegExp {
	return new RegExp('(' + field + '\\s*' + FILTER_OPERATORS_REGEX + '\\s*)([\'"]?)([^\'"&|()]+\\1?)?', 'ig')
}

export function transformFilterStringForApi(
	filter: string,
	labelResolver: (title: string) => number | null,
	projectResolver: (title: string) => number | null,
): string {
	
	if (filter.trim() === '') {
		return ''
	}
	
	// Transform labels to ids
	LABEL_FIELDS.forEach(field => {
		const pattern = getFieldPattern(field)

		let match: RegExpExecArray | null
		while ((match = pattern.exec(filter)) !== null) {
			// eslint-disable-next-line @typescript-eslint/no-unused-vars
			const [matched, prefix, operator, space, keyword] = match
			if (keyword) {
				const labelId = labelResolver(keyword.trim())
				if (labelId !== null) {
					filter = filter.replace(keyword, String(labelId))
				}
			}
		}
	})
	// Transform projects to ids
	PROJECT_FIELDS.forEach(field => {
		const pattern = getFieldPattern(field)

		let match: RegExpExecArray | null
		while ((match = pattern.exec(filter)) !== null) {
			// eslint-disable-next-line @typescript-eslint/no-unused-vars
			const [matched, prefix, operator, space, keyword] = match
			if (keyword) {
				const projectId = projectResolver(keyword.trim())
				if (projectId !== null) {
					filter = filter.replace(keyword, String(projectId))
				}
			}
		}
	})

	// Transform all attributes to snake case
	AVAILABLE_FILTER_FIELDS.forEach(f => {
		filter = filter.replace(f, snakeCase(f))
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
		filter = filter.replace(snakeCase(f), f)
	})
	
	// Transform labels to their titles
	LABEL_FIELDS.forEach(field => {
		const pattern = getFieldPattern(field)

		let match: RegExpExecArray | null
		while ((match = pattern.exec(filter)) !== null) {
			// eslint-disable-next-line @typescript-eslint/no-unused-vars
			const [matched, prefix, operator, space, keyword] = match
			if (keyword) {
				const labelTitle = labelResolver(Number(keyword.trim()))
				if (labelTitle !== null) {
					filter = filter.replace(keyword, labelTitle)
				}
			}
		}
	})

	// Transform projects to ids
	PROJECT_FIELDS.forEach(field => {
		const pattern = getFieldPattern(field)

		let match: RegExpExecArray | null
		while ((match = pattern.exec(filter)) !== null) {
			// eslint-disable-next-line @typescript-eslint/no-unused-vars
			const [matched, prefix, operator, space, keyword] = match
			if (keyword) {
				const project = projectResolver(Number(keyword.trim()))
				if (project !== null) {
					filter = filter.replace(keyword, project)
				}
			}
		}
	})

	return filter
}
