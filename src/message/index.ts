import {i18n} from '@/i18n'
import {notify} from '@kyvg/vue3-notification'

export function getErrorText(r): string {
	let data = undefined
	if (r?.response?.data) {
		data = r.response.data
	}
	
	if (r?.reason?.response?.data) {
		data = r.reason.response.data
	}
	
	if (data) {
		if(data.code) {
			const path = `error.${data.code}`
			const message = i18n.global.t(path)

			// If message and path are equal no translation exists for that error code
			if (path !== message) {
				return message
			}
		}

		if (data.message) {
			return data.message
		}
	}

	return r.message
}

export function error(e, actions = []) {
	notify({
		type: 'error',
		title: i18n.global.t('error.error'),
		text: [getErrorText(e)],
		actions: actions,
	})
}

export function success(e, actions = []) {
	notify({
		type: 'success',
		title: i18n.global.t('error.success'),
		text: [getErrorText(e)],
		data: {
			actions: actions,
		},
	})
}