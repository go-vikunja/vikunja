import type {Priority} from '@/constants/priorities'
import {parseDateMath} from '@/helpers/time/parseDateMath'

// `before` includes overdue tasks; overdue only is `before` with `to: 'now'`.
export type DueFilter =
	| {kind: 'any'}
	| {kind: 'before', to: string}
	| {kind: 'between', from: string, to: string}

export interface SimpleFilter {
	showDone: boolean
	due: DueFilter
	labelIds: number[]
	assignees: string[]
	projectIds: number[]
	minPriority: Priority | null
}
const simpleUsername = /^[^\s,()&|=<>"'\\]+$/

// The backend ignores quotes, splits values at &|() and treats \ as an escape, so unsafe names can't be escaped.
export function isSimpleUsername(username: string): boolean {
	return simpleUsername.test(username)
}

// Absolute dates are datemath too, but they stay quoted ISO strings.
export function isDateMath(value: string): boolean {
	return value.startsWith('now') && !/\s/.test(value) && parseDateMath(value) !== null
}

function quoteDate(value: string): string {
	return isDateMath(value) ? value : JSON.stringify(value)
}

export function toQuery(filter: SimpleFilter): string {
	const clauses: string[] = []
	if (!filter.showDone) {
		clauses.push('done = false')
	}
	const {due} = filter
	if (due.kind === 'between') {
		clauses.push(`due_date > ${quoteDate(due.from)}`)
	}
	if (due.kind !== 'any') {
		clauses.push(`due_date < ${quoteDate(due.to)}`)
	}
	if (filter.labelIds.length) {
		clauses.push(`labels in ${filter.labelIds.join(', ')}`)
	}
	if (filter.assignees.length) {
		clauses.push(`assignees in ${filter.assignees.join(', ')}`)
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
	if (isDateMath(value)) {
		return value
	}
	if (value.startsWith('"') && value.endsWith('"')) {
		const date = value.slice(1, -1)
		if (!Number.isNaN(Date.parse(date)) && new Date(date).toISOString() === date) {
			return date
		}
	}
	return null
}

export function emptySimpleFilter(): SimpleFilter {
	return {
		showDone: true,
		due: {kind: 'any'},
		labelIds: [],
		assignees: [],
		projectIds: [],
		minPriority: null,
	}
}

export function fromQuery(query: string): SimpleFilter | null {
	const filter = emptySimpleFilter()
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
			if (ids[1] === 'labels') {
				filter.labelIds = values
			} else {
				filter.projectIds = values
			}
			continue
		}
		const assignees = /^assignees in (.+)$/.exec(clause)
		if (assignees) {
			const values = assignees[1].split(', ')
			if (!values.every(isSimpleUsername)) {
				return null
			}
			filter.assignees = values
			continue
		}
		const priority = /^priority >= ([0-5])$/.exec(clause)
		if (!priority) {
			return null
		}
		filter.minPriority = Number(priority[1]) as Priority
	}
	if (to !== undefined) {
		filter.due = from !== undefined
			? {
				kind: 'between',
				from,
				to,
			} : {
				kind: 'before',
				to,
			}
	}
	return toQuery(filter) === query ? filter : null
}
