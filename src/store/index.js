import Vue from 'vue'
import Vuex from 'vuex'
Vue.use(Vuex)

import config from './modules/config'
import auth from './modules/auth'
import namespaces from './modules/namespaces'
import {ERROR_MESSAGE, IS_FULLPAGE, LOADING, ONLINE} from './mutation-types'

export const store = new Vuex.Store({
	modules: {
		config,
		auth,
		namespaces,
	},
	state: {
		loading: false,
		errorMessage: '',
		online: true,
		isFullpage: false,
	},
	mutations: {
		[LOADING](state, loading) {
			state.loading = loading
		},
		[ERROR_MESSAGE](state, error) {
			state.errorMessage = error
		},
		[ONLINE](state, online) {
			state.online = online
		},
		[IS_FULLPAGE](state, fullpage) {
			state.isFullpage = fullpage
		}
	},
})