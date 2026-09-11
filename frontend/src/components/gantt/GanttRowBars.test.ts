import {afterEach, beforeEach, describe, expect, it} from 'vitest'
import {mount} from '@vue/test-utils'
import {createPinia, setActivePinia} from 'pinia'
import {createI18n} from 'vue-i18n'

import GanttRowBars from './GanttRowBars.vue'
import en from '@/i18n/lang/en.json'
import {i18n as globalI18n} from '@/i18n'
import {useAuthStore} from '@/stores/auth'
import {formatDate} from '@/helpers/time/formatDate'
import type {GanttBarModel} from '@/composables/useGanttBar'

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

function setLocaleAndTimezone(locale: string, timezone: string) {
	globalI18n.global.locale.value = locale as never
	useAuthStore().setUserSettings({timezone} as never)
}

const START = new Date('2026-09-07T12:00:00Z')
const END = new Date('2026-09-10T12:00:00Z')

function makeBar(): GanttBarModel {
	return {
		id: '1',
		start: new Date(START),
		end: new Date(END),
		meta: {
			label: 'Test task',
			hasActualDates: true,
			dateType: 'both',
		},
	}
}

function mountBars(bars: GanttBarModel[]) {
	return mount(GanttRowBars, {
		props: {
			bars,
			totalWidth: 500,
			dateFromDate: new Date('2026-09-01T00:00:00Z'),
			dateToDate: new Date('2026-09-15T00:00:00Z'),
			dayWidthPixels: 30,
			isDragging: false,
			isResizing: false,
			dragState: null,
			focusedRow: null,
			focusedCell: null,
			rowId: 'row-0',
			isParent: false,
			isCollapsed: false,
		},
		global: {plugins: [i18n]},
	})
}

describe('GanttRowBars Jalali aria', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
	})

	afterEach(() => {
		globalI18n.global.locale.value = 'en'
	})

	it('uses the central Jalali formatter in Asia/Tehran', () => {
		setLocaleAndTimezone('fa-IR', 'Asia/Tehran')
		const wrapper = mountBars([makeBar()])
		const bar = wrapper.find('.gantt-bar')
		const aria = bar.attributes('aria-label') ?? ''

		expect(aria).toContain(formatDate(START, 'LL'))
		expect(aria).toContain(formatDate(END, 'LL'))
		expect(aria).not.toContain(START.toLocaleDateString())
	})

	it('uses the central Jalali formatter in UTC', () => {
		setLocaleAndTimezone('fa-IR', 'UTC')
		const wrapper = mountBars([makeBar()])
		const bar = wrapper.find('.gantt-bar')
		const aria = bar.attributes('aria-label') ?? ''

		expect(aria).toContain(formatDate(START, 'LL'))
		expect(aria).toContain(formatDate(END, 'LL'))
	})

	it('keeps Gregorian aria byte-identical', () => {
		setLocaleAndTimezone('en', 'UTC')
		const wrapper = mountBars([makeBar()])
		const bar = wrapper.find('.gantt-bar')
		const aria = bar.attributes('aria-label') ?? ''

		expect(aria).toContain(START.toLocaleDateString())
		expect(aria).toContain(END.toLocaleDateString())
	})

	it('keeps bar positioning identical across locales (no pixel shift)', () => {
		setLocaleAndTimezone('en', 'UTC')
		const enWrapper = mountBars([makeBar()])
		const enBar = enWrapper.find('.gantt-bar')
		const enX = enBar.attributes('x')
		const enWidth = enBar.attributes('width')

		setLocaleAndTimezone('fa-IR', 'Asia/Tehran')
		const faWrapper = mountBars([makeBar()])
		const faBar = faWrapper.find('.gantt-bar')
		expect(faBar.attributes('x')).toBe(enX)
		expect(faBar.attributes('width')).toBe(enWidth)
	})

	it('keeps instants intact (no conversion on the model)', () => {
		setLocaleAndTimezone('fa-IR', 'Asia/Tehran')
		const bar = makeBar()
		mountBars([bar])
		expect(bar.start.toISOString()).toBe(START.toISOString())
		expect(bar.end.toISOString()).toBe(END.toISOString())
	})
})
