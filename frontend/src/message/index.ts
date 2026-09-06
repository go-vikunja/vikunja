import {i18n} from '@/i18n'
import {notify} from '@kyvg/vue3-notification'

import {ERROR_CODE_FEATURE_DISABLED_FOR_USER, ERROR_CODE_LIMIT_REACHED} from '@/constants/entitlements'
import {useConfigStore} from '@/stores/config'

export function getErrorText(r): string {
	const data = r?.reason?.response?.data || r?.response?.data || r

	if (data?.code) {
		const path = `error.${data.code}`
		let message = i18n.global.t(path, data.i18n_params ?? {})

		if (data?.code && data?.message && (data.code === 4016 || data.code === 4017 || data.code === 4018 || data.code === 4019 || data.code === 4024)) {
			message += '\n' + data.message
		}

		// If message and path are equal no translation exists for that error code
		if (path !== message) {
			return message
		}
	}
	
	// v2 errors are RFC 9457 problem+json, which carries `detail` instead of `message`.
	let message = data?.message || data?.detail || r.message
	
	if (typeof r.cause?.message !== 'undefined') {
		message += ' ' + r.cause.message
	}

	return message
}

export function translatedError(key: string): Error {
	return new Error(i18n.global.t(key))
}

export interface Action {
	title: string,
	callback: () => void,
}

// Empty when the instance has no upgrade url, so self-hosted users never see a prompt.
export function upgradeActions(): Action[] {
	const upgradeUrl = useConfigStore().upgradeUrl
	if (!upgradeUrl) {
		return []
	}
	return [{
		title: i18n.global.t('entitlement.upgrade'),
		callback: () => window.open(upgradeUrl, '_blank', 'noopener'),
	}]
}

function isEntitlementError(e): boolean {
	const code = (e?.reason?.response?.data || e?.response?.data || e)?.code
	return code === ERROR_CODE_LIMIT_REACHED || code === ERROR_CODE_FEATURE_DISABLED_FOR_USER
}

export function error(e, actions: Action[] = []) {
	notify({
		type: 'error',
		title: i18n.global.t('error.error'),
		text: getErrorText(e),
		ignoreDuplicates: true,
		data: {
			actions: isEntitlementError(e) ? [...actions, ...upgradeActions()] : actions,
		},
	})
}

export function success(e, actions: Action[] = []) {
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
