import Vue from 'vue'
import Router from 'vue-router'

import HomeComponent from '@/components/Home'
import LoginComponent from '@/components/Login'

Vue.use(Router)

export default new Router({
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeComponent
    },
    {
      path: '/login',
      name: 'login',
      component: LoginComponent
    }
  ]
})