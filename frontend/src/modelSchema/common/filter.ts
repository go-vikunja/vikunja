import type {TypeOf} from 'zod'
import {z, nativeEnum, array, boolean, object, number} from 'zod'

export enum SORT_BY {
	ID = 'id',
	DONE = 'done',
	TITLE = 'title',
	PRIORITY = 'priority',
	DONE_AT = 'done_at',
	DUE_DATE = 'due_date',
	START_DATE = 'start_date',
	END_DATE = 'end_date',
	PERCENT_DONE = 'percent_done',
	CREATED = 'created',
	UPDATED = 'updated',
	POSITION = 'position',
	KANBAN_POSITION = 'kanban_position',
 }

export enum ORDER_BY {
	ASC = 'asc',
	DESC = 'desc',
	NONE = 'none',
}

export enum FILTER_BY {
	DONE = 'done',
	DUE_DATE = 'due_date',
	START_DATE = 'start_date',
	END_DATE = 'end_date',
	NAMESPACE = 'namespace',
	ASSIGNEES = 'assignees',
	LIST_ID = 'list_id',
	BUCKET_ID = 'bucket_id',
	PRIORITY = 'priority',
	PERCENT_DONE = 'percent_done',
	LABELS = 'labels',
	UNDEFINED = 'undefined', 	// FIXME: Why do we have a value that is undefined as string?
}

export enum FILTER_COMPARATOR {
	EQUALS = 'equals',
	LESS = 'less',
	GREATER = 'greater',
	GREATER_EQUALS = 'greater_equals',
	LESS_EQUALS = 'less_equals',
	IN = 'in',
}

export enum FILTER_CONCAT {
	AND = 'and',
	OR = 'or',
	IN = 'in',
}

const TASKS_PER_BUCKET = 25

export const FilterSchema = object({
	sortBy: array(nativeEnum(SORT_BY)).default([SORT_BY.DONE, SORT_BY.ID]), // FIXME: create from taskSchema,
	// fixme default order seem so also be `desc`
	// see line from ListTable:
	// 	if (typeof order === 'undefined' || order === 'none') {
	orderBy: array(nativeEnum(ORDER_BY)).default([ORDER_BY.ASC, ORDER_BY.DESC]),
	// FIXME: create from taskSchema
	filterBy: array(nativeEnum(FILTER_BY)).default([FILTER_BY.DONE]),
	// FIXME: create from taskSchema
	// FIXME: might need to preprocess values, e.g. date.
	// see line from 'filters.vue':
	// params.filter_value = params.filter_value.map(v => v instanceof Date ? v.toISOString() : v)
	filterValue: array(z.enum(['false'])).default(['false']),
	// FIXME: is `in` value correct?
	// found in `quick-actions.vue`:
	// params.filter_comparator.push('in')
	filterComparator: array(nativeEnum(FILTER_COMPARATOR)).default([FILTER_COMPARATOR.EQUALS]),
	filterConcat: z.nativeEnum(FILTER_CONCAT).default(FILTER_CONCAT.AND),
	filterIncludeNulls: boolean().default(true),
	perPage: number().default(TASKS_PER_BUCKET), // FIXME: is perPage is just available for the bucket endpoint?
})

export type IFilter = TypeOf<typeof FilterSchema>