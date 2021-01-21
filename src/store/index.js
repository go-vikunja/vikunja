import Vue from 'vue'
import Vuex from 'vuex'
import {
	CURRENT_LIST,
	ERROR_MESSAGE,
	HAS_TASKS,
	KEYBOARD_SHORTCUTS_ACTIVE,
	LOADING,
	LOADING_MODULE,
	MENU_ACTIVE,
	ONLINE,
} from './mutation-types'
import config from './modules/config'
import auth from './modules/auth'
import namespaces from './modules/namespaces'
import kanban from './modules/kanban'
import tasks from './modules/tasks'
import lists from './modules/lists'
import attachments from './modules/attachments'

import ListService from '../services/list'
import {setTitle} from '@/helpers/setTitle'

Vue.use(Vuex)

export const store = new Vuex.Store({
	modules: {
		config,
		auth,
		namespaces,
		kanban,
		tasks,
		lists,
		attachments,
	},
	state: {
		loading: false,
		loadingModule: null,
		errorMessage: '',
		online: true,
		// This is used to highlight the current list in menu for all list related views
		currentList: {id: 0},
		background: '',
		hasTasks: false,
		menuActive: true,
		keyboardShortcutsActive: false,
	},
	mutations: {
		[LOADING](state, loading) {
			state.loading = loading
		},
		[LOADING_MODULE](state, module) {
			state.loadingModule = module
		},
		[ERROR_MESSAGE](state, error) {
			state.errorMessage = error
		},
		[ONLINE](state, online) {
			state.online = online
		},
		[CURRENT_LIST](state, currentList) {

			setTitle(currentList.title)

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

			// Server updates don't return the right. Therefore the right is reset after updating the list which is
			// confusing because all the buttons will disappear in that case. To prevent this, we're keeping the right
			// when updating the list in global state.
			if(typeof state.currentList.maxRight !== 'undefined') {
				currentList.maxRight = state.currentList.maxRight
			}
			state.currentList = currentList
		},
		[HAS_TASKS](state, hasTasks) {
			state.hasTasks = hasTasks
		},
		[MENU_ACTIVE](state, menuActive) {
			state.menuActive = menuActive
		},
		toggleMenu(state) {
			state.menuActive = !state.menuActive
		},
		[KEYBOARD_SHORTCUTS_ACTIVE](state, active) {
			state.keyboardShortcutsActive = active
		},
	},
})