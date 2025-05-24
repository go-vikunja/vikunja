import {useAuthStore} from '@/stores/auth'
// TODO: only import needed languages
import FlatpickrLanguages from 'flatpickr/dist/l10n'
import type { key } from 'flatpickr/dist/types/locale'
import { computed } from 'vue'

export function useFlatpickrLanguage() {
	const authStore = useAuthStore()

	return computed(() => {
		const userLanguage = authStore.settings.language
		if (!userLanguage) {
			return FlatpickrLanguages.en
		}

		const langPair = userLanguage.split('-')
		const code = userLanguage === 'vi-VN' ? 'vn' : 'en'
		const language = FlatpickrLanguages?.[langPair?.[0] as key] || FlatpickrLanguages[code]
		language.firstDayOfWeek = authStore.settings.weekStart ?? language.firstDayOfWeek
		return language
	})
}
