import {
	queryOptions,
	skipToken,
	type Query,
} from '@tanstack/vue-query'
import {tasksOccurrencesPreview} from '@/client/generated'
import type {
	PreviewAnswer,
	TaskRepeatWritable,
} from '@/client/generated'
import {ApiError} from '@/client/problem'

export const PREVIEW_COUNT = 10

const RULE_REFUSED_CODE = 2002

// The editor previews an existing task, so the zone comes from task_id; timezone is never sent, or the
// preview could disagree with the save.
type PreviewAnchor =
	| {due_date: string, initial_due_time?: never}
	| {initial_due_time: string, due_date?: never}

export type OccurrencePreviewInput = PreviewAnchor & {
	repeat: TaskRepeatWritable,
	task_id: number,
	repeats_from_current_date?: boolean,
	count?: number,
}

export type PreviewView = {
	dates: string[],
	source?: PreviewAnswer['due_date_source'],
	reason?: string,
	zone: string,
	evaluatedAt: string,
	forecast: boolean,
}

export type PreviewErrorClass = 'ruleHint' | 'auth' | 'unavailable'

// Picks exactly the fields the server reads, so equal requests give equal keys and nothing else reaches the body.
export function previewBody(input: OccurrencePreviewInput): OccurrencePreviewInput {
	const body = {
		repeat: input.repeat,
		due_date: input.due_date,
		initial_due_time: input.initial_due_time,
		task_id: input.task_id,
		repeats_from_current_date: input.repeats_from_current_date,
		count: input.count,
	}
	return Object.fromEntries(Object.entries(body).filter(([, value]) => value !== undefined)) as OccurrencePreviewInput
}

export const occurrenceKeys = {
	all: ['occurrence-preview'] as const,
	task: (taskId: number) => ['occurrence-preview', taskId] as const,
	preview: (input: OccurrencePreviewInput | null) => ['occurrence-preview', input?.task_id ?? null, input ? previewBody(input) : null] as const,
}

type PreviewKey = ReturnType<typeof occurrenceKeys.preview>

export function toPreviewView(answer: PreviewAnswer, repeatsFromCurrentDate: boolean): PreviewView {
	return {
		dates: answer.due_date ? [answer.due_date, ...(answer.occurrences ?? [])] : [],
		source: answer.due_date_source,
		reason: answer.reason,
		zone: answer.timezone,
		evaluatedAt: answer.evaluated_at,
		forecast: repeatsFromCurrentDate,
	}
}

export function classifyPreviewError(cause: unknown): PreviewErrorClass {
	if (!(cause instanceof ApiError)) return 'unavailable'
	if (cause.status === 400 && cause.code === RULE_REFUSED_CODE) return 'ruleHint'
	if (cause.status === 401) return 'auth'
	return 'unavailable'
}

// Only a request the server calls malformed is our bug; 403/404 are expected for some shares.
export function shouldLogFault(cause: unknown): boolean {
	if (!(cause instanceof ApiError)) return false
	return cause.status === 422 || (cause.status === 400 && cause.code !== RULE_REFUSED_CODE)
}

export function occurrencePreviewQuery(input: OccurrencePreviewInput | null) {
	const taskId = input?.task_id ?? null
	return queryOptions<PreviewView, Error, PreviewView, PreviewKey>({
		queryKey: occurrenceKeys.preview(input),
		queryFn: input
			? async ({signal}) => {
				const body = previewBody(input)
				try {
					const {data} = await tasksOccurrencesPreview({body, signal})
					return toPreviewView(data, body.repeats_from_current_date === true)
				} catch (cause) {
					if (shouldLogFault(cause)) console.error('Occurrence preview request refused as malformed', cause)
					throw cause
				}
			}
			: skipToken,
		// Each answer depends on the server clock, and a refused rule stays refused until the user edits it.
		retry: false,
		staleTime: 0,
		gcTime: 30_000,
		// Offline must fail as "unavailable" now, not pause and answer for an old draft later.
		networkMode: 'always',
		meta: {handlesError: true},
		// Not keepPreviousData: it would show another task's dates.
		placeholderData: (
			previousData: PreviewView | undefined,
			previousQuery: Query<PreviewView, Error, PreviewView, PreviewKey> | undefined,
		) => {
			if (!previousQuery || taskId === null) return undefined
			const [, previousTaskId] = previousQuery.queryKey
			return previousTaskId === taskId ? previousData : undefined
		},
	})
}
