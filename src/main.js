import Vue from 'vue'
import App from './App.vue'
import router from './router'
import auth from './auth'

Vue.config.productionTip = false

// Check the user's auth status when the app starts
auth.checkAuth()

new Vue({
    router,
  render: h => h(App)
}).$mount('#app')
