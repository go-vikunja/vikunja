<template>
	<Card :title="$t('admin.inviteLinks.title')">
		<XButton
			class="mbe-4"
			@click="openCreate"
		>
			{{ $t('admin.inviteLinks.create') }}
		</XButton>
		<p v-if="loading">
			{{ $t('misc.loading') }}
		</p>
		<p v-else-if="links.length === 0">
			{{ $t('admin.inviteLinks.empty') }}
		</p>
		<div
			v-else
			class="overflow-x-auto"
		>
			<table class="table has-actions is-striped is-hoverable is-fullwidth">
				<thead>
					<tr>
						<th>{{ $t('admin.inviteLinks.name') }}</th>
						<th>{{ $t('admin.inviteLinks.teams') }}</th>
						<th>{{ $t('admin.inviteLinks.uses') }}</th>
						<th>{{ $t('admin.inviteLinks.expiresAt') }}</th>
						<th>{{ $t('admin.inviteLinks.createdBy') }}</th>
						<th>{{ $t('task.attributes.created') }}</th>
						<th />
					</tr>
				</thead>
				<tbody>
					<tr
						v-for="link in links"
						:key="link.id"
					>
						<td>{{ link.name }}</td>
						<td>{{ (link.teams ?? []).map(team => team.name).join(', ') }}</td>
						<td>{{ link.uses }} / {{ link.max_uses ?? '∞' }}</td>
						<td>
							<TimeDisplay
								v-if="link.expires_at"
								:date="link.expires_at"
							/>
							<template v-else>
								{{ $t('admin.inviteLinks.neverExpires') }}
							</template>
						</td>
						<td>
							<User
								v-if="link.created_by"
								:user="link.created_by"
								:avatar-size="24"
							/>
							<template v-else>
								{{ $t('admin.inviteLinks.creatorDeleted') }}
							</template>
						</td>
						<td>
							<TimeDisplay
								v-if="link.created"
								:date="link.created"
							/>
						</td>
						<td class="actions">
							<XButton
								variant="secondary"
								danger
								@click="pendingDelete = link"
							>
								{{ $t('misc.delete') }}
							</XButton>
						</td>
					</tr>
				</tbody>
			</table>
		</div>
		<PaginationEmit
			v-if="totalPages > 1"
			:total-pages="totalPages"
			:current-page="page"
			@pageChanged="goToPage"
		/>

		<Modal
			v-if="createOpen"
			variant="hint-modal"
			:aria-label="$t('admin.inviteLinks.create')"
			@close="closeCreate"
		>
			<Card
				class="has-no-shadow"
				:title="$t('admin.inviteLinks.create')"
			>
				<template v-if="createdUrl">
					<Message
						variant="warning"
						class="mbe-4"
					>
						{{ $t('admin.inviteLinks.shownOnce') }}
					</Message>
					<FormField :label="$t('admin.inviteLinks.url')">
						<template #default="{id}">
							<input
								:id="id"
								:value="createdUrl"
								class="input"
								readonly
								@focus="selectLink"
							>
						</template>
					</FormField>
					<XButton @click="copy(createdUrl)">
						{{ $t(copied ? 'admin.inviteLinks.copied' : 'admin.inviteLinks.copy') }}
					</XButton>
				</template>
				<form
					v-else
					id="invite-create-form"
					@submit.prevent="submitCreate"
				>
					<FormField
						id="invite-name"
						v-model="form.name"
						:label="$t('admin.inviteLinks.name')"
						required
						maxlength="250"
					/>
					<FormField :label="$t('admin.inviteLinks.teams')">
						<template #default="{id}">
							<Multiselect
								:id="id"
								v-model="selectedTeams"
								multiple
								show-empty
								label="name"
								:aria-label="$t('admin.inviteLinks.teams')"
								:search-results="teamResults"
								:loading="loadingTeams"
								@search="teamSearch = $event"
							/>
						</template>
					</FormField>
					<FormField
						id="invite-max-uses"
						v-model="form.maxUses"
						type="number"
						min="1"
						step="1"
						:label="$t('admin.inviteLinks.maxUses')"
						:placeholder="$t('admin.inviteLinks.unlimited')"
					/>
					<FormField
						id="invite-expiry"
						v-model="form.expiresAt"
						type="datetime-local"
						:min="minimumExpiry"
						:label="$t('admin.inviteLinks.expiresAt')"
					/>
					<FormCheckbox
						v-model="form.skipEmailConfirm"
						:label="$t('admin.inviteLinks.skipEmailConfirm')"
					/>
				</form>
				<template #footer>
					<XButton
						variant="tertiary"
						:disabled="creating"
						@click="closeCreate"
					>
						{{ $t(createdUrl ? 'misc.close' : 'misc.cancel') }}
					</XButton>
					<XButton
						v-if="!createdUrl"
						type="submit"
						form="invite-create-form"
						:loading="creating"
						:disabled="!form.name.trim()"
					>
						{{ $t('admin.inviteLinks.create') }}
					</XButton>
				</template>
			</Card>
		</Modal>
		<Modal
			v-if="pendingDelete"
			:loading="deleting"
			@close="pendingDelete = null"
			@submit="deleteLink"
		>
			<template #header>
				{{ $t('admin.inviteLinks.delete') }}
			</template>
			<template #text>
				{{ $t('admin.inviteLinks.deleteConfirm', {name: pendingDelete.name}) }}
			</template>
		</Modal>
	</Card>
</template>

<script setup lang="ts">
import {computed, reactive, ref} from 'vue'
import {useClipboard} from '@vueuse/core'
import {formatDate} from '@/helpers/time/formatDate'
import {useI18n} from 'vue-i18n'
import Card from '@/components/misc/Card.vue'
import Modal from '@/components/misc/Modal.vue'
import Message from '@/components/misc/Message.vue'
import FormField from '@/components/input/FormField.vue'
import FormCheckbox from '@/components/input/FormCheckbox.vue'
import Multiselect from '@/components/input/Multiselect.vue'
import TimeDisplay from '@/components/misc/TimeDisplay.vue'
import User from '@/components/misc/User.vue'
import PaginationEmit from '@/components/misc/PaginationEmit.vue'
import {useQuery} from '@tanstack/vue-query'
import {
	adminInvitesQuery,
	adminTeamsQuery,
	useCreateAdminInviteMutation,
	useDeleteAdminInviteMutation,
} from '@/client/queries/admin'
import type {UserInviteLink, InviteLinkTeam} from '@/client/generated'
import {useConfigStore} from '@/stores/config'
import {useTitle} from '@/composables/useTitle'
import {error} from '@/message'

const {t} = useI18n()
useTitle(() => t('admin.inviteLinks.title'))
const configStore = useConfigStore()
const {copy, copied} = useClipboard({legacy: true})
const page = ref(1)
const {data, isPending: loading} = useQuery(computed(() => adminInvitesQuery(page.value)))
const links = computed(() => data.value?.items ?? [])
const totalPages = computed(() => data.value?.total_pages ?? 0)
const createOpen = ref(false)
const createMutation = useCreateAdminInviteMutation()
const {isPending: creating} = createMutation
const createdUrl = ref('')
const pendingDelete = ref<UserInviteLink | null>(null)
const deleteMutation = useDeleteAdminInviteMutation()
const {isPending: deleting} = deleteMutation
const selectedTeams = ref<InviteLinkTeam[]>([])
const teamSearch = ref('')
const {data: teamData, isFetching: loadingTeams} = useQuery(computed(() => ({
	...adminTeamsQuery(teamSearch.value),
	enabled: createOpen.value,
})))
const teamResults = computed(() => teamData.value ?? [])
const minimumExpiry = ref('')
const form = reactive({name: '', maxUses: '' as string | number, expiresAt: '', skipEmailConfirm: false})

function goToPage(nextPage = page.value) { page.value = nextPage }

function openCreate() {
	Object.assign(form, {name: '', maxUses: '', expiresAt: '', skipEmailConfirm: false})
	selectedTeams.value = []
	createdUrl.value = ''
	minimumExpiry.value = formatDate(new Date(Date.now() + 60_000), 'YYYY-MM-DD[T]HH:mm')
	createOpen.value = true
	teamSearch.value = ''
}

function closeCreate() {
	if (creating.value) return
	createOpen.value = false
	createdUrl.value = ''
}

function selectLink(event: FocusEvent) {
	(event.target as HTMLInputElement).select()
}

async function submitCreate() {
	if (creating.value) return
	const expiresAt = form.expiresAt ? new Date(form.expiresAt) : null
	if (expiresAt && expiresAt.getTime() <= Date.now()) {
		error({message: t('admin.inviteLinks.futureExpiry')})
		return
	}
	try {
		const link = await createMutation.mutateAsync({
			name: form.name,
			team_ids: selectedTeams.value.map(team => team.id).filter((id): id is number => id !== undefined),
			max_uses: form.maxUses === '' ? null : Number(form.maxUses),
			expires_at: expiresAt?.toISOString() ?? null,
			skip_email_confirm: form.skipEmailConfirm,
		})
		const base = configStore.frontend_url || new URL(import.meta.env.BASE_URL, window.location.origin).toString()
		createdUrl.value = new URL(
			`register#invite-link=${encodeURIComponent(link.token!)}`,
			base.endsWith('/') ? base : `${base}/`,
		).toString()
		goToPage(1)
	} catch { /* Mutation reports the error. */ }
	finally { createMutation.reset() }
}

async function deleteLink() {
	if (pendingDelete.value?.id === undefined || deleting.value) return
	try {
		const nextPage = links.value.length === 1 && page.value > 1 ? page.value - 1 : page.value
		await deleteMutation.mutateAsync(pendingDelete.value.id)
		pendingDelete.value = null
		goToPage(nextPage)
	} catch { /* Mutation reports the error. */ }
}

</script>
