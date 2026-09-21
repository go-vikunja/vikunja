import {createPinia, setActivePinia} from 'pinia'
import {beforeEach, describe, expect, it, vi} from 'vitest'

import {AUTH_TYPES} from '@/constants/auth'
import {queryClient} from '@/client/queryClient'
import {projectKeys, type ProjectResponse} from '@/client/queries/projects'

const auth = vi.hoisted(() => ({
	token: null as string | null,
	post: vi.fn(),
}))

const sdk = vi.hoisted(() => ({
	projectsBackgroundGet: vi.fn(),
	projectsRead: vi.fn(),
}))

vi.mock('@/client/generated', async (importOriginal) => ({
	...await importOriginal<typeof import('@/client/generated')>(),
	...sdk,
}))

vi.mock('@/helpers/auth', () => ({
	getToken: () => auth.token,
	refreshToken: vi.fn(),
	removeToken: vi.fn(() => {
		auth.token = null
	}),
	saveToken: vi.fn((token: string) => {
		auth.token = token
	}),
}))

vi.mock('@/helpers/fetcher', () => ({
	AuthenticatedHTTPFactory: () => fakeHttp(),
	HTTPFactory: () => fakeHttp(),
}))

function fakeHttp() {
	return {
		post: auth.post,
		get: vi.fn(),
		request: vi.fn(),
		interceptors: {
			request: {use: vi.fn()},
			response: {use: vi.fn()},
		},
	}
}

vi.mock('@/router', () => ({
	default: {push: vi.fn(), isReady: vi.fn().mockResolvedValue(undefined)},
}))

vi.mock('@/helpers/redirectToProvider', () => ({
	getRedirectUrlFromCurrentFrontendPath: vi.fn(),
	redirectToProvider: vi.fn(),
	redirectToProviderOnLogout: vi.fn(),
}))

vi.mock('@/composables/useWebSocket', () => ({
	useWebSocket: () => ({
		disconnect: vi.fn(),
		closeStaleConnection: vi.fn(),
	}),
}))

vi.mock('vue-i18n', async (importOriginal) => ({
	...await importOriginal<typeof import('vue-i18n')>(),
	useI18n: () => ({t: (key: string) => key}),
}))

vi.mock('@/helpers/checkAndSetApiUrl', async (importOriginal) => ({
	...await importOriginal<typeof import('@/helpers/checkAndSetApiUrl')>(),
	checkAndSetApiUrl: vi.fn().mockResolvedValue('http://localhost/api/v1'),
}))

vi.mock('@/helpers/desktopAuth', () => ({
	isDesktopApp: () => false,
}))

vi.mock('@/helpers/getBlobFromBlurHash', () => ({
	getBlobFromBlurHash: vi.fn(),
}))

vi.mock('@/composables/useMenuActive', async () => {
	const {ref} = await import('vue')
	return {
		useMenuActive: () => ({
			menuActive: ref(false),
			setMenuActive: vi.fn(),
		}),
	}
})

import {useAuthStore} from './auth'
import {useBaseStore} from './base'

function project(id: number): ProjectResponse {
	return {
		id,
		title: 'Project',
		description: '',
		hex_color: '',
		identifier: '',
		is_archived: false,
		is_favorite: false,
		parent_project_id: 0,
		position: 0,
		views: [],
	} as unknown as ProjectResponse
}

function seedProjectWithBackground(id: number) {
	queryClient.setQueryData(projectKeys.detail(id), {
		...project(id),
		background_information: {id: 8},
		background_blur_hash: '',
	})
	sdk.projectsBackgroundGet.mockResolvedValue({data: new Blob(['image'])})
	window.URL.createObjectURL = vi.fn().mockReturnValue('blob:new-background')
}

describe('base store identity reset', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		auth.token = null
		auth.post.mockReset()
		queryClient.clear()
		Object.values(sdk).forEach(mock => mock.mockReset())
		window.URL.revokeObjectURL = vi.fn()
	})

	it('displays the cached background of the current project', async () => {
		const store = useBaseStore()
		await store.appReady
		seedProjectWithBackground(42)

		store.setCurrentProject(project(42))

		await vi.waitFor(() => expect(store.background).toBe('blob:new-background'))
		expect(sdk.projectsBackgroundGet).toHaveBeenCalledWith({path: {project: 42}})
		expect(sdk.projectsRead).not.toHaveBeenCalled()

		store.setCurrentProject(null)

		await vi.waitFor(() => expect(store.background).toBe(''))
	})

	it('keeps the current view id when the project is already current', async () => {
		const store = useBaseStore()
		await store.appReady

		store.setCurrentProject(project(42), 5)
		store.setCurrentProjectIfNotSet(project(42))

		expect(store.currentProjectViewId).toBe(5)

		store.setCurrentProjectIfNotSet(project(43))

		expect(store.currentProjectId).toBe(43)
		expect(store.currentProjectViewId).toBeUndefined()
	})

	it.each([
		{label: 'user switch', next: {id: 2, type: AUTH_TYPES.USER}, resets: true},
		{label: 'link share with the same numeric id', next: {id: 1, type: AUTH_TYPES.LINK_SHARE}, resets: true},
		{label: 'same identity', next: {id: 1, type: AUTH_TYPES.USER}, resets: false},
	])('$label resets the background, current project and tasks flag: $resets', async ({next, resets}) => {
		const authStore = useAuthStore()
		const baseStore = useBaseStore()
		await baseStore.appReady

		authStore.setUser({id: 1, type: AUTH_TYPES.USER} as never, false)
		seedProjectWithBackground(42)

		baseStore.setCurrentProject(project(42))
		baseStore.setHasTasks(true)

		expect(baseStore.currentProjectId).toBe(42)
		await vi.waitFor(() => expect(baseStore.background).toBe('blob:new-background'))

		authStore.setUser(next as never, false)

		if (!resets) {
			expect(baseStore.background).toBe('blob:new-background')
			expect(baseStore.currentProjectId).toBe(42)
			expect(baseStore.hasTasks).toBe(true)
			return
		}

		expect(baseStore.currentProjectId).toBe(0)
		expect(baseStore.hasTasks).toBe(false)
		await vi.waitFor(() => expect(baseStore.background).toBe(''))
	})
})
