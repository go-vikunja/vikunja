import Vue from 'vue'
import Vuex from 'vuex'

Vue.use(Vuex)

import {CURRENT_LIST, ERROR_MESSAGE, HAS_TASKS, IS_FULLPAGE, LOADING, ONLINE} from './mutation-types'
import config from './modules/config'
import auth from './modules/auth'
import namespaces from './modules/namespaces'
import kanban from './modules/kanban'
import tasks from './modules/tasks'
import lists from './modules/lists'
import ListService from '../services/list'

export const store = new Vuex.Store({
	modules: {
		config,
		auth,
		namespaces,
		kanban,
		tasks,
		lists,
	},
	state: {
		loading: false,
		errorMessage: '',
		online: true,
		isFullpage: false,
		// This is used to highlight the current list in menu for all list related views
		currentList: {id: 0},
		background: '',
		hasTasks: false,
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
		},
		[CURRENT_LIST](state, currentList) {
			// Not sure if this is the right way to do it but hey, it works
			if (
				// List changed
				currentList.id !== state.currentList.id ||
				// The current list got a new background and didn't have one previously
				(
					currentList.backgroundInformation &&
					!state.currentList.backgroundInformation
				) ||
				// The current list got a new background and had one previously
				(
					currentList.backgroundInformation &&
					currentList.backgroundInformation.unsplashId &&
					state.currentList &&
					state.currentList.backgroundInformation &&
					state.currentList.backgroundInformation.unsplashId &&
					currentList.backgroundInformation.unsplashId !== state.currentList.backgroundInformation.unsplashId
				) ||
				// The new list has a background which is not an unsplash one and did not have one previously
				(
					currentList.backgroundInformation &&
					currentList.backgroundInformation.type &&
					state.currentList &&
					state.currentList.backgroundInformation &&
					state.currentList.backgroundInformation.type
				)
			) {
				if (currentList.backgroundInformation) {
					const listService = new ListService()
					listService.background(currentList)
						.then(b => {
							state.background = b
						})
						.catch(e => {
							console.error('Error getting background image for list', currentList.id, e)
						})
				} else {
					state.background = null
				}
			}

			state.currentList = currentList
		},
		[HAS_TASKS](state, hasTasks) {
			state.hasTasks = hasTasks
		}
	},
})