import {computed, ref, watch} from 'vue'
import type dayjs from 'dayjs'

import {i18n, type ISOLanguage, type SupportedLocale} from '@/i18n'

export const DAYJS_LOCALE_MAPPING = {
	'de-de': 'de',
	'de-swiss': 'de-ch',
	'ru-ru': 'ru',
	'fr-fr': 'fr',
	'vi-vn': 'vi',
	'it-it': 'it',
	'cs-cz': 'cs',
	'pl-pl': 'pl',
	'nl-nl': 'nl',
	'pt-pt': 'pt',
	'zh-cn': 'zh-cn',
	'no-no': 'nb',
	'es-es': 'es',
	'da-dk': 'da',
	'ja-jp': 'ja',
	'hu-hu': 'hu',
	'ar-sa': 'ar-sa',
	'sl-si': 'sl',
	'pt-br': 'pt',
	'hr-hr': 'hr',
	'uk-ua': 'uk',
	'lt-lt': 'lt',
	'bg-bg': 'bg',
	'ko-kr': 'ko',
	'tr-tr': 'tr',
	'fi-fi': 'fi',
	'he-il': 'he',
} as Record<SupportedLocale, ISOLanguage>

export const DAYJS_LANGUAGE_IMPORTS = {
	'de-de': () => import('dayjs/locale/de'),
	'de-swiss': () => import('dayjs/locale/de-ch'),
	'ru-ru': () => import('dayjs/locale/ru'),
	'fr-fr': () => import('dayjs/locale/fr'),
	'vi-vn': () => import('dayjs/locale/vi'),
	'it-it': () => import('dayjs/locale/it'),
	'cs-cz': () => import('dayjs/locale/cs'),
	'pl-pl': () => import('dayjs/locale/pl'),
	'nl-nl': () => import('dayjs/locale/nl'),
	'pt-pt': () => import('dayjs/locale/pt'),
	'zh-cn': () => import('dayjs/locale/zh-cn'),
	'no-no': () => import('dayjs/locale/nb'),
	'es-es': () => import('dayjs/locale/es'),
	'da-dk': () => import('dayjs/locale/da'),
	'ja-jp': () => import('dayjs/locale/ja'),
	'hu-hu': () => import('dayjs/locale/hu'),
	'ar-sa': () => import('dayjs/locale/ar-sa'),
	'sl-si': () => import('dayjs/locale/sl'),
	'pt-br': () => import('dayjs/locale/pt-br'),
	'hr-hr': () => import('dayjs/locale/hr'),
	'uk-ua': () => import('dayjs/locale/uk'),
	'lt-lt': () => import('dayjs/locale/lt'),
	'bg-bg': () => import('dayjs/locale/bg'),
	'ko-kr': () => import('dayjs/locale/ko'),
	'tr-tr': () => import('dayjs/locale/tr'),
	'fi-fi': () => import('dayjs/locale/fi'),
	'he-il': () => import('dayjs/locale/he'),
} as Record<SupportedLocale, () => Promise<ILocale>>

export async function loadDayJsLocale(language: SupportedLocale) {
	if (language === 'en') {
		return
	}

	await DAYJS_LANGUAGE_IMPORTS[language.toLowerCase()]()
}

export function useDayjsLanguageSync(dayjsGlobal: typeof dayjs) {

	const dayjsLanguageLoaded = ref(false)
	watch(
		() => i18n.global.locale.value,
		async (currentLanguage: string) => {
			if (!dayjsGlobal) {
				return
			}
			const dayjsLanguageCode = DAYJS_LOCALE_MAPPING[currentLanguage.toLowerCase()] || currentLanguage.toLowerCase()
			dayjsLanguageLoaded.value = dayjsGlobal.locale() === dayjsLanguageCode
			if (dayjsLanguageLoaded.value) {
				return
			}
			await loadDayJsLocale(currentLanguage)
			dayjsGlobal.locale(dayjsLanguageCode)
			dayjsLanguageLoaded.value = true
		},
		{immediate: true},
	)

	// we export the loading state since that's easier to work with
	const isLoading = computed(() => !dayjsLanguageLoaded.value)

	return isLoading
}
