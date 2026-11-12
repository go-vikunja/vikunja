import type {
	PreviewAnswer,
	PreviewInputWritable,
} from '@/client/generated'

// Answers copied from the server's preview test rows. Nothing here is computed: each instant
// is the row's expected value written out, and each entry names the server test it comes from.
// The HTTP rows send timezone to pin the zone; the editor never sends it, so it is left out of the
// request here and appears only as the answer's timezone.

export type PreviewRefusal = {
	status: number,
	code?: number,
}

export type PreviewFixture = {
	row: string,
	request: PreviewInputWritable,
	answer: PreviewAnswer | PreviewRefusal,
}

export const previewFixtures: PreviewFixture[] = [
	{
		// pkg/webtests/huma_task_occurrences_test.go TestOccurrencePreview_InitialDueRows "P-1 DAILY 09:00, now 10:00"; Berlin, now 2026-10-28 10:00 CET.
		row: 'P-1',
		request: {
			repeats_rrule: 'FREQ=DAILY',
			initial_due_time: '09:00',
			repeats_from_current_date: false,
			count: 2,
		},
		answer: {
			due_date: '2026-10-29T09:00:00+01:00',
			due_date_source: 'generated',
			occurrences: ['2026-10-30T09:00:00+01:00', '2026-10-31T09:00:00+01:00'],
			timezone: 'Europe/Berlin',
			evaluated_at: '2026-10-28T10:00:00+01:00',
		},
	},
	{
		// Same test, "P-5 COUNT=1, seed before now"; Berlin, now 2026-10-28 10:00 CET.
		row: 'P-5',
		request: {
			repeats_rrule: 'FREQ=DAILY;COUNT=1',
			initial_due_time: '09:00',
			repeats_from_current_date: false,
			count: 10,
		},
		answer: {
			due_date: '2026-10-29T09:00:00+01:00',
			due_date_source: 'generated',
			occurrences: [],
			reason: 'count_exhausted',
			timezone: 'Europe/Berlin',
			evaluated_at: '2026-10-28T10:00:00+01:00',
		},
	},
	{
		// Same test, "P-7b now <= UNTIL < first candidate"; Berlin, now 2026-10-28 10:00 CET.
		row: 'P-7b',
		request: {
			repeats_rrule: 'FREQ=DAILY;UNTIL=20261029T075959Z',
			initial_due_time: '09:00',
			repeats_from_current_date: false,
			count: 3,
		},
		answer: {
			occurrences: [],
			reason: 'no_occurrence',
			timezone: 'Europe/Berlin',
			evaluated_at: '2026-10-28T10:00:00+01:00',
		},
	},
	{
		// Same test, "P-7c UNTIL < now"; Berlin, now 2026-10-28 10:00 CET.
		row: 'P-7c',
		request: {
			repeats_rrule: 'FREQ=DAILY;UNTIL=20261027T000000Z',
			initial_due_time: '09:00',
			repeats_from_current_date: false,
			count: 3,
		},
		answer: {
			occurrences: [],
			reason: 'until_passed',
			timezone: 'Europe/Berlin',
			evaluated_at: '2026-10-28T10:00:00+01:00',
		},
	},
	{
		// Same test, "P-13 HOURLY from completion, 18:00 at 10:00"; Berlin, now 2026-10-28 10:00 CET.
		row: 'P-13',
		request: {
			repeats_rrule: 'FREQ=HOURLY;INTERVAL=1',
			initial_due_time: '18:00',
			repeats_from_current_date: true,
			count: 3,
		},
		answer: {
			due_date: '2026-10-28T18:00:00+01:00',
			due_date_source: 'generated',
			occurrences: ['2026-10-28T19:00:00+01:00', '2026-10-28T20:00:00+01:00', '2026-10-28T21:00:00+01:00'],
			timezone: 'Europe/Berlin',
			evaluated_at: '2026-10-28T10:00:00+01:00',
		},
	},
	{
		// TestOccurrencePreview_ExplicitDueMatchesQuery (P-10); caller zone Berlin, now 2026-10-28 10:00 CET. The due date is echoed to the second.
		row: 'P-10',
		request: {
			repeats_rrule: 'FREQ=WEEKLY;BYDAY=MO,TH',
			due_date: '2026-10-29T09:00:00.4+01:00',
			count: 4,
		},
		answer: {
			due_date: '2026-10-29T09:00:00+01:00',
			due_date_source: 'explicit',
			occurrences: [
				'2026-11-02T09:00:00+01:00',
				'2026-11-05T09:00:00+01:00',
				'2026-11-09T09:00:00+01:00',
				'2026-11-12T09:00:00+01:00',
			],
			timezone: 'Europe/Berlin',
			evaluated_at: '2026-10-28T10:00:00+01:00',
		},
	},
	{
		// pkg/models/task_initial_due_oracle_test.go TestRepeatPreviewMatrixPreviewParity over testdata/repeat_preview_matrix.json
		// "monthly implicit from the 15th"; zone UTC, evaluated one second before the due date.
		row: 'matrix: monthly implicit from the 15th',
		request: {
			repeat: {freq: 'monthly'},
			due_date: '2093-01-15T12:00:00Z',
			count: 3,
		},
		answer: {
			due_date: '2093-01-15T12:00:00Z',
			due_date_source: 'explicit',
			occurrences: ['2093-02-15T12:00:00Z', '2093-03-15T12:00:00Z', '2093-04-15T12:00:00Z'],
			timezone: 'UTC',
			evaluated_at: '2093-01-15T11:59:59Z',
		},
	},
	{
		// Same test and matrix, "yearly implicit"; zone America/New_York, evaluated one second before the due date.
		row: 'matrix: yearly implicit',
		request: {
			repeat: {freq: 'yearly'},
			due_date: '2091-03-14T09:00:00-04:00',
			count: 2,
		},
		answer: {
			due_date: '2091-03-14T09:00:00-04:00',
			due_date_source: 'explicit',
			occurrences: ['2092-03-14T09:00:00-04:00', '2093-03-14T09:00:00-04:00'],
			timezone: 'America/New_York',
			evaluated_at: '2091-03-14T08:59:59-04:00',
		},
	},
	{
		// TestOccurrencePreview_StatusMatrix "P-8 due_date and initial_due_time".
		row: 'P-8',
		request: {
			repeats_rrule: 'FREQ=DAILY',
			due_date: '2026-10-29T09:00:00+01:00',
			initial_due_time: '09:00',
			count: 2,
		},
		answer: {status: 422},
	},
	{
		// TestOccurrencePreview_StatusMatrix "P-8 malformed initial_due_time 24:00".
		row: 'P-8',
		request: {
			repeats_rrule: 'FREQ=DAILY',
			initial_due_time: '24:00',
			count: 2,
		},
		answer: {status: 422},
	},
	{
		// TestOccurrencePreview_StatusMatrix "rule a save refuses (initial_due_time)" (status matrix: rule a task save would refuse).
		row: 'status matrix: refused rule',
		request: {
			repeats_rrule: 'FREQ=HOURLY;BYHOUR=9',
			initial_due_time: '09:00',
			count: 2,
		},
		answer: {status: 400, code: 2002},
	},
	{
		// TestOccurrencePreview_ZoneContextAccess (P-9) "task_id unreadable is refused as the read is".
		row: 'P-9',
		request: {
			repeats_rrule: 'FREQ=DAILY',
			due_date: '2026-10-29T09:00:00Z',
			task_id: 34,
			count: 1,
		},
		answer: {status: 403},
	},
]

// PreviewAnswer types occurrences as nullable although the server always sends an array. This is P-5's
// answer with that one field set to null, kept out of answerFor so the row above still answers P-5.
export const occurrencesNullAnswer: PreviewAnswer = {
	due_date: '2026-10-29T09:00:00+01:00',
	due_date_source: 'generated',
	occurrences: null,
	reason: 'count_exhausted',
	timezone: 'Europe/Berlin',
	evaluated_at: '2026-10-28T10:00:00+01:00',
}

export function isRefusal(answer: PreviewFixture['answer']): answer is PreviewRefusal {
	return 'status' in answer
}

function sameValue(a: unknown, b: unknown): boolean {
	if (a === b) return true
	if (typeof a !== 'object' || typeof b !== 'object' || a === null || b === null) return false
	if (Array.isArray(a) !== Array.isArray(b)) return false
	const defined = (value: object) => Object.entries(value).filter(([, entry]) => entry !== undefined)
	const left = defined(a)
	const right = new Map(defined(b))
	return left.length === right.size && left.every(([key, entry]) => right.has(key) && sameValue(entry, right.get(key)))
}

export function answerFor(request: PreviewInputWritable): PreviewFixture['answer'] | undefined {
	return previewFixtures.find(fixture => sameValue(fixture.request, request))?.answer
}
