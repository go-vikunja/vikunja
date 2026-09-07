import {computed} from 'vue'

import {i18n} from '@/i18n'
import {useAuthStore} from '@/stores/auth'
import {
	formatJalaliDate,
	formatJalaliParts,
	isJalaliLocale,
	normalizePersianDigits,
	resolveUserTimezone,
	toPersianDigits,
} from '@/helpers/time/jalali'

/**
 * Thin reactive access to the Jalali helpers. No business logic lives here:
 * formatting, parsing and digit rules stay in helpers/time/jalali.ts so they
 * remain usable without Vue.
 */
export function useJalaliCalendar() {
	const authStore = useAuthStore()
	const locale = computed(() => i18n.global.locale.value)
	const isJalali = computed(() => isJalaliLocale(locale.value))
	// Display/selection timezone: the user's configured timezone, falling back
	// to the browser zone. Reactive so pickers follow settings changes.
	const timeZone = computed(() => resolveUserTimezone(authStore.settings.timezone))

	return {
		locale,
		isJalali,
		timeZone,
		isJalaliLocale,
		formatJalaliDate,
		formatJalaliParts,
		normalizePersianDigits,
		toPersianDigits,
	}
}
