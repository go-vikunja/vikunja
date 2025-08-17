import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'

import {ParsedTaskText, parseTaskText, PrefixMode} from './parseTaskText'
import {parseDate} from '../helpers/time/parseDate'
import {calculateDayInterval} from '../helpers/time/calculateDayInterval'
import {PRIORITIES} from '@/constants/priorities'
import {MILLISECONDS_A_DAY} from '@/constants/date'
import type {IRepeatAfter} from '@/types/IRepeatAfter'

describe('Parse Task Text', () => {
	beforeEach(() => {
		vi.useFakeTimers()
	})

	afterEach(() => {
		vi.useRealTimers()
	})

	it('should return text with no intents as is', () => {
		expect(parseTaskText('Lorem Ipsum').text).toBe('Lorem Ipsum')
	})

	it('should not parse text when disabled', () => {
		const text = 'Lorem Ipsum today *label +project !2 @user'
		const result = parseTaskText(text, PrefixMode.Disabled)

		expect(result.text).toBe(text)
	})

	it('should parse text in todoist mode when configured', () => {
		const result = parseTaskText('Lorem Ipsum today @label #project !2 +user', PrefixMode.Todoist)

		expect(result.text).toBe('Lorem Ipsum  +user')
		const now = new Date()
		expect(result?.date?.getFullYear()).toBe(now.getFullYear())
		expect(result?.date?.getMonth()).toBe(now.getMonth())
		expect(result?.date?.getDate()).toBe(now.getDate())
		expect(result.labels).toHaveLength(1)
		expect(result.labels[0]).toBe('label')
		expect(result.project).toBe('project')
		expect(result.priority).toBe(2)
		expect(result.assignees).toHaveLength(1)
		expect(result.assignees[0]).toBe('user')
	})

	it('should ignore email addresses', () => {
		const text = 'Lorem Ipsum email@example.com'
		const result = parseTaskText(text)

		expect(result.text).toBe(text)
	})

	describe('Date Parsing', () => {
		it('should not return any date if none was provided', () => {
			const result = parseTaskText('Lorem Ipsum')

			expect(result.text).toBe('Lorem Ipsum')
			expect(result.date).toBeNull()
		})
		it('should ignore casing', () => {
			const result = parseTaskText('Lorem Ipsum ToDay')

			expect(result.text).toBe('Lorem Ipsum')
			const now = new Date()
			expect(result?.date?.getFullYear()).toBe(now.getFullYear())
			expect(result?.date?.getMonth()).toBe(now.getMonth())
			expect(result?.date?.getDate()).toBe(now.getDate())
		})
		it('should recognize today', () => {
			const result = parseTaskText('Lorem Ipsum today')

			expect(result.text).toBe('Lorem Ipsum')
			const now = new Date()
			expect(result?.date?.getFullYear()).toBe(now.getFullYear())
			expect(result?.date?.getMonth()).toBe(now.getMonth())
			expect(result?.date?.getDate()).toBe(now.getDate())
		})
		it('should recognize tonight', () => {
			const result = parseTaskText('Lorem Ipsum tonight')

			expect(result.text).toBe('Lorem Ipsum')
			const now = new Date()
			expect(result?.date?.getFullYear()).toBe(now.getFullYear())
			expect(result?.date?.getMonth()).toBe(now.getMonth())
			expect(result?.date?.getDate()).toBe(now.getDate())
			expect(result?.date?.getHours()).toBe(21)
		})
		describe('should recognize today with a time', () => {
			const cases = {
				'at 15:00': '15:0',
				'@ 15:00': '15:0',
				'at 15:30': '15:30',
				'@ 3pm': '15:0',
				'at 3pm': '15:0',
				'at 3 pm': '15:0',
				'at 3am': '3:0',
				'at 3:12 am': '3:12',
				'at 3:12 pm': '15:12',
			} as const
			
			for (const c in cases) {
				it(`should recognize today with a time ${c}`, () => {
					const result = parseTaskText(`Lorem Ipsum today ${c}`)

					expect(result.text).toBe('Lorem Ipsum')
					const now = new Date()
					expect(result?.date?.getFullYear()).toBe(now.getFullYear())
					expect(result?.date?.getMonth()).toBe(now.getMonth())
					expect(result?.date?.getDate()).toBe(now.getDate())
					expect(`${result?.date?.getHours()}:${result?.date?.getMinutes()}`).toBe(cases[c as keyof typeof cases])
					expect(result?.date?.getSeconds()).toBe(0)
				})
			}
		})
		it('should recognize tomorrow', () => {
			const result = parseTaskText('Lorem Ipsum tomorrow')

			expect(result.text).toBe('Lorem Ipsum')
			const tomorrow = new Date()
			tomorrow.setDate(tomorrow.getDate() + 1)
			expect(result?.date?.getFullYear()).toBe(tomorrow.getFullYear())
			expect(result?.date?.getMonth()).toBe(tomorrow.getMonth())
			expect(result?.date?.getDate()).toBe(tomorrow.getDate())
		})
		it('should recognize Tomorrow', () => {
			const result = parseTaskText('Lorem Ipsum Tomorrow')

			expect(result.text).toBe('Lorem Ipsum')
			const tomorrow = new Date()
			tomorrow.setDate(tomorrow.getDate() + 1)
			expect(result?.date?.getFullYear()).toBe(tomorrow.getFullYear())
			expect(result?.date?.getMonth()).toBe(tomorrow.getMonth())
			expect(result?.date?.getDate()).toBe(tomorrow.getDate())
		})
		it('should recognize next monday', () => {
			const result = parseTaskText('Lorem Ipsum next monday')

			const untilNextMonday = calculateDayInterval('nextMonday')

			expect(result.text).toBe('Lorem Ipsum')
			const nextMonday = new Date()
			nextMonday.setDate(nextMonday.getDate() + untilNextMonday)
			expect(result?.date?.getFullYear()).toBe(nextMonday.getFullYear())
			expect(result?.date?.getMonth()).toBe(nextMonday.getMonth())
			expect(result?.date?.getDate()).toBe(nextMonday.getDate())
		})
		it('should recognize next monday on the beginning of the sentence', () => {
			const result = parseTaskText('next monday Lorem Ipsum')

			const untilNextMonday = calculateDayInterval('nextMonday')

			expect(result.text).toBe('Lorem Ipsum')
			const nextMonday = new Date()
			nextMonday.setDate(nextMonday.getDate() + untilNextMonday)
			expect(result?.date?.getFullYear()).toBe(nextMonday.getFullYear())
			expect(result?.date?.getMonth()).toBe(nextMonday.getMonth())
			expect(result?.date?.getDate()).toBe(nextMonday.getDate())
		})
		it('should recognize next monday and ignore casing', () => {
			const result = parseTaskText('Lorem Ipsum nExt Monday')

			const untilNextMonday = calculateDayInterval('nextMonday')

			expect(result.text).toBe('Lorem Ipsum')
			const nextMonday = new Date()
			nextMonday.setDate(nextMonday.getDate() + untilNextMonday)
			expect(result?.date?.getFullYear()).toBe(nextMonday.getFullYear())
			expect(result?.date?.getMonth()).toBe(nextMonday.getMonth())
			expect(result?.date?.getDate()).toBe(nextMonday.getDate())
		})
		it('should recognize this weekend', () => {
			const result = parseTaskText('Lorem Ipsum this weekend')

			const untilThisWeekend = calculateDayInterval('thisWeekend')

			expect(result.text).toBe('Lorem Ipsum')
			const thisWeekend = new Date()
			thisWeekend.setDate(thisWeekend.getDate() + untilThisWeekend)
			expect(result?.date?.getFullYear()).toBe(thisWeekend.getFullYear())
			expect(result?.date?.getMonth()).toBe(thisWeekend.getMonth())
			expect(result?.date?.getDate()).toBe(thisWeekend.getDate())
		})
		it('should recognize later this week', () => {
			const result = parseTaskText('Lorem Ipsum later this week')

			const untilLaterThisWeek = calculateDayInterval('laterThisWeek')

			expect(result.text).toBe('Lorem Ipsum')
			const laterThisWeek = new Date()
			laterThisWeek.setDate(laterThisWeek.getDate() + untilLaterThisWeek)
			expect(result?.date?.getFullYear()).toBe(laterThisWeek.getFullYear())
			expect(result?.date?.getMonth()).toBe(laterThisWeek.getMonth())
			expect(result?.date?.getDate()).toBe(laterThisWeek.getDate())
		})
		it('should recognize later next week', () => {
			const result = parseTaskText('Lorem Ipsum later next week')

			const untilLaterNextWeek = calculateDayInterval('laterNextWeek')

			expect(result.text).toBe('Lorem Ipsum')
			const laterNextWeek = new Date()
			laterNextWeek.setDate(laterNextWeek.getDate() + untilLaterNextWeek)
			expect(result?.date?.getFullYear()).toBe(laterNextWeek.getFullYear())
			expect(result?.date?.getMonth()).toBe(laterNextWeek.getMonth())
			expect(result?.date?.getDate()).toBe(laterNextWeek.getDate())
		})
		it('should recognize next week', () => {
			const result = parseTaskText('Lorem Ipsum next week')

			const untilNextWeek = calculateDayInterval('nextWeek')

			expect(result.text).toBe('Lorem Ipsum')
			const nextWeek = new Date()
			nextWeek.setDate(nextWeek.getDate() + untilNextWeek)
			expect(result?.date?.getFullYear()).toBe(nextWeek.getFullYear())
			expect(result?.date?.getMonth()).toBe(nextWeek.getMonth())
			expect(result?.date?.getDate()).toBe(nextWeek.getDate())
		})
		it('should recognize next month', () => {
			const result = parseTaskText('Lorem Ipsum next month')

			expect(result.text).toBe('Lorem Ipsum')
			const nextMonth = new Date()
			nextMonth.setDate(1)
			nextMonth.setMonth(nextMonth.getMonth() + 1)
			expect(result?.date?.getFullYear()).toBe(nextMonth.getFullYear())
			expect(result?.date?.getMonth()).toBe(nextMonth.getMonth())
			expect(result?.date?.getDate()).toBe(nextMonth.getDate())
		})
		it('should recognize a date', () => {
			const result = parseTaskText('Lorem Ipsum 06/26/2021')

			expect(result.text).toBe('Lorem Ipsum')
			const date = new Date()
			date.setFullYear(2021, 5, 26)
			expect(result?.date?.getFullYear()).toBe(date.getFullYear())
			expect(result?.date?.getMonth()).toBe(date.getMonth())
			expect(result?.date?.getDate()).toBe(date.getDate())
		})
		it('should recognize end of month', () => {
			const result = parseTaskText('Lorem Ipsum end of month')

			expect(result.text).toBe('Lorem Ipsum')
			const curDate = new Date()
			const date = new Date(curDate.getFullYear(), curDate.getMonth() + 1, 0)
			expect(result?.date?.getFullYear()).toBe(date.getFullYear())
			expect(result?.date?.getMonth()).toBe(date.getMonth())
			expect(result?.date?.getDate()).toBe(date.getDate())
		})


		it('should recognize weekdays with time', () => {
			const result = parseTaskText('Lorem Ipsum thu at 14:00')

			expect(result.text).toBe('Lorem Ipsum')
			const nextThursday = new Date()
			nextThursday.setDate(nextThursday.getDate() + ((4 + 7 - nextThursday.getDay()) % 7))
			expect(`${result?.date?.getFullYear()}-${result?.date?.getMonth()}-${result?.date?.getDate()}`).toBe(`${nextThursday.getFullYear()}-${nextThursday.getMonth()}-${nextThursday.getDate()}`)
			expect(`${result?.date?.getHours()}:${result?.date?.getMinutes()}`).toBe('14:0')
		})
		it('should recognize dates of the month in the past but next month', () => {
			const time = new Date(2022, 0, 15)
			vi.setSystemTime(time)

			const result = parseTaskText(`Lorem Ipsum ${time.getDate() - 1}th`)

			expect(result.text).toBe('Lorem Ipsum')
			expect(result?.date?.getDate()).toBe(time.getDate() - 1)
			expect(result?.date?.getMonth()).toBe(time.getMonth() + 1)
		})
		it('should recognize dates of the month in the past but next month when february is the next month', () => {
			const jan = new Date(2022, 0, 30)
			vi.setSystemTime(jan)

			const result = parseTaskText(`Lorem Ipsum ${jan.getDate() - 1}th`)

			const expectedDate = new Date(2022, 2, jan.getDate() - 1)
			expect(result.text).toBe('Lorem Ipsum')
			expect(result?.date?.getDate()).toBe(expectedDate.getDate())
			expect(result?.date?.getMonth()).toBe(expectedDate.getMonth())
		})
		it('should recognize dates of the month in the past but next month when the next month has less days than this one', () => {
			const mar = new Date(2022, 2, 32)
			vi.setSystemTime(mar)

			const result = parseTaskText('Lorem Ipsum 31st')

			const expectedDate = new Date(2022, 4, 31)
			expect(result.text).toBe('Lorem Ipsum')
			expect(result?.date?.getDate()).toBe(expectedDate.getDate())
			expect(result?.date?.getMonth()).toBe(expectedDate.getMonth())
		})
		it('should recognize dates of the month in the future', () => {
			const nextDay = new Date(+new Date() + MILLISECONDS_A_DAY)
			const result = parseTaskText(`Lorem Ipsum ${nextDay.getDate()}nd`)

			expect(result.text).toBe('Lorem Ipsum')
			expect(result?.date?.getDate()).toBe(nextDay.getDate())
		})
		it('should only recognize weekdays with a space before or after them 1', () => {
			const result = parseTaskText('Lorem Ipsum renewed')

			expect(result.text).toBe('Lorem Ipsum renewed')
			expect(result.date).toBeNull()
		})
		it('should only recognize weekdays with a space before or after them 2', () => {
			const result = parseTaskText('Lorem Ipsum github')

			expect(result.text).toBe('Lorem Ipsum github')
			expect(result.date).toBeNull()
		})
		describe('Should not recognize weekdays in words', () => {
			const cases = [
				'renewed',
				'github',
				'fix monitor stand',
				'order wedding cake',
				'investigate thumping noise',
				'iron frilly napkins',
				'take photo of saturn',
				'fix sunglasses',
				'monitor blood pressure',
				'Monitor blood pressure',
				'buy almonds',
			]

			cases.forEach(c => {
				it(`should not recognize text with ${c} at the beginning as weekday`, () => {
					const result = parseTaskText(`${c} dolor sit amet`)

					expect(result.text).toBe(`${c} dolor sit amet`)
					expect(result.date).toBeNull()
				})
				it(`should not recognize text with ${c} at the end as weekday`, () => {
					const result = parseTaskText(`Lorem Ipsum ${c}`)

					expect(result.text).toBe(`Lorem Ipsum ${c}`)
					expect(result.date).toBeNull()
				})
				it(`should not recognize text with ${c} as weekday`, () => {
					const result = parseTaskText(`Lorem Ipsum ${c} dolor`)

					expect(result.text).toBe(`Lorem Ipsum ${c} dolor`)
					expect(result.date).toBeNull()
				})
			})
		})
		it('should not recognize date number with no spacing around them', () => {
			const result = parseTaskText('Lorem Ispum v1.1.1')

			expect(result.text).toBe('Lorem Ispum v1.1.1')
			expect(result.date).toBeNull()
		})
		it('should not recognize dates in urls', () => {
			const text = 'https://some-url.org/blog/2019/1/233526-some-more-text'
			const result = parseTaskText(text)

			expect(result.text).toBe(text)
			expect(result.date).toBeNull()
		})

		describe('Parse weekdays', () => {

			const days = {
				'monday': 1,
				'Monday': 1,
				'mon': 1,
				'Mon': 1,
				'tuesday': 2,
				'Tuesday': 2,
				'tue': 2,
				'Tue': 2,
				'wednesday': 3,
				'Wednesday': 3,
				'wed': 3,
				'Wed': 3,
				'thursday': 4,
				'Thursday': 4,
				'thu': 4,
				'Thu': 4,
				'friday': 5,
				'Friday': 5,
				'fri': 5,
				'Fri': 5,
				'saturday': 6,
				'Saturday': 6,
				'sat': 6,
				'Sat': 6,
				'sunday': 7,
				'Sunday': 7,
				'sun': 7,
				'Sun': 7,
			} as Record<string, number>

			const prefix = [
				'next ',
				'',
			]

			prefix.forEach(p => {
				for (const d in days) {
					it(`should recognize ${p}${d}`, () => {
						const result = parseTaskText(`Lorem Ipsum ${p}${d}`)

						const next = new Date()
						const distance = (days[d] + 7 - next.getDay()) % 7
						next.setDate(next.getDate() + distance)

						expect(result.text).toBe('Lorem Ipsum')
						expect(result?.date?.getFullYear()).toBe(next.getFullYear())
						expect(result?.date?.getMonth()).toBe(next.getMonth())
						expect(result?.date?.getDate()).toBe(next.getDate())
					})
					it(`should recognize ${p}${d} at the beginning of the text`, () => {
						const result = parseTaskText(`${p}${d} Lorem Ipsum`)

						const next = new Date()
						const distance = (days[d] + 7 - next.getDay()) % 7
						next.setDate(next.getDate() + distance)

						expect(result.text).toBe('Lorem Ipsum')
						expect(result?.date?.getFullYear()).toBe(next.getFullYear())
						expect(result?.date?.getMonth()).toBe(next.getMonth())
						expect(result?.date?.getDate()).toBe(next.getDate())
					})
				}
			})

			// This tests only standalone days are recognized and not things like "github", "monitor" or "renewed".
			// We're not using real words here to generate tests for all days on the fly.
			for (const d in days) {
				it(`should not recognize ${d} with a space before it but none after it`, () => {
					const text = `Lorem Ipsum ${d}ipsum`
					const result = parseTaskText(text)

					expect(result.text).toBe(text)
					expect(result.date).toBeNull()
				})
				it(`should not recognize ${d} with a space after it but none before it`, () => {
					const text = `Lorem ipsum${d} dolor`
					const result = parseTaskText(text)

					expect(result.text).toBe(text)
					expect(result.date).toBeNull()
				})
				it(`should not recognize ${d} with no space before or after it`, () => {
					const text = `Lorem Ipsum lorem${d}ipsum`
					const result = parseTaskText(text)

					expect(result.text).toBe(text)
					expect(result.date).toBeNull()
				})
			}
		})

		describe('Parse date from text', () => {
			const now = new Date()
			now.setFullYear(2021, 5, 24)

			const cases = {
				'06/08/2021': '2021-6-8',
				'6/7/21': '2021-6-7',
				'27/07/2021,': null,
				'2021/07/06': '2021-7-6',
				'2021-07-06': '2021-7-6',
				'27 jan': '2022-1-27',
				'27/1': '2022-1-27',
				'27/01': '2022-1-27',
				'16/12': '2021-12-16',
				'01/27': '2022-1-27',
				'1/27': '2022-1-27',
				'jan 27': '2022-1-27',
				'Jan 27': '2022-1-27',
				'january 27': '2022-1-27',
				'January 27': '2022-1-27',
				'feb 21': '2022-2-21',
				'Feb 21': '2022-2-21',
				'february 21': '2022-2-21',
				'February 21': '2022-2-21',
				'mar 21': '2022-3-21',
				'Mar 21': '2022-3-21',
				'march 21': '2022-3-21',
				'March 21': '2022-3-21',
				'apr 21': '2022-4-21',
				'Apr 21': '2022-4-21',
				'april 21': '2022-4-21',
				'April 21': '2022-4-21',
				'may 21': '2022-5-21',
				'May 21': '2022-5-21',
				'jun 21': '2022-6-21',
				'Jun 21': '2022-6-21',
				'june 21': '2022-6-21',
				'June 21': '2022-6-21',
				'21st June': '2021-6-21',
				'jul 21': '2021-7-21',
				'Jul 21': '2021-7-21',
				'july 21': '2021-7-21',
				'July 21': '2021-7-21',
				'aug 21': '2021-8-21',
				'Aug 21': '2021-8-21',
				'august 21': '2021-8-21',
				'August 21': '2021-8-21',
				'sep 21': '2021-9-21',
				'Sep 21': '2021-9-21',
				'september 21': '2021-9-21',
				'September 21': '2021-9-21',
				'oct 21': '2021-10-21',
				'Oct 21': '2021-10-21',
				'october 21': '2021-10-21',
				'October 21': '2021-10-21',
				'nov 21': '2021-11-21',
				'Nov 21': '2021-11-21',
				'november 21': '2021-11-21',
				'November 21': '2021-11-21',
				'dec 21': '2021-12-21',
				'Dec 21': '2021-12-21',
				'december 21': '2021-12-21',
				'01.02.2021': '2021-2-1',
				'01.02': '2022-2-1',
				'01.10': '2021-10-1',
				'01.02.25': '2025-2-1',
			} as Record<string, string | null>

			for (const c in cases) {
				const assertResult = ({date, text}: ParsedTaskText) => {
					if (date === null && cases[c] === null) {
						expect(date).toBeNull()
						return
					}

					expect(`${date?.getFullYear()}-${date?.getMonth() + 1}-${date?.getDate()}`).toBe(cases[c])
					expect(text.trim()).toBe('Lorem Ipsum')
				}
				
				it(`should parse '${c}' as '${cases[c]}' with the date at the end`, () => {
					assertResult(parseTaskText(`Lorem Ipsum ${c}`, PrefixMode.Default, now))
				})
				it(`should parse '${c}' as '${cases[c]}' with the date at the beginning`, () => {
					assertResult(parseTaskText(`${c} Lorem Ipsum`, PrefixMode.Default, now))
				})
			}
		})

		describe('Parse date from text in', () => {
			const now = new Date()
			now.setFullYear(2021, 5, 24)
			now.setHours(12)
			now.setMinutes(0)
			now.setSeconds(0)

			beforeEach(() => {
				vi.useFakeTimers()
				vi.setSystemTime(now)
			})

			afterEach(() => {
				vi.useRealTimers()
			})

			const cases = {
				'Lorem Ipsum in 1 hour': '2021-6-24 13:0',
				'in 2 hours': '2021-6-24 14:0',
				'in 1 day': '2021-6-25 12:0',
				'in 2 days': '2021-6-26 12:0',
				'in 1 week': '2021-7-1 12:0',
				'in 2 weeks': '2021-7-8 12:0',
				'in 4 weeks': '2021-7-22 12:0',
				'in 1 month': '2021-7-24 12:0',
				'in 3 months': '2021-9-24 12:0',
				'Something in 5 days at 10:00': '2021-6-29 10:0',
				'Something 17th at 10:00': '2021-7-17 10:0',
				'Something sep 17 at 10:00': '2021-9-17 10:0',
				'Something sep 17th at 10:00': '2021-9-17 10:0',
				'Something at 10:00 in 5 days': '2021-6-29 10:0',
				'Something at 10:00 17th': '2021-7-17 10:0',
				'Something at 10:00 sep 17th': '2021-9-17 10:0',
			} as Record<string, string>

			for (const c in cases) {
				it(`should parse '${c}' as '${cases[c]}'`, () => {
					const {date} = parseDate(c, now)
					if (date === null && cases[c] === null) {
						expect(date).toBeNull()
						return
					}

					expect(`${date.getFullYear()}-${date.getMonth() + 1}-${date.getDate()} ${date.getHours()}:${date.getMinutes()}`).toBe(cases[c])
				})
			}

			it('should replace the text in title case', () => {
				const {date, newText} = parseDate('Some task Mar 8th', now)

				expect(`${date.getFullYear()}-${date.getMonth() + 1}-${date.getDate()} ${date.getHours()}:${date.getMinutes()}`).toBe('2021-3-8 12:0')
				expect(newText).toBe('Some task')
			})

			it('should replace the text in lowercase', () => {
				const {date, newText} = parseDate('Some task mar 8th', now)

				expect(`${date.getFullYear()}-${date.getMonth() + 1}-${date.getDate()} ${date.getHours()}:${date.getMinutes()}`).toBe('2021-3-8 12:0')
				expect(newText).toBe('Some task')
			})
		})
	})

	describe('Labels', () => {
		it('should parse labels', () => {
			const result = parseTaskText('Lorem Ipsum *label1 *label2')

			expect(result.text).toBe('Lorem Ipsum')
			expect(result.labels).toHaveLength(2)
			expect(result.labels[0]).toBe('label1')
			expect(result.labels[1]).toBe('label2')
		})
		it('should parse labels from the start', () => {
			const result = parseTaskText('*label1 Lorem Ipsum *label2')

			expect(result.text).toBe('Lorem Ipsum')
			expect(result.labels).toHaveLength(2)
			expect(result.labels[0]).toBe('label1')
			expect(result.labels[1]).toBe('label2')
		})
		it('should resolve duplicate labels', () => {
			const result = parseTaskText('Lorem Ipsum *label1 *label1 *label2')

			expect(result.text).toBe('Lorem Ipsum')
			expect(result.labels).toHaveLength(2)
			expect(result.labels[0]).toBe('label1')
			expect(result.labels[1]).toBe('label2')
		})
		it('should correctly parse labels with spaces in them', () => {
			const result = parseTaskText('Lorem *\'label with space\' Ipsum')

			expect(result.text).toBe('Lorem Ipsum')
			expect(result.labels).toHaveLength(1)
			expect(result.labels[0]).toBe('label with space')
		})
		it('should correctly parse labels with spaces in them and "', () => {
			const result = parseTaskText('Lorem *"label with space" Ipsum')

			expect(result.text).toBe('Lorem Ipsum')
			expect(result.labels).toHaveLength(1)
			expect(result.labels[0]).toBe('label with space')
		})
		it('should not parse labels called date expressions as dates', () => {
			const result = parseTaskText('Lorem Ipsum *today')

			expect(result.text).toBe('Lorem Ipsum')
			expect(result.labels).toHaveLength(1)
			expect(result.labels[0]).toBe('today')
		})
		it('should parse labels with parentheses and remove them from text', () => {
			const result = parseTaskText('a *"a (a)"')

			expect(result.text).toBe('a')
			expect(result.labels).toHaveLength(1)
			expect(result.labels[0]).toBe('a (a)')
		})
		it('should parse labels with parentheses from the start', () => {
			const result = parseTaskText('*"a (a)" a')

			expect(result.text).toBe('a')
			expect(result.labels).toHaveLength(1)
			expect(result.labels[0]).toBe('a (a)')
		})
	})

	describe('Project', () => {
		it('should parse a project', () => {
			const result = parseTaskText('Lorem Ipsum +project')

			expect(result.text).toBe('Lorem Ipsum')
			expect(result.project).toBe('project')
		})
		it('should parse a project with a space in it', () => {
			const result = parseTaskText('Lorem Ipsum +\'project with long name\'')

			expect(result.text).toBe('Lorem Ipsum')
			expect(result.project).toBe('project with long name')
		})
		it('should parse a project with a space in it and "', () => {
			const result = parseTaskText('Lorem Ipsum +"project with long name"')

			expect(result.text).toBe('Lorem Ipsum')
			expect(result.project).toBe('project with long name')
		})
		it('should parse only the first project', () => {
			const result = parseTaskText('Lorem Ipsum +project1 +project2 +project3')

			expect(result.text).toBe('Lorem Ipsum +project2 +project3')
			expect(result.project).toBe('project1')
		})
		it('should parse a project that\'s called like a date as project', () => {
			const result = parseTaskText('Lorem Ipsum +today')

			expect(result.text).toBe('Lorem Ipsum')
			expect(result.project).toBe('today')
		})
	})

	describe('Priority', () => {
		for (const p in PRIORITIES) {
			it(`should parse priority ${p}`, () => {
				const result = parseTaskText(`Lorem Ipsum !${PRIORITIES[p]}`)

				expect(result.text).toBe('Lorem Ipsum')
				expect(result.priority).toBe(PRIORITIES[p])
			})
		}
		it('should not parse an invalid priority', () => {
			const result = parseTaskText('Lorem Ipsum !9999')

			expect(result.text).toBe('Lorem Ipsum !9999')
			expect(result.priority).toBe(null)
		})
		it('should not parse an invalid priority but use the first valid one it finds', () => {
			const result = parseTaskText('Lorem Ipsum !9999 !1')

			expect(result.text).toBe('Lorem Ipsum !9999')
			expect(result.priority).toBe(1)
		})
	})

	describe('Assignee', () => {
		it('should parse an assignee', () => {
			const text = 'Lorem Ipsum @user'
			const result = parseTaskText(text)

			expect(result.text).toBe(text)
			expect(result.assignees).toHaveLength(1)
			expect(result.assignees[0]).toBe('user')
		})
		it('should parse multiple assignees', () => {
			const text = 'Lorem Ipsum @user1 @user2 @user3'
			const result = parseTaskText(text)

			expect(result.text).toBe(text)
			expect(result.assignees).toHaveLength(3)
			expect(result.assignees[0]).toBe('user1')
			expect(result.assignees[1]).toBe('user2')
			expect(result.assignees[2]).toBe('user3')
		})
		it('should parse avoid duplicate assignees', () => {
			const text = 'Lorem Ipsum @user1 @user1 @user2'
			const result = parseTaskText(text)

			expect(result.text).toBe(text)
			expect(result.assignees).toHaveLength(2)
			expect(result.assignees[0]).toBe('user1')
			expect(result.assignees[1]).toBe('user2')
		})
		it('should parse an assignee with a space in it', () => {
			const text = 'Lorem Ipsum @\'user with long name\''
			const result = parseTaskText(text)

			expect(result.text).toBe(text)
			expect(result.assignees).toHaveLength(1)
			expect(result.assignees[0]).toBe('user with long name')
		})
		it('should parse an assignee with a space in it and "', () => {
			const text = 'Lorem Ipsum @"user with long name"'
			const result = parseTaskText(text)

			expect(result.text).toBe(text)
			expect(result.assignees).toHaveLength(1)
			expect(result.assignees[0]).toBe('user with long name')
		})
		it('should parse an assignee who is called like a date as assignee', () => {
			const text = 'Lorem Ipsum @today'
			const result = parseTaskText(text)

			expect(result.text).toBe(text)
			expect(result.assignees).toHaveLength(1)
			expect(result.assignees[0]).toBe('today')
		})
		it('should recognize an email address', () => {
			const text = 'Lorem Ipsum @email@example.com'
			const result = parseTaskText(text)

			expect(result.text).toBe('Lorem Ipsum @email@example.com')
			expect(result.assignees).toHaveLength(1)
			expect(result.assignees[0]).toBe('email@example.com')
		})
	})

	describe('Recurring Dates', () => {
		const cases = {
			'every 1 hour': {type: 'hours', amount: 1},
			'every hour': {type: 'hours', amount: 1},
			'every 5 hours': {type: 'hours', amount: 5},
			'every 12 hours': {type: 'hours', amount: 12},
			'every day': {type: 'days', amount: 1},
			'every 1 day': {type: 'days', amount: 1},
			'every 2 days': {type: 'days', amount: 2},
			'every week': {type: 'weeks', amount: 1},
			'every 1 week': {type: 'weeks', amount: 1},
			'every 3 weeks': {type: 'weeks', amount: 3},
			'every month': {type: 'months', amount: 1},
			'every 1 month': {type: 'months', amount: 1},
			'every 2 months': {type: 'months', amount: 2},
			'every year': {type: 'years', amount: 1},
			'every 1 year': {type: 'years', amount: 1},
			'every 4 years': {type: 'years', amount: 4},
			'every one hour': {type: 'hours', amount: 1}, // maybe unnesecary but better to include it for completeness sake
			'every two hours': {type: 'hours', amount: 2},
			'every three hours': {type: 'hours', amount: 3},
			'every four hours': {type: 'hours', amount: 4},
			'every five hours': {type: 'hours', amount: 5},
			'every six hours': {type: 'hours', amount: 6},
			'every seven hours': {type: 'hours', amount: 7},
			'every eight hours': {type: 'hours', amount: 8},
			'every nine hours': {type: 'hours', amount: 9},
			'every ten hours': {type: 'hours', amount: 10},
			'annually': {type: 'years', amount: 1},
			'biannually': {type: 'months', amount: 6},
			'semiannually': {type: 'months', amount: 6},
			'biennially': {type: 'years', amount: 2},
			'daily': {type: 'days', amount: 1},
			'hourly': {type: 'hours', amount: 1},
			'monthly': {type: 'months', amount: 1},
			'weekly': {type: 'weeks', amount: 1},
			'yearly': {type: 'years', amount: 1},
		} as Record<string, IRepeatAfter>

		for (const c in cases) {
			it(`should parse ${c} as recurring date every ${cases[c].amount} ${cases[c].type}`, () => {
				const result = parseTaskText(`Lorem Ipsum ${c}`)

				expect(result.text).toBe('Lorem Ipsum')
				expect(result?.repeats?.type).toBe(cases[c].type)
				expect(result?.repeats?.amount).toBe(cases[c].amount)
			})
		 	
			it(`should parse ${c} as recurring date every ${cases[c].amount} ${cases[c].type} at 11:42`, () => {
				const result = parseTaskText(`Lorem Ipsum ${c} at 11:42`)

				expect(result.text).toBe('Lorem Ipsum')
				expect(result?.repeats?.type).toBe(cases[c].type)
				expect(result?.repeats?.amount).toBe(cases[c].amount)
				const now = new Date()
				expect(`${result?.date?.getFullYear()}-${result?.date?.getMonth()}-${result?.date?.getDate()}`).toBe(`${now.getFullYear()}-${now.getMonth()}-${now.getDate()}`)
				expect(`${result?.date?.getHours()}:${result?.date?.getMinutes()}`).toBe('11:42')
			})
		}

		const wordCases = [
			'annually',
			'biannually',
			'semiannually',
			'biennially',
			'daily',
			'hourly',
			'monthly',
			'weekly',
			'yearly',
		]

		wordCases.forEach(c => {
			it(`should ignore recurring periods if they are part of a word ${c}`, () => {
				const result = parseTaskText(`Lorem Ipsum word${c}notword`)

				expect(result.text).toBe(`Lorem Ipsum word${c}notword`)
				expect(result?.repeats).toBeNull()
			})
		})
	})
})
