import {describe, it, expect} from 'vitest'
import {parseScopesFromQuery} from './parseScopesFromQuery'

describe('parseScopesFromQuery', () => {
	it('returns empty object for null input', () => {
		expect(parseScopesFromQuery(null)).toEqual({})
	})

	it('returns empty object for undefined input', () => {
		expect(parseScopesFromQuery(undefined)).toEqual({})
	})

	it('returns empty object for empty string', () => {
		expect(parseScopesFromQuery('')).toEqual({})
	})

	it('parses a single scope', () => {
		expect(parseScopesFromQuery('tasks:read')).toEqual({
			tasks: ['read'],
		})
	})

	it('parses multiple scopes in the same group', () => {
		expect(parseScopesFromQuery('tasks:read,tasks:write')).toEqual({
			tasks: ['read', 'write'],
		})
	})

	it('parses scopes across different groups', () => {
		expect(parseScopesFromQuery('tasks:read,projects:write,labels:read')).toEqual({
			tasks: ['read'],
			projects: ['write'],
			labels: ['read'],
		})
	})

	it('handles whitespace around commas', () => {
		expect(parseScopesFromQuery('tasks:read , projects:write')).toEqual({
			tasks: ['read'],
			projects: ['write'],
		})
	})

	it('handles whitespace around colons', () => {
		expect(parseScopesFromQuery('tasks : read')).toEqual({
			tasks: ['read'],
		})
	})

	it('ignores malformed entries without colons', () => {
		expect(parseScopesFromQuery('tasks:read,invalidentry,projects:write')).toEqual({
			tasks: ['read'],
			projects: ['write'],
		})
	})

	it('ignores entries with empty group', () => {
		expect(parseScopesFromQuery(':read,tasks:write')).toEqual({
			tasks: ['write'],
		})
	})

	it('ignores entries with empty permission', () => {
		expect(parseScopesFromQuery('tasks:,projects:write')).toEqual({
			projects: ['write'],
		})
	})

	it('handles trailing commas', () => {
		expect(parseScopesFromQuery('tasks:read,')).toEqual({
			tasks: ['read'],
		})
	})

	it('handles leading commas', () => {
		expect(parseScopesFromQuery(',tasks:read')).toEqual({
			tasks: ['read'],
		})
	})
})
