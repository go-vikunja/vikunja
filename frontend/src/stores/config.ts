import {computed, reactive, toRefs} from 'vue'
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

	const migratorsEnabled = computed(() => state.availableMigrators?.length > 0)
	const apiBase = computed(() => {
		const {host, protocol, href} = parseURL(window.API_URL)

		const cleanHref = href ? (href.endsWith('/') 
			? href.slice(0, -1) 
			: href) : ''
		return `${protocol}//${host}${cleanHref ? `/${cleanHref}` : ''}`
	})

	function setConfig(config: ConfigState) {
		Object.assign(state, config)
	}

	async function update(): Promise<boolean> {
		const HTTP = HTTPFactory()
		const {data: config} = await HTTP.get('info')

		if (typeof config.version === 'undefined') {
			throw new InvalidApiUrlProvidedError()
		}

		setConfig(objectToCamelCase(config))
		return !!config
	}

	return {
		...toRefs(state),

		migratorsEnabled,
		apiBase,
		setConfig,
		update,
	}

})

// support hot reloading
if (import.meta.hot) {
	import.meta.hot.accept(acceptHMRUpdate(useConfigStore, import.meta.hot))
}
