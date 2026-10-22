import {describe, expect, it} from 'vitest'

import {parseDateMath} from './parseDateMath'

// Wednesday
const now = new Date(2026, 9, 21, 14, 30, 15)

describe('parseDateMath', () => {
	it.each([
		['now', new Date(2026, 9, 21, 14, 30, 15)],
		['now/d', new Date(2026, 9, 21)],
		['now/d+1d', new Date(2026, 9, 22)],
		['now/w', new Date(2026, 9, 19)],
		['now/w-1w', new Date(2026, 9, 12)],
		['now/w+2w', new Date(2026, 10, 2)],
		['now/M', new Date(2026, 9, 1)],
		['now/M+1M', new Date(2026, 10, 1)],
		['now/y+1y', new Date(2027, 0, 1)],
		['now+7d', new Date(2026, 9, 28, 14, 30, 15)],
		[' now/d ', new Date(2026, 9, 21)],
	])('evaluates %s', (expression, expected) => {
		expect(parseDateMath(expression, now)).toEqual(expected)
	})

	it('rounds a Sunday back to the previous Monday', () => {
		expect(parseDateMath('now/w', new Date(2026, 9, 25, 10))).toEqual(new Date(2026, 9, 19))
	})

	it.each(['', 'now+', 'now/x', 'now/b', 'now||+1d', 'nope||+1d'])('returns null for %s', expression => {
		expect(parseDateMath(expression, now)).toBeNull()
	})
})

// Ported from go-datemath's datemath_test.go, minus cases for options the API doesn't set (round-up, custom
// fiscal year, week start or business days). Go evaluates those in UTC, so expectations without an offset
// are local wall time here.
describe('parseDateMath matches go-datemath', () => {
	const goNow = new Date('2014-11-18T14:27:32.000')

	it.each([
		// basic dates
		['2014', '2014-01-01T00:00:00.000'],
		['2014-05', '2014-05-01T00:00:00.000'],
		['2014-05-30', '2014-05-30T00:00:00.000'],
		['2014-05-30T20', '2014-05-30T20:00:00.000'],
		['2014-05-30T20:21', '2014-05-30T20:21:00.000'],
		['2014-05-30T20:21:35', '2014-05-30T20:21:35.000'],
		['2014-05-30T20:21:35.123', '2014-05-30T20:21:35.123'],

		// basic math
		['2014-11-18||+y', '2015-11-18T00:00:00.000'],
		['2014-11-18||+fy', '2015-11-18T00:00:00.000'],
		['2014-11-18||-2y', '2012-11-18T00:00:00.000'],
		['2014-11-18||+1fQ', '2015-02-18T00:00:00.000'],
		['2014-11-18||+3M', '2015-02-18T00:00:00.000'],
		['2014-11-18||-M', '2014-10-18T00:00:00.000'],
		['2014-11-18||+1w', '2014-11-25T00:00:00.000'],
		['2014-11-18||-3w', '2014-10-28T00:00:00.000'],
		['2014-11-18||+22d', '2014-12-10T00:00:00.000'],
		['2014-11-18||-423d', '2013-09-21T00:00:00.000'],
		['2014-11-18T14||+13h', '2014-11-19T03:00:00.000'],
		['2014-11-18T14||-1h', '2014-11-18T13:00:00.000'],
		['2014-11-18T14||+13H', '2014-11-19T03:00:00.000'],
		['2014-11-18T14||-1H', '2014-11-18T13:00:00.000'],
		['2014-11-18T14:27||+10240m', '2014-11-25T17:07:00.000'],
		['2014-11-18T14:27||-10m', '2014-11-18T14:17:00.000'],
		['2014-11-18T14:27:32||+60s', '2014-11-18T14:28:32.000'],
		['2014-11-18T14:27:32||-3600s', '2014-11-18T13:27:32.000'],
		['2014-11-19T14:27:32||/w', '2014-11-17T00:00:00.000'],
		['2014-11-01T14:27:32||/w', '2014-10-27T00:00:00.000'],

		// lenient math
		['2014-05-30T20:21||', '2014-05-30T20:21:00.000'],

		// multiple adjustments
		['2014-11-18||+1M-1M', '2014-11-18T00:00:00.000'],
		['2014-11-18||+1M-1m', '2014-12-17T23:59:00.000'],
		['2014-11-18||-1m+1M', '2014-12-17T23:59:00.000'],
		['2014-11-18||+1M/M', '2014-12-01T00:00:00.000'],
		['2014-11-18||+1M/M+1h', '2014-12-01T01:00:00.000'],

		// now
		['now', '2014-11-18T14:27:32.000'],
		['now+M', '2014-12-18T14:27:32.000'],
		['now/Q', '2014-10-01T00:00:00.000'],

		// fiscal years and quarters, default start
		['now/fy', '2014-01-01T00:00:00.000'],
		['now/fQ', '2014-10-01T00:00:00.000'],

		// epoch times
		['04:52:20', '1970-01-01T04:52:20.000'],

		// timestamps
		['1418248078000', '2014-12-10T21:47:58.000Z'],
		['32484216259000', '2999-05-20T17:24:19.000Z'],
		['253382837059000', '9999-05-20T17:24:19.000Z'],
		['1418248078000||/m', '2014-12-10T21:47:00.000Z'],

		// timezones
		['2014-05-30T20:21+03:00', '2014-05-30T17:21:00.000Z'],
		['2014-05-30T20:21Z', '2014-05-30T20:21:00.000Z'],

		// business days
		['now+5b', '2014-11-25T14:27:32.000'],
		['now-5b', '2014-11-11T14:27:32.000'],

		// weekdays
		['2020-03-12||/w', '2020-03-09T00:00:00.000'],
	])('evaluates %s', (expression, expected) => {
		expect(parseDateMath(expression, goNow)).toEqual(new Date(expected))
	})

	it.each([
		['now/m', '2014-11-18T14:27:32.000Z', '2014-11-18T14:27:00.000Z'],
		['now/Q', '2014-03-18T14:27:32.000', '2014-01-01T00:00:00.000'],
		['now/fQ', '2022-01-04T10:10:10.000', '2022-01-01T00:00:00.000'],
		['now-15m', '2014-05-30T20:21:35.123', '2014-05-30T20:06:35.123'],
		['now/w', '2014-05-30T20:21:35.123', '2014-05-26T00:00:00.000'],
		['now+1M', '2014-05-30T20:21:35.123', '2014-06-30T20:21:35.123'],
		['2014-05-31||+1M/w', '2014-05-30T20:21:35.123', '2014-06-30T00:00:00.000'],
	])('evaluates %s at %s', (expression, at, expected) => {
		expect(parseDateMath(expression, new Date(at))).toEqual(new Date(expected))
	})

	it.each(['2014-05-35T20:21:21Z', 'npe', 'now-nope', 'definitely-nope'])('rejects %s', expression => {
		expect(parseDateMath(expression, goNow)).toBeNull()
	})
})
