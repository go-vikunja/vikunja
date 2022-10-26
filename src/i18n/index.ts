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

	// do not change language to the current one
	if (i18n.global.locale === lang) {
		return
	}

	// If the language hasn't been loaded yet
	if (!i18n.global.availableLocales.includes(lang)) {
		const messages = await import(`./lang/${lang}.json`)
		i18n.global.setLocaleMessage(lang, messages.default)
	}

	return setI18nLanguage(lang)
}

export function getCurrentLanguage(): SupportedLocale {
	const savedLanguage = localStorage.getItem('language')
	if (savedLanguage !== null) {
		return savedLanguage
	}

	const browserLanguage = navigator.language

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