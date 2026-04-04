import {setActivePinia, createPinia} from 'pinia'
import {describe, it, expect, beforeEach, vi} from 'vitest'

import {useAuthStore} from './auth'

import type {IUserSettings} from '@/modelTypes/IUserSettings'
import {PrefixMode} from '@/modules/quickAddMagic'
import {DATE_DISPLAY} from '@/constants/dateDisplay'
import {TIME_FORMAT} from '@/constants/timeFormat'
import {RELATION_KIND} from '@/types/IRelationKind'

vi.mock('vue-router', () => ({
	useRouter: () => ({
		push: vi.fn(),
	}),
}))

vi.mock('vue-i18n', () => ({
	useI18n: () => ({
		t: (key: string) => key,
	}),
	createI18n: () => ({
		global: {
			t: (key: string) => key,
			locale: {value: 'en'},
		},
	}),
}))

vi.mock('@/i18n', () => ({
	getBrowserLanguage: () => 'en',
	i18n: {
		global: {
			t: (key: string) => key,
			locale: {value: 'en'},
		},
	},
	setLanguage: vi.fn(),
}))

vi.mock('@/router', () => ({
	default: {
		push: vi.fn(),
		resolve: vi.fn(() => ({href: '/'})),
		currentRoute: {value: {query: {}}},
	},
}))

vi.mock('@/stores/config', () => ({
	useConfigStore: () => ({
		auth: {
			local: {enabled: true},
			openidConnect: {providers: [], enabled: false, redirectUrl: ''},
		},
		userDeletionEnabled: false,
	}),
}))

vi.mock('@/helpers/fetcher', () => ({
	AuthenticatedHTTPFactory: () => ({
		get: vi.fn().mockResolvedValue({data: {}}),
		post: vi.fn().mockResolvedValue({data: {}}),
		put: vi.fn().mockResolvedValue({data: {}}),
	}),
	HTTPFactory: () => ({
		get: vi.fn().mockResolvedValue({data: {}}),
		post: vi.fn().mockResolvedValue({data: {}}),
	}),
}))

vi.mock('@/helpers/auth', () => ({
	getToken: () => null,
	refreshToken: vi.fn(),
	removeToken: vi.fn(),
	saveToken: vi.fn(),
}))

vi.mock('@/message', () => ({
	success: vi.fn(),
	error: vi.fn(),
}))

vi.mock('@/helpers/redirectToProvider', () => ({
	getRedirectUrlFromCurrentFrontendPath: () => '',
	redirectToProvider: vi.fn(),
	redirectToProviderOnLogout: vi.fn(),
}))

vi.mock('@/stores/helper', () => ({
	setModuleLoading: vi.fn((_, fn) => fn()),
}))

vi.mock('@/models/user', () => {
	class UserModel {
		constructor(data: any = {}) {
			Object.assign(this, data)
		}
	}
	return {
		default: UserModel,
		getDisplayName: (user: any) => user?.name || user?.username || '',
		fetchAvatarBlobUrl: vi.fn().mockResolvedValue(''),
		invalidateAvatarCache: vi.fn(),
	}
})

function makeUserSettings(overrides: Partial<IUserSettings> = {}): IUserSettings {
	return {
		name: 'Test User',
		emailRemindersEnabled: true,
		discoverableByName: false,
		discoverableByEmail: false,
		overdueTasksRemindersEnabled: true,
		overdueTasksRemindersTime: undefined,
		defaultProjectId: undefined,
		weekStart: 0,
		timezone: 'UTC',
		language: null,
		frontendSettings: {
			playSoundWhenDone: true,
			quickAddMagicMode: PrefixMode.Default,
			colorSchema: 'auto',
			allowIconChanges: true,
			filterIdUsedOnOverview: null,
			dateDisplay: DATE_DISPLAY.RELATIVE,
			timeFormat: TIME_FORMAT.HOURS_24,
			defaultTaskRelationType: RELATION_KIND.RELATED,
			backgroundBrightness: 100,
			alwaysShowBucketTaskCount: false,
			showLastViewed: true,
			sidebarWidth: null,
			commentSortOrder: 'asc',
			desktopQuickEntryShortcut: 'CmdOrCtrl+Shift+A',
			...overrides.frontendSettings,
		},
		extraSettingsLinks: {},
		...overrides,
	} as IUserSettings
}

describe('auth store', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
	})

	describe('showLastViewed setting', () => {
		it('should default showLastViewed to true when no settings are provided', () => {
			const store = useAuthStore()
			expect(store.settings.frontendSettings.showLastViewed).toBe(true)
		})

		it('should default showLastViewed to true when API returns empty frontendSettings', () => {
			const store = useAuthStore()
			const settingsFromApi = makeUserSettings()
			delete (settingsFromApi as any).frontendSettings.showLastViewed
			store.setUserSettings(settingsFromApi)
			expect(store.settings.frontendSettings.showLastViewed).toBe(true)
		})

		it('should respect showLastViewed=false from API settings', () => {
			const store = useAuthStore()
			const settingsFromApi = makeUserSettings({
				frontendSettings: {
					showLastViewed: false,
				},
			})
			store.setUserSettings(settingsFromApi)
			expect(store.settings.frontendSettings.showLastViewed).toBe(false)
		})

		it('should respect showLastViewed=true from API settings', () => {
			const store = useAuthStore()
			const settingsFromApi = makeUserSettings({
				frontendSettings: {
					showLastViewed: true,
				},
			})
			store.setUserSettings(settingsFromApi)
			expect(store.settings.frontendSettings.showLastViewed).toBe(true)
		})

		it('should preserve other frontend settings when showLastViewed is set', () => {
			const store = useAuthStore()
			const settingsFromApi = makeUserSettings({
				frontendSettings: {
					showLastViewed: false,
					colorSchema: 'dark',
					playSoundWhenDone: false,
				},
			})
			store.setUserSettings(settingsFromApi)
			expect(store.settings.frontendSettings.showLastViewed).toBe(false)
			expect(store.settings.frontendSettings.colorSchema).toBe('dark')
			expect(store.settings.frontendSettings.playSoundWhenDone).toBe(false)
		})
	})
})
