import {computed, reactive, toRefs} from 'vue'
import {defineStore, acceptHMRUpdate} from 'pinia'
import {parseURL} from 'ufo'

import {HTTPFactory} from '@/helpers/fetcher'
import {objectToCamelCase} from '@/helpers/case'

import type {IProvider} from '@/types/IProvider'

export interface ConfigState {
	version: string,
	frontendUrl: string,
	motd: string,
	linkSharingEnabled: boolean,
	maxFileSize: string,
	registrationEnabled: boolean,
	availableMigrators: [],
	taskAttachmentsEnabled: boolean,
	totpEnabled: boolean,
	enabledBackgroundProviders: [],
	legal: {
		imprintUrl: string,
		privacyPolicyUrl: string,
	},
	caldavEnabled: boolean,
	userDeletionEnabled: boolean,
	taskCommentsEnabled: boolean,
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
	async function update() {
		const HTTP = HTTPFactory()
		const {data: config} = await HTTP.get('info')
		setConfig(objectToCamelCase(config))
		return config
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