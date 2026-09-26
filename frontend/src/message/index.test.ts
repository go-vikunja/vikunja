import {describe, it, expect} from 'vitest'

import {getErrorText} from './index'

describe('getErrorText', () => {
	it('interpolates i18n_params into the translated error message', () => {
		const text = getErrorText({
			code: 14002,
			message: 'The permission frobnicate of group tasks is invalid.',
			i18n_params: {permission: 'frobnicate', group: 'tasks'},
		})

		expect(text).toContain('frobnicate')
		expect(text).toContain('tasks')
	})

	it('falls back to empty placeholders when i18n_params is missing, without crashing', () => {
		const text = getErrorText({
			code: 14002,
			message: 'The permission frobnicate of group tasks is invalid.',
		})

		expect(text).not.toContain('{permission}')
		expect(text).not.toContain('{group}')
		expect(text).not.toContain('undefined')
	})

	it('falls back to the message when there is no error code', () => {
		const text = getErrorText({
			message: 'Something went wrong',
		})

		expect(text).toBe('Something went wrong')
	})

	it('falls back to the message for an unknown error code', () => {
		const text = getErrorText({
			code: 99999,
			message: 'some backend message',
		})

		expect(text).toBe('some backend message')
	})

	it('appends the server message of a request error given as cause', () => {
		const text = getErrorText(Object.assign(new Error('Error while refreshing user info:'), {
			cause: {
				status: 500,
				message: 'Internal server error',
				detail: 'Internal server error',
			},
		}))

		expect(text).toBe('Error while refreshing user info: Internal server error')
	})

	it('reads a direct problem response', () => {
		const text = getErrorText({
			code: 99999,
			detail: 'direct problem detail',
		})

		expect(text).toBe('direct problem detail')
	})
})
