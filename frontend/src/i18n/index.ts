import {createI18n} from 'vue-i18n'
import type {PluralizationRule} from 'vue-i18n'
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
	'zh-CN': '中文',
	'no-NO': 'Norsk Bokmål',
	'es-ES': 'Español',
	'da-DK': 'Dansk',
	'ja-JP': '日本語',
	'hu-HU': 'Magyar',
	'ar-SA': 'اَلْعَرَبِيَّةُ',
	'sl-SI': 'Slovenščina',
	'pt-BR': 'Português Brasileiro',
	'hr-HR': 'Hrvatski',
	'uk-UA': 'Українська',
	// IMPORTANT: Also add new languages to useDayjsLanguageSync
} as const

export type SupportedLocale = keyof typeof SUPPORTED_LOCALES

export const DEFAULT_LANGUAGE: SupportedLocale= 'en'

export type ISOLanguage = string

// we load all messages async
export const i18n = createI18n({
	fallbackLocale: DEFAULT_LANGUAGE,
	legacy: false,
	pluralRules: {
		'ru-RU': (choice: number, choicesLength: number, orgRule?: PluralizationRule) => {
			if (choicesLength !== 3) {
				return orgRule ? orgRule(choice, choicesLength) : 0
			}
			const n = Math.abs(choice) % 100
			if (n > 10 && n < 20) {
				return 2
			}
			if (n % 10 === 1) {
				return 0
			}
			if (n % 10 >= 2 && n % 10 <= 4) {
				return 1
			}
			return 2
		},
	},
	messages: {
		[DEFAULT_LANGUAGE]: langEN,
	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	} as Record<SupportedLocale, any>,
})

export async function setLanguage(lang: SupportedLocale): Promise<SupportedLocale | undefined> {
	if (!lang) {
		throw new Error()
	}

	// do not change language to the current one
	if (i18n.global.locale.value === lang) {
		return
	}

	// If the language hasn't been loaded yet
	if (!i18n.global.availableLocales.includes(lang)) {
		try {
			const messages = await import(`./lang/${lang}.json`)
			i18n.global.setLocaleMessage(lang, messages.default)
		} catch (e) {
			console.error(`Failed to load language ${lang}:`, e)
			return setLanguage(getBrowserLanguage())
		}
	}

	i18n.global.locale.value = lang
	document.documentElement.lang = lang
	return lang
}

export function getBrowserLanguage(): SupportedLocale {
	const browserLanguage = navigator.language

	const language = Object.keys(SUPPORTED_LOCALES).find(langKey => {
		return langKey === browserLanguage || langKey.startsWith(browserLanguage + '-')
	}) as SupportedLocale | undefined

	return language || DEFAULT_LANGUAGE
}
