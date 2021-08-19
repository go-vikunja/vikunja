import {CONFIG} from '../mutation-types'
import {HTTPFactory} from '@/http-common'
import {objectToCamelCase} from '@/helpers/case'
import {redirectToProvider} from '../../helpers/redirectToProvider'

export default {
	namespaced: true,
	state: () => ({
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
	mutations: {
		[CONFIG](state, config) {
			state.version = config.version
			state.frontendUrl = config.frontend_url
			state.motd = config.motd
			state.linkSharingEnabled = config.link_sharing_enabled
			state.maxFileSize = config.max_file_size
			state.registrationEnabled = config.registration_enabled
			state.availableMigrators = config.available_migrators
			state.taskAttachmentsEnabled = config.task_attachments_enabled
			state.totpEnabled = config.totp_enabled
			state.enabledBackgroundProviders = config.enabled_background_providers
			state.legal.imprintUrl = config.legal.imprint_url
			state.legal.privacyPolicyUrl = config.legal.privacy_policy_url
			state.caldavEnabled = config.caldav_enabled
			state.userDeletionEnabled = config.user_deletion_enabled
			state.taskCommentsEnabled = config.task_comments_enabled
			const auth = objectToCamelCase(config.auth)
			state.auth.local.enabled = auth.local.enabled
			state.auth.openidConnect.enabled = auth.openidConnect.enabled
			state.auth.openidConnect.redirectUrl = auth.openidConnect.redirectUrl
			state.auth.openidConnect.providers = auth.openidConnect.providers
		},
	},
	actions: {
		update(ctx) {
			const HTTP = HTTPFactory()

			return HTTP.get('info')
				.then(r => {
					ctx.commit(CONFIG, r.data)
					return Promise.resolve(r)
				})
				.catch(e => Promise.reject(e))
		},
		redirectToProviderIfNothingElseIsEnabled(ctx) {
			if (ctx.state.auth.local.enabled === false &&
				ctx.state.auth.openidConnect.enabled &&
				ctx.state.auth.openidConnect.providers &&
				ctx.state.auth.openidConnect.providers.length === 1 &&
				window.location.pathname.startsWith('/login') // Kinda hacky, but prevents an endless loop.
			) {
				redirectToProvider(ctx.state.auth.openidConnect.providers[0], ctx.state.auth.openidConnect.redirectUrl)
			}
		},
	},
}