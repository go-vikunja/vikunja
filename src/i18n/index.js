import {createI18n} from 'vue-i18n'
import langEN from './lang/en.json'

export const i18n = createI18n({
	locale: 'en', // set locale
	fallbackLocale: 'en',
	globalInjection: true,
	messages: {
		en: langEN,
	},
})

export const availableLanguages = {
	en: 'English',
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
}

const loadedLanguages = ['en'] // our default language that is preloaded

const setI18nLanguage = lang => {
	i18n.global.locale = lang
	document.querySelector('html').setAttribute('lang', lang)
	return lang
}

export const loadLanguageAsync = lang => {
	if (!lang) {
		return
	}

	if (
		// If the same language
		i18n.global.locale === lang ||
		// If the language was already loaded
		loadedLanguages.includes(lang)
	) {
		return setI18nLanguage(lang)
	}

	// If the language hasn't been loaded yet
	return import(`./lang/${lang}.json`).then(
		messages => {
			i18n.global.setLocaleMessage(lang, messages.default)
			loadedLanguages.push(lang)
			return setI18nLanguage(lang)
		},
	)
}

export const getCurrentLanguage = () => {
	const savedLanguage = localStorage.getItem('language')
	if (savedLanguage !== null) {
		return savedLanguage
	}

	let browserLanguage = navigator.language || navigator.userLanguage

	for (let k in availableLanguages) {
		if (browserLanguage[k] === browserLanguage || k.startsWith(browserLanguage + '-')) {
			return k
		}
	}

	return 'en'
}

export const saveLanguage = lang => {
	localStorage.setItem('language', lang)
	setLanguage()
}

export const setLanguage = () => {
	loadLanguageAsync(getCurrentLanguage())
}
