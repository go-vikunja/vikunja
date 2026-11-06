import type {Priority} from '@/constants/priorities'

export interface DueFilter {
	range: {
		from: string
		to: string
	} | null
	includeOverdue: boolean
}

export interface SimpleFilter {
	showDone: boolean
	due: DueFilter
	labelIds: number[]
	assignees: string[]
	projectIds: number[]
	minPriority: Priority | null
}
const dateMath = /^now(?:[+-]\d+[smhdwMy]|\/[smhdwMy])*$/
const usernameList = /^[^\s,()&|=<>"']+(?:, [^\s,()&|=<>"']+)*$/
const isoDate = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$/

function quoteDate(value: string): string {
	return dateMath.test(value) ? value : JSON.stringify(value)
}

export function toQuery(filter: SimpleFilter): string {
	const clauses: string[] = []
	if (!filter.showDone) {
		clauses.push('done = false')
	}
	const {range, includeOverdue} = filter.due
	if (range) {
		if (!includeOverdue) {
			clauses.push(`due_date > ${quoteDate(range.from)}`)
		}
		clauses.push(`due_date < ${quoteDate(range.to)}`)
	} else if (includeOverdue) {
		clauses.push('due_date < now')
	}
	if (filter.labelIds.length) {
		clauses.push(`labels in ${filter.labelIds.join(', ')}`)
	}
	if (filter.assignees.length) {
		const value = filter.assignees.join(', ')
		clauses.push(`assignees in ${usernameList.test(value) ? value : JSON.stringify(value)}`)
	}
	if (filter.projectIds.length) {
		clauses.push(`project in ${filter.projectIds.join(', ')}`)
	}
	if (filter.minPriority !== null) {
		clauses.push(`priority >= ${filter.minPriority}`)
	}
	return clauses.join(' && ')
}

function parseDate(value: string): string | null {
	if (dateMath.test(value)) {
		return value
	}
	if (value.startsWith('"') && value.endsWith('"')) {
		const date = value.slice(1, -1)
		if (isoDate.test(date) && !Number.isNaN(Date.parse(date)) && new Date(date).toISOString() === date) {
			return date
		}
	}
	return null
}

export function fromQuery(query: string): SimpleFilter | null {
	const filter: SimpleFilter = {
		showDone: true,
		due: {
			range: null,
			includeOverdue: false,
		},
		labelIds: [],
		assignees: [],
		projectIds: [],
		minPriority: null,
	}
	let from: string | undefined
	let to: string | undefined
	for (const clause of query === '' ? [] : query.split(' && ')) {
		if (clause === 'done = false') {
			filter.showDone = false
			continue
		}
		const date = /^due_date ([<>]) (.+)$/.exec(clause)
		if (date) {
			const value = parseDate(date[2])
			if (value === null) {
				return null
			}
			if (date[1] === '>') {
				from = value
			} else {
				to = value
			}
			continue
		}
		const ids = /^(labels|project) in ([1-9]\d*(?:, [1-9]\d*)*)$/.exec(clause)
		if (ids) {
			const values = ids[2].split(', ').map(Number)
			if (values.some(id => !Number.isSafeInteger(id))) {
				return null
			}
			if (ids[1] === 'labels') {
				filter.labelIds = values
			} else {
				filter.projectIds = values
			}
			continue
		}
		const assignees = /^assignees in (.+)$/.exec(clause)
		if (assignees) {
			let value = assignees[1]
			if (value.startsWith('"')) {
				try {
					value = JSON.parse(value)
				} catch {
					return null
				}
				if (typeof value !== 'string') {
					return null
				}
			} else if (!usernameList.test(value)) {
				return null
			}
			filter.assignees = value.split(', ')
			continue
		}
		const priority = /^priority >= ([0-5])$/.exec(clause)
		if (!priority) {
			return null
		}
		filter.minPriority = Number(priority[1]) as Priority
	}
	if (from !== undefined && to === undefined) {
		return null
	}
	if (to !== undefined) {
		filter.due = from !== undefined
			? {
				range: {
					from,
					to,
				},
				includeOverdue: false,
			} : {
				range: to === 'now' ? null : {
					from: '',
					to,
				},
				includeOverdue: true,
			}
	}
	return toQuery(filter) === query ? filter : null
}
