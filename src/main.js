import Vue from 'vue'
import App from './App.vue'
import router from './router'
import auth from './auth'

import './vikunja.scss'

Vue.config.productionTip = false

// Notifications
import Notifications from 'vue-notification'
Vue.use(Notifications)

// Icons
import { library } from '@fortawesome/fontawesome-svg-core'
import { faSignOutAlt } from '@fortawesome/free-solid-svg-icons'
import { faPlus } from '@fortawesome/free-solid-svg-icons'
import { faListOl } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

library.add(faSignOutAlt)
library.add(faPlus)
library.add(faListOl)

Vue.component('icon', FontAwesomeIcon)

// Check the user's auth status when the app starts
auth.checkAuth()

new Vue({
    router,
  render: h => h(App)
}).$mount('#app')
