import {describe, it, expect} from 'vitest'
import {parseValidationErrors} from './parseValidationErrors'

describe('parseValidationErrors', () => {
	it('returns empty object when no invalid_fields present', () => {
		const error = {
			message: 'invalid data',
			code: 2002,
		}

		const result = parseValidationErrors(error)

		expect(result).toEqual({})
	})

	it('parses single field error', () => {
		const error = {
			message: 'invalid data',
			code: 2002,
			invalid_fields: ['email: email is not a valid email address'],
		}

		const result = parseValidationErrors(error)

		expect(result).toEqual({
			email: 'email is not a valid email address',
		})
	})

	it('parses multiple field errors', () => {
		const error = {
			message: 'invalid data',
			code: 2002,
			invalid_fields: [
				'email: email is not a valid email address',
				'username: username must not contain spaces',
			],
		}

		const result = parseValidationErrors(error)

		expect(result).toEqual({
			email: 'email is not a valid email address',
			username: 'username must not contain spaces',
		})
	})

	it('handles fields without colon separator', () => {
		const error = {
			message: 'invalid data',
			code: 2002,
			invalid_fields: ['something went wrong'],
		}

		const result = parseValidationErrors(error)

		// Fields without colon are ignored (can't map to specific field)
		expect(result).toEqual({})
	})

	it('handles errors with whitespace around field names', () => {
		const error = {
			message: 'invalid data',
			code: 2002,
			invalid_fields: ['email : not a valid email'],
		}

		const result = parseValidationErrors(error)

		expect(result).toEqual({
			email: 'not a valid email',
		})
	})

	it('returns empty object for null/undefined error', () => {
		expect(parseValidationErrors(null)).toEqual({})
		expect(parseValidationErrors(undefined)).toEqual({})
	})

	it('handles last occurrence when same field appears multiple times', () => {
		const error = {
			message: 'invalid data',
			code: 2002,
			invalid_fields: [
				'email: first error',
				'email: second error',
			],
		}

		const result = parseValidationErrors(error)

		expect(result).toEqual({
			email: 'second error',
		})
	})
})
