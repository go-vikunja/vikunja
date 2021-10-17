import { createApp, configureCompat } from 'vue'

configureCompat({
	COMPONENT_V_MODEL: false,
	COMPONENT_ASYNC: false,
	RENDER_FUNCTION: false,
	WATCH_ARRAY: false, // TODO: check this again; this might lead to some problemes
	TRANSITION_GROUP_ROOT: false,
})

import App from './App.vue'
import router from './router'

import {error, success} from './message'

declare global {
	interface Window {
		API_URL: string;
	}
}

import {formatDate, formatDateShort, formatDateLong, formatDateSince} from '@/helpers/time/formatDate'
// @ts-ignore
import {VERSION} from './version.json'

// Add CSS
import './styles/vikunja.scss'
// Notifications
import Notifications from '@kyvg/vue3-notification'

// PWA
import './registerServiceWorker'

// Shortcuts
import shortkey from '@/plugins/shortkey'
// Vuex
import {store} from './store'
// i18n
import {i18n} from './i18n'

console.info(`Vikunja frontend version ${VERSION}`)

// Check if we have an api url in local storage and use it if that's the case
const apiUrlFromStorage = localStorage.getItem('API_URL')
if (apiUrlFromStorage !== null) {
	window.API_URL = apiUrlFromStorage
}

// Make sure the api url does not contain a / at the end
if (window.API_URL.substr(window.API_URL.length - 1, window.API_URL.length) === '/') {
	window.API_URL = window.API_URL.substr(0, window.API_URL.length - 1)
}

const app = createApp(App)

app.use(Notifications)



app.use(shortkey, {prevent: ['input', 'textarea', '.input', '[contenteditable]']})

// directives
import focus from './directives/focus'
import tooltip from './directives/tooltip'
app.directive('focus', focus)
app.directive('tooltip', tooltip)

// global components
import FontAwesomeIcon from './icons'
import Button from './components/input/button.vue'
import Modal from './components/modal/modal.vue'
import Card from './components/misc/card.vue'
app.component('icon', FontAwesomeIcon)
app.component('x-button', Button)
app.component('modal', Modal)
app.component('card', Card)

// Mixins
import {getNamespaceTitle} from './helpers/getNamespaceTitle'
import {getListTitle} from './helpers/getListTitle'
import {colorIsDark} from './helpers/color/colorIsDark'
import {setTitle} from './helpers/setTitle'
app.mixin({
	methods: {
		formatDateSince,
		format: formatDate,
		formatDate: formatDateLong,
		formatDateShort: formatDateShort,
		getNamespaceTitle,
		getListTitle,
		colorIsDark,
		setTitle,
	},
})

app.config.errorHandler = (err, vm, info) => {
	// if (import.meta.env.PROD) {
		// error(err)
	// } else {
		// console.error(err, vm, info)
		error(err)
	// }
}

if (import.meta.env.DEV) {
	app.config.warnHandler = (msg, vm, info) => { 
		error(msg)
	}

	// https://stackoverflow.com/a/52076738/15522256
	window.addEventListener('error', (err) => {
		error(err)
		throw err
	})
	
	
	window.addEventListener('unhandledrejection', (err) => {
		// event.promise contains the promise object
		// event.reason contains the reason for the rejection
		error(err)
		throw err
	})
}

app.config.globalProperties.$message = {
	error,
	success,
}

app.use(router)
app.use(store)
app.use(i18n)

app.mount('#app')