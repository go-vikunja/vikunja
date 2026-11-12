import {
	beforeEach,
	describe,
	expect,
	it,
	vi,
} from 'vitest'
import {
	hashKey,
	QueryClient,
} from '@tanstack/vue-query'
import {ApiError} from '@/client/problem'
import type {PreviewAnswer} from '@/client/generated'
import {
	classifyPreviewError,
	occurrenceKeys,
	occurrencePreviewQuery,
	shouldLogFault,
	toPreviewView,
	type OccurrencePreviewInput,
} from './occurrencePreview'
import {
	answerFor,
	isRefusal,
	occurrencesNullAnswer,
	previewFixtures,
} from './occurrencePreview.fixtures'

const sdk = vi.hoisted(() => ({
	tasksOccurrencesPreview: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)

function answerOf(row: string): PreviewAnswer {
	const answer = previewFixtures.find(fixture => fixture.row === row)?.answer
	if (!answer || isRefusal(answer)) throw new Error(`no answer for ${row}`)
	return answer
}

const undated: OccurrencePreviewInput = {
	repeat: {freq: 'daily'},
	initial_due_time: '09:00',
	task_id: 7,
	repeats_from_current_date: false,
	count: 10,
}

function apiError(status: number, code?: number) {
	return new ApiError({status, code, detail: `HTTP ${status}`})
}

let client: QueryClient
beforeEach(() => {
	vi.clearAllMocks()
	client = new QueryClient()
})

describe('occurrencePreviewQuery', () => {
	it('sends one anchor and the task, never a timezone, and forwards the abort signal', async () => {
		sdk.tasksOccurrencesPreview.mockResolvedValue({data: answerOf('P-1')})
		const smuggled = {...undated, timezone: 'Europe/Berlin', project_id: 3} as OccurrencePreviewInput

		await client.fetchQuery(occurrencePreviewQuery(smuggled))

		expect(sdk.tasksOccurrencesPreview).toHaveBeenCalledWith({
			body: {
				repeat: {freq: 'daily'},
				initial_due_time: '09:00',
				task_id: 7,
				repeats_from_current_date: false,
				count: 10,
			},
			signal: expect.any(AbortSignal),
		})
	})

	it('sends an explicit due date instead of a time when the task is dated', async () => {
		sdk.tasksOccurrencesPreview.mockResolvedValue({data: answerOf('P-10')})
		const dated: OccurrencePreviewInput = {
			repeat: {freq: 'weekly', by_day: ['mo', 'th']},
			due_date: '2026-10-29T09:00:00+01:00',
			task_id: 7,
		}

		await client.fetchQuery(occurrencePreviewQuery(dated))

		const [{body}] = sdk.tasksOccurrencesPreview.mock.calls[0]
		expect(body).toEqual({
			repeat: {freq: 'weekly', by_day: ['mo', 'th']},
			due_date: '2026-10-29T09:00:00+01:00',
			task_id: 7,
		})
		expect(body).not.toHaveProperty('initial_due_time')
	})

	it('aborts the request when the query is cancelled', async () => {
		let received: AbortSignal | undefined
		sdk.tasksOccurrencesPreview.mockImplementation(({signal}: {signal: AbortSignal}) => {
			received = signal
			return new Promise(() => {})
		})
		void client.fetchQuery(occurrencePreviewQuery(undated)).catch(() => {})
		await vi.waitFor(() => expect(received).toBeDefined())

		await client.cancelQueries({queryKey: occurrenceKeys.task(7)})

		expect(received?.aborted).toBe(true)
	})

	it('keys equal requests equally and every server-visible input differently', () => {
		const key = hashKey(occurrenceKeys.preview(undated))
		expect(hashKey(occurrenceKeys.preview({...undated}))).toBe(key)
		expect(hashKey(occurrenceKeys.preview({...undated, due_date: undefined} as OccurrencePreviewInput))).toBe(key)
		for (const changed of [
			{...undated, repeat: {freq: 'weekly'}},
			{...undated, initial_due_time: '10:00'},
			{...undated, task_id: 8},
			{...undated, repeats_from_current_date: true},
			{...undated, count: 5},
			{repeat: undated.repeat, due_date: '2026-10-29T09:00:00Z', task_id: 7, repeats_from_current_date: false, count: 10},
		] as OccurrencePreviewInput[]) {
			expect(hashKey(occurrenceKeys.preview(changed))).not.toBe(key)
		}
		expect(occurrenceKeys.preview(undated).slice(0, 2)).toEqual(occurrenceKeys.task(7))
	})

	it('overrides the app defaults: no retry, never fresh, short cache, always online, own error handling', () => {
		const options = occurrencePreviewQuery(undated)
		expect(options.retry).toBe(false)
		expect(options.staleTime).toBe(0)
		expect(options.gcTime).toBeLessThanOrEqual(60_000)
		expect(options.networkMode).toBe('always')
		expect(options.meta).toEqual({handlesError: true})
	})

	it('keeps the previous answer as placeholder for the same task only', () => {
		const previous = toPreviewView(answerOf('P-1'), false)
		const placeholder = occurrencePreviewQuery(undated).placeholderData as (data: unknown, query: unknown) => unknown

		expect(placeholder(previous, {queryKey: occurrenceKeys.preview({...undated, initial_due_time: '08:00'})})).toBe(previous)
		expect(placeholder(previous, {queryKey: occurrenceKeys.preview({...undated, task_id: 8})})).toBeUndefined()
		expect(placeholder(previous, undefined)).toBeUndefined()
	})

	it('does not request anything without input', () => {
		expect(occurrencePreviewQuery(null).enabled).not.toBe(true)
		expect(typeof occurrencePreviewQuery(null).queryFn).not.toBe('function')
	})

	it('logs a fault for a malformed request and rethrows', async () => {
		const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
		const refused = apiError(422)
		sdk.tasksOccurrencesPreview.mockRejectedValue(refused)

		await expect(client.fetchQuery(occurrencePreviewQuery(undated))).rejects.toBe(refused)

		expect(consoleError).toHaveBeenCalledOnce()
		consoleError.mockRestore()
	})

	it('does not log an expected refusal', async () => {
		const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
		sdk.tasksOccurrencesPreview.mockRejectedValue(apiError(403))

		await expect(client.fetchQuery(occurrencePreviewQuery(undated))).rejects.toBeInstanceOf(ApiError)

		expect(consoleError).not.toHaveBeenCalled()
		consoleError.mockRestore()
	})
})

describe('toPreviewView', () => {
	it('lists the due date first, then the occurrences, as returned', () => {
		expect(toPreviewView(answerOf('P-1'), false)).toEqual({
			dates: ['2026-10-29T09:00:00+01:00', '2026-10-30T09:00:00+01:00', '2026-10-31T09:00:00+01:00'],
			source: 'generated',
			reason: undefined,
			zone: 'Europe/Berlin',
			evaluatedAt: '2026-10-28T10:00:00+01:00',
			forecast: false,
		})
	})

	it('reads null occurrences as none after the due date', () => {
		expect(toPreviewView(occurrencesNullAnswer, false).dates).toEqual(['2026-10-29T09:00:00+01:00'])
	})

	it('has no dates and keeps the reason when no due date was selected', () => {
		const view = toPreviewView(answerOf('P-7c'), false)
		expect(view.dates).toEqual([])
		expect(view.source).toBeUndefined()
		expect(view.reason).toBe('until_passed')
	})

	it('keeps a reason beside a due date', () => {
		expect(toPreviewView(answerOf('P-5'), false)).toMatchObject({
			dates: ['2026-10-29T09:00:00+01:00'],
			reason: 'count_exhausted',
		})
	})

	it('marks a from-completion rule as a forecast', () => {
		expect(toPreviewView(answerOf('P-13'), true).forecast).toBe(true)
	})

	it('keeps monthly and yearly explicit answers as returned', () => {
		expect(toPreviewView(answerOf('matrix: monthly implicit from the 15th'), false)).toMatchObject({
			dates: ['2093-01-15T12:00:00Z', '2093-02-15T12:00:00Z', '2093-03-15T12:00:00Z', '2093-04-15T12:00:00Z'],
			source: 'explicit',
			zone: 'UTC',
		})
		expect(toPreviewView(answerOf('matrix: yearly implicit'), false).dates).toEqual([
			'2091-03-14T09:00:00-04:00',
			'2092-03-14T09:00:00-04:00',
			'2093-03-14T09:00:00-04:00',
		])
	})
})

describe('classifyPreviewError', () => {
	it.each([
		['400 with 2002', apiError(400, 2002), 'ruleHint', false],
		['401', apiError(401), 'auth', false],
		['400 without a code', apiError(400), 'unavailable', true],
		['400 with 2003', apiError(400, 2003), 'unavailable', true],
		['422', apiError(422), 'unavailable', true],
		['403', apiError(403), 'unavailable', false],
		['404', apiError(404), 'unavailable', false],
		['429', apiError(429), 'unavailable', false],
		['500', apiError(500), 'unavailable', false],
		['503', apiError(503), 'unavailable', false],
		['2002 on another status', apiError(422, 2002), 'unavailable', true],
		['a plain error', new Error('boom'), 'unavailable', false],
		['a network failure', new TypeError('Failed to fetch'), 'unavailable', false],
		['a problem body that is not an ApiError', {status: 400, code: 2002}, 'unavailable', false],
		['nothing', undefined, 'unavailable', false],
	])('%s', (_name, cause, kind, logged) => {
		expect(classifyPreviewError(cause)).toBe(kind)
		expect(shouldLogFault(cause)).toBe(logged)
	})

	it('maps every refusal in the fixtures', () => {
		const kinds = previewFixtures
			.filter(fixture => isRefusal(fixture.answer))
			.map(fixture => {
				const {status, code} = fixture.answer as {status: number, code?: number}
				return [fixture.row, classifyPreviewError(apiError(status, code))]
			})
		expect(kinds).toEqual([
			['P-8', 'unavailable'],
			['P-8', 'unavailable'],
			['status matrix: refused rule', 'ruleHint'],
			['P-9', 'unavailable'],
		])
	})
})

describe('answerFor', () => {
	it('answers a request equal to a row, ignoring absent fields', () => {
		expect(answerFor({
			repeats_rrule: 'FREQ=DAILY',
			initial_due_time: '09:00',
			repeats_from_current_date: false,
			count: 2,
			task_id: undefined,
		})).toBe(answerOf('P-1'))
	})

	it('answers nothing for a request no row covers', () => {
		expect(answerFor({repeats_rrule: 'FREQ=DAILY', initial_due_time: '09:00', count: 3})).toBeUndefined()
	})
})
