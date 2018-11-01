import Vue from 'vue'
import Router from 'vue-router'

import HomeComponent from '@/components/Home'
// User Handling
import LoginComponent from '@/components/user/Login'
import RegisterComponent from '@/components/user/Register'
import PasswordResetComponent from '@/components/user/PasswordReset'
import GetPasswordResetComponent from '@/components/user/RequestPasswordReset'
// List Handling
import ShowListComponent from '@/components/lists/ShowList'
import NewListComponent from '@/components/lists/NewList'
import EditListComponent from '@/components/lists/EditList'
// Namespace Handling
import NewNamespaceComponent from '@/components/namespaces/NewNamespace'
import EditNamespaceComponent from '@/components/namespaces/EditNamespace'
// Team Handling
import ListTeamsComponent from '@/components/teams/ListTeams'
import EditTeamComponent from '@/components/teams/EditTeam'
import NewTeamComponent from '@/components/teams/NewTeam'

Vue.use(Router)

export default new Router({
  mode:'history',
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
      path: '/get-password-reset',
      name: 'getPasswordReset',
      component: GetPasswordResetComponent
    },
    {
      path: '/password-reset',
      name: 'passwordReset',
      component: PasswordResetComponent
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
    },
    {
      path: '/lists/:id/edit',
      name: 'editList',
      component: EditListComponent
    },
    {
      path: '/namespaces/:id/list',
      name: 'newList',
      component: NewListComponent
    },
    {
      path: '/namespaces/new',
      name: 'newNamespace',
      component: NewNamespaceComponent
    },
    {
      path: '/namespaces/:id/edit',
      name: 'editNamespace',
      component: EditNamespaceComponent
    },
    {
      path: '/teams',
      name: 'listTeams',
      component: ListTeamsComponent
    },
    {
      path: '/teams/new',
      name: 'newTeam',
      component: NewTeamComponent
    },
    {
      path: '/teams/:id/edit',
      name: 'editTeam',
      component: EditTeamComponent
    },
  ]
})