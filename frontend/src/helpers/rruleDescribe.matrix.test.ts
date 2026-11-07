import {describe, expect, it} from 'vitest'
import {readFileSync} from 'node:fs'
import {join} from 'node:path'
import {createI18n} from 'vue-i18n'

import type {TaskRepeat} from '@/client/generated'
import {describeRepeatRule, type DescribeOptions} from './rruleDescribe'

interface MatrixRow {
	name: string
	user_zone: string
	repeat: TaskRepeat
	due: string
	preview_eligible: boolean
	summary: string
}

// The backend's golden matrix also pins the English summary the editor shows for each rule.
const matrix = JSON.parse(readFileSync(join(process.cwd(), '../pkg/models/testdata/repeat_preview_matrix.json'), 'utf8')) as {version: number, rows: MatrixRow[]}
const rows = matrix.rows.filter(row => row.preview_eligible)

const en = JSON.parse(readFileSync(join(process.cwd(), 'src/i18n/lang/en.json'), 'utf8'))
type Translator = (key: string, named: Record<string, unknown>, plural?: number) => string
// vue-i18n's composer type over untyped messages exceeds the compiler's instantiation depth.
const i18n = createI18n({legacy: false, locale: 'en', messages: {en}, missingWarn: false, fallbackWarn: false}) as unknown as {global: {t: Translator}}
const t = i18n.global.t
const translate: DescribeOptions['t'] = (key, params, plural) => plural === undefined
	? t(key, params ?? {})
	: t(key, params ?? {}, plural)

describe('golden repeat matrix summaries', () => {
	it('has the rows the agreement depends on', () => {
		expect(matrix.version).toBe(1)
		expect(rows.length).toBeGreaterThanOrEqual(27)
		for (const row of rows) {
			expect(row.summary, row.name).not.toBe('')
		}
	})

	it.each(rows.map(row => [row.name, row] as const))('summarises %s', (_name, row) => {
		const summary = describeRepeatRule(row.repeat, {
			t: translate,
			locale: 'en',
			weekStart: 1,
			zone: row.user_zone,
			dueDate: new Date(row.due),
		})
		// ICU versions differ on the space before AM/PM (U+202F or a plain space); the fixture uses a plain one.
		expect(summary.replace(/[ \u202F]/g, ' ')).toBe(row.summary)
	})
})
