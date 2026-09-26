import {beforeEach, describe, expect, it} from 'vitest'
import {createPinia, setActivePinia} from 'pinia'

import {PrefixMode} from './prefixes'
import {parseTaskTextWithTokens, type MagicToken} from './tokens'

const now = new Date(2026, 8, 23, 10, 0)

function highlighted(text: string, mode = PrefixMode.Default) {
	return parseTaskTextWithTokens(text, mode, now).tokens
		.map(({type, from, to}: MagicToken) => [type, text.slice(from, to)])
}

describe('parseTaskTextWithTokens', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
	})

	it('locates every kind of magic', () => {
		expect(highlighted('Buy milk *groceries +shopping !3 @alice every day tomorrow at 15:00')).toEqual([
			['label', '*groceries'],
			['project', '+shopping'],
			['priority', '!3'],
			['assignee', '@alice'],
			['repeat', 'every day'],
			['date', 'tomorrow'],
			['date', 'at 15:00'],
		])
	})

	it('includes quotes of quoted items', () => {
		expect(highlighted('Task *"my label" +\'my project\'')).toEqual([
			['label', '*"my label"'],
			['project', '+\'my project\''],
		])
	})

	it('highlights every occurrence the parser strips', () => {
		expect(highlighted('Task *home and *home again')).toEqual([
			['label', '*home'],
			['label', '*home'],
		])
	})

	it('does not highlight a priority the parser rejects', () => {
		expect(highlighted('Task !9')).toEqual([])
	})

	it('does not confuse date words inside labels', () => {
		expect(highlighted('Task *today tomorrow')).toEqual([
			['label', '*today'],
			['date', 'tomorrow'],
		])
	})

	it('locates dates with a month and day', () => {
		expect(highlighted('Task 17th oct')).toEqual([
			['date', '17th'],
			['date', 'oct'],
		])
	})

	it('uses the configured prefix mode', () => {
		expect(highlighted('Task @groceries #shopping +alice', PrefixMode.Todoist)).toEqual([
			['label', '@groceries'],
			['project', '#shopping'],
			['assignee', '+alice'],
		])
	})

	it('returns nothing when disabled', () => {
		expect(highlighted('Task *label tomorrow', PrefixMode.Disabled)).toEqual([])
	})

	it('returns nothing for quoted text', () => {
		expect(highlighted('"Task *label tomorrow"')).toEqual([])
	})

	it('returns nothing for magic-only text', () => {
		expect(highlighted('*label tomorrow')).toEqual([])
	})
})
