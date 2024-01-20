import {useAuthStore} from '@/stores/auth'
import FlatpickrLanguages from 'flatpickr/dist/l10n'
import type { CustomLocale, key } from 'flatpickr/dist/types/locale'

export function getFlatpickrLanguage(): CustomLocale {
	const authStore = useAuthStore()
	const lang = authStore.settings.language
	const langPair = lang.split('-')
	let language = FlatpickrLanguages[lang === 'vi-vn' ? 'vn' : 'en']
	if (langPair.length > 0 && FlatpickrLanguages[langPair[0] as key] !== undefined) {
		language = FlatpickrLanguages[langPair[0] as key]
	}
	language.firstDayOfWeek = authStore.settings.weekStart ?? language.firstDayOfWeek
	return language
}