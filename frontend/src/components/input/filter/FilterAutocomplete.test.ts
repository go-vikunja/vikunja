import {describe, it, expect} from 'vitest'
import {calculateReplacementRange} from './FilterAutocomplete'

describe('FilterAutocomplete', () => {
	describe('calculateReplacementRange', () => {
		describe('single value replacement', () => {
			it('should use startPos and endPos for replacement boundaries', () => {
				const context = {
					keyword: 'Work To Do',
					startPos: 11, // position after "project in "
					endPos: 21,   // position after "Work To Do"
				}

				const result = calculateReplacementRange(context, 'in')

				expect(result.replaceFrom).toBe(11)
				expect(result.replaceTo).toBe(21)
				expect(result.replaceTo - result.replaceFrom).toBe(context.keyword.length)
			})

			it('should handle single-word values correctly', () => {
				const context = {
					keyword: 'Inbox',
					startPos: 11,
					endPos: 16,
				}

				const result = calculateReplacementRange(context, 'in')

				expect(result.replaceTo - result.replaceFrom).toBe(5) // "Inbox".length
			})

			it('should handle equals operator', () => {
				const context = {
					keyword: 'MyProject',
					startPos: 10,
					endPos: 19,
				}

				const result = calculateReplacementRange(context, '=')

				expect(result.replaceFrom).toBe(10)
				expect(result.replaceTo).toBe(19)
			})
		})

		describe('multi-value operator replacement', () => {
			it('should only replace text after last comma for multi-value operators', () => {
				const context = {
					keyword: 'Inbox, Work To Do',
					startPos: 11,
					endPos: 28, // 11 + 17
				}

				const result = calculateReplacementRange(context, 'in')

				// lastCommaIndex = 5 (position of comma in "Inbox, Work To Do")
				// textAfterComma = " Work To Do" (11 chars)
				// leadingSpaces = 1
				// replaceFrom = 11 + 5 + 1 + 1 = 18
				expect(result.replaceFrom).toBe(18)
				expect(result.replaceTo).toBe(28)
			})

			it('should handle multiple commas correctly', () => {
				const context = {
					keyword: 'One, Two, Three',
					startPos: 11,
					endPos: 26,
				}

				const result = calculateReplacementRange(context, 'in')

				// lastCommaIndex = 8 (position of second comma in "One, Two, Three")
				// textAfterComma = " Three" (6 chars)
				// leadingSpaces = 1
				// replaceFrom = 11 + 8 + 1 + 1 = 21
				expect(result.replaceFrom).toBe(21)
				expect(result.replaceTo).toBe(26)
			})

			it('should handle ?= operator as multi-value', () => {
				const context = {
					keyword: 'Label1, Label2',
					startPos: 10,
					endPos: 24,
				}

				const result = calculateReplacementRange(context, '?=')

				// lastCommaIndex = 6
				// textAfterComma = " Label2" (7 chars)
				// leadingSpaces = 1
				// replaceFrom = 10 + 6 + 1 + 1 = 18
				expect(result.replaceFrom).toBe(18)
				expect(result.replaceTo).toBe(24)
			})

			it('should handle no spaces after comma', () => {
				const context = {
					keyword: 'A,B,C',
					startPos: 5,
					endPos: 10,
				}

				const result = calculateReplacementRange(context, 'in')

				// lastCommaIndex = 3 (position of second comma)
				// textAfterComma = "C" (1 char)
				// leadingSpaces = 0
				// replaceFrom = 5 + 3 + 1 + 0 = 9
				expect(result.replaceFrom).toBe(9)
				expect(result.replaceTo).toBe(10)
			})

			it('should not modify range for single values even with in operator', () => {
				const context = {
					keyword: 'SingleProject',
					startPos: 11,
					endPos: 24,
				}

				const result = calculateReplacementRange(context, 'in')

				// No comma in keyword, so full range should be used
				expect(result.replaceFrom).toBe(11)
				expect(result.replaceTo).toBe(24)
			})
		})

		describe('non-multi-value operators', () => {
			it('should not modify range for = operator even with commas in value', () => {
				const context = {
					keyword: 'Value, With, Commas',
					startPos: 10,
					endPos: 29,
				}

				const result = calculateReplacementRange(context, '=')

				// = is not a multi-value operator, so full range should be used
				expect(result.replaceFrom).toBe(10)
				expect(result.replaceTo).toBe(29)
			})

			it('should not modify range for != operator', () => {
				const context = {
					keyword: 'A, B',
					startPos: 10,
					endPos: 14,
				}

				const result = calculateReplacementRange(context, '!=')

				expect(result.replaceFrom).toBe(10)
				expect(result.replaceTo).toBe(14)
			})
		})
	})
})
