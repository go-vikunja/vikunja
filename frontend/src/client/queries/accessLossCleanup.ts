import type {QueryClient} from '@tanstack/vue-query'
import {linkShareKeys} from './linkShares'
import {webhookKeys} from './webhooks'

export type AccessLossCleanup = {
	keep_webhooks: boolean
	revoke_link_shares: boolean
}

// Asking first fails with 409 and the affected counts instead of applying the defaults.
export type AccessLossChoice = AccessLossCleanup | {confirm_access_loss: true}

export function defaultAccessLossCleanup(): AccessLossCleanup {
	return {keep_webhooks: false, revoke_link_shares: false}
}

// The cleanup reaches every project the users lost, not just the one on screen.
export function invalidateAccessLossCleanup(client: QueryClient) {
	return Promise.all([
		client.invalidateQueries({queryKey: linkShareKeys.all}),
		client.invalidateQueries({queryKey: webhookKeys.projects}),
	])
}
