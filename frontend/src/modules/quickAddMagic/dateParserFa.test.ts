import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'
import {setActivePinia, createPinia} from 'pinia'

import {parsePersianDate, getPersianRepeats} from './dateParserFa'
import {parseDate} from './dateParser'
import {i18n} from '@/i18n'
import {useAuthStore} from '@/stores/auth'
import {REPEAT_TYPES} from '@/types/IRepeatAfter'

const FIXED_NOW = new Date('2026-09-07T12:00:00Z')
const LEAP_NOW = new Date('2025-03-10T12:00:00Z')
const TZ = 'UTC'

function setDefaultDueTime(defaultDueTime?: string) {
	const authStore = useAuthStore()
	authStore.setUserSettings({
		...authStore.settings,
		// @ts-expect-error - same helper shape as the existing quick-add tests (baseline)
		frontendSettings: {
			...authStore.settings.frontendSettings,
			defaultDueTime,
		},
	})
}

function expectFaDate(input: string, expectedISO: string, expectedNewText: string, now: Date = FIXED_NOW) {
	const result = parsePersianDate(input, now, {locale: 'fa-IR', timeZone: TZ})
	expect(result).not.toBeNull()
	expect(result?.date?.toISOString()).toBe(expectedISO)
	expect(result?.newText).toBe(expectedNewText)
}

function expectFaNull(input: string, now: Date = FIXED_NOW) {
	expect(parsePersianDate(input, now, {locale: 'fa-IR', timeZone: TZ})).toBeNull()
}

describe('Persian quick-add date parsing', () => {
	beforeEach(() => {
		vi.useFakeTimers()
		vi.setSystemTime(FIXED_NOW)
		setActivePinia(createPinia())
		useAuthStore()
		setDefaultDueTime('12:00')
		i18n.global.locale.value = 'fa-IR'
	})

	afterEach(() => {
		i18n.global.locale.value = 'en'
		vi.useRealTimers()
	})

	describe('fixed keywords', () => {
		const cases = [
			{input: 'کار امروز', expectedGregorianISO: '2026-09-07T12:00:00.000Z', expectedNewText: 'کار'},
			{input: 'امروز کار', expectedGregorianISO: '2026-09-07T12:00:00.000Z', expectedNewText: 'کار'},
			{input: 'کار فردا', expectedGregorianISO: '2026-09-08T12:00:00.000Z', expectedNewText: 'کار'},
			{input: 'فردا کار', expectedGregorianISO: '2026-09-08T12:00:00.000Z', expectedNewText: 'کار'},
			{input: 'کار پس‌فردا', expectedGregorianISO: '2026-09-09T12:00:00.000Z', expectedNewText: 'کار'},
			{input: 'کار پس فردا', expectedGregorianISO: '2026-09-09T12:00:00.000Z', expectedNewText: 'کار'},
			{input: 'پس‌فردا کار', expectedGregorianISO: '2026-09-09T12:00:00.000Z', expectedNewText: 'کار'},
		]
		for (const c of cases) {
			it(`parses '${c.input}'`, () => {
				expectFaDate(c.input, c.expectedGregorianISO, c.expectedNewText)
			})
		}
	})

	describe('weekdays', () => {
		const cases = [
			{input: 'کار شنبه', expectedGregorianISO: '2026-09-12T12:00:00.000Z', expectedNewText: 'کار'},
			{input: 'کار یکشنبه', expectedGregorianISO: '2026-09-13T12:00:00.000Z', expectedNewText: 'کار'},
			{input: 'کار دوشنبه', expectedGregorianISO: '2026-09-07T12:00:00.000Z', expectedNewText: 'کار'},
			{input: 'کار سه‌شنبه', expectedGregorianISO: '2026-09-08T12:00:00.000Z', expectedNewText: 'کار'},
			{input: 'کار چهارشنبه', expectedGregorianISO: '2026-09-09T12:00:00.000Z', expectedNewText: 'کار'},
			{input: 'کار پنجشنبه', expectedGregorianISO: '2026-09-10T12:00:00.000Z', expectedNewText: 'کار'},
			{input: 'کار جمعه', expectedGregorianISO: '2026-09-11T12:00:00.000Z', expectedNewText: 'کار'},
			{input: 'شنبه کار', expectedGregorianISO: '2026-09-12T12:00:00.000Z', expectedNewText: 'کار'},
			{input: 'کار یک‌شنبه', expectedGregorianISO: '2026-09-13T12:00:00.000Z', expectedNewText: 'کار'},
			{input: 'کار سه شنبه', expectedGregorianISO: '2026-09-08T12:00:00.000Z', expectedNewText: 'کار'},
		]
		for (const c of cases) {
			it(`parses '${c.input}'`, () => {
				expectFaDate(c.input, c.expectedGregorianISO, c.expectedNewText)
			})
		}
	})

	describe('month boundaries', () => {
		it('parses آخر ماه as the last day of Shahrivar', () => {
			expectFaDate('کار آخر ماه', '2026-09-22T12:00:00.000Z', 'کار')
		})

		it('parses اول ماه بعد with Esfand wrap handling', () => {
			expectFaDate('کار اول ماه بعد', '2026-09-23T12:00:00.000Z', 'کار')
		})

		it('parses اول ماه بعد at the start', () => {
			expectFaDate('اول ماه بعد کار', '2026-09-23T12:00:00.000Z', 'کار')
		})
	})

	describe('day with Jalali month names', () => {
		const cases = [
			{input: 'کار ۱۷ فروردین', expectedGregorianISO: '2027-04-06T12:00:00.000Z', expectedNewText: 'کار'},
			{input: 'کار ۱۷ اردیبهشت', expectedGregorianISO: '2027-05-07T12:00:00.000Z', expectedNewText: 'کار'},
			{input: 'کار ۱۷ خرداد', expectedGregorianISO: '2027-06-07T12:00:00.000Z', expectedNewText: 'کار'},
			{input: 'کار ۱۷ تیر', expectedGregorianISO: '2027-07-08T12:00:00.000Z', expectedNewText: 'کار'},
			{input: 'کار ۱۷ مرداد', expectedGregorianISO: '2027-08-08T12:00:00.000Z', expectedNewText: 'کار'},
			{input: 'کار ۱۷ شهریور', expectedGregorianISO: '2026-09-08T12:00:00.000Z', expectedNewText: 'کار'},
			{input: 'کار ۱۷ مهر', expectedGregorianISO: '2026-10-09T12:00:00.000Z', expectedNewText: 'کار'},
			{input: 'کار ۱۷ آبان', expectedGregorianISO: '2026-11-08T12:00:00.000Z', expectedNewText: 'کار'},
			{input: 'کار ۱۷ آذر', expectedGregorianISO: '2026-12-08T12:00:00.000Z', expectedNewText: 'کار'},
			{input: 'کار ۱۷ دی', expectedGregorianISO: '2027-01-07T12:00:00.000Z', expectedNewText: 'کار'},
			{input: 'کار ۱۷ بهمن', expectedGregorianISO: '2027-02-06T12:00:00.000Z', expectedNewText: 'کار'},
			{input: 'کار ۱۷ اسفند', expectedGregorianISO: '2027-03-08T12:00:00.000Z', expectedNewText: 'کار'},
		]
		for (const c of cases) {
			it(`parses '${c.input}'`, () => {
				expectFaDate(c.input, c.expectedGregorianISO, c.expectedNewText)
			})
		}

		it('parses ASCII digits for day with month', () => {
			expectFaDate('کار 17 مرداد', '2027-08-08T12:00:00.000Z', 'کار')
		})

		it('parses Arabic-Indic digits for day with month', () => {
			expectFaDate('کار ١٧ مرداد', '2027-08-08T12:00:00.000Z', 'کار')
		})

		it('rolls passed Jalali dates to the next year', () => {
			expectFaDate('کار ۱۰ شهریور', '2027-09-01T12:00:00.000Z', 'کار')
		})
	})

	describe('numeric Jalali dates', () => {
		it('parses slash-separated ASCII numerics', () => {
			expectFaDate('کار 1405/06/17', '2026-09-08T12:00:00.000Z', 'کار')
		})

		it('parses dash-separated numerics', () => {
			expectFaDate('کار 1405-06-17', '2026-09-08T12:00:00.000Z', 'کار')
		})

		it('parses dot-separated numerics', () => {
			expectFaDate('کار 1405.06.17', '2026-09-08T12:00:00.000Z', 'کار')
		})

		it('parses Persian digits', () => {
			expectFaDate('کار ۱۴۰۵/۰۶/۱۷', '2026-09-08T12:00:00.000Z', 'کار')
		})

		it('parses Arabic-Indic digits', () => {
			expectFaDate('کار ١٤٠٥/٠٦/١٧', '2026-09-08T12:00:00.000Z', 'کار')
		})

		it('parses numerics at the start', () => {
			expectFaDate('1405/06/17 کار', '2026-09-08T12:00:00.000Z', 'کار')
		})

		it('rejects years outside the 1300-1500 window', () => {
			expectFaNull('کار 1501/01/01')
			expectFaNull('کار 1299/01/01')
			expectFaNull('کار 2026/09/07')
		})

		it('rejects impossible Jalali dates', () => {
			expectFaNull('کار 1405/13/01')
			expectFaNull('کار 1405/06/32')
		})
	})

	describe('time handling', () => {
		it('parses ساعت with Persian digits', () => {
			expectFaDate('کار فردا ساعت ۱۵:۳۰', '2026-09-08T15:30:00.000Z', 'کار')
		})

		it('parses ساعت with ASCII digits', () => {
			expectFaDate('کار فردا ساعت 15:30', '2026-09-08T15:30:00.000Z', 'کار')
		})

		it('parses ساعت with Arabic-Indic digits', () => {
			expectFaDate('کار فردا ساعت ١٥:٣٠', '2026-09-08T15:30:00.000Z', 'کار')
		})

		it('parses time-only ساعت for today', () => {
			expectFaDate('کار ساعت ۱۵:۳۰', '2026-09-07T15:30:00.000Z', 'کار')
		})

		it('keeps @user assignees intact', () => {
			const result = parsePersianDate('کار فردا @user', FIXED_NOW, {locale: 'fa-IR', timeZone: TZ})
			expect(result?.date?.toISOString()).toBe('2026-09-08T12:00:00.000Z')
			expect(result?.newText).toContain('@user')
		})

		it('parses mixed-language فردا at 15:00', () => {
			expectFaDate('کار فردا at 15:00', '2026-09-08T15:00:00.000Z', 'کار')
		})
	})

	describe('negatives', () => {
		it('ignores prose numbers', () => {
			expectFaNull('خرید ۲ عدد نان')
		})

		it('rejects mid-text numerics without a following time', () => {
			expectFaNull('گزارش 1405/06/16 مهم')
		})

		it('accepts mid-text numerics followed by a time', () => {
			const result = parsePersianDate('گزارش 1405/06/16 ساعت 15:00 پایان', FIXED_NOW, {locale: 'fa-IR', timeZone: TZ})
			expect(result?.date?.toISOString()).toBe('2026-09-07T15:00:00.000Z')
		})

		it('rejects شنبه‌ها plurals with ZWNJ', () => {
			expectFaNull('کار شنبه‌ها')
		})

		it('rejects شنبه ها plurals with space', () => {
			expectFaNull('کار شنبه ها')
		})

		it('rejects impossible month days', () => {
			expectFaNull('کار ۳۱ آبان')
		})
	})

	describe('Esfand leap handling', () => {
		it('parses ۲۹ اسفند in a common year', () => {
			expectFaDate('کار ۲۹ اسفند', '2027-03-20T12:00:00.000Z', 'کار')
		})

		it('rejects ۳۰ اسفند in a common year', () => {
			expectFaNull('کار ۳۰ اسفند')
		})

		it('parses ۳۰ اسفند in a leap year', () => {
			vi.setSystemTime(LEAP_NOW)
			const result = parsePersianDate('کار ۳۰ اسفند', LEAP_NOW, {locale: 'fa-IR', timeZone: TZ})
			expect(result?.date?.toISOString()).toBe('2025-03-20T12:00:00.000Z')
			expect(result?.newText).toBe('کار')
		})
	})

	describe('locale gate via parseDate', () => {
		it('ignores Persian under the en locale', () => {
			const result = parseDate('کار فردا', FIXED_NOW, 'en', TZ)
			expect(result.date).toBeNull()
			expect(result.newText).toBe('کار فردا')
		})

		it('parses Persian under the fa-IR locale', () => {
			const result = parseDate('کار فردا', FIXED_NOW, 'fa-IR', TZ)
			expect(result.date?.toISOString()).toBe('2026-09-08T12:00:00.000Z')
			expect(result.newText).toBe('کار')
		})

		it('falls through to English when no Persian matches', () => {
			const result = parseDate('meeting tomorrow', FIXED_NOW, 'fa-IR', TZ)
			expect(result.date).not.toBeNull()
			expect(result.newText).toBe('meeting')
		})
	})

	describe('Persian repeats', () => {
		it('parses هر ماه and ماهانه as months', () => {
			expect(getPersianRepeats('کار هر ماه').repeats).toEqual({amount: 1, type: REPEAT_TYPES.Months})
			expect(getPersianRepeats('کار ماهانه').repeats).toEqual({amount: 1, type: REPEAT_TYPES.Months})
			expect(getPersianRepeats('کار هر ماه').textWithoutMatched).toBe('کار')
		})

		it('parses هر سال and سالانه as years', () => {
			expect(getPersianRepeats('کار هر سال').repeats).toEqual({amount: 1, type: REPEAT_TYPES.Years})
			expect(getPersianRepeats('کار سالانه').repeats).toEqual({amount: 1, type: REPEAT_TYPES.Years})
		})

		it('parses هر هفته and هفتگی as weeks', () => {
			expect(getPersianRepeats('کار هر هفته').repeats).toEqual({amount: 1, type: REPEAT_TYPES.Weeks})
			expect(getPersianRepeats('کار هفتگی').repeats).toEqual({amount: 1, type: REPEAT_TYPES.Weeks})
		})

		it('parses هر روز and روزانه as days', () => {
			expect(getPersianRepeats('کار هر روز').repeats).toEqual({amount: 1, type: REPEAT_TYPES.Days})
			expect(getPersianRepeats('کار روزانه').repeats).toEqual({amount: 1, type: REPEAT_TYPES.Days})
		})

		it('returns null when no repeat words are present', () => {
			const result = getPersianRepeats('کار ماه')
			expect(result.repeats).toBeNull()
			expect(result.textWithoutMatched).toBe('کار ماه')
		})
	})
})
