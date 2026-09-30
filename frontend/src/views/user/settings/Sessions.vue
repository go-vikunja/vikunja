<script setup lang="ts">
import {
	ref,
	computed,
	watch,
} from 'vue'
import {useI18n} from 'vue-i18n'
import {useRouteQuery} from '@vueuse/router'

import {useTitle} from '@/composables/useTitle'
import {useAuthStore} from '@/stores/auth'
import {formatDateSince} from '@/helpers/time/formatDate'
import {useQuery} from '@tanstack/vue-query'
import {sessionsQuery, useDeleteSessionMutation} from '@/client/queries/sessions'
import {normalizePageNumber} from '@/client/queries/pagination'
import type {Session} from '@/client/generated'
import ErrorMessage from '@/components/misc/Error.vue'
import Pagination from '@/components/misc/Pagination.vue'

const {t} = useI18n({useScope: 'global'})
useTitle(() => `${t('user.settings.sessions.title')} - ${t('user.settings.title')}`)

const authStore = useAuthStore()
const page = useRouteQuery('page', '1', {transform: normalizePageNumber})
const sessionQuery = useQuery(computed(() => sessionsQuery(page.value)))
const sessions = computed(() => sessionQuery.data.value?.items ?? [])
const totalPages = computed(() => sessionQuery.data.value?.total_pages ?? 0)
watch([page, totalPages], ([current, last]) => {
	if (last > 0 && current > last) page.value = last
}, {immediate: true})
const deleteMutation = useDeleteSessionMutation()

const showDeleteModal = ref(false)
const sessionToDelete = ref<Session | null>(null)

function confirmDelete(session: Session) {
	sessionToDelete.value = session
	showDeleteModal.value = true
}

async function deleteSession() {
	if (deleteMutation.isPending.value) return
	if (!sessionToDelete.value?.id) return

	try {
		await deleteMutation.mutateAsync(sessionToDelete.value.id)
	} catch { return }
	showDeleteModal.value = false
	sessionToDelete.value = null
}
</script>

<template>
	<Card
		:title="$t('user.settings.sessions.title')"
		:loading="sessionQuery.isPending.value"
	>
		<p class="mbe-4">
			{{ $t('user.settings.sessions.description') }}
		</p>

		<ErrorMessage v-if="sessionQuery.isError.value" />

		<div
			v-if="sessions.length > 0"
			class="has-horizontal-overflow loader-container"
			:class="{'is-loading': sessionQuery.isPlaceholderData.value}"
		>
			<table class="table">
				<thead>
					<tr>
						<th>{{ $t('user.settings.sessions.deviceInfo') }}</th>
						<th>{{ $t('user.settings.sessions.ipAddress') }}</th>
						<th>{{ $t('user.settings.sessions.lastActive') }}</th>
						<th class="has-text-end">
							{{ $t('misc.actions') }}
						</th>
					</tr>
				</thead>
				<tbody>
					<tr
						v-for="session in sessions"
						:key="session.id"
					>
						<td>
							{{ session.device_info }}
							<span
								v-if="session.id === authStore.currentSessionId"
								class="tag is-primary mis-2"
							>
								{{ $t('user.settings.sessions.current') }}
							</span>
						</td>
						<td>{{ session.ip_address }}</td>
						<td>{{ formatDateSince(session.last_active) }}</td>
						<td class="has-text-end">
							<XButton
								v-if="session.id !== authStore.currentSessionId"
								variant="secondary"
								@click="confirmDelete(session)"
							>
								{{ $t('misc.delete') }}
							</XButton>
						</td>
					</tr>
				</tbody>
			</table>
		</div>

		<p v-else-if="!sessionQuery.isPending.value && !sessionQuery.isError.value">
			{{ $t('user.settings.sessions.noOtherSessions') }}
		</p>

		<Pagination
			:total-pages="totalPages"
			:current-page="page"
		/>

		<Modal
			:enabled="showDeleteModal"
			@close="showDeleteModal = false"
			@submit="deleteSession()"
		>
			<template #header>
				{{ $t('user.settings.sessions.delete.header') }}
			</template>

			<template #text>
				<p>
					{{ $t('user.settings.sessions.delete.text') }}
				</p>
			</template>
		</Modal>
	</Card>
</template>
