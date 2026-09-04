import {describe, it, expect} from 'vitest'
import {AxiosError} from 'axios'

import {shouldDropEvent} from './sentryFilters'

describe('shouldDropEvent', () => {
	it('drops a plain AxiosError', () => {
		expect(shouldDropEvent(new AxiosError('Request failed'))).toBe(true)
	})

	it('drops an error wrapping an AxiosError as cause', () => {
		const wrapped = new Error('Error renewing token: ', {cause: new AxiosError('Request failed')})

		expect(shouldDropEvent(wrapped)).toBe(true)
	})

	it('drops an error with an AxiosError two levels deep', () => {
		const inner = new Error('inner', {cause: new AxiosError('Request failed')})

		expect(shouldDropEvent(new Error('outer', {cause: inner}))).toBe(true)
	})

	it('drops an error-like object with code and message', () => {
		expect(shouldDropEvent({code: 'ECONNABORTED', message: 'timeout'})).toBe(true)
	})

	it('keeps a plain error', () => {
		expect(shouldDropEvent(new Error('something actually broke'))).toBe(false)
	})

	it('keeps a plain error wrapping another plain error', () => {
		expect(shouldDropEvent(new Error('outer', {cause: new Error('inner')}))).toBe(false)
	})

	it('keeps undefined', () => {
		expect(shouldDropEvent(undefined)).toBe(false)
	})

	it('does not loop on a cause cycle', () => {
		const a = new Error('a')
		const b = new Error('b', {cause: a})
		a.cause = b

		expect(shouldDropEvent(a)).toBe(false)
	})
})
