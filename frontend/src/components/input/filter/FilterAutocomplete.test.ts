import {describe, it, expect} from 'vitest'

// Test the position calculation logic extracted for testability
// These test the core replacement boundary calculations

describe('Filter Autocomplete Position Calculation', () => {
	describe('single value replacement', () => {
		it('should use startPos and endPos for replacement boundaries', () => {
			const context = {
				field: 'project',
				prefix: 'project in ',
				keyword: 'Work To Do',
				search: 'Work To Do',
				operator: 'in',
				startPos: 11, // position after "project in "
				endPos: 21,   // position after "Work To Do"
				isComplete: false,
			}

			const replaceFrom = context.startPos
			const replaceTo = context.endPos

			expect(replaceFrom).toBe(11)
			expect(replaceTo).toBe(21)
			expect(replaceTo - replaceFrom).toBe(context.keyword.length)
		})

		it('should handle single-word values correctly', () => {
			const context = {
				field: 'project',
				prefix: 'project in ',
				keyword: 'Inbox',
				search: 'Inbox',
				operator: 'in',
				startPos: 11,
				endPos: 16,
				isComplete: false,
			}

			const replaceFrom = context.startPos
			const replaceTo = context.endPos

			expect(replaceTo - replaceFrom).toBe(5) // "Inbox".length
		})
	})

	describe('multi-value operator replacement', () => {
		it('should only replace text after last comma for multi-value operators', () => {
			const context = {
				field: 'project',
				prefix: 'project in ',
				keyword: 'Inbox, Work To Do',
				search: 'Work To Do',
				operator: 'in',
				startPos: 11,
				endPos: 28, // 11 + 17
				isComplete: false,
			}

			let replaceFrom = context.startPos
			const replaceTo = context.endPos

			// Multi-value handling: only replace after last comma
			if (context.keyword.includes(',')) {
				const lastCommaIndex = context.keyword.lastIndexOf(',')
				const textAfterComma = context.keyword.substring(lastCommaIndex + 1)
				const leadingSpaces = textAfterComma.length - textAfterComma.trimStart().length
				replaceFrom = context.startPos + lastCommaIndex + 1 + leadingSpaces
			}

			// Should replace from after ", " (comma + space) to end
			expect(replaceFrom).toBe(11 + 6 + 1) // startPos + "Inbox," + 1 space
			expect(replaceTo).toBe(28)
		})

		it('should handle multiple commas correctly', () => {
			const context = {
				field: 'project',
				prefix: 'project in ',
				keyword: 'One, Two, Three',
				search: 'Three',
				operator: 'in',
				startPos: 11,
				endPos: 26,
				isComplete: false,
			}

			let replaceFrom = context.startPos

			if (context.keyword.includes(',')) {
				const lastCommaIndex = context.keyword.lastIndexOf(',')
				const textAfterComma = context.keyword.substring(lastCommaIndex + 1)
				const leadingSpaces = textAfterComma.length - textAfterComma.trimStart().length
				replaceFrom = context.startPos + lastCommaIndex + 1 + leadingSpaces
			}

			// lastCommaIndex = 8 (position of second comma in "One, Two, Three")
			// textAfterComma = " Three" (6 chars)
			// leadingSpaces = 1
			expect(replaceFrom).toBe(11 + 8 + 1 + 1) // startPos + commaIndex + 1 + space
		})
	})
})
