import {PRIORITIES} from '@/constants/priorities'
import {getItemsFromPrefix} from './prefixParser'

export const getPriority = (text: string, prefix: string): number | null => {
	const ps = getItemsFromPrefix(text, prefix)
	if (ps.length === 0) {
		return null
	}

	for (const p of ps) {
		// Only a whole number is a priority: "!2fa" or "!3rd" is part of the title
		if (!/^\d+$/.test(p)) {
			continue
		}
		for (const pi of Object.values(PRIORITIES)) {
			if (pi === parseInt(p)) {
				return parseInt(p)
			}
		}
	}

	return null
}
