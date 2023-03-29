import './polyfills'
import {defineSetupVue3} from '@histoire/plugin-vue'
import {i18n} from './i18n'

// import './histoire.css' // Import global CSS
import './styles/global.scss'

import {createPinia} from 'pinia'

import cypress from '@/directives/cypress'

import FontAwesomeIcon from '@/components/misc/Icon'
import XButton from '@/components/input/button.vue'
import Modal from '@/components/misc/modal.vue'
import Card from '@/components/misc/card.vue'

export const setupVue3 = defineSetupVue3(({ app }) => {
	// Add Pinia store
  const pinia = createPinia()
  app.use(pinia)
	app.use(i18n)

	app.directive('cy', cypress)

	app.component('icon', FontAwesomeIcon)
	app.component('XButton', XButton)
	app.component('modal', Modal)
	app.component('card', Card)
})