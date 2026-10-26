import {afterEach, describe, expect, it} from 'vitest'
import {readFileSync} from 'node:fs'
import {join} from 'node:path'
import {createI18n} from 'vue-i18n'

import type {TaskRepeat} from '@/client/generated'
import {previewDates} from './rrulePreview'
import {describeRepeatRule, type DescribeOptions} from './rruleDescribe'

interface MatrixStep {
	due: string
	count: number | null
	cleared: boolean
	done: boolean
}

interface MatrixRow {
	name: string
	user_zone: string
	server_zone: string
	repeat: TaskRepeat
	due: string
	completions: number
	preview_eligible: boolean
	summary: string
	expect: MatrixStep[]
}

// The backend's golden matrix: its chains come from running real completion, so the preview must match them.
const matrix = JSON.parse(readFileSync(join(process.cwd(), '../pkg/models/testdata/repeat_preview_matrix.json'), 'utf8')) as {version: number, rows: MatrixRow[]}
const rows = matrix.rows.filter(row => row.preview_eligible)
const zonesDiffer = rows.filter(row => row.user_zone !== row.server_zone)

const en = JSON.parse(readFileSync(join(process.cwd(), 'src/i18n/lang/en.json'), 'utf8'))
type Translator = (key: string, named: Record<string, unknown>, plural?: number) => string
// vue-i18n's composer type over untyped messages exceeds the compiler's instantiation depth.
const i18n = createI18n({legacy: false, locale: 'en', messages: {en}, missingWarn: false, fallbackWarn: false}) as unknown as {global: {t: Translator}}
const t = i18n.global.t
const translate: DescribeOptions['t'] = (key, params, plural) => plural === undefined
	? t(key, params ?? {})
	: t(key, params ?? {}, plural)

const PREVIEW_LENGTH = 3

// The preview is the current due followed by the dues completion assigns, until a completion leaves the task done.
function expectedInstants(row: MatrixRow): number[] {
	const next: number[] = []
	for (const step of row.expect) {
		if (step.done) {
			break
		}
		next.push(Date.parse(step.due))
	}
	return [Date.parse(row.due), ...next].slice(0, PREVIEW_LENGTH)
}

function previewInstants(row: MatrixRow): number[] | null {
	const due = new Date(row.due)
	const now = new Date(due.getTime() - 1000)
	return previewDates(row.repeat, due, now, false, row.user_zone).dates?.map(d => d.getTime()) ?? null
}

function iso(instants: number[] | null): string[] | null {
	return instants?.map(ms => new Date(ms).toISOString()) ?? null
}

const originalTZ = process.env.TZ
afterEach(() => {
	process.env.TZ = originalTZ
})

describe('golden repeat matrix', () => {
	it('has the rows the agreement depends on', () => {
		expect(matrix.version).toBe(1)
		expect(rows.length).toBeGreaterThanOrEqual(27)
		expect(zonesDiffer.length).toBeGreaterThan(0)
		// Without two compared steps a row only checks the due it was given.
		for (const row of rows) {
			const ended = row.expect.some(step => step.done)
			expect(ended || row.expect.length >= PREVIEW_LENGTH - 1, row.name).toBe(true)
		}
	})

	describe.each(rows.map(row => [row.name, row] as const))('%s', (_name, row) => {
		it('previews the dues completion assigns', () => {
			process.env.TZ = row.user_zone
			expect(iso(previewInstants(row))).toEqual(iso(expectedInstants(row)))
		})

		it('summarises the rule', () => {
			const summary = describeRepeatRule(row.repeat, {
				t: translate,
				locale: 'en',
				weekStart: 1,
				zone: row.user_zone,
				dueDate: new Date(row.due),
			})
			// ICU versions differ on the space before AM/PM (U+202F or a plain space); the fixture uses a plain one.
			expect(summary.replace(/[  ]/g, ' ')).toBe(row.summary)
		})
	})

	it.each(zonesDiffer.map(row => [row.name, row] as const))('previews %s the same with the process zone in UTC', (_name, row) => {
		process.env.TZ = 'UTC'
		expect(iso(previewInstants(row))).toEqual(iso(expectedInstants(row)))
	})
})
