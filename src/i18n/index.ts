import {computed, ref, watch} from 'vue'
import {createI18n} from 'vue-i18n'
import langEN from './lang/en.json'

export const SUPPORTED_LOCALES = {
	'en': 'English',
	'de-DE': 'Deutsch',
	'de-swiss': 'Schwizertütsch',
	'ru-RU': 'Русский',
	'fr-FR': 'Français',
	'vi-VN': 'Tiếng Việt',
	'it-IT': 'Italiano',
	'cs-CZ': 'Čeština',
	'pl-PL': 'Polski',
	'nl-NL': 'Nederlands',
	'pt-PT': 'Português',
	'zh-CN': 'Chinese',
} as Record<string, string>

export type SupportedLocale = keyof typeof SUPPORTED_LOCALES

export const DEFAULT_LANGUAGE: SupportedLocale= 'en'

export type ISOLanguage = string

export const DAYJS_LOCALE_MAPPING = {
	'de-swiss': 'de-AT',
} as Record<SupportedLocale, ISOLanguage>

export const i18n = createI18n({
	locale: DEFAULT_LANGUAGE, // set locale
	fallbackLocale: DEFAULT_LANGUAGE,
	legacy: true,
	globalInjection: true,
	allowComposition: true,
	inheritLocale: true,
	messages: {
		en: langEN,
	} as Record<SupportedLocale, any>,
})

function setI18nLanguage(lang: SupportedLocale): SupportedLocale {
	i18n.global.locale = lang
	document.documentElement.lang = lang
	return lang
}

export async function loadLanguageAsync(lang: SupportedLocale) {
	if (!lang) {
		throw new Error()
	}

	if (
		// If the same language
		i18n.global.locale === lang ||
		// If the language was already loaded
		i18n.global.availableLocales.includes(lang)
	) {
		return setI18nLanguage(lang)
	}

	// If the language hasn't been loaded yet
	const messages = await import(`./lang/${lang}.json`)
	i18n.global.setLocaleMessage(lang, messages.default)
	return setI18nLanguage(lang)
}

export function getCurrentLanguage(): SupportedLocale {
	const savedLanguage = localStorage.getItem('language')
	if (savedLanguage !== null) {
		return savedLanguage
	}

	const browserLanguage = navigator.language || navigator.userLanguage

	const language: SupportedLocale | undefined = Object.keys(SUPPORTED_LOCALES).find(langKey => {
		return langKey === browserLanguage || langKey.startsWith(browserLanguage + '-')
	})

	return language || DEFAULT_LANGUAGE
}

export function saveLanguage(lang: SupportedLocale) {
	localStorage.setItem('language', lang)
	setLanguage()
}

export function setLanguage() {
	return loadLanguageAsync(getCurrentLanguage())
}

import type dayjs from 'dayjs'

export function useDayjsLanguageSync(dayjsGlobal: typeof dayjs) {
	const dayjsLanguageLoaded = ref(false)
	watch(
		() => i18n.global.locale,
		async (currentLanguage: string) => {
			if (!dayjsGlobal) {
				return
			}
			const dayjsLanguageCode = DAYJS_LOCALE_MAPPING[currentLanguage.toLowerCase()] || currentLanguage.toLowerCase()
			dayjsLanguageLoaded.value = dayjsGlobal.locale() === dayjsLanguageCode
			if (dayjsLanguageLoaded.value) {
				return
			}
			console.log('foo')
			await import(`../../node_modules/dayjs/locale/${dayjsLanguageCode}.js`)
			console.log('bar')
			dayjsGlobal.locale(dayjsLanguageCode)
			dayjsLanguageLoaded.value = true
		},
		{immediate: true},
	)

	// we export the loading state since that's easier to work with
	const isLoading = computed(() => !dayjsLanguageLoaded.value)

	return isLoading
}