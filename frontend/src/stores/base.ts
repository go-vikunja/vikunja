import {computed, readonly, ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'
import {defineStore, acceptHMRUpdate} from 'pinia'
import {useQuery} from '@tanstack/vue-query'

import {checkAndSetApiUrl, ERROR_NO_API_URL, InvalidApiUrlProvidedError, NoApiUrlProvidedError} from '@/helpers/checkAndSetApiUrl'
import {isDesktopApp} from '@/helpers/desktopAuth'

import {useMenuActive} from '@/composables/useMenuActive'

import {useAuthStore} from '@/stores/auth'
import router from '@/router'
import {queryClient} from '@/client/queryClient'
import {projectQuery, type ProjectResponse} from '@/client/queries/projects'
import {useProjectBackground} from '@/composables/useProjectBackground'

export const useBaseStore = defineStore('base', () => {
	const authStore = useAuthStore()

	const {t} = useI18n()

	const ready = ref(false)
	const error = ref('')
	const loading = computed(() => !ready.value && error.value === '')

	// This is used to highlight the current project in menu for all project related views
	const currentProjectId = ref(0)
	const currentProjectViewId = ref<number | undefined>(undefined)

	// Pseudo projects (favorites, saved filters) have no detail endpoint and no background.
	const currentProjectQuery = useQuery(computed(() => ({
		...projectQuery(currentProjectId.value),
		enabled: currentProjectId.value > 0,
	})), queryClient)
	const currentProject = computed(() => currentProjectQuery.data.value ?? null)
	const {background: backgroundUrl, blurHashUrl} = useProjectBackground(currentProject)
	const background = computed(() => backgroundUrl.value ?? '')
	const blurHash = computed(() => blurHashUrl.value ?? '')

	const hasTasks = ref(false)
	const keyboardShortcutsActive = ref(false)
	const quickActionsActive = ref(false)
	const logoVisible = ref(true)
	const updateAvailable = ref(false)

	function setCurrentProject(project: Pick<ProjectResponse, 'id'> | null, viewId?: number) {
		currentProjectId.value = project?.id ?? 0
		setCurrentProjectViewId(viewId)
	}

	// A task opened over its kanban board must keep that board's view id.
	function setCurrentProjectIfNotSet(project: Pick<ProjectResponse, 'id'>) {
		if (currentProjectId.value !== project.id) {
			setCurrentProject(project)
		}
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

	// SPA logout keeps this store alive, so reset on identity change.
	watch(() => authStore.identityKey, () => {
		setCurrentProject(null)
		setHasTasks(false)
	}, {flush: 'sync'})

	function setLogoVisible(visible: boolean) {
		logoVisible.value = visible
	}
	
	function setUpdateAvailable(value: boolean) {
		updateAvailable.value = value
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
		setCurrentProjectIfNotSet,
		setCurrentProjectViewId,
		setHasTasks,
		setKeyboardShortcutsActive,
		setQuickActionsActive,
		setLogoVisible,
		setUpdateAvailable,

		...useMenuActive(),
	}
})

// support hot reloading
if (import.meta.hot) {
	import.meta.hot.accept(acceptHMRUpdate(useBaseStore, import.meta.hot))
}
