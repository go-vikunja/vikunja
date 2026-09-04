import {describe, it, expect} from 'vitest'
import {AxiosError} from 'axios'

import {shouldDropEvent} from './sentryFilters'

// Object.assign instead of `new Error(msg, {cause})`: the vitest tsconfig
// targets a lib without the two-argument Error constructor.
function errorWithCause(message: string, cause: unknown): Error {
	return Object.assign(new Error(message), {cause})
}

describe('shouldDropEvent', () => {
	it('drops a plain AxiosError', () => {
		expect(shouldDropEvent(new AxiosError('Request failed'))).toBe(true)
	})

	it('drops an error wrapping an AxiosError as cause', () => {
		expect(shouldDropEvent(errorWithCause('Error renewing token: ', new AxiosError('Request failed')))).toBe(true)
	})

	it('drops an error with an AxiosError two levels deep', () => {
		const inner = errorWithCause('inner', new AxiosError('Request failed'))

		expect(shouldDropEvent(errorWithCause('outer', inner))).toBe(true)
	})

	it('drops an error-like object with code and message', () => {
		expect(shouldDropEvent({code: 'ECONNABORTED', message: 'timeout'})).toBe(true)
	})

	it('keeps a plain error', () => {
		expect(shouldDropEvent(new Error('something actually broke'))).toBe(false)
	})

	it('keeps a plain error wrapping another plain error', () => {
		expect(shouldDropEvent(errorWithCause('outer', new Error('inner')))).toBe(false)
	})

	it('keeps undefined', () => {
		expect(shouldDropEvent(undefined)).toBe(false)
	})

	it('does not loop on a cause cycle', () => {
		const a = new Error('a')
		const b = errorWithCause('b', a)
		Object.assign(a, {cause: b})

		expect(shouldDropEvent(a)).toBe(false)
	})
})

describe('shouldDropEvent with chunk load errors', () => {
	const messages = [
		'Failed to fetch dynamically imported module: https://try.vikunja.io/assets/ProjectList-abc123.js',
		'error loading dynamically imported module: https://try.vikunja.io/assets/ProjectList-abc123.js',
		'Importing a module script failed.',
		'Unable to preload CSS for /assets/ProjectList-abc123.css',
	]

	it.each(messages)('drops the exception %s', message => {
		expect(shouldDropEvent(new Error(message))).toBe(true)
	})

	it.each(messages)('drops the event message %s', message => {
		expect(shouldDropEvent(undefined, {message})).toBe(true)
	})

	it.each(messages)('drops the event exception value %s', message => {
		expect(shouldDropEvent(undefined, {exception: {values: [{value: message}]}})).toBe(true)
	})

	it('keeps an unrelated event message', () => {
		expect(shouldDropEvent(undefined, {message: 'something actually broke'})).toBe(false)
	})
})
