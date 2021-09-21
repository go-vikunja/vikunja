import Vue from 'vue'
import App from './App.vue'
import router from './router'

declare global {
	interface Window {
		API_URL: string;
	}
}

import {formatDate, formatDateSince} from '@/helpers/time/formatDate'
// @ts-ignore
import {VERSION} from './version.json'

// Register the modal
// @ts-ignore
import Modal from './components/modal/modal'
// Add CSS
import './styles/vikunja.scss'
// Notifications
import Notifications from 'vue-notification'
// PWA
import './registerServiceWorker'

// Shortcuts
// @ts-ignore - no types available
import vueShortkey from 'vue-shortkey'
// Mixins
import message from './message'
import {colorIsDark} from './helpers/color/colorIsDark'
import {setTitle} from './helpers/setTitle'
import {getNamespaceTitle} from './helpers/getNamespaceTitle'
import {getListTitle} from './helpers/getListTitle'
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

Vue.component('modal', Modal)

Vue.config.productionTip = false

Vue.use(Notifications)

import FontAwesomeIcon from './icons'
Vue.component('icon', FontAwesomeIcon)

Vue.use(vueShortkey, {prevent: ['input', 'textarea', '.input']})

import focus from './directives/focus'
Vue.directive('focus', focus)

import tooltip from './directives/tooltip'

// @ts-ignore
Vue.directive('tooltip', tooltip)

// @ts-ignore
import Button from './components/input/button'
Vue.component('x-button', Button)

// @ts-ignore
import Card from './components/misc/card'
Vue.component('card', Card)

Vue.mixin({
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
		error(e, actions = []) {
			return message.error(e, this, (p: VueI18n.Path) => this.$t(p), actions)
		},
		success(s, actions = []) {
			return message.success(s, this, (p: VueI18n.Path) => this.$t(p), actions)
		},
		colorIsDark: colorIsDark,
		setTitle: setTitle,
	},
})

new Vue({
	router,
	store,
	i18n,
	render: h => h(App),
}).$mount('#app')
