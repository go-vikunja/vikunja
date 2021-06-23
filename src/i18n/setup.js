import Vue from 'vue'
import VueI18n from 'vue-i18n'

Vue.use(VueI18n)

export const i18n = new VueI18n({
	locale: 'en', // set locale
	fallbackLocale: 'en',
	messages: {
		en: require('./lang/en.json'),
	},
})

export const availableLanguages = {
	en: 'English',
	de: 'Deutsch',
}

const loadedLanguages = ['en'] // our default language that is preloaded

const setI18nLanguage = lang => {
	i18n.locale = lang
	document.querySelector('html').setAttribute('lang', lang)
	return lang
}

export const loadLanguageAsync = lang => {
	// If the same language
	if (i18n.locale === lang) {
		return Promise.resolve(setI18nLanguage(lang))
	}

	// If the language was already loaded
	if (loadedLanguages.includes(lang)) {
		return Promise.resolve(setI18nLanguage(lang))
	}

	// If the language hasn't been loaded yet
	return import(/* webpackChunkName: "lang-[request]" */ `@/i18n/lang/${lang}.json`).then(
		messages => {
			i18n.setLocaleMessage(lang, messages.default)
			loadedLanguages.push(lang)
			return setI18nLanguage(lang)
		}
	)
}

export const getCurrentLanguage = () => {
	const savedLanguage = localStorage.getItem('language')
	if(savedLanguage !== null) {
		return savedLanguage
	}

	let browserLanguage = navigator.language || navigator.userLanguage

	if (browserLanguage.startsWith('en-')) {
		browserLanguage = 'en'
	}

	if (typeof availableLanguages[browserLanguage] !== 'undefined') {
		return browserLanguage
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
