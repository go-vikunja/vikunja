import {i18n} from '@/i18n'
import {notify} from '@kyvg/vue3-notification'

export function getErrorText(r): string {
	const data = r?.reason?.response?.data || r?.response?.data

	if (data?.code) {
		const path = `error.${data.code}`
		let message = i18n.global.t(path)

		if (data?.code && data?.message && (data.code === 4016 || data.code === 4017 || data.code === 4018 || data.code === 4019 || data.code === 4024)) {
			message += '\n' + data.message
		}

		// If message and path are equal no translation exists for that error code
		if (path !== message) {
			return message
		}
	}
	
	let message = data?.message || r.message
	
	if (typeof r.cause?.message !== 'undefined') {
		message += ' ' + r.cause.message
	}

	return message
}

export interface Action {
	title: string,
	callback: () => void,
}

export function error(e, actions: Action[] = []) {
	notify({
		type: 'error',
		title: i18n.global.t('error.error'),
		text: [getErrorText(e)],
		actions: actions,
	})
}

export function success(e, actions: Action[] = []) {
	notify({
		type: 'success',
		title: i18n.global.t('error.success'),
		text: [getErrorText(e)],
		data: {
			actions: actions,
		},
	})
}
