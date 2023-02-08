import { readonly, ref} from 'vue'
import {defineStore, acceptHMRUpdate} from 'pinia'

import {getBlobFromBlurHash} from '@/helpers/getBlobFromBlurHash'

import ListModel from '@/models/list'
import ListService from '../services/list'
import {checkAndSetApiUrl} from '@/helpers/checkAndSetApiUrl'

import {useMenuActive} from '@/composables/useMenuActive'

import {useAuthStore} from '@/stores/auth'
import type {IList} from '@/modelTypes/IList'

export const useBaseStore = defineStore('base', () => {
	const loading = ref(false)
	const ready = ref(false)

	// This is used to highlight the current list in menu for all list related views
	const currentList = ref<IList | null>(new ListModel({
		id: 0,
		isArchived: false,
	}))
	const background = ref('')
	const blurHash = ref('')

	const hasTasks = ref(false)
	const keyboardShortcutsActive = ref(false)
	const quickActionsActive = ref(false)
	const logoVisible = ref(true)

	function setLoading(newLoading: boolean) {
		loading.value = newLoading
	}

	function setCurrentList(newCurrentList: IList | null) {
		// Server updates don't return the right. Therefore, the right is reset after updating the list which is
		// confusing because all the buttons will disappear in that case. To prevent this, we're keeping the right
		// when updating the list in global state.
		if (
			typeof currentList.value?.maxRight !== 'undefined' &&
			newCurrentList !== null &&
			(
				typeof newCurrentList.maxRight === 'undefined' ||
				newCurrentList.maxRight === null
			)
		) {
			newCurrentList.maxRight = currentList.value.maxRight
		}
		currentList.value = newCurrentList
	}

	function setHasTasks(newHasTasks: boolean) {
		hasTasks.value = newHasTasks
	}

	function setKeyboardShortcutsActive(value: boolean) {
		keyboardShortcutsActive.value = value
	}

	function setQuickActionsActive(value: boolean) {
		quickActionsActive.value = value
	}

	function setBackground(newBackground: string) {
		background.value = newBackground
	}

	function setBlurHash(newBlurHash: string) {
		blurHash.value = newBlurHash
	}

	function setLogoVisible(visible: boolean) {
		logoVisible.value = visible
	}
	
	function setReady(value: boolean) {
		ready.value = value
	}

	async function handleSetCurrentList(
		{list, forceUpdate = false}: {list: IList | null, forceUpdate?: boolean},
	) {
		if (list === null) {
			setCurrentList({})
			setBackground('')
			setBlurHash('')
			return
		}

		// The forceUpdate parameter is used only when updating a list background directly because in that case 
		// the current list stays the same, but we want to show the new background right away.
		if (list.id !== currentList.value?.id || forceUpdate) {
			if (list.backgroundInformation) {
				try {
					const blurHash = await getBlobFromBlurHash(list.backgroundBlurHash)
					if (blurHash) {
						setBlurHash(window.URL.createObjectURL(blurHash))
					}

					const listService = new ListService()
					const background = await listService.background(list)
					setBackground(background)
				} catch (e) {
					console.error('Error getting background image for list', list.id, e)
				}
			}
		}

		if (
			typeof list.backgroundInformation === 'undefined' ||
			list.backgroundInformation === null
		) {
			setBackground('')
			setBlurHash('')
		}

		setCurrentList(list)
	}

	const authStore = useAuthStore()
	async function loadApp() {
		await checkAndSetApiUrl(window.API_URL)
		await authStore.checkAuth()
		ready.value = true
	}

	return {
		loading: readonly(loading),
		ready: readonly(ready),
		currentList: readonly(currentList),
		background: readonly(background),
		blurHash: readonly(blurHash),
		hasTasks: readonly(hasTasks),
		keyboardShortcutsActive: readonly(keyboardShortcutsActive),
		quickActionsActive: readonly(quickActionsActive),
		logoVisible: readonly(logoVisible),

		setLoading,
		setReady,
		setCurrentList,
		setHasTasks,
		setKeyboardShortcutsActive,
		setQuickActionsActive,
		setBackground,
		setBlurHash,
		setLogoVisible,

		handleSetCurrentList,
		loadApp,

		...useMenuActive(),
	}
})

// support hot reloading
if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useBaseStore, import.meta.hot))
}