import {parseTaskText} from './parseTaskText'
import {getDateFromText, getDateFromTextIn} from '../helpers/time/parseDate'
import {calculateDayInterval} from '../helpers/time/calculateDayInterval'
import priorities from '../models/priorities.json'

describe('Parse Task Text', () => {
	it('should return text with no intents as is', () => {
		expect(parseTaskText('Lorem Ipsum').text).toBe('Lorem Ipsum')
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
			expect(result.date.getFullYear()).toBe(now.getFullYear())
			expect(result.date.getMonth()).toBe(now.getMonth())
			expect(result.date.getDate()).toBe(now.getDate())
		})
		it('should recognize today', () => {
			const result = parseTaskText('Lorem Ipsum today')

			expect(result.text).toBe('Lorem Ipsum')
			const now = new Date()
			expect(result.date.getFullYear()).toBe(now.getFullYear())
			expect(result.date.getMonth()).toBe(now.getMonth())
			expect(result.date.getDate()).toBe(now.getDate())
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
			}

			for (const c in cases) {
				it('should recognize today with a time ' + c, () => {
					const result = parseTaskText('Lorem Ipsum today ' + c)

					expect(result.text).toBe('Lorem Ipsum')
					const now = new Date()
					expect(result.date.getFullYear()).toBe(now.getFullYear())
					expect(result.date.getMonth()).toBe(now.getMonth())
					expect(result.date.getDate()).toBe(now.getDate())
					expect(`${result.date.getHours()}:${result.date.getMinutes()}`).toBe(cases[c])
					expect(result.date.getSeconds()).toBe(0)
				})
			}
		})
		it('should recognize tomorrow', () => {
			const result = parseTaskText('Lorem Ipsum tomorrow')

			expect(result.text).toBe('Lorem Ipsum')
			const tomorrow = new Date()
			tomorrow.setDate(tomorrow.getDate() + 1)
			expect(result.date.getFullYear()).toBe(tomorrow.getFullYear())
			expect(result.date.getMonth()).toBe(tomorrow.getMonth())
			expect(result.date.getDate()).toBe(tomorrow.getDate())
		})
		it('should recognize next monday', () => {
			const result = parseTaskText('Lorem Ipsum next monday')

			const untilNextMonday = calculateDayInterval('nextMonday')

			expect(result.text).toBe('Lorem Ipsum')
			const nextMonday = new Date()
			nextMonday.setDate(nextMonday.getDate() + untilNextMonday)
			expect(result.date.getFullYear()).toBe(nextMonday.getFullYear())
			expect(result.date.getMonth()).toBe(nextMonday.getMonth())
			expect(result.date.getDate()).toBe(nextMonday.getDate())
		})
		it('should recognize this weekend', () => {
			const result = parseTaskText('Lorem Ipsum this weekend')

			const untilThisWeekend = calculateDayInterval('thisWeekend')

			expect(result.text).toBe('Lorem Ipsum')
			const thisWeekend = new Date()
			thisWeekend.setDate(thisWeekend.getDate() + untilThisWeekend)
			expect(result.date.getFullYear()).toBe(thisWeekend.getFullYear())
			expect(result.date.getMonth()).toBe(thisWeekend.getMonth())
			expect(result.date.getDate()).toBe(thisWeekend.getDate())
		})
		it('should recognize later this week', () => {
			const result = parseTaskText('Lorem Ipsum later this week')

			const untilLaterThisWeek = calculateDayInterval('laterThisWeek')

			expect(result.text).toBe('Lorem Ipsum')
			const laterThisWeek = new Date()
			laterThisWeek.setDate(laterThisWeek.getDate() + untilLaterThisWeek)
			expect(result.date.getFullYear()).toBe(laterThisWeek.getFullYear())
			expect(result.date.getMonth()).toBe(laterThisWeek.getMonth())
			expect(result.date.getDate()).toBe(laterThisWeek.getDate())
		})
		it('should recognize later next week', () => {
			const result = parseTaskText('Lorem Ipsum later next week')

			const untilLaterNextWeek = calculateDayInterval('laterNextWeek')

			expect(result.text).toBe('Lorem Ipsum')
			const laterNextWeek = new Date()
			laterNextWeek.setDate(laterNextWeek.getDate() + untilLaterNextWeek)
			expect(result.date.getFullYear()).toBe(laterNextWeek.getFullYear())
			expect(result.date.getMonth()).toBe(laterNextWeek.getMonth())
			expect(result.date.getDate()).toBe(laterNextWeek.getDate())
		})
		it('should recognize next week', () => {
			const result = parseTaskText('Lorem Ipsum next week')

			const untilNextWeek = calculateDayInterval('nextWeek')

			expect(result.text).toBe('Lorem Ipsum')
			const nextWeek = new Date()
			nextWeek.setDate(nextWeek.getDate() + untilNextWeek)
			expect(result.date.getFullYear()).toBe(nextWeek.getFullYear())
			expect(result.date.getMonth()).toBe(nextWeek.getMonth())
			expect(result.date.getDate()).toBe(nextWeek.getDate())
		})
		it('should recognize next month', () => {
			const result = parseTaskText('Lorem Ipsum next month')

			expect(result.text).toBe('Lorem Ipsum')
			const nextMonth = new Date()
			nextMonth.setDate(1)
			nextMonth.setMonth(nextMonth.getMonth() + 1)
			expect(result.date.getFullYear()).toBe(nextMonth.getFullYear())
			expect(result.date.getMonth()).toBe(nextMonth.getMonth())
			expect(result.date.getDate()).toBe(nextMonth.getDate())
		})
		it('should recognize a date', () => {
			const result = parseTaskText('Lorem Ipsum 06/26/2021')

			expect(result.text).toBe('Lorem Ipsum')
			const date = new Date()
			date.setFullYear(2021, 5, 26)
			expect(result.date.getFullYear()).toBe(date.getFullYear())
			expect(result.date.getMonth()).toBe(date.getMonth())
			expect(result.date.getDate()).toBe(date.getDate())
		})
		it('should recognize end of month', () => {
			const result = parseTaskText('Lorem Ipsum end of month')

			expect(result.text).toBe('Lorem Ipsum')
			const curDate = new Date()
			const date = new Date(curDate.getFullYear(), curDate.getMonth() + 1, 0)
			expect(result.date.getFullYear()).toBe(date.getFullYear())
			expect(result.date.getMonth()).toBe(date.getMonth())
			expect(result.date.getDate()).toBe(date.getDate())
		})
		it('should recognize weekdays', () => {
			const result = parseTaskText('Lorem Ipsum thu')

			expect(result.text).toBe('Lorem Ipsum')
			const nextThursday = new Date()
			nextThursday.setDate(nextThursday.getDate() + ((4 + 7 - nextThursday.getDay()) % 7))
			expect(`${result.date.getFullYear()}-${result.date.getMonth()}-${result.date.getDate()}`).toBe(`${nextThursday.getFullYear()}-${nextThursday.getMonth()}-${nextThursday.getDate()}`)
		})
		it('should recognize weekdays with time', () => {
			const result = parseTaskText('Lorem Ipsum thu at 14:00')

			expect(result.text).toBe('Lorem Ipsum')
			const nextThursday = new Date()
			nextThursday.setDate(nextThursday.getDate() + ((4 + 7 - nextThursday.getDay()) % 7))
			expect(`${result.date.getFullYear()}-${result.date.getMonth()}-${result.date.getDate()}`).toBe(`${nextThursday.getFullYear()}-${nextThursday.getMonth()}-${nextThursday.getDate()}`)
			expect(`${result.date.getHours()}:${result.date.getMinutes()}`).toBe('14:0')
		})
		it('should recognize dates of the month in the past but next month', () => {
			const date = new Date()
			date.setDate(date.getDate() - 1)
			const result = parseTaskText(`Lorem Ipsum ${date.getDate()}nd`)

			expect(result.text).toBe('Lorem Ipsum')
			expect(result.date.getDate()).toBe(date.getDate())
			expect(result.date.getMonth()).toBe(date.getMonth() + 1)
		})
		it('should recognize dates of the month in the future', () => {
			const date = new Date()
			const result = parseTaskText(`Lorem Ipsum ${date.getDate() + 1}nd`)

			expect(result.text).toBe('Lorem Ipsum')
			expect(result.date.getDate()).toBe(date.getDate() + 1)
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

		describe('Parse date from text', () => {
			const now = new Date()
			now.setFullYear(2021, 5, 24)

			const cases = {
				'Lorem Ipsum 06/08/2021 ad': '2021-6-8',
				'Lorem Ipsum 6/7/21 ad': '2021-6-7',
				'27/07/2021,': null,
				'2021/07/06,': '2021-7-6',
				'2021-07-06': '2021-7-6',
				'27 jan': '2022-1-27',
				'27/1': '2022-1-27',
				'27/01': '2022-1-27',
				'16/12': '2021-12-16',
				'01/27': '2022-1-27',
				'1/27': '2022-1-27',
				'Jan 27': '2022-1-27',
				'jan 27': '2022-1-27',
				'feb 21': '2022-2-21',
				'mar 21': '2022-3-21',
				'apr 21': '2022-4-21',
				'may 21': '2022-5-21',
				'jun 21': '2022-6-21',
				'jul 21': '2021-7-21',
				'aug 21': '2021-8-21',
				'sep 21': '2021-9-21',
				'oct 21': '2021-10-21',
				'nov 21': '2021-11-21',
				'dec 21': '2021-12-21',
			}

			for (const c in cases) {
				it(`should parse '${c}' as '${cases[c]}'`, () => {
					const {date} = getDateFromText(c, now)
					if (date === null && cases[c] === null) {
						expect(date).toBeNull()
						return
					}

					expect(`${date.getFullYear()}-${date.getMonth() + 1}-${date.getDate()}`).toBe(cases[c])
				})
			}
		})

		describe('Parse date from text in', () => {
			const now = new Date()
			now.setFullYear(2021, 5, 24)
			now.setHours(12)
			now.setMinutes(0)
			now.setSeconds(0)

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
			}

			for (const c in cases) {
				it(`should parse '${c}' as '${cases[c]}'`, () => {
					const {date} = getDateFromTextIn(c, now)
					if (date === null && cases[c] === null) {
						expect(date).toBeNull()
						return
					}

					expect(`${date.getFullYear()}-${date.getMonth() + 1}-${date.getDate()} ${date.getHours()}:${date.getMinutes()}`).toBe(cases[c])
				})
			}
		})
	})

	describe('Labels', () => {
		it('should parse labels', () => {
			const result = parseTaskText('Lorem Ipsum @label1 @label2')

			expect(result.text).toBe('Lorem Ipsum')
			expect(result.labels).toHaveLength(2)
			expect(result.labels[0]).toBe('label1')
			expect(result.labels[1]).toBe('label2')
		})
		it('should parse labels from the start', () => {
			const result = parseTaskText('@label1 Lorem Ipsum @label2')

			expect(result.text).toBe('Lorem Ipsum')
			expect(result.labels).toHaveLength(2)
			expect(result.labels[0]).toBe('label1')
			expect(result.labels[1]).toBe('label2')
		})
		it('should resolve duplicate labels', () => {
			const result = parseTaskText('Lorem Ipsum @label1 @label1 @label2')

			expect(result.text).toBe('Lorem Ipsum')
			expect(result.labels).toHaveLength(2)
			expect(result.labels[0]).toBe('label1')
			expect(result.labels[1]).toBe('label2')
		})
		it('should correctly parse labels with spaces in them', () => {
			const result = parseTaskText(`Lorem @'label with space' Ipsum`)

			expect(result.text).toBe('Lorem Ipsum')
			expect(result.labels).toHaveLength(1)
			expect(result.labels[0]).toBe('label with space')
		})
		it('should correctly parse labels with spaces in them and "', () => {
			const result = parseTaskText('Lorem @"label with space" Ipsum')

			expect(result.text).toBe('Lorem Ipsum')
			expect(result.labels).toHaveLength(1)
			expect(result.labels[0]).toBe('label with space')
		})
	})

	describe('List', () => {
		it('should parse a list', () => {
			const result = parseTaskText('Lorem Ipsum #list')

			expect(result.text).toBe('Lorem Ipsum')
			expect(result.list).toBe('list')
		})
		it('should parse a list with a space in it', () => {
			const result = parseTaskText(`Lorem Ipsum #'list with long name'`)

			expect(result.text).toBe('Lorem Ipsum')
			expect(result.list).toBe('list with long name')
		})
		it('should parse a list with a space in it and "', () => {
			const result = parseTaskText(`Lorem Ipsum #"list with long name"`)

			expect(result.text).toBe('Lorem Ipsum')
			expect(result.list).toBe('list with long name')
		})
		it('should parse only the first list', () => {
			const result = parseTaskText(`Lorem Ipsum #list1 #list2 #list3`)

			expect(result.text).toBe('Lorem Ipsum #list2 #list3')
			expect(result.list).toBe('list1')
		})
	})

	describe('Priority', () => {
		for (const p in priorities) {
			it(`should parse priority ${p}`, () => {
				const result = parseTaskText(`Lorem Ipsum !${priorities[p]}`)

				expect(result.text).toBe('Lorem Ipsum')
				expect(result.priority).toBe(priorities[p])
			})
		}
		it(`should not parse an invalid priority`, () => {
			const result = parseTaskText(`Lorem Ipsum !9999`)

			expect(result.text).toBe('Lorem Ipsum !9999')
			expect(result.priority).toBe(null)
		})
		it(`should not parse an invalid priority but use the first valid one it finds`, () => {
			const result = parseTaskText(`Lorem Ipsum !9999 !1`)

			expect(result.text).toBe('Lorem Ipsum !9999')
			expect(result.priority).toBe(1)
		})
	})

	describe('Assignee', () => {
		it('should parse an assignee', () => {
			const result = parseTaskText('Lorem Ipsum +user')

			expect(result.text).toBe('Lorem Ipsum')
			expect(result.assignees).toHaveLength(1)
			expect(result.assignees[0]).toBe('user')
		})
		it('should parse multiple assignees', () => {
			const result = parseTaskText('Lorem Ipsum +user1 +user2 +user3')

			expect(result.text).toBe('Lorem Ipsum')
			expect(result.assignees).toHaveLength(3)
			expect(result.assignees[0]).toBe('user1')
			expect(result.assignees[1]).toBe('user2')
			expect(result.assignees[2]).toBe('user3')
		})
		it('should parse avoid duplicate assignees', () => {
			const result = parseTaskText('Lorem Ipsum +user1 +user1 +user2')

			expect(result.text).toBe('Lorem Ipsum')
			expect(result.assignees).toHaveLength(2)
			expect(result.assignees[0]).toBe('user1')
			expect(result.assignees[1]).toBe('user2')
		})
		it('should parse an assignee with a space in it', () => {
			const result = parseTaskText(`Lorem Ipsum +'user with long name'`)

			expect(result.text).toBe('Lorem Ipsum')
			expect(result.assignees).toHaveLength(1)
			expect(result.assignees[0]).toBe('user with long name')
		})
		it('should parse an assignee with a space in it and "', () => {
			const result = parseTaskText(`Lorem Ipsum +"user with long name"`)

			expect(result.text).toBe('Lorem Ipsum')
			expect(result.assignees).toHaveLength(1)
			expect(result.assignees[0]).toBe('user with long name')
		})
	})
})
