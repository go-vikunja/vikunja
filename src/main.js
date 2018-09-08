import Vue from 'vue'
import App from './App.vue'
import router from './router'
import auth from './auth'

import '../node_modules/bulma/bulma.sass'

Vue.config.productionTip = false

// Notifications
import Notifications from 'vue-notification'
Vue.use(Notifications)

// Check the user's auth status when the app starts
auth.checkAuth()

new Vue({
    router,
  render: h => h(App)
}).$mount('#app')
