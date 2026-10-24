import type {TaskCollection} from '@/client/generated'
import type {TaskResponse} from '@/client/queries/tasks'
import {trimQuotes} from '@/helpers/filters'

export type BucketFilterField = 'assignees' | 'labels'

export type BucketFilterClause = {
	field: BucketFilterField
	negated: boolean
	values: string[]
}

export type BucketFilterEdit = {
	field: BucketFilterField
	value: string
	add: boolean
}

type TaskValues = Pick<TaskResponse, 'assignees' | 'labels'>

const CLAUSE = /^(assignees|labels)\s*(\?!=|!=|\?=|=|not in\b|in\b)\s*(.+)$/

function parseClause(clause: string): BucketFilterClause | null {
	const match = clause.trim().match(CLAUSE)
	if (!match) return null
	const [, field, operator, list] = match
	const negated = ['!=', '?!=', 'not in'].includes(operator)
	const unquoted = trimQuotes(list)
	const values = (/["']/.test(unquoted) ? list : unquoted).split(',').map(trimQuotes).filter(Boolean)
	if (values.length === 0 || (values.length > 1 && !negated)) return null
	if (field === 'labels' && !values.every(value => /^\d+$/.test(value))) return null
	return {
		field: field as BucketFilterField,
		negated,
		values,
	}
}

export function parseBucketFilter(filter?: Pick<TaskCollection, 'filter' | 's'>): BucketFilterClause[] | null {
	const source = filter?.filter?.trim()
	if (!source || filter?.s || /\|\||[()]/.test(source)) return null
	const clauses = source.split('&&').map(parseClause)
	return clauses.every(clause => clause !== null) ? clauses as BucketFilterClause[] : null
}

export function findFilterValue(task: TaskValues, field: BucketFilterField, value: string) {
	return field === 'assignees'
		? task.assignees.find(user => user.username === value)
		: task.labels.find(label => label.id === Number(value))
}

function sameValue(field: BucketFilterField, a: string, b: string) {
	return field === 'labels' ? Number(a) === Number(b) : a === b
}

export function matchesBucketFilter(task: TaskValues, clauses: BucketFilterClause[], includeNulls = false): boolean {
	return clauses.every(({field, negated, values}) => {
		const hit = values.some(value => findFilterValue(task, field, value) !== undefined)
		return negated
			? !hit
			: hit || (includeNulls && task[field].length === 0)
	})
}

export function planBucketFilterMove(
	task: TaskValues,
	from: BucketFilterClause[] | null,
	to: BucketFilterClause[],
): BucketFilterEdit[] {
	const edits: BucketFilterEdit[] = []
	const planned = (field: BucketFilterField, value: string) =>
		edits.some(edit => edit.field === field && sameValue(field, edit.value, value))
	const requiredByTarget = (field: BucketFilterField, value: string) =>
		to.some(clause => !clause.negated && clause.field === field && sameValue(field, clause.values[0], value))
	const has = (field: BucketFilterField, value: string) => findFilterValue(task, field, value) !== undefined

	for (const {field, negated, values} of to) {
		for (const value of values) {
			if (negated === has(field, value) && !planned(field, value)) {
				edits.push({
					field,
					value,
					add: !negated,
				})
			}
		}
	}

	for (const {field, negated, values: [value]} of from ?? []) {
		if (!negated && !requiredByTarget(field, value) && has(field, value) && !planned(field, value)) {
			edits.push({
				field,
				value,
				add: false,
			})
		}
	}
	return edits
}
