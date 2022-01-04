import {computed} from 'vue'
import {useNow} from '@vueuse/core'

const TRANSLATION_KEY_PREFIX = 'home.welcome'

export function hourToSalutation(now: Date) {
	const hours = now.getHours()

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

export function useDateTimeSalutation() {
	const now = useNow()
	return computed(() => hourToSalutation(now.value))
}