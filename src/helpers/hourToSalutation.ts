import {useNow} from '@vueuse/core'
import {Ref} from 'vue'

const TRANSLATION_KEY_PREFIX = 'home.welcome'

export function hourToSalutation(now: Date | Ref<Date> = useNow()): String {
	const hours = now instanceof Date ? new Date(now).getHours() : new Date(now.value).getHours()

	if (hours < 5) {
		return `${TRANSLATION_KEY_PREFIX}Night`
	}

	if (hours < 11) {
		return `${TRANSLATION_KEY_PREFIX}Morning`
	}

	if (hours < 18) {
		return `${TRANSLATION_KEY_PREFIX}Day`
	}

	if (hours < 23) {
		return `${TRANSLATION_KEY_PREFIX}Evening`
	}

	return `${TRANSLATION_KEY_PREFIX}Night`
}
