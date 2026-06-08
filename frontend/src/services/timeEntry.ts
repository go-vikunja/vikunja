import {AuthenticatedHTTPFactory, getApiBaseUrl} from '@/helpers/fetcher'
import {objectToCamelCase, objectToSnakeCase} from '@/helpers/case'

import type {ITimeEntry} from '@/modelTypes/ITimeEntry'

// Time tracking is the first frontend feature on /api/v2, while the shared
// AuthenticatedHTTPFactory pins baseURL to /api/v1. We hand axios absolute v2
// URLs to bypass that. Bespoke and intentionally a bit dirty — to be folded
// into the proper service layer once the frontend moves fully onto v2.
function v2Url(path: string): string {
	const v2Base = getApiBaseUrl().replace(/\/api\/v1\/$/, '/api/v2/')
	return new URL(v2Base + path, window.location.origin).toString()
}

export function parseTimeEntry(raw: Record<string, unknown>): ITimeEntry {
	const e = objectToCamelCase(raw)
	const end = e.endTime as string | null | undefined
	return {
		id: e.id,
		userId: e.userId,
		taskId: e.taskId ?? 0,
		projectId: e.projectId ?? 0,
		startTime: new Date(e.startTime),
		// null end_time = a running timer.
		endTime: end ? new Date(end) : null,
		comment: e.comment ?? '',
		created: new Date(e.created),
		updated: new Date(e.updated),
		maxPermission: e.maxPermission ?? null,
	}
}

export interface TimeEntryListParams {
	filter?: string
	filterTimezone?: string
	q?: string
	page?: number
	perPage?: number
}

export interface TimeEntryListResult {
	items: ITimeEntry[]
	total: number
	page: number
	perPage: number
	totalPages: number
}

export function useTimeEntryService() {
	const http = AuthenticatedHTTPFactory()

	async function getAll(params: TimeEntryListParams = {}): Promise<TimeEntryListResult> {
		const {data} = await http.get(v2Url('time-entries'), {
			params: {
				filter: params.filter,
				filter_timezone: params.filterTimezone,
				q: params.q,
				page: params.page,
				per_page: params.perPage,
			},
		})
		return {
			items: (data.items ?? []).map(parseTimeEntry),
			total: data.total,
			page: data.page,
			perPage: data.per_page,
			totalPages: data.total_pages,
		}
	}

	async function create(entry: Partial<ITimeEntry>): Promise<ITimeEntry> {
		const {data} = await http.post(v2Url('time-entries'), objectToSnakeCase(entry))
		return parseTimeEntry(data)
	}

	async function update(entry: Partial<ITimeEntry> & {id: number}): Promise<ITimeEntry> {
		const {data} = await http.put(v2Url(`time-entries/${entry.id}`), objectToSnakeCase(entry))
		return parseTimeEntry(data)
	}

	async function remove(id: number): Promise<void> {
		await http.delete(v2Url(`time-entries/${id}`))
	}

	async function stopTimer(): Promise<ITimeEntry> {
		const {data} = await http.post(v2Url('time-entries/timer/stop'))
		return parseTimeEntry(data)
	}

	return {getAll, create, update, remove, stopTimer}
}
