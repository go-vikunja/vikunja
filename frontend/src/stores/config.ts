import {computed, reactive, toRefs} from 'vue'
import {acceptHMRUpdate, defineStore} from 'pinia'
import {parseURL} from 'ufo'

import {getApiV2BaseUrl} from '@/helpers/fetcher'
import {info, type VikunjaInfos, type AuthInfo} from '@/client/generated'
import {createClient} from '@/client/generated/client'

import type {IProvider} from '@/types/IProvider'
import type {ProFeature} from '@/constants/proFeatures'
import {InvalidApiUrlProvidedError} from '@/helpers/checkAndSetApiUrl'

export type ConfigState = Required<Omit<VikunjaInfos,
 '$schema' | 'auth' | 'legal' | 'available_migrators' | 'enabled_background_providers' | 'enabled_pro_features'
>> & {
 available_migrators: string[]
 enabled_background_providers: string[]
 enabled_pro_features: string[]
 legal: Required<NonNullable<VikunjaInfos['legal']>>
 auth: {
  local: Required<NonNullable<AuthInfo['local']>>
  ldap: Required<NonNullable<AuthInfo['ldap']>>
  openid_connect: {enabled: boolean; providers: IProvider[]}
 }
}

const publicClient = createClient({throwOnError: true})

export const useConfigStore = defineStore('config', () => {
	const state: ConfigState = reactive({
		// These are the api defaults.
		version: '',
		email_reminders_enabled: true,
		frontend_url: '',
		motd: '',
		link_sharing_enabled: true,
		max_file_size: '20MB',
		max_items_per_page: 50,
		available_migrators: [],
		task_attachments_enabled: true,
		totp_enabled: true,
		enabled_background_providers: [],
		legal: {
			imprint_url: '',
			privacy_policy_url: '',
		},
		caldav_enabled: false,
		user_deletion_enabled: true,
		task_comments_enabled: true,
		demo_mode_enabled: false,
		webhooks_enabled: false,
		auth: {
			local: {
				enabled: true,
				registration_enabled: true,
			},
			ldap: {
				enabled: false,
			},
			openid_connect: {
				enabled: false,
				providers: [],
			},
		},
		public_teams_enabled: false,
		allow_icon_changes: true,
		enabled_pro_features: [],
		concurrent_writes: false,
	})

	const migratorsEnabled = computed(() => state.available_migrators?.length > 0)
	const apiBase = computed(() => {
		const {host, protocol, pathname} = parseURL(window.API_URL)

		// Strip the /api/v1 suffix (and optional trailing slash) to get the deployment base.
		const basePath = pathname
			.replace(/\/api\/v1\/?$/, '')
			.replace(/\/+$/, '')
		return `${protocol}//${host}${basePath}`
	})

	function setConfig(config: VikunjaInfos) {
		Object.assign(state, config, {
			available_migrators: config.available_migrators ?? [],
			enabled_background_providers: config.enabled_background_providers ?? [],
			enabled_pro_features: config.enabled_pro_features ?? [],
			legal: {...state.legal, ...config.legal},
			auth: {
				local: {...state.auth.local, ...config.auth?.local},
				ldap: {...state.auth.ldap, ...config.auth?.ldap},
				openid_connect: {
					enabled: config.auth?.openid_connect?.enabled ?? false,
					providers: (config.auth?.openid_connect?.providers ?? []).map(provider => ({
						...provider,
						name: provider.name ?? '',
						key: provider.key ?? '',
						auth_url: provider.auth_url ?? '',
						client_id: provider.client_id ?? '',
						logout_url: provider.logout_url ?? '',
						scope: provider.scope ?? 'openid email profile',
					})),
				},
			},
		})
	}

	function isProFeatureEnabled(name: ProFeature): boolean {
		return state.enabled_pro_features?.includes(name) ?? false
	}

	async function update(): Promise<boolean> {
		const {data: config} = await info({
			client: publicClient,
			baseUrl: getApiV2BaseUrl().replace(/\/$/, ''),
		})

		if (typeof config.version === 'undefined') {
			throw new InvalidApiUrlProvidedError()
		}

		setConfig(config)
		return !!config
	}

	return {
		...toRefs(state),

		migratorsEnabled,
		apiBase,
		setConfig,
		isProFeatureEnabled,
		update,
	}

})

// support hot reloading
if (import.meta.hot) {
	import.meta.hot.accept(acceptHMRUpdate(useConfigStore, import.meta.hot))
}
