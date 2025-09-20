import {i18n} from '@/i18n'
import {notify} from '@kyvg/vue3-notification'

export function getErrorText(r: unknown): string {
	// Type guards for the error object
	const hasNestedData = (obj: unknown): obj is { reason?: { response?: { data?: unknown } }, response?: { data?: unknown } } => {
		return typeof obj === 'object' && obj !== null
	}

	const hasMessage = (obj: unknown): obj is { message?: string } => {
		return typeof obj === 'object' && obj !== null
	}

	const hasCause = (obj: unknown): obj is { cause?: { message?: string } } => {
		return typeof obj === 'object' && obj !== null && 'cause' in obj
	}

	const isDataWithCode = (obj: unknown): obj is { code?: number, message?: string } => {
		return typeof obj === 'object' && obj !== null
	}

	if (!hasNestedData(r)) {
		return String(r || 'Unknown error')
	}

	const data = r.reason?.response?.data || r.response?.data

	if (isDataWithCode(data) && data.code) {
		const path = `error.${data.code}`
		// @ts-expect-error: Complex vue-i18n type inference issue
		const translatedMessage: string = String(i18n.global.t(path))

		if (data.code && data.message && (data.code === 4016 || data.code === 4017 || data.code === 4018 || data.code === 4019 || data.code === 4024)) {
			return translatedMessage + '\n' + data.message
		}

		// If message and path are equal no translation exists for that error code
		if (path !== translatedMessage) {
			return translatedMessage
		}
	}

	let message = (isDataWithCode(data) ? data.message : undefined) || (hasMessage(r) ? r.message : undefined) || 'Unknown error'

	if (hasCause(r) && typeof r.cause?.message !== 'undefined') {
		message += ' ' + r.cause.message
	}

	return message
}

export interface Action {
	title: string,
	callback: () => void,
}

export function error(e: unknown, actions: Action[] = []) {
	notify({
		type: 'error',
		title: String(i18n.global.t('error.error')),
		text: getErrorText(e),
		data: { actions },
	})
}

export function success(e: unknown, actions: Action[] = []) {
	notify({
		type: 'success',
		title: String(i18n.global.t('error.success')),
		text: getErrorText(e),
		data: { actions },
	})
}
