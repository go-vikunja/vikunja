import {defineStore, acceptHMRUpdate} from 'pinia'

import {getBlobFromBlurHash} from '@/helpers/getBlobFromBlurHash'

import ListModel from '@/models/list'
import ListService from '../services/list'
import {checkAndSetApiUrl} from '@/helpers/checkAndSetApiUrl'

import {useAuthStore} from '@/stores/auth'
import type {IList} from '@/modelTypes/IList'

export interface RootStoreState {
	loading: boolean,

	currentList: IList,
	background: string,
	blurHash: string,

	hasTasks: boolean,
	menuActive: boolean,
	keyboardShortcutsActive: boolean,
	quickActionsActive: boolean,
	logoVisible: boolean,
}

export const useBaseStore = defineStore('base', {
	state: () : RootStoreState => ({
		loading: false,

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
		logoVisible: true,
	}),

	actions: {
		setLoading(loading: boolean) {
			this.loading = loading
		},

		setCurrentList(currentList: IList) {
			// Server updates don't return the right. Therefore, the right is reset after updating the list which is
			// confusing because all the buttons will disappear in that case. To prevent this, we're keeping the right
			// when updating the list in global state.
			if (
				typeof this.currentList.maxRight !== 'undefined' &&
				(
					typeof currentList.maxRight === 'undefined' ||
					currentList.maxRight === null
				)
			) {
				currentList.maxRight = this.currentList.maxRight
			}
			this.currentList = currentList
		},

		setHasTasks(hasTasks: boolean) {
			this.hasTasks = hasTasks
		},

		setMenuActive(menuActive: boolean) {
			this.menuActive = menuActive
		},

		toggleMenu() {
			this.menuActive = !this.menuActive
		},

		setKeyboardShortcutsActive(active: boolean) {
			this.keyboardShortcutsActive = active
		},

		setQuickActionsActive(active: boolean) {
			this.quickActionsActive = active
		},

		setBackground(background: string) {
			this.background = background
		},

		setBlurHash(blurHash: string) {
			this.blurHash = blurHash
		},

		setLogoVisible(visible: boolean) {
			this.logoVisible = visible
		},

		async handleSetCurrentList({list, forceUpdate = false} : {list: IList, forceUpdate: boolean}) {
			if (list === null) {
				this.setCurrentList({})
				this.setBackground('')
				this.setBlurHash('')
				return
			}

			// The forceUpdate parameter is used only when updating a list background directly because in that case 
			// the current list stays the same, but we want to show the new background right away.
			if (list.id !== this.currentList.id || forceUpdate) {
				if (list.backgroundInformation) {
					try {
						const blurHash = await getBlobFromBlurHash(list.backgroundBlurHash)
						if (blurHash) {
							this.setBlurHash(window.URL.createObjectURL(blurHash))
						}

						const listService = new ListService()
						const background = await listService.background(list)
						this.setBackground(background)
					} catch (e) {
						console.error('Error getting background image for list', list.id, e)
					}
				}
			}

			if (typeof list.backgroundInformation === 'undefined' || list.backgroundInformation === null) {
				this.setBackground('')
				this.setBlurHash('')
			}

			this.setCurrentList(list)
		},

		async loadApp() {
			await checkAndSetApiUrl(window.API_URL)
			useAuthStore().checkAuth()
		},
	},
})

// support hot reloading
if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useBaseStore, import.meta.hot))
}