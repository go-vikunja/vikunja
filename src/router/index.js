import Vue from 'vue'
import Router from 'vue-router'

import HomeComponent from '@/components/Home'
// User Handling
import LoginComponent from '@/components/user/Login'
import RegisterComponent from '@/components/user/Register'
// List Handling
import ShowListComponent from '@/components/lists/ShowList'

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
    },
    {
      path: '/register',
      name: 'register',
      component: RegisterComponent
    },
    {
      path: '/lists/:id',
      name: 'showList',
      component: ShowListComponent
    }
  ]
})