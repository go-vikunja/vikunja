import {createStore} from 'vuex'
import {getBlobFromBlurHash} from '../helpers/getBlobFromBlurHash'
import {
	BACKGROUND,
	BLUR_HASH,
	CURRENT_LIST,
	HAS_TASKS,
	KEYBOARD_SHORTCUTS_ACTIVE,
	LOADING,
	LOADING_MODULE,
	MENU_ACTIVE,
	QUICK_ACTIONS_ACTIVE,
} from './mutation-types'
import config from './modules/config'
import auth from './modules/auth'
import namespaces from './modules/namespaces'
import kanban from './modules/kanban'
import tasks from './modules/tasks'
import lists from './modules/lists'
import attachments from './modules/attachments'
import labels from './modules/labels'

import ListModel from '@/models/list'

import ListService from '../services/list'
import {checkAndSetApiUrl} from '@/helpers/checkAndSetApiUrl'

export const store = createStore({
	strict: import.meta.env.DEV,
	modules: {
		config,
		auth,
		namespaces,
		kanban,
		tasks,
		lists,
		attachments,
		labels,
	},
	state: {
		loading: false,
		loadingModule: null,
		// This is used to highlight the current list in menu for all list related views
		currentList: new ListModel({
			id: 0,
			isArchived: false,
		}),
		background: '',
		blurHash: '',
		hasTasks: false,
		menuActive: true,
		keyboardShortcutsActive: false,
		quickActionsActive: false,
	},
	mutations: {
		[LOADING](state, loading) {
			state.loading = loading
		},
		[LOADING_MODULE](state, module) {
			state.loadingModule = module
		},
		[CURRENT_LIST](state, currentList) {
			// Server updates don't return the right. Therefore the right is reset after updating the list which is
			// confusing because all the buttons will disappear in that case. To prevent this, we're keeping the right
			// when updating the list in global state.
			if (typeof state.currentList.maxRight !== 'undefined' && (typeof currentList.maxRight === 'undefined' || currentList.maxRight === null)) {
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
		[QUICK_ACTIONS_ACTIVE](state, active) {
			state.quickActionsActive = active
		},
		[BACKGROUND](state, background) {
			state.background = background
		},
		[BLUR_HASH](state, blurHash) {
			state.blurHash = blurHash
		},
	},
	actions: {
		async [CURRENT_LIST]({state, commit}, {list, forceUpdate = false}) {

			if (list === null) {
				commit(CURRENT_LIST, {})
				commit(BACKGROUND, null)
				commit(BLUR_HASH, null)
				return
			}

			// The forceUpdate parameter is used only when updating a list background directly because in that case 
			// the current list stays the same, but we want to show the new background right away.
			if (list.id !== state.currentList.id || forceUpdate) {
				if (list.backgroundInformation) {
					try {
						const blurHash = await getBlobFromBlurHash(list.backgroundBlurHash)
						if (blurHash) {
							commit(BLUR_HASH, window.URL.createObjectURL(blurHash))
						}

						const listService = new ListService()
						const background = await listService.background(list)
						commit(BACKGROUND, background)
					} catch (e) {
						console.error('Error getting background image for list', list.id, e)
					}
				}
			}

			if (typeof list.backgroundInformation === 'undefined' || list.backgroundInformation === null) {
				commit(BACKGROUND, null)
				commit(BLUR_HASH, null)
			}

			commit(CURRENT_LIST, list)
		},
		async loadApp({dispatch}) {
			await checkAndSetApiUrl(window.API_URL)
			await dispatch('auth/checkAuth')
		},
	},
})
