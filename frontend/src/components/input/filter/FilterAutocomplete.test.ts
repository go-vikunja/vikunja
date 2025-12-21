import {describe, it, expect} from 'vitest'
import {calculateReplacementRange} from './FilterAutocomplete'

describe('FilterAutocomplete', () => {
	describe('calculateReplacementRange', () => {
		// Note: calculateReplacementRange adds +1 to convert string indices to ProseMirror positions
		// In ProseMirror, position 0 is before the document, text starts at position 1

		describe('single value replacement', () => {
			it('should use startPos and endPos for replacement boundaries with +1 offset for ProseMirror', () => {
				const context = {
					keyword: 'Work To Do',
					startPos: 11, // position after "project in "
					endPos: 21,   // position after "Work To Do"
				}

				const result = calculateReplacementRange(context, 'in')

				expect(result.replaceFrom).toBe(12) // 11 + 1 for ProseMirror offset
				expect(result.replaceTo).toBe(22)   // 21 + 1 for ProseMirror offset
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

				expect(result.replaceFrom).toBe(11) // 10 + 1
				expect(result.replaceTo).toBe(20)   // 19 + 1
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
				// replaceFrom = 11 + 5 + 1 + 1 + 1 = 19 (extra +1 for ProseMirror offset)
				expect(result.replaceFrom).toBe(19)
				expect(result.replaceTo).toBe(29) // 28 + 1
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
				// replaceFrom = 11 + 8 + 1 + 1 + 1 = 22 (extra +1 for ProseMirror offset)
				expect(result.replaceFrom).toBe(22)
				expect(result.replaceTo).toBe(27) // 26 + 1
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
				// replaceFrom = 10 + 6 + 1 + 1 + 1 = 19 (extra +1 for ProseMirror offset)
				expect(result.replaceFrom).toBe(19)
				expect(result.replaceTo).toBe(25) // 24 + 1
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
				// replaceFrom = 5 + 3 + 1 + 0 + 1 = 10 (extra +1 for ProseMirror offset)
				expect(result.replaceFrom).toBe(10)
				expect(result.replaceTo).toBe(11) // 10 + 1
			})

			it('should not modify range for single values even with in operator', () => {
				const context = {
					keyword: 'SingleProject',
					startPos: 11,
					endPos: 24,
				}

				const result = calculateReplacementRange(context, 'in')

				// No comma in keyword, so full range should be used (with +1 offset)
				expect(result.replaceFrom).toBe(12) // 11 + 1
				expect(result.replaceTo).toBe(25)   // 24 + 1
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

				// = is not a multi-value operator, so full range should be used (with +1 offset)
				expect(result.replaceFrom).toBe(11) // 10 + 1
				expect(result.replaceTo).toBe(30)   // 29 + 1
			})

			it('should not modify range for != operator', () => {
				const context = {
					keyword: 'A, B',
					startPos: 10,
					endPos: 14,
				}

				const result = calculateReplacementRange(context, '!=')

				expect(result.replaceFrom).toBe(11) // 10 + 1
				expect(result.replaceTo).toBe(15)   // 14 + 1
			})
		})

		describe('closing quote handling', () => {
			it('should extend replaceTo by 1 when hasClosingQuote is true', () => {
				const context = {
					keyword: 'Work To Do',
					startPos: 12, // position after opening quote in 'project = "'
					endPos: 22,
				}

				const result = calculateReplacementRange(context, '=', true)

				expect(result.replaceFrom).toBe(13) // 12 + 1
				expect(result.replaceTo).toBe(24)   // 22 + 1 + 1 (extra 1 for closing quote)
			})

			it('should not extend replaceTo when hasClosingQuote is false', () => {
				const context = {
					keyword: 'Work To Do',
					startPos: 12,
					endPos: 22,
				}

				const result = calculateReplacementRange(context, '=', false)

				expect(result.replaceFrom).toBe(13) // 12 + 1
				expect(result.replaceTo).toBe(23)   // 22 + 1 (no extra for closing quote)
			})

			it('should default hasClosingQuote to false when not provided', () => {
				const context = {
					keyword: 'Work To Do',
					startPos: 12,
					endPos: 22,
				}

				const result = calculateReplacementRange(context, '=')

				expect(result.replaceTo).toBe(23) // 22 + 1 (no extra for closing quote)
			})

			it('should handle closing quote with multi-value operators', () => {
				const context = {
					keyword: 'Inbox, Work To Do',
					startPos: 12,
					endPos: 29,
				}

				const result = calculateReplacementRange(context, 'in', true)

				// lastCommaIndex = 5
				// textAfterComma = " Work To Do" (11 chars)
				// leadingSpaces = 1
				// replaceFrom = 12 + 5 + 1 + 1 + 1 = 20
				expect(result.replaceFrom).toBe(20)
				// replaceTo = 29 + 1 + 1 = 31 (extra 1 for closing quote)
				expect(result.replaceTo).toBe(31)
			})
		})
	})
})
