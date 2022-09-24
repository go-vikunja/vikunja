import {defineStore, acceptHMRUpdate} from 'pinia'
import {parseURL} from 'ufo'

import {HTTPFactory} from '@/http-common'
import {objectToCamelCase} from '@/helpers/case'

export interface ConfigState {
	version: string,
	frontendUrl: string,
	motd: string,
	linkSharingEnabled: boolean,
	maxFileSize: '20MB',
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
			providers: [],
		},
	},
}

export const useConfigStore = defineStore('config', {
	state: (): ConfigState => ({
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
	}),
	getters: {
		migratorsEnabled: (state) => state.availableMigrators?.length > 0,
		apiBase() {
			const {host, protocol} = parseURL(window.API_URL)
			return protocol + '//' + host
		},
	},
	actions: {
		setConfig(config: ConfigState) {
			Object.assign(this, config)
		},
		async update() {
			const HTTP = HTTPFactory()
			const {data: config} = await HTTP.get('info')
			this.setConfig(objectToCamelCase(config))
			return config
		},
	},
})

// support hot reloading
if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useConfigStore, import.meta.hot))
}