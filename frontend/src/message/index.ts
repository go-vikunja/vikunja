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

export type ToastOptions = {
	takeFocus?: boolean,
}

export type ToastData = {
	actions: Action[],
	takeFocus: boolean,
}

// Long enough to reach the action; Notification.vue pauses it while the toast is hovered or focused
export const ACTION_TOAST_DURATION = 10_000

function toast(type: 'error' | 'success', title: string, e: unknown, actions: Action[], options: ToastOptions = {}) {
	const data: ToastData = {
		actions,
		takeFocus: options.takeFocus ?? false,
	}
	notify({
		type,
		title,
		text: getErrorText(e),
		// Duplicates merge on title and text, which would drop this toast's own actions.
		ignoreDuplicates: actions.length === 0,
		// Notification.vue runs the pausable timer instead
		duration: actions.length > 0 ? -1 : undefined,
		data,
	})
}

export function error(e: unknown, actions: Action[] = []) {
	toast('error', i18n.global.t('error.error'), e, actions)
}

export function success(e: unknown, actions: Action[] = [], options: ToastOptions = {}) {
	toast('success', i18n.global.t('error.success'), e, actions, options)
}
