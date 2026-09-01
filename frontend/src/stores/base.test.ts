import {createPinia, setActivePinia} from 'pinia'
import {beforeEach, describe, expect, it, vi} from 'vitest'

import {AUTH_TYPES} from '@/modelTypes/IUser'
import type {ProjectResponse} from '@/client/queries/projects'

const auth = vi.hoisted(() => ({
	token: null as string | null,
	post: vi.fn(),
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
	useWebSocket: () => ({disconnect: vi.fn()}),
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

vi.mock('@/client/queries/projectBackgrounds', () => ({
	projectBackgroundQuery: vi.fn(),
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

describe('base store identity reset', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		auth.token = null
		auth.post.mockReset()
		window.URL.revokeObjectURL = vi.fn()
	})

	it('clears the background, blur hash and current project when the auth identity changes', async () => {
		const authStore = useAuthStore()
		const baseStore = useBaseStore()
		await baseStore.appReady

		authStore.setUser({id: 1, type: AUTH_TYPES.USER} as never, false)

		baseStore.setCurrentProject(project(42))
		baseStore.setBackground('blob:old-background')
		baseStore.setBlurHash('blob:old-blur')

		expect(baseStore.currentProjectId).toBe(42)

		authStore.setUser({id: 2, type: AUTH_TYPES.USER} as never, false)

		expect(baseStore.background).toBe('')
		expect(baseStore.blurHash).toBe('')
		expect(baseStore.currentProjectId).toBe(0)
		expect(window.URL.revokeObjectURL).toHaveBeenCalledWith('blob:old-background')
		expect(window.URL.revokeObjectURL).toHaveBeenCalledWith('blob:old-blur')
	})

	it('does not reset state when the same user is set again', async () => {
		const authStore = useAuthStore()
		const baseStore = useBaseStore()
		await baseStore.appReady

		authStore.setUser({id: 1, type: AUTH_TYPES.USER} as never, false)
		baseStore.setCurrentProject(project(42))
		baseStore.setBackground('blob:old-background')

		authStore.setUser({id: 1, type: AUTH_TYPES.USER} as never, false)

		expect(baseStore.currentProjectId).toBe(42)
		expect(baseStore.background).toBe('blob:old-background')
	})
})
