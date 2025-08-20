import {computed, reactive, readonly, ref, toRefs} from 'vue'
import {acceptHMRUpdate, defineStore} from 'pinia'
import {parseURL} from 'ufo'

import {HTTPFactory} from '@/helpers/fetcher'
import {objectToCamelCase} from '@/helpers/case'

import type {IProvider} from '@/types/IProvider'
import type {MIGRATORS} from '@/views/migrate/migrators'
import {InvalidApiUrlProvidedError} from '@/helpers/checkAndSetApiUrl'

export interface ConfigState {
	version: string,
	frontendUrl: string,
	motd: string,
	linkSharingEnabled: boolean,
	maxFileSize: string,
	availableMigrators: Array<keyof typeof MIGRATORS>,
	taskAttachmentsEnabled: boolean,
	totpEnabled: boolean,
	enabledBackgroundProviders: Array<'unsplash' | 'upload'>,
	legal: {
		imprintUrl: string,
		privacyPolicyUrl: string,
	},
	caldavEnabled: boolean,
	userDeletionEnabled: boolean,
	taskCommentsEnabled: boolean,
	demoModeEnabled: boolean,
	auth: {
		local: {
			enabled: boolean,
			registrationEnabled: boolean,
		},
		ldap: {
			enabled: boolean,
		},
		openidConnect: {
			enabled: boolean,
			redirectUrl: string,
			providers: IProvider[],
		},
	},
	publicTeamsEnabled: boolean,
}

export const useConfigStore = defineStore('config', () => {
	const state: ConfigState = reactive({
		// These are the api defaults.
		version: '',
		frontendUrl: '',
		motd: '',
		linkSharingEnabled: true,
		maxFileSize: '20MB',
		availableMigrators: [],
		taskAttachmentsEnabled: true,
		totpEnabled: true,
		enabledBackgroundProviders: [],
		legal: {
			imprintUrl: '',
			privacyPolicyUrl: '',
		},
		caldavEnabled: false,
		userDeletionEnabled: true,
		taskCommentsEnabled: true,
		demoModeEnabled: false,
		auth: {
			local: {
				enabled: true,
				registrationEnabled: true,
			},
			ldap: {
				enabled: false,
			},
			openidConnect: {
				enabled: false,
				redirectUrl: '',
				providers: [],
			},
		},
		publicTeamsEnabled: false,
	})

	const apiUrl = ref('')

	function setApiUrl (url: string) {
		apiUrl.value = url
	}

	const migratorsEnabled = computed(() => state.availableMigrators?.length > 0)
	const apiBase = computed(() => {
		if (!apiUrl.value) return ''

		if (apiUrl.value.endsWith('/')) {
			return apiUrl.value.slice(0, -1)
		}

		return apiUrl.value
	})

	function setConfig(config: ConfigState) {
		Object.assign(state, config)
	}

	async function update(): Promise<boolean> {
		const HTTP = HTTPFactory()
		const {data: config} = await HTTP.get(`${apiBase.value}/info`, {
			headers: { 'Accept': 'application/json' },
		})

		if (typeof config.version === 'undefined') {
			throw new InvalidApiUrlProvidedError()
		}

		setConfig(objectToCamelCase(config))
		return !!config
	}

	return {
		...toRefs(state),

		apiUrl: readonly(apiUrl),
		migratorsEnabled,
		apiBase,
		setConfig,
		setApiUrl,
		update,
	}

})

// support hot reloading
if (import.meta.hot) {
	import.meta.hot.accept(acceptHMRUpdate(useConfigStore, import.meta.hot))
}
