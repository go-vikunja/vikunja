import {computed, reactive, toRefs} from 'vue'
import {defineStore, acceptHMRUpdate} from 'pinia'
import {parseURL} from 'ufo'

import {HTTPFactory} from '@/helpers/fetcher'
import {objectToCamelCase} from '@/helpers/case'

import type {IProvider} from '@/types/IProvider'
import type {MIGRATORS} from '@/views/migrate/migrators'

export interface ConfigState {
	version: string,
	frontendUrl: string,
	motd: string,
	linkSharingEnabled: boolean,
	maxFileSize: string,
	registrationEnabled: boolean,
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
		},
		openidConnect: {
			enabled: boolean,
			redirectUrl: string,
			providers: IProvider[],
		},
	},
}

export const useConfigStore = defineStore('config', () => {
	const state = reactive({
		// These are the api defaults.
		version: '',
		frontendUrl: '',
		motd: '',
		linkSharingEnabled: true,
		maxFileSize: '20MB',
		registrationEnabled: true,
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
			},
			openidConnect: {
				enabled: false,
				redirectUrl: '',
				providers: [],
			},
		},
	})

	const migratorsEnabled = computed(() => state.availableMigrators?.length > 0)
	const apiBase = computed(() => {
		const {host, protocol} = parseURL(window.API_URL)
		return protocol + '//' + host
	})

	function setConfig(config: ConfigState) {
		Object.assign(state, config)
	}
	async function update(): Promise<boolean> {
		const HTTP = HTTPFactory()
		const {data: config} = await HTTP.get('info')
		if (typeof config.version === 'undefined') {
			return false
		}
		setConfig(objectToCamelCase(config))
		const success = !!config
		return success
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