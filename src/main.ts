import {createApp} from 'vue'

import pinia from './pinia'
import router from './router'
import App from './App.vue'

import {error, success} from './message'

import {VERSION} from './version.json'

// Notifications
import Notifications from '@kyvg/vue3-notification'

// PWA
import './registerServiceWorker'

// i18n
import {i18n} from './i18n'

declare global {
	interface Window {
		API_URL: string;
		SENTRY_ENABLED: boolean;
		SENTRY_DSN: string;
	}
}

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

// directives
import focus from '@/directives/focus'
// @ts-ignore The export does exist, ts just doesn't find it.
import { VTooltip } from 'v-tooltip'
import 'v-tooltip/dist/v-tooltip.css'
import shortcut from '@/directives/shortcut'
import cypress from '@/directives/cypress'

app.directive('focus', focus)
app.directive('tooltip', VTooltip)
app.directive('shortcut', shortcut)
app.directive('cy', cypress)

// global components
import FontAwesomeIcon from '@/components/misc/Icon'
import Button from '@/components/input/button.vue'
import Modal from '@/components/misc/modal.vue'
import Card from '@/components/misc/card.vue'

app.component('icon', FontAwesomeIcon)
app.component('x-button', Button)
app.component('modal', Modal)
app.component('card', Card)

app.config.errorHandler = (err, vm, info) => {
	if (import.meta.env.DEV) {
		console.error(err, vm, info)
	}
	error(err)
}

if (import.meta.env.DEV) {
	app.config.warnHandler = (msg) => {
		error(msg)
		throw(msg)
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

if (window.SENTRY_ENABLED) {
	import('./sentry').then(sentry => sentry.default(app, router))
}

app.use(pinia)
app.use(router)
app.use(i18n)

app.mount('#app')