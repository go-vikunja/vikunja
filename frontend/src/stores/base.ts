import {computed, onScopeDispose, readonly, ref} from 'vue'
import {useI18n} from 'vue-i18n'
import {defineStore, acceptHMRUpdate} from 'pinia'

import {getBlobFromBlurHash} from '@/helpers/getBlobFromBlurHash'

import {checkAndSetApiUrl, ERROR_NO_API_URL, InvalidApiUrlProvidedError, NoApiUrlProvidedError} from '@/helpers/checkAndSetApiUrl'
import {isDesktopApp} from '@/helpers/desktopAuth'

import {useMenuActive} from '@/composables/useMenuActive'

import {useAuthStore} from '@/stores/auth'
import router from '@/router'
import type {ProjectResponse} from '@/client/queries/projects'
import ProjectService from '@/services/project'
import type {IProject} from '@/modelTypes/IProject'

type CurrentProjectInput = ProjectResponse | IProject

function getBackgroundInformation(project: CurrentProjectInput) {
	return 'background_information' in project
		? project.background_information
		: (project as IProject).backgroundInformation
}

function getBackgroundBlurHash(project: CurrentProjectInput) {
	return 'background_blur_hash' in project
		? project.background_blur_hash ?? ''
		: (project as IProject).backgroundBlurHash
}

export const useBaseStore = defineStore('base', () => {
	const authStore = useAuthStore()

	const {t} = useI18n()

	const ready = ref(false)
	const error = ref('')
	const loading = computed(() => !ready.value && error.value === '')

	// This is used to highlight the current project in menu for all project related views
	const currentProjectId = ref(0)
	const currentProjectViewId = ref<number | undefined>(undefined)
	const background = ref('')
	const blurHash = ref('')

	const hasTasks = ref(false)
	const keyboardShortcutsActive = ref(false)
	const quickActionsActive = ref(false)
	const logoVisible = ref(true)
	const updateAvailable = ref(false)

	function setCurrentProject(newCurrentProject: CurrentProjectInput | null, currentViewId?: number) {
		if (currentProjectId.value !== (newCurrentProject?.id ?? 0)) {
			setBackground('')
			setBlurHash('')
		}
		currentProjectId.value = newCurrentProject?.id ?? 0
		setCurrentProjectViewId(currentViewId)
	}

	function setCurrentProjectViewId(viewId?: number) {
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
		if (background.value && background.value !== newBackground) {
			window.URL.revokeObjectURL(background.value)
		}
		background.value = newBackground
	}

	function setBlurHash(newBlurHash: string) {
		if (blurHash.value && blurHash.value !== newBlurHash) {
			window.URL.revokeObjectURL(blurHash.value)
		}
		blurHash.value = newBlurHash
	}

	function setLogoVisible(visible: boolean) {
		logoVisible.value = visible
	}
	
	function setUpdateAvailable(value: boolean) {
		updateAvailable.value = value
	}

	async function handleSetCurrentProject(
		{project, forceUpdate = false, currentProjectViewId: viewId = undefined}: {
			project: CurrentProjectInput | null
			forceUpdate?: boolean
			currentProjectViewId?: number
		},
	) {
		if (!project) {
			setCurrentProject(null)
			return
		}
		const previousProjectId = currentProjectId.value
		setCurrentProject(project, viewId)

		if (project.id !== previousProjectId || forceUpdate) {
			const backgroundInformation = getBackgroundInformation(project)
			if (backgroundInformation) {
				try {
					const preview = await getBlobFromBlurHash(getBackgroundBlurHash(project))
					if (currentProjectId.value !== project.id) {
						return
					}
					setBlurHash(preview ? window.URL.createObjectURL(preview) : '')
					const projectService = new ProjectService()
					const image = await projectService.background({
						id: project.id,
						backgroundInformation,
					})
					if (currentProjectId.value === project.id) {
						setBackground(image)
					}
				} catch (e) {
					console.error('Error getting background image for project', project.id, e)
				}
			}
		}

		if (
			typeof getBackgroundInformation(project) === 'undefined' ||
			getBackgroundInformation(project) === null
		) {
			setBackground('')
			setBlurHash('')
		}
	}

	async function handleSetCurrentProjectIfNotSet(project: CurrentProjectInput) {
		if (currentProjectId.value !== project.id) {
			await handleSetCurrentProject({project})
		}
	}

	async function hydrateConfig() {
		try {
			if (isDesktopApp()) {
				// On desktop, ignore the default window.API_URL (set by index.html)
				// and only use a previously stored API URL from localStorage.
				const storedApiUrl = localStorage.getItem('API_URL')
				if (storedApiUrl) {
					// Hydrate /info before marking ready; otherwise pro-feature gates stay off for returning desktop users.
					await checkAndSetApiUrl(storedApiUrl)
					await authStore.checkAuth()
				}
				return
			}

			await checkAndSetApiUrl(window.API_URL)
			await authStore.checkAuth()
		} catch (e: unknown) {
			if (e instanceof NoApiUrlProvidedError) {
				error.value = ERROR_NO_API_URL
				return
			}
			if (e instanceof InvalidApiUrlProvidedError) {
				error.value = t('apiConfig.error')
				return
			}
			error.value = String(e instanceof Error ? e.message : e)
		}
	}

	// Exposed so router guards can await config/auth hydration on direct
	// navigation without deadlocking on router.isReady().
	const appReady = hydrateConfig()

	async function loadApp() {
		// Re-hydrates (used when the user selects a new API URL from Ready.vue).
		await hydrateConfig()
		await router.isReady()
		ready.value = true
	}

	// Initial load: wait on the in-flight hydration, then mark ready once
	// the router has settled.
	appReady.then(async () => {
		await router.isReady()
		ready.value = true
	})

	onScopeDispose(() => {
		setBackground('')
		setBlurHash('')
	})

	return {
		error: readonly(error),
		loading: readonly(loading),
		ready: readonly(ready),
		loadApp,
		appReady,

		currentProjectId: readonly(currentProjectId),
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
		handleSetCurrentProjectIfNotSet,

		...useMenuActive(),
	}
})

// support hot reloading
if (import.meta.hot) {
	import.meta.hot.accept(acceptHMRUpdate(useBaseStore, import.meta.hot))
}
