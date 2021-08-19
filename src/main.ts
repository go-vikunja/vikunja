import { createApp } from 'vue'

import App from './App.vue'
import router from './router'

import {error, success} from './message'

declare global {
	interface Window {
		API_URL: string;
	}
}

import {formatDate, formatDateSince} from '@/helpers/time/formatDate'
// @ts-ignore
import {VERSION} from './version.json'

// Add CSS
import './styles/vikunja.scss'
// Notifications
import Notifications from 'vue-notification'
// PWA
import './registerServiceWorker'

// Shortcuts
// @ts-ignore - no types available
import vueShortkey from 'vue-shortkey'
// Vuex
import {store} from './store'
// i18n
import VueI18n from 'vue-i18n' // types
import {i18n} from './i18n/setup'

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

Vue.use(Notifications)


Vue.use(vueShortkey, {prevent: ['input', 'textarea', '.input', '[contenteditable]']})

app.config.globalProperties.$message = {
	error(e, actions = []) {
		return error(e, Vue.prototype, actions)
	},
	success(s, actions = []) {
		return success(s, Vue.prototype, actions)
	},
}

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
import message from './message'
import {getNamespaceTitle} from './helpers/getNamespaceTitle'
import {getListTitle} from './helpers/getListTitle'
import {colorIsDark} from './helpers/color/colorIsDark'
import {setTitle} from './helpers/setTitle'
app.mixin({
	methods: {
		formatDateSince(date) {
			return formatDateSince(date, (p: VueI18n.Path, params?: VueI18n.Values) => this.$t(p, params))
		},
		formatDate(date) {
			return formatDate(date, 'PPPPpppp', this.$t('date.locale'))
		},
		formatDateShort(date) {
			return formatDate(date, 'PPpp', this.$t('date.locale'))
		},
		getNamespaceTitle(n) {
			return getNamespaceTitle(n, (p: VueI18n.Path) => this.$t(p))
		},
		getListTitle(l) {
			return getListTitle(l, (p: VueI18n.Path) => this.$t(p))
		},
		colorIsDark: colorIsDark,
		setTitle: setTitle,
	},
})

app.use(router)
app.use(store)
app.use(i18n)

app.mount('#app')