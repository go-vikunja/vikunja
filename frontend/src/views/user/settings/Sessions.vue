<script setup lang="ts">
import {ref, shallowReactive} from 'vue'
import {useI18n} from 'vue-i18n'

import {useTitle} from '@/composables/useTitle'
import {useAuthStore} from '@/stores/auth'
import {success} from '@/message'
import {formatDateSince} from '@/helpers/time/formatDate'
import SessionService from '@/services/session'
import type {ISession} from '@/modelTypes/ISession'

const {t} = useI18n({useScope: 'global'})
useTitle(() => `${t('user.settings.sessions.title')} - ${t('user.settings.title')}`)

const authStore = useAuthStore()
const service = shallowReactive(new SessionService())
const sessions = ref<ISession[]>([])

const showDeleteModal = ref(false)
const sessionToDelete = ref<ISession | null>(null)

service.getAll().then((result: ISession[]) => {
	sessions.value = result
})

function confirmDelete(session: ISession) {
	sessionToDelete.value = session
	showDeleteModal.value = true
}

async function deleteSession() {
	if (!sessionToDelete.value) return

	await service.delete(sessionToDelete.value)
	sessions.value = sessions.value.filter(({id}) => id !== sessionToDelete.value?.id)
	showDeleteModal.value = false
	sessionToDelete.value = null
	success({message: t('user.settings.sessions.deleteSuccess')})
}
</script>

<template>
	<Card :title="$t('user.settings.sessions.title')">
		<p class="mbe-4">
			{{ $t('user.settings.sessions.description') }}
		</p>

		<table
			v-if="sessions.length > 0"
			class="table"
		>
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
						{{ session.deviceInfo }}
						<span
							v-if="session.id === authStore.currentSessionId"
							class="tag is-primary mis-2"
						>
							{{ $t('user.settings.sessions.current') }}
						</span>
					</td>
					<td>{{ session.ipAddress }}</td>
					<td>{{ formatDateSince(session.lastActive) }}</td>
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

		<p v-else>
			{{ $t('user.settings.sessions.noOtherSessions') }}
		</p>

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
