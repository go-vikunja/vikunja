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
	'no-NO': 'Norsk Bokmål',
} as const

export type SupportedLocale = keyof typeof SUPPORTED_LOCALES

export const DEFAULT_LANGUAGE: SupportedLocale= 'en'

export type ISOLanguage = string

// we load all messages async
export const i18n = createI18n({
	fallbackLocale: DEFAULT_LANGUAGE,
	legacy: false,
	messages: {
		[DEFAULT_LANGUAGE]: langEN,
	} as Record<SupportedLocale, any>,
})

export async function setLanguage(lang: SupportedLocale = getCurrentLanguage()): Promise<SupportedLocale | undefined> {
	if (!lang) {
		throw new Error()
	}

	// do not change language to the current one
	if (i18n.global.locale.value === lang) {
		return
	}

	// If the language hasn't been loaded yet
	if (!i18n.global.availableLocales.includes(lang)) {
		const messages = await import(`./lang/${lang}.json`)
		i18n.global.setLocaleMessage(lang, messages.default)
	}

	i18n.global.locale.value = lang
	document.documentElement.lang = lang
	return lang
}

export function getCurrentLanguage(): SupportedLocale {
	const savedLanguage = localStorage.getItem('language') as SupportedLocale | null
	if (savedLanguage !== null) {
		return savedLanguage
	}

	const browserLanguage = navigator.language

	const language = Object.keys(SUPPORTED_LOCALES).find(langKey => {
		return langKey === browserLanguage || langKey.startsWith(browserLanguage + '-')
	}) as SupportedLocale | undefined

	return language || DEFAULT_LANGUAGE
}

export async function saveLanguage(lang: SupportedLocale) {
	localStorage.setItem('language', lang)
	await setLanguage()
}