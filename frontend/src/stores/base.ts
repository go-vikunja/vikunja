import {ref, computed, readonly} from 'vue'
import {useI18n} from 'vue-i18n'
import {defineStore, acceptHMRUpdate} from 'pinia'

import {getBlobFromBlurHash} from '@/helpers/getBlobFromBlurHash'

import ProjectModel from '@/models/project'
import ProjectService from '@/services/project'
import {checkAndSetApiUrl, ERROR_NO_API_URL, InvalidApiUrlProvidedError, NoApiUrlProvidedError} from '@/helpers/checkAndSetApiUrl'

import {useMenuActive} from '@/composables/useMenuActive'

import {useAuthStore} from '@/stores/auth'
import type {IProject} from '@/modelTypes/IProject'
import type {Permission} from '@/constants/permissions'
import type {IProjectView} from '@/modelTypes/IProjectView'

export const useBaseStore = defineStore('base', () => {
	const authStore = useAuthStore()
	
	const {t} = useI18n()

	const ready = ref(false)
	const error = ref('')
	const loading = computed(() => !ready.value && error.value === '')

	// This is used to highlight the current project in menu for all project related views
	const currentProject = ref<IProject | null>(new ProjectModel({
		id: 0,
		isArchived: false,
	}))
	const currentProjectViewId = ref<IProjectView['id'] | undefined>(undefined)
	const background = ref('')
	const blurHash = ref('')

	const hasTasks = ref(false)
	const keyboardShortcutsActive = ref(false)
	const quickActionsActive = ref(false)
	const logoVisible = ref(true)
	const updateAvailable = ref(false)

	function setCurrentProject(newCurrentProject: IProject | null, currentViewId?: IProjectView['id']) {
		// Server updates don't return the permission. Therefore, the permission is reset after updating the project which is
		// confusing because all the buttons will disappear in that case. To prevent this, we're keeping the permission
		// when updating the project in global state.
		let maxPermission: Permission | null = newCurrentProject?.maxPermission || null
		if (
			typeof currentProject.value?.maxPermission !== 'undefined' &&
			newCurrentProject !== null &&
			(
				typeof newCurrentProject.maxPermission === 'undefined' ||
				newCurrentProject.maxPermission === null
			)
		) {
			maxPermission = currentProject.value.maxPermission
		}
		if (newCurrentProject === null) {
			currentProject.value = null
			return
		}
		currentProject.value = {
			...newCurrentProject,
			maxPermission,
		}
		setCurrentProjectViewId(currentViewId)
	}
	
	function setCurrentProjectViewId(viewId?: IProjectView['id']) {
		currentProjectViewId.value = viewId
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
	
	function setUpdateAvailable(value: boolean) {
		updateAvailable.value = value
	}

	async function handleSetCurrentProject(
		{project, forceUpdate = false, currentProjectViewId = undefined}: {project: IProject | null, forceUpdate?: boolean, currentProjectViewId?: IProjectView['id']},
	) {
		if (project === null || typeof project === 'undefined') {
			setCurrentProject({})
			setBackground('')
			setBlurHash('')
			return
		}

		// The forceUpdate parameter is used only when updating a project background directly because in that case 
		// the current project stays the same, but we want to show the new background right away.
		if (project.id !== currentProject.value?.id || forceUpdate) {
			if (project.backgroundInformation) {
				try {
					const blurHash = await getBlobFromBlurHash(project.backgroundBlurHash)
					if (blurHash) {
						setBlurHash(window.URL.createObjectURL(blurHash))
					}

					const projectService = new ProjectService()
					const background = await projectService.background(project)
					setBackground(background)
				} catch (e) {
					console.error('Error getting background image for project', project.id, e)
				}
			}
		}

		if (
			typeof project.backgroundInformation === 'undefined' ||
			project.backgroundInformation === null
		) {
			setBackground('')
			setBlurHash('')
		}

		setCurrentProject(project, currentProjectViewId)
	}

	async function loadApp() {
		try {
			await checkAndSetApiUrl(window.API_URL)
			await authStore.checkAuth()
			ready.value = true
		} catch (e: unknown) {
			if (e instanceof NoApiUrlProvidedError) {
				error.value = ERROR_NO_API_URL
				return
			}
			if (e instanceof InvalidApiUrlProvidedError) {
				error.value = t('apiConfig.error')
				return
			}
			error.value = String(e.message)
		}
	}

	loadApp()

	return {
		error: readonly(error),
		loading: readonly(loading),
		ready: readonly(ready),
		loadApp,

		currentProject: readonly(currentProject),
		currentProjectViewId: readonly(currentProjectViewId),
		background: readonly(background),
		blurHash: readonly(blurHash),
		hasTasks: readonly(hasTasks),
		keyboardShortcutsActive: readonly(keyboardShortcutsActive),
		quickActionsActive: readonly(quickActionsActive),
		logoVisible: readonly(logoVisible),
		updateAvailable: readonly(updateAvailable),

		setCurrentProject,
		setCurrentProjectViewId,
		setHasTasks,
		setKeyboardShortcutsActive,
		setQuickActionsActive,
		setBackground,
		setBlurHash,
		setLogoVisible,
		setUpdateAvailable,

		handleSetCurrentProject,

		...useMenuActive(),
	}
})

// support hot reloading
if (import.meta.hot) {
	import.meta.hot.accept(acceptHMRUpdate(useBaseStore, import.meta.hot))
}
