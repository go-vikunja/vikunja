const TRANSLATION_KEY_PREFIX = 'home.welcome'

export function hourToSalutation(now: Date = new Date()): String {
	const hours = new Date(now).getHours()

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
