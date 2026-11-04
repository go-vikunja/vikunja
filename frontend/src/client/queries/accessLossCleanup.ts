import type {QueryClient} from '@tanstack/vue-query'
import type {ProjectUsersDeleteData} from '@/client/generated'
import {linkShareKeys} from './linkShares'
import {webhookKeys} from './webhooks'

export type AccessLossCleanup = Required<NonNullable<ProjectUsersDeleteData['query']>>

// Asking first fails with 409 and the affected counts instead of applying the defaults.
export type AccessLossChoice = AccessLossCleanup | {confirm_access_loss: true}

export function defaultAccessLossCleanup(): AccessLossCleanup {
	return {
		keep_webhooks: false,
		revoke_link_shares: false,
	}
}

// The cleanup reaches every project the users lost, not just the one on screen.
export function invalidateAccessLossCleanup(client: QueryClient, refetchType: 'active' | 'none' = 'active') {
	return Promise.all([
		client.invalidateQueries({queryKey: linkShareKeys.all, refetchType}),
		client.invalidateQueries({queryKey: webhookKeys.projects, refetchType}),
	])
}
