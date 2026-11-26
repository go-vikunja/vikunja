<script setup lang="ts">
import {computed, ref} from 'vue'
import {useI18n} from 'vue-i18n'
import {useNow} from '@vueuse/core'
import {useRouteQuery} from '@vueuse/router'
import {useQuery} from '@tanstack/vue-query'
import {connectedAppsQuery, useDeleteApiTokenMutation} from '@/client/queries/apiTokens'
import {normalizePageNumber} from '@/client/queries/pagination'
import type {ApiToken} from '@/client/generated'
import {useClampedPage} from '@/composables/useClampedPage'
import {isApiTokenExpired, oauthClientHost} from '@/helpers/apiToken'
import {formatDateSince, formatDisplayDate} from '@/helpers/time/formatDate'
import XButton from '@/components/input/Button.vue'
import ErrorMessage from '@/components/misc/Error.vue'
import Pagination from '@/components/misc/Pagination.vue'

const ACCESS_LABELS = {
	read: 'project.share.permission.read',
	write: 'project.share.permission.readWrite',
} as const satisfies Record<NonNullable<ApiToken['oauth_access']>, string>

const {t} = useI18n({useScope: 'global'})
const page = useRouteQuery('page', '1', {transform: normalizePageNumber})
const query = useQuery(computed(() => connectedAppsQuery(page.value)))
useClampedPage(page, query)
const apps = computed(() => (query.data.value?.items ?? []).map(token => {
	const host = oauthClientHost(token)
	return {
		token,
		host,
		name: token.title || host || token.oauth_client_id,
	}
}))
const totalPages = computed(() => query.data.value?.total_pages ?? 0)
const deleteMutation = useDeleteApiTokenMutation()
const appToRevoke = ref<{token: ApiToken, name?: string}>()
const now = useNow({interval: 60_000})

async function revoke() {
	const app = appToRevoke.value
	if (!app?.token.id || deleteMutation.isPending.value) return
	try {
		await deleteMutation.mutateAsync(app.token.id)
	} catch { return }
	if (appToRevoke.value === app) appToRevoke.value = undefined
}
</script>

<template>
	<Card
		:title="t('user.settings.mcp.connectedApps.title')"
		:loading="query.isPending.value"
	>
		<p>{{ t('user.settings.mcp.connectedApps.description') }}</p>
		<ErrorMessage v-if="query.isError.value" />
		<div
			v-else-if="apps.length"
			class="has-horizontal-overflow"
		>
			<table class="table">
				<thead>
					<tr>
						<th>{{ t('user.settings.mcp.connectedApps.app') }}</th>
						<th>{{ t('user.auth.oauthConsent.access') }}</th>
						<th>{{ t('user.settings.mcp.connectedApps.connected') }}</th>
						<th>{{ t('user.settings.apiTokens.attributes.expiresAt') }}</th>
						<th class="has-text-end">
							{{ t('misc.actions') }}
						</th>
					</tr>
				</thead>
				<tbody>
					<tr
						v-for="app in apps"
						:key="app.token.id"
						:class="{'is-expired': isApiTokenExpired(app.token, now.getTime())}"
					>
						<td>
							<span class="app-name">{{ app.name }}</span>
							<span
								v-if="app.host"
								class="app-host"
							>
								{{ t('user.auth.oauthConsent.publisher') }}
								<strong>{{ app.host }}</strong>
							</span>
						</td>
						<td>{{ app.token.oauth_access ? t(ACCESS_LABELS[app.token.oauth_access]) : '' }}</td>
						<td>{{ formatDisplayDate(app.token.created) }}</td>
						<td>
							{{ formatDisplayDate(app.token.expires_at) }}
							<p
								v-if="isApiTokenExpired(app.token, now.getTime())"
								class="has-text-danger"
							>
								{{ t('user.settings.apiTokens.expired', {ago: formatDateSince(app.token.expires_at)}) }}
							</p>
						</td>
						<td class="has-text-end">
							<XButton
								variant="secondary"
								@click="appToRevoke = app"
							>
								{{ t('user.settings.mcp.connectedApps.revoke') }}
							</XButton>
						</td>
					</tr>
				</tbody>
			</table>
		</div>
		<p v-else-if="!query.isPending.value">
			{{ t('user.settings.mcp.connectedApps.none') }}
		</p>
		<Pagination
			:total-pages="totalPages"
			:current-page="page"
		/>
		<Modal
			:enabled="appToRevoke !== undefined"
			@close="appToRevoke = undefined"
			@submit="revoke"
		>
			<template #header>
				{{ t('user.settings.mcp.connectedApps.revokeHeader') }}
			</template>
			<template #text>
				<p v-if="appToRevoke">
					{{ t('user.settings.mcp.connectedApps.revokeText', {app: appToRevoke.name}) }}
				</p>
			</template>
		</Modal>
	</Card>
</template>

<style scoped lang="scss">
.is-expired {
	opacity: .6;
}

.app-name,
.app-host {
	display: block;
	overflow-wrap: anywhere;
}

.app-host {
	color: var(--grey-500);
	font-size: .875rem;
}
</style>
