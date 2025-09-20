import {i18n} from '@/i18n'
import {notify} from '@kyvg/vue3-notification'

export function getErrorText(r: any): string {
	const data = r?.reason?.response?.data || r?.response?.data

	if (data?.code) {
		const path = `error.${data.code}`
		// eslint-disable-next-line @typescript-eslint/ban-ts-comment
		// @ts-ignore: Complex vue-i18n type inference issue
		const translatedMessage: string = String(i18n.global.t(path))

		if (data?.code && data?.message && (data.code === 4016 || data.code === 4017 || data.code === 4018 || data.code === 4019 || data.code === 4024)) {
			return translatedMessage + '\n' + data.message
		}

		// If message and path are equal no translation exists for that error code
		if (path !== translatedMessage) {
			return translatedMessage
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

export function error(e: any, actions: Action[] = []) {
	notify({
		type: 'error',
		// @ts-ignore: Complex vue-i18n type inference issue
		title: String(i18n.global.t('error.error')),
		text: getErrorText(e),
		data: { actions },
	})
}

export function success(e: any, actions: Action[] = []) {
	notify({
		type: 'success',
		// @ts-ignore: Complex vue-i18n type inference issue
		title: String(i18n.global.t('error.success')),
		text: getErrorText(e),
		data: { actions },
	})
}
