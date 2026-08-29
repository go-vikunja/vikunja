import {describe, it, expect} from 'vitest'

import {getErrorText} from './index'

describe('getErrorText', () => {
	it('interpolates i18n_params into the translated error message', () => {
		const text = getErrorText({
			response: {
				data: {
					code: 14002,
					message: 'The permission frobnicate of group tasks is invalid.',
					i18n_params: {permission: 'frobnicate', group: 'tasks'},
				},
			},
		})

		expect(text).toContain('frobnicate')
		expect(text).toContain('tasks')
	})

	it('falls back to empty placeholders when i18n_params is missing, without crashing', () => {
		const text = getErrorText({
			response: {
				data: {
					code: 14002,
					message: 'The permission frobnicate of group tasks is invalid.',
				},
			},
		})

		expect(text).not.toContain('{permission}')
		expect(text).not.toContain('{group}')
		expect(text).not.toContain('undefined')
	})

	it('falls back to data.message when there is no error code', () => {
		const text = getErrorText({
			response: {
				data: {
					message: 'Something went wrong',
				},
			},
		})

		expect(text).toBe('Something went wrong')
	})

	it('falls back to data.message for an unknown error code', () => {
		const text = getErrorText({
			response: {
				data: {
					code: 99999,
					message: 'some backend message',
				},
			},
		})

		expect(text).toBe('some backend message')
	})
})
