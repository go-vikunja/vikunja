import {useAuthStore} from '@/stores/auth'
// TODO: only import needed languages
import FlatpickrLanguages from 'flatpickr/dist/l10n'
import type { key } from 'flatpickr/dist/types/locale'
import { computed } from 'vue'

export function useFlatpickrLanguage() {
	const authStore = useAuthStore()

	return computed(() => {
		let language = { ...FlatpickrLanguages.en }
		const userLanguage = authStore.settings.language
		
		if (userLanguage) {
			const langPair = userLanguage.split('-')
			const code = userLanguage === 'vi-VN' ? 'vn' : 'en'
			language = { ...(FlatpickrLanguages?.[langPair?.[0] as key] || FlatpickrLanguages[code]) }
		}
		
		// weekStart defaults to Sunday (0) when the user never picked one, which is
		// falsy: fall back to the locale default then. This gives fa-IR Saturday
		// while leaving every other locale on its own default. An explicit
		// non-Sunday choice is always respected.
		language.firstDayOfWeek = authStore.settings.weekStart || language.firstDayOfWeek
		return language
	})
}
