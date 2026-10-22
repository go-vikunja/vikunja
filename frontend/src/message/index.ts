import {i18n} from '@/i18n'
import {notify} from '@kyvg/vue3-notification'

function errorRecord(value: unknown): Record<string, unknown> {
	return value !== null && typeof value === 'object' ? value as Record<string, unknown> : {}
}

export function getErrorText(reason: unknown): string {
	const original = errorRecord(reason)
	// The dev rejection handler passes the event rather than its reason.
	const data = errorRecord(original.reason ?? reason)
	const code = data.code

	if (typeof code === 'number' && code) {
		const path = `error.${code}`
		let message = i18n.global.t(path, errorRecord(data.i18n_params))
		if (typeof data.message === 'string' && [4016, 4017, 4018, 4019, 4024].includes(code)) {
			message += '\n' + data.message
		}
		if (path !== message) return message
	}

	// v2 problem responses carry detail instead of message.
	const fallback = data.message || data.detail || original.message || original.reason || reason
	let message = typeof fallback === 'string' ? fallback : ''
	const cause = errorRecord(original.cause)
	const causeMessage = cause.detail ?? cause.message
	if (typeof causeMessage === 'string') message += ' ' + causeMessage
	return message
}

export function translatedError(key: string): Error {
	return new Error(i18n.global.t(key))
}

export interface Action {
	title: string,
	callback: () => void,
}

export function error(e: unknown, actions: Action[] = []) {
	notify({
		type: 'error',
		title: i18n.global.t('error.error'),
		text: getErrorText(e),
		ignoreDuplicates: true,
		data: {
			actions: actions,
		},
	})
}

export function success(e: unknown, actions: Action[] = []) {
	notify({
		type: 'success',
		title: i18n.global.t('error.success'),
		text: getErrorText(e),
		ignoreDuplicates: true,
		data: {
			actions: actions,
		},
	})
}
