import {describe, expect, it} from 'vitest'
import {readdirSync, readFileSync} from 'node:fs'
import {join} from 'node:path'

import type {TaskRepeat} from '@/client/generated'
import {repeatPresets, type RepeatUnit} from '@/helpers/rrule'
import {describeRepeatRule} from '@/helpers/rruleDescribe'

type Messages = {[key: string]: string | Messages}

// The test environment gives import.meta.url an http scheme, so the source root comes from the working directory.
const SRC = join(process.cwd(), 'src')
// Read as text: importing it would give the build's precompiled message functions instead of the source strings.
const en: Messages = JSON.parse(readFileSync(join(SRC, 'i18n/lang/en.json'), 'utf8'))
const ROOTS = ['task.repeat']
const KEY_PATH = `(?:${ROOTS.map(root => root.replace('.', '\\.')).join('|')})(?:\\.[A-Za-z0-9_]+)*`
const LITERAL_KEY = new RegExp(`(['"\`])(${KEY_PATH})\\1`, 'g')
const TEMPLATE_PREFIX = new RegExp(`\`(${KEY_PATH})\\.\\$\\{`, 'g')
const KEY_CONSTANT = /const KEY = '([^']+)'/

const UNITS: RepeatUnit[] = ['hour', 'day', 'week', 'month', 'year']
// The custom form passes a count to these, so they need a singular and a plural form.
const COUNTED_KEYS = [...UNITS.map(unit => `task.repeat.unit.${unit}`), 'task.repeat.custom.times']

function resolve(path: string): string | Messages | undefined {
	return path.split('.').reduce<string | Messages | undefined>(
		(node, part) => typeof node === 'object' ? node[part] : undefined,
		en,
	)
}

function leaves(node: Messages, prefix: string): string[] {
	return Object.entries(node).flatMap(([key, value]) => typeof value === 'string'
		? [`${prefix}.${key}`]
		: leaves(value, `${prefix}.${key}`))
}

function sourceFiles(dir: string): string[] {
	return readdirSync(dir, {withFileTypes: true}).flatMap(entry => {
		const path = join(dir, entry.name)
		if (entry.isDirectory()) {
			return entry.name === 'lang' ? [] : sourceFiles(path)
		}
		const isSource = /\.(ts|vue)$/.test(entry.name) && !/\.(test|spec)\.ts$/.test(entry.name) && !entry.name.endsWith('.story.vue')
		return isSource ? [path] : []
	})
}

function scanSources() {
	const literals = new Set<string>()
	const prefixes = new Set<string>()
	for (const file of sourceFiles(SRC)) {
		let text = readFileSync(file, 'utf8')
		const constant = KEY_CONSTANT.exec(text)
		if (constant) {
			// The constant is a namespace prefix, not a key of its own.
			text = text.replace(constant[0], '').split('${KEY}').join(constant[1])
		}
		for (const match of text.matchAll(LITERAL_KEY)) {
			literals.add(match[2])
		}
		for (const match of text.matchAll(TEMPLATE_PREFIX)) {
			prefixes.add(match[1])
		}
	}
	return {literals, prefixes}
}

const FRIDAY = new Date('2090-10-20T12:00:00Z')

function r(fields: Record<string, unknown>): TaskRepeat {
	return fields as unknown as TaskRepeat
}

// Every sentence shape the describer has, so that every describe key it can ask for is asked for.
const DESCRIBE_CASES: [TaskRepeat, Date | null][] = [
	[r({freq: 'hourly'}), FRIDAY],
	[r({freq: 'daily'}), FRIDAY],
	[r({freq: 'daily', by_day: ['mo', 'we']}), FRIDAY],
	[r({freq: 'weekly'}), null],
	[r({freq: 'weekly'}), FRIDAY],
	[r({freq: 'weekly', by_day: ['mo', 'tu', 'we', 'th', 'fr']}), FRIDAY],
	[r({freq: 'weekly', count: 3}), FRIDAY],
	[r({freq: 'weekly', until: '2090-12-31T23:59:59Z'}), FRIDAY],
	[r({freq: 'monthly'}), null],
	[r({freq: 'monthly', by_month_day: [1]}), FRIDAY],
	[r({freq: 'monthly', by_month_day: [2]}), FRIDAY],
	[r({freq: 'monthly', by_month_day: [3]}), FRIDAY],
	[r({freq: 'monthly', by_month_day: [4]}), FRIDAY],
	[r({freq: 'monthly', by_month_day: [-1]}), FRIDAY],
	[r({freq: 'monthly', by_day: ['2tu']}), FRIDAY],
	[r({freq: 'monthly', by_day: ['-1fr']}), FRIDAY],
	[r({freq: 'monthly', by_day: ['mo', 'tu', 'we', 'th', 'fr'], by_set_pos: [-1]}), FRIDAY],
	[r({freq: 'yearly'}), null],
	[r({freq: 'yearly'}), FRIDAY],
	...['secondly', 'minutely', 'hourly', 'daily', 'weekly', 'monthly', 'yearly'].flatMap((freq): [TaskRepeat, Date][] => [
		[r({freq, interval: 2, by_week_no: [1]}), FRIDAY],
		[r({freq, by_week_no: [1]}), FRIDAY],
	]),
	[r({freq: 'weekly', by_day: ['mo'], wkst: 'su'}), FRIDAY],
	[r({freq: 'monthly', by_day: ['mo', 'tu']}), FRIDAY],
	[r({freq: 'monthly', by_day: ['1mo', '-1fr', '-2tu']}), FRIDAY],
	[r({freq: 'monthly', by_month_day: [1, 15]}), FRIDAY],
	[r({freq: 'yearly', by_month: [3, 9], count: 2}), null],
	[r({freq: 'yearly', by_month: [3, 9], count: 0}), null],
	[r({freq: 'yearly', by_month: [3, 9], until: '2091-01-01T00:00:00Z'}), null],
	[r({freq: 'yearly', by_year_day: [100]}), FRIDAY],
	[r({freq: 'daily', by_hour: [9]}), FRIDAY],
	[r({freq: 'yearly', read_only: true}), FRIDAY],
]

function describerKeys() {
	const asked = new Map<string, boolean>()
	const t = (key: string, _params?: Record<string, unknown>, plural?: number) => {
		asked.set(key, (asked.get(key) ?? false) || plural !== undefined)
		return key
	}
	for (const [repeat, dueDate] of DESCRIBE_CASES) {
		describeRepeatRule(repeat, {t, locale: 'en', weekStart: 1, zone: 'UTC', dueDate})
	}
	return asked
}

function isPluralMessage(message: unknown): boolean {
	return typeof message === 'string' && message.split('|').length === 2 && message.split('|').every(form => form.trim() !== '')
}

describe('repeat translation keys', () => {
	const {literals, prefixes} = scanSources()
	const describeAsked = describerKeys()
	const presetKeys = [new Date('2090-10-20T12:00:00Z'), new Date('2090-10-30T12:00:00Z')]
		.flatMap(anchor => repeatPresets(anchor, 'UTC').map(preset => preset.labelKey))

	it('finds the keys the editor uses in the source', () => {
		expect(literals).toContain('task.repeat.customOption')
		expect(literals).toContain('task.repeat.error.noOccurrence')
		expect(literals).toContain('task.repeat.sheetTitle')
		expect(literals).toContain('task.repeat.lockedNote')
		expect(prefixes).toContain('task.repeat.unit')
		expect(prefixes).toContain('task.repeat.describe.part.every')
	})

	it('resolves every literal key to a message, never to a namespace', () => {
		const missing = [...literals].filter(key => typeof resolve(key) !== 'string')
		expect(missing).toEqual([])
	})

	it('resolves every key prefix built at runtime to a namespace', () => {
		const notNamespaces = [...prefixes].filter(prefix => {
			const node = resolve(prefix)
			return typeof node !== 'object' || Object.keys(node).length === 0
		})
		expect(notNamespaces).toEqual([])
	})

	it('has a message for every key the describer and the presets ask for', () => {
		const asked = [...describeAsked.keys(), ...presetKeys, ...COUNTED_KEYS]
		expect(asked.filter(key => typeof resolve(key) !== 'string')).toEqual([])
	})

	it('has no repeat message that nothing asks for', () => {
		const used = new Set([...literals, ...describeAsked.keys(), ...presetKeys, ...COUNTED_KEYS])
		const unused = leaves(resolve('task.repeat') as Messages, 'task.repeat').filter(key => !used.has(key))
		expect(unused).toEqual([])
	})

	it('gives every counted message a singular and a plural form', () => {
		const counted = [...[...describeAsked].filter(([, plural]) => plural).map(([key]) => key), ...COUNTED_KEYS]
		expect(counted.length).toBeGreaterThan(COUNTED_KEYS.length)
		expect(counted.filter(key => !isPluralMessage(resolve(key)))).toEqual([])
	})

	it('only uses the plural separator in well-formed plural messages', () => {
		const malformed = leaves(resolve('task.repeat') as Messages, 'task.repeat')
			.filter(key => (resolve(key) as string).includes('|') && !isPluralMessage(resolve(key)))
		expect(malformed).toEqual([])
	})
})
