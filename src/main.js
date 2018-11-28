import Vue from 'vue'
import App from './App.vue'
import router from './router'
import auth from './auth'

// Register the modal
import Modal from './components/modal/Modal'
Vue.component('modal', Modal)

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
import { faTasks } from '@fortawesome/free-solid-svg-icons'
import { faCog } from '@fortawesome/free-solid-svg-icons'
import { faAngleRight } from '@fortawesome/free-solid-svg-icons'
import { faLayerGroup } from '@fortawesome/free-solid-svg-icons'
import { faTrashAlt } from '@fortawesome/free-solid-svg-icons'
import { faUsers } from '@fortawesome/free-solid-svg-icons'
import { faUser } from '@fortawesome/free-solid-svg-icons'
import { faLock } from '@fortawesome/free-solid-svg-icons'
import { faPen } from '@fortawesome/free-solid-svg-icons'
import { faTimes } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

library.add(faSignOutAlt)
library.add(faPlus)
library.add(faListOl)
library.add(faTasks)
library.add(faCog)
library.add(faAngleRight)
library.add(faLayerGroup)
library.add(faTrashAlt)
library.add(faUsers)
library.add(faUser)
library.add(faLock)
library.add(faPen)
library.add(faTimes)

Vue.component('icon', FontAwesomeIcon)

// Check the user's auth status when the app starts
auth.checkAuth()

new Vue({
    router,
  render: h => h(App)
}).$mount('#app')
