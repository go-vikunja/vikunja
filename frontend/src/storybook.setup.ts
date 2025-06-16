import {setup} from '@storybook/vue3'
import {createPinia} from 'pinia'
import {i18n} from './i18n'

import cypress from '@/directives/cypress'
import FontAwesomeIcon from '@/components/misc/Icon'
import XButton from '@/components/input/button.vue'
import Modal from '@/components/misc/Modal.vue'
import Card from '@/components/misc/Card.vue'

import './styles/global.scss'

setup(app => {
    const pinia = createPinia()
    app.use(pinia)
    app.use(i18n)

    app.directive('cy', cypress)

    app.component('Icon', FontAwesomeIcon)
    app.component('XButton', XButton)
    app.component('Modal', Modal)
    app.component('Card', Card)
})

