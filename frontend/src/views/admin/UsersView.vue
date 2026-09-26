<template>
	<Card>
		<div class="admin-users">
			<div class="admin-users__toolbar">
				<FormInput
					v-model="searchTerm"
					type="text"
					:placeholder="$t('admin.searchUsersPlaceholder')"
					@input="onSearch"
				/>
				<XButton
					variant="primary"
					@click="openCreate"
				>
					{{ $t('admin.users.addUser') }}
				</XButton>
			</div>

			<p v-if="loading">
				{{ $t('misc.loading') }}
			</p>
			<template v-else>
				<table class="table has-actions is-striped is-hoverable is-fullwidth">
					<thead>
						<tr>
							<th>{{ $t('misc.id') }}</th>
							<th>{{ $t('user.auth.username') }}</th>
							<th>{{ $t('user.auth.email') }}</th>
							<th>{{ $t('admin.users.issuer') }}</th>
							<th>{{ $t('admin.users.status') }}</th>
							<th>{{ $t('task.attributes.created') }}</th>
							<th />
						</tr>
					</thead>
					<tbody>
						<tr
							v-for="u in users"
							:key="u.id"
						>
							<td>{{ u.id }}</td>
							<td>{{ u.username }}</td>
							<td>{{ u.email }}</td>
							<td>{{ u.auth_provider || $t('admin.users.issuerLocal') }}</td>
							<td>{{ statusLabel(u.status) }}</td>
							<td>
								<TimeDisplay :date="u.created" />
							</td>
							<td class="actions">
								<XButton
									variant="secondary"
									@click="openDetails(u)"
								>
									{{ $t('admin.users.details') }}
								</XButton>
							</td>
						</tr>
					</tbody>
				</table>
				<PaginationEmit
					v-if="totalPages > 1"
					:total-pages="totalPages"
					:current-page="currentPage"
					@pageChanged="goToPage"
				/>
			</template>

			<Modal
				v-if="detailTarget && !pendingDelete"
				variant="hint-modal"
				@close="closeDetail"
			>
				<Card
					class="has-no-shadow"
					:title="$t('admin.users.detailsTitle', {username: detailTarget.username})"
				>
					<dl class="admin-users__meta">
						<dt>{{ $t('misc.id') }}</dt>
						<dd>{{ detailTarget.id }}</dd>
						<dt>{{ $t('user.auth.email') }}</dt>
						<dd>{{ detailTarget.email }}</dd>
						<dt>{{ $t('admin.users.issuer') }}</dt>
						<dd>
							{{ detailTarget.auth_provider || $t('admin.users.issuerLocal') }}
						</dd>
						<template v-if="detailTarget.issuer?.startsWith('http')">
							<dt>{{ $t('admin.users.issuerUrl') }}</dt>
							<dd class="admin-users__issuer-url-value">
								{{ detailTarget.issuer }}
							</dd>
						</template>
						<template v-if="detailTarget.subject">
							<dt>{{ $t('admin.users.subject') }}</dt>
							<dd class="admin-users__subject">
								{{ detailTarget.subject }}
							</dd>
						</template>
						<dt>{{ $t('task.attributes.created') }}</dt>
						<dd>
							<TimeDisplay :date="detailTarget.created" />
						</dd>
						<dt>{{ $t('task.attributes.updated') }}</dt>
						<dd>
							<TimeDisplay :date="detailTarget.updated" />
						</dd>
					</dl>

					<FormCheckbox
						v-model="editable.is_admin"
						:label="$t('admin.users.isAdminLabel')"
					/>

					<FormField :label="$t('admin.users.status')">
						<template #default="{id}">
							<FormSelect
								:id="id"
								v-model.number="editable.status"
								:options="statusOptions"
							/>
						</template>
					</FormField>

					<template v-if="!detailTarget.auth_provider">
						<FormField :label="$t('admin.users.newPasswordLabel')">
							<template #default="{id}">
								<FormInput
									:id="id"
									v-model="newPassword"
									type="password"
									autocomplete="new-password"
								/>
							</template>
						</FormField>
						<div class="admin-users__password-actions">
							<XButton
								variant="secondary"
								:disabled="!newPassword || settingPassword"
								:loading="settingPassword"
								@click="setPassword"
							>
								{{ $t('admin.users.setPassword') }}
							</XButton>
							<XButton
								variant="secondary"
								:disabled="sendingResetEmail"
								:loading="sendingResetEmail"
								@click="sendResetEmail"
							>
								{{ $t('admin.users.sendResetEmail') }}
							</XButton>
						</div>
					</template>

					<template #footer>
						<XButton
							variant="tertiary"
							@click="closeDetail"
						>
							{{ $t('misc.cancel') }}
						</XButton>
						<XButton
							v-if="detailTarget.id !== currentUserId"
							variant="secondary"
							:danger="true"
							@click="pendingDelete = detailTarget"
						>
							{{ $t('misc.delete') }}
						</XButton>
						<XButton
							variant="primary"
							:disabled="!hasChanges || saving"
							:loading="saving"
							@click="saveChanges"
						>
							{{ $t('admin.users.saveButton') }}
						</XButton>
					</template>
				</Card>
			</Modal>

			<Modal
				v-if="createOpen"
				variant="hint-modal"
				@close="closeCreate"
			>
				<Card
					class="has-no-shadow"
					:title="$t('admin.users.createTitle')"
				>
					<FormField :label="$t('user.auth.username')">
						<template #default="{id}">
							<FormInput
								:id="id"
								v-model="createForm.username"
								type="text"
								required
							/>
						</template>
					</FormField>
					<FormField :label="$t('user.auth.email')">
						<template #default="{id}">
							<FormInput
								:id="id"
								v-model="createForm.email"
								type="email"
								required
							/>
						</template>
					</FormField>
					<FormField :label="$t('admin.users.nameLabel')">
						<template #default="{id}">
							<FormInput
								:id="id"
								v-model="createForm.name"
								type="text"
							/>
						</template>
					</FormField>
					<FormField :label="$t('user.auth.password')">
						<template #default="{id}">
							<FormInput
								:id="id"
								v-model="createForm.password"
								type="password"
								autocomplete="new-password"
								required
							/>
						</template>
					</FormField>
					<FormField :label="$t('user.settings.general.language')">
						<template #default="{id}">
							<FormInput
								:id="id"
								v-model="createForm.language"
								type="text"
							/>
						</template>
					</FormField>
					<FormCheckbox
						v-model="createForm.is_admin"
						:label="$t('admin.users.isAdminLabel')"
					/>
					<FormCheckbox
						v-model="createForm.skip_email_confirm"
						:label="$t('admin.users.skipEmailConfirm')"
					/>

					<template #footer>
						<XButton
							variant="tertiary"
							@click="closeCreate"
						>
							{{ $t('misc.cancel') }}
						</XButton>
						<XButton
							variant="primary"
							:disabled="creating || !createForm.username || !createForm.email || !createForm.password"
							:loading="creating"
							@click="submitCreate"
						>
							{{ $t('admin.users.createSubmit') }}
						</XButton>
					</template>
				</Card>
			</Modal>

			<Modal
				v-if="pendingDelete"
				variant="hint-modal"
				@close="cancelDelete"
			>
				<Card
					class="has-no-shadow"
					:title="$t('admin.users.confirmDeleteTitle')"
				>
					<p>{{ $t('admin.users.confirmDeleteIntro', {username: pendingDelete.username}) }}</p>
					<p>{{ $t('admin.users.deleteModeScheduledHelp') }}</p>
					<p>{{ $t('admin.users.deleteModeNowHelp') }}</p>

					<template #footer>
						<XButton
							variant="tertiary"
							@click="cancelDelete"
						>
							{{ $t('misc.cancel') }}
						</XButton>
						<XButton
							variant="secondary"
							:loading="deleting && deleteMode === 'scheduled'"
							:disabled="deleting"
							@click="doDelete('scheduled')"
						>
							{{ $t('admin.users.deleteModeScheduled') }}
						</XButton>
						<XButton
							variant="primary"
							:danger="true"
							:loading="deleting && deleteMode === 'now'"
							:disabled="deleting"
							@click="doDelete('now')"
						>
							{{ $t('admin.users.deleteModeNow') }}
						</XButton>
					</template>
				</Card>
			</Modal>
		</div>
	</Card>
</template>

<script setup lang="ts">
import {ref, computed, reactive, watch} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import type {AdminUser, CreateUserBodyWritable} from '@/client/generated'
import {
	adminUsersQuery,
	useCreateAdminUserMutation,
	useUpdateAdminUserMutation,
	useSetAdminUserPasswordMutation,
	useResetAdminUserPasswordMutation,
	useDeleteAdminUserMutation,
	type DeleteUserMode,
} from '@/client/queries/admin'
import {useDebounceFn} from '@vueuse/core'
import {useI18n} from 'vue-i18n'
import {useAuthStore} from '@/stores/auth'
import Card from '@/components/misc/Card.vue'
import Modal from '@/components/misc/Modal.vue'
import PaginationEmit from '@/components/misc/PaginationEmit.vue'
import XButton from '@/components/input/Button.vue'
import FormField from '@/components/input/FormField.vue'
import FormInput from '@/components/input/FormInput.vue'
import FormSelect from '@/components/input/FormSelect.vue'
import FormCheckbox from '@/components/input/FormCheckbox.vue'
import TimeDisplay from '@/components/misc/TimeDisplay.vue'

type AdminUserRow = AdminUser & {id: number}
const {t} = useI18n({useScope: 'global'})
const authStore = useAuthStore()
const currentUserId = computed(() => authStore.info?.id)
const searchTerm = ref('')
const search = ref('')
const currentPage = ref(1)
const {data, isPending: loading} = useQuery(computed(() => adminUsersQuery(search.value, currentPage.value)))
const users = computed(() => (data.value?.items ?? []).filter((u): u is AdminUserRow => u.id !== undefined))
const totalPages = computed(() => data.value?.total_pages ?? 1)
const createMutation = useCreateAdminUserMutation()
const updateMutation = useUpdateAdminUserMutation()
const passwordMutation = useSetAdminUserPasswordMutation()
const resetMutation = useResetAdminUserPasswordMutation()
const deleteMutation = useDeleteAdminUserMutation()
const {isPending: creating} = createMutation
const {isPending: saving} = updateMutation
const {isPending: settingPassword} = passwordMutation
const {isPending: sendingResetEmail} = resetMutation
const {isPending: deleting} = deleteMutation
const detailTarget = ref<AdminUserRow | null>(null)
const pendingDelete = ref<AdminUserRow | null>(null)
const deleteMode = ref<DeleteUserMode | null>(null)
const createOpen = ref(false)
const editable = reactive({is_admin: false, status: 0})
const newPassword = ref('')
function emptyCreateForm(): Required<CreateUserBodyWritable> {
	return {
		username: '',
		email: '',
		name: '',
		password: '',
		language: '',
		is_admin: false,
		skip_email_confirm: false,
	}
}
const createForm = reactive(emptyCreateForm())
const hasChanges = computed(() => detailTarget.value
	&& (editable.is_admin !== !!detailTarget.value.is_admin || editable.status !== detailTarget.value.status))
watch(detailTarget, u => {
	newPassword.value = ''
	if (!u) return
	editable.is_admin = !!u.is_admin
	editable.status = u.status ?? 0
})
function statusLabel(status: number | undefined): string {
	switch (status) {
		case 0: return t('admin.users.statusActive')
		case 1: return t('admin.users.statusEmailConfirmation')
		case 2: return t('admin.users.statusDisabled')
		case 3: return t('admin.users.statusLocked')
		default: return status === undefined ? '' : String(status)
	}
}
const statusOptions = computed(() => [0, 1, 2, 3].map(value => ({value, label: statusLabel(value)})))
function goToPage(page: number) { currentPage.value = page }
const onSearch = useDebounceFn(() => {
	currentPage.value = 1
	search.value = searchTerm.value
}, 300)
function openDetails(u: AdminUserRow) { detailTarget.value = u }
function closeDetail() { detailTarget.value = null }
function openCreate() {
	Object.assign(createForm, emptyCreateForm())
	createOpen.value = true
}
function closeCreate() {
	if (creating.value) return
	createOpen.value = false
	Object.assign(createForm, emptyCreateForm())
}
async function submitCreate() {
	try {
		await createMutation.mutateAsync({...createForm, language: createForm.language || undefined})
		closeCreate()
	} catch { /* Mutation reports the error. */ }
	finally { createMutation.reset() }
}
async function saveChanges() {
	const target = detailTarget.value
	if (!target) return
	try {
		await updateMutation.mutateAsync({
			id: target.id,
			is_admin: editable.is_admin !== !!target.is_admin ? editable.is_admin : undefined,
			status: editable.status !== target.status ? editable.status : undefined,
		})
		if (detailTarget.value?.id === target.id) detailTarget.value = null
	} catch { /* Mutation reports the error. */ }
}
async function setPassword() {
	const target = detailTarget.value
	if (!target || !newPassword.value) return
	try {
		await passwordMutation.mutateAsync({id: target.id, password: newPassword.value})
		if (detailTarget.value?.id === target.id) newPassword.value = ''
	} catch { /* Mutation reports the error. */ }
	finally { passwordMutation.reset() }
}
function sendResetEmail() {
	if (detailTarget.value) resetMutation.mutate({id: detailTarget.value.id, username: detailTarget.value.username})
}
function cancelDelete() {
	if (deleting.value) return
	pendingDelete.value = null
	deleteMode.value = null
}
async function doDelete(mode: DeleteUserMode) {
	const target = pendingDelete.value
	if (!target || deleting.value) return
	deleteMode.value = mode
	try {
		await deleteMutation.mutateAsync({id: target.id, mode, username: target.username})
		pendingDelete.value = null
		detailTarget.value = null
	} catch { /* Mutation reports the error. */ }
	finally { deleteMode.value = null }
}
</script>

<style lang="scss" scoped>
.admin-users__toolbar {
	display: flex;
	gap: 0.5rem;
	margin-block-end: 1rem;
}

.admin-users__meta {
	display: grid;
	grid-template-columns: auto 1fr;
	column-gap: 1rem;
	row-gap: 0.25rem;
	margin-block-end: 1rem;

	dt {
		font-weight: 600;
		color: var(--grey-700);
	}

	dd {
		margin: 0;
	}
}

.admin-users__password-actions {
	display: flex;
	gap: 0.5rem;
	margin-block-end: 1rem;
}

.admin-users__issuer-url {
	margin-inline-start: 0.35rem;
	color: var(--grey-600);
	font-size: 0.85rem;
	word-break: break-all;
}

.admin-users__issuer-url-value,
.admin-users__subject {
	font-family: monospace;
	font-size: 0.85rem;
	word-break: break-all;
}
</style>
