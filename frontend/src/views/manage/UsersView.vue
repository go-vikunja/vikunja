<template>
	<Card>
		<div class="manage-users">
			<div class="manage-users__toolbar">
				<FormInput
					v-model="searchTerm"
					type="text"
					:placeholder="$t('manage.users.searchPlaceholder')"
				/>
				<FormSelect
					v-model="sourceFilter"
					:options="sourceOptions"
				/>
				<FormSelect
					v-model="statusFilter"
					:options="statusFilterOptions"
				/>
				<XButton
					variant="primary"
					@click="openCreate"
				>
					{{ $t('manage.users.add') }}
				</XButton>
			</div>

			<p v-if="isPending">
				{{ $t('misc.loading') }}
			</p>
			<p
				v-else-if="isError"
				class="has-text-danger"
			>
				{{ $t('manage.users.loadError') }}
			</p>
			<template v-else>
				<table class="table has-actions is-striped is-hoverable is-fullwidth">
					<thead>
						<tr>
							<th>{{ $t('user.auth.username') }}</th>
							<th>{{ $t('manage.users.name') }}</th>
							<th>{{ $t('user.auth.email') }}</th>
							<th>{{ $t('manage.users.jobTitle') }}</th>
							<th>{{ $t('manage.users.department') }}</th>
							<th>{{ $t('manage.users.source') }}</th>
							<th>{{ $t('manage.users.status') }}</th>
							<th />
						</tr>
					</thead>
					<tbody>
						<tr
							v-for="u in users"
							:key="u.id"
						>
							<td>
								{{ u.username }}
								<span
									v-if="u.is_admin"
									class="tag is-info"
								>{{ $t('manage.users.admin') }}</span>
							</td>
							<td>{{ u.name }}</td>
							<td>{{ u.email }}</td>
							<td>{{ u.job_title }}</td>
							<td>{{ u.department }}</td>
							<td>{{ sourceLabel(u.source) }}</td>
							<td>
								{{ statusLabel(u.status) }}
								<span
									v-if="u.must_change_password"
									class="tag is-warning"
								>{{ $t('manage.users.mustChangePassword') }}</span>
							</td>
							<td class="actions">
								<XButton
									variant="secondary"
									@click="openEdit(u)"
								>
									{{ $t('manage.users.manage') }}
								</XButton>
							</td>
						</tr>
						<tr v-if="users.length === 0">
							<td colspan="8">
								{{ $t('manage.users.empty') }}
							</td>
						</tr>
					</tbody>
				</table>
				<PaginationEmit
					v-if="totalPages > 1"
					:total-pages="totalPages"
					:current-page="page"
					@pageChanged="p => page = p"
				/>
			</template>
		</div>
	</Card>

	<Modal
		v-if="editTarget"
		variant="hint-modal"
		@close="closeEdit"
	>
		<Card
			class="has-no-shadow"
			:title="$t('manage.users.editTitle', {username: editTarget.username})"
		>
			<Message
				v-if="!editTarget.profile_editable"
				class="mbe-4"
			>
				{{ $t('manage.users.notEditable', {source: sourceLabel(editTarget.source)}) }}
			</Message>

			<FormField :label="$t('manage.users.name')">
				<template #default="{id}">
					<FormInput
						:id="id"
						v-model="form.name"
						:disabled="!editTarget.profile_editable"
						type="text"
					/>
				</template>
			</FormField>
			<FormField :label="$t('user.auth.email')">
				<template #default="{id}">
					<FormInput
						:id="id"
						v-model="form.email"
						:disabled="!editTarget.profile_editable"
						type="email"
					/>
				</template>
			</FormField>
			<FormField :label="$t('user.settings.general.language')">
				<template #default="{id}">
					<FormInput
						:id="id"
						v-model="form.language"
						:disabled="!editTarget.profile_editable"
						type="text"
					/>
				</template>
			</FormField>
			<FormField :label="$t('manage.users.jobTitle')">
				<template #default="{id}">
					<FormInput
						:id="id"
						:model-value="editTarget.job_title"
						disabled
						readonly
						type="text"
					/>
				</template>
			</FormField>
			<FormField :label="$t('manage.users.department')">
				<template #default="{id}">
					<FormInput
						:id="id"
						:model-value="editTarget.department"
						disabled
						readonly
						type="text"
					/>
				</template>
			</FormField>
			<p class="help mbe-4">
				{{ $t('manage.users.titleDepartmentHelp') }}
			</p>

			<hr>

			<FormCheckbox
				v-model="form.isAdmin"
				:disabled="editTarget.id === currentUserId"
				:label="$t('manage.users.isAdminLabel')"
			/>
			<FormField :label="$t('manage.users.status')">
				<template #default="{id}">
					<FormSelect
						:id="id"
						v-model.number="form.status"
						:options="statusOptions"
					/>
				</template>
			</FormField>

			<template v-if="editTarget.profile_editable">
				<FormField :label="$t('manage.users.newPassword')">
					<template #default="{id}">
						<FormInput
							:id="id"
							v-model="newPassword"
							type="password"
							autocomplete="new-password"
						/>
					</template>
				</FormField>
				<FormCheckbox
					v-model="requireChange"
					:label="$t('manage.users.requireChange')"
				/>
				<XButton
					variant="secondary"
					class="mbe-4"
					:disabled="!newPassword || passwordMutation.isPending.value"
					:loading="passwordMutation.isPending.value"
					@click="setPassword"
				>
					{{ $t('manage.users.setPassword') }}
				</XButton>
			</template>

			<template #footer>
				<XButton
					variant="tertiary"
					@click="closeEdit"
				>
					{{ $t('misc.cancel') }}
				</XButton>
				<XButton
					variant="secondary"
					@click="openTransfer"
				>
					{{ $t('manage.users.transferProjects') }}
				</XButton>
				<XButton
					v-if="editTarget.id !== currentUserId"
					variant="secondary"
					:danger="true"
					@click="pendingDelete = editTarget"
				>
					{{ $t('misc.delete') }}
				</XButton>
				<XButton
					variant="primary"
					:disabled="!hasChanges || saving"
					:loading="saving"
					@click="save"
				>
					{{ $t('misc.save') }}
				</XButton>
			</template>
		</Card>
	</Modal>

	<Modal
		v-if="createOpen"
		variant="hint-modal"
		@close="createOpen = false"
	>
		<Card
			class="has-no-shadow"
			:title="$t('manage.users.createTitle')"
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
			<FormField :label="$t('manage.users.name')">
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
			<FormCheckbox
				v-model="createForm.requireChange"
				:label="$t('manage.users.requireChange')"
			/>
			<FormCheckbox
				v-model="createForm.isAdmin"
				:label="$t('manage.users.isAdminLabel')"
			/>

			<template #footer>
				<XButton
					variant="tertiary"
					@click="createOpen = false"
				>
					{{ $t('misc.cancel') }}
				</XButton>
				<XButton
					variant="primary"
					:disabled="createMutation.isPending.value || !createForm.username || !createForm.email || !createForm.password"
					:loading="createMutation.isPending.value"
					@click="create"
				>
					{{ $t('manage.users.createSubmit') }}
				</XButton>
			</template>
		</Card>
	</Modal>

	<Modal
		v-if="pendingDelete"
		variant="hint-modal"
		@close="pendingDelete = null"
	>
		<Card
			class="has-no-shadow"
			:title="$t('manage.users.deleteTitle')"
		>
			<p>{{ $t('manage.users.deleteIntro', {username: pendingDelete.username}) }}</p>
			<Message
				class="mbs-4"
				variant="warning"
			>
				{{ $t('manage.users.deleteTransferHint') }}
			</Message>
			<template #footer>
				<XButton
					variant="tertiary"
					@click="pendingDelete = null"
				>
					{{ $t('misc.cancel') }}
				</XButton>
				<XButton
					variant="secondary"
					:disabled="deleteMutation.isPending.value"
					@click="remove('scheduled')"
				>
					{{ $t('manage.users.deleteScheduled') }}
				</XButton>
				<XButton
					variant="primary"
					:danger="true"
					:disabled="deleteMutation.isPending.value"
					@click="remove('now')"
				>
					{{ $t('manage.users.deleteNow') }}
				</XButton>
			</template>
		</Card>
	</Modal>

	<Modal
		v-if="transferOpen && editTarget"
		variant="hint-modal"
		@close="transferOpen = false"
	>
		<Card
			class="has-no-shadow"
			:title="$t('manage.users.transferTitle', {username: editTarget.username})"
		>
			<p
				v-if="ownedProjects.isPending.value"
			>
				{{ $t('misc.loading') }}
			</p>
			<template v-else>
				<p v-if="(ownedProjects.data.value ?? []).length === 0">
					{{ $t('manage.users.transferNone') }}
				</p>
				<template v-else>
					<p>{{ $t('manage.users.transferIntro', {count: ownedProjects.data.value?.length ?? 0}) }}</p>
					<ul class="manage-users__projects">
						<li
							v-for="p in ownedProjects.data.value"
							:key="p.id"
						>
							<FormCheckbox
								v-model="selectedProjects[p.id]"
								:label="p.title + (p.is_archived ? ' (' + $t('manage.users.archived') + ')' : '')"
							/>
						</li>
					</ul>
					<FormField :label="$t('manage.users.newOwner')">
						<template #default="{id}">
							<FormSelect
								:id="id"
								v-model.number="newOwnerId"
								:options="ownerOptions"
							/>
						</template>
					</FormField>
					<p class="help">
						{{ $t('manage.users.transferNote') }}
					</p>
				</template>
			</template>
			<template #footer>
				<XButton
					variant="tertiary"
					@click="transferOpen = false"
				>
					{{ $t('misc.cancel') }}
				</XButton>
				<XButton
					variant="primary"
					:disabled="!newOwnerId || chosenProjectIds.length === 0 || transferMutation.isPending.value"
					:loading="transferMutation.isPending.value"
					@click="transfer"
				>
					{{ $t('manage.users.transfer') }}
				</XButton>
			</template>
		</Card>
	</Modal>
</template>

<script setup lang="ts">
import {computed, reactive, ref, watch} from 'vue'
import {useDebounceFn} from '@vueuse/core'
import {useI18n} from 'vue-i18n'
import {useMutation, useQuery} from '@tanstack/vue-query'

import Card from '@/components/misc/Card.vue'
import Modal from '@/components/misc/Modal.vue'
import Message from '@/components/misc/Message.vue'
import PaginationEmit from '@/components/misc/PaginationEmit.vue'
import XButton from '@/components/input/Button.vue'
import FormField from '@/components/input/FormField.vue'
import FormInput from '@/components/input/FormInput.vue'
import FormSelect from '@/components/input/FormSelect.vue'
import FormCheckbox from '@/components/input/FormCheckbox.vue'

import {
	createManagedUserMutation,
	deleteManagedUserMutation,
	manageUserProjectsQuery,
	manageUsersQuery,
	setManagedUserAdminMutation,
	setManagedUserPasswordMutation,
	setManagedUserStatusMutation,
	transferProjectsMutation,
	updateManagedUserMutation,
} from '@/client/queries/manage'
import type {ManagedUser, ManageUserSource} from '@/client/queries/manage'
import {error, success} from '@/message'
import {useAuthStore} from '@/stores/auth'
import {useTitle} from '@/composables/useTitle'

const {t} = useI18n({useScope: 'global'})
useTitle(() => t('manage.users.title'))

const authStore = useAuthStore()
const currentUserId = computed(() => authStore.info?.id)

// ---- list ----------------------------------------------------------------------------------
const searchTerm = ref('')
const debouncedSearch = ref('')
const sourceFilter = ref<ManageUserSource | ''>('')
const statusFilter = ref<number | -1>(-1)
const page = ref(1)

const applySearch = useDebounceFn((value: string) => {
	debouncedSearch.value = value
}, 300)
watch(searchTerm, value => applySearch(value))
// A narrower filter must not strand the view on a page that no longer exists.
watch([debouncedSearch, sourceFilter, statusFilter], () => {
	page.value = 1
})

const query = useQuery(computed(() => manageUsersQuery({
	page: page.value,
	q: debouncedSearch.value || undefined,
	status: statusFilter.value === -1 ? undefined : statusFilter.value,
	source: sourceFilter.value || undefined,
})))
const users = computed(() => query.data.value?.items ?? [])
const totalPages = computed(() => query.data.value?.total_pages ?? 1)
const isPending = query.isPending
const isError = query.isError

function sourceLabel(source: ManageUserSource): string {
	return t(`manage.users.sources.${source}`)
}

function statusLabel(status: number): string {
	switch (status) {
		case 0: return t('admin.users.statusActive')
		case 1: return t('admin.users.statusEmailConfirmation')
		case 2: return t('admin.users.statusDisabled')
		case 3: return t('admin.users.statusLocked')
		default: return String(status)
	}
}

const sourceOptions = computed(() => [
	{value: '', label: t('manage.users.allSources')},
	...(['local', 'entra', 'import', 'ldap', 'other'] as ManageUserSource[]).map(s => ({value: s, label: sourceLabel(s)})),
])
const statusOptions = computed(() => [0, 1, 2, 3].map(s => ({value: s, label: statusLabel(s)})))
const statusFilterOptions = computed(() => [
	{value: -1, label: t('manage.users.allStatuses')},
	...statusOptions.value,
])

// ---- edit ----------------------------------------------------------------------------------
const editTarget = ref<ManagedUser | null>(null)
const form = reactive({name: '', email: '', language: '', isAdmin: false, status: 0})
const newPassword = ref('')
const requireChange = ref(true)
const saving = ref(false)

const updateMutation = useMutation(updateManagedUserMutation())
const adminMutation = useMutation(setManagedUserAdminMutation())
const statusMutation = useMutation(setManagedUserStatusMutation())
const passwordMutation = useMutation(setManagedUserPasswordMutation())

function openEdit(u: ManagedUser) {
	editTarget.value = u
	form.name = u.name
	form.email = u.email
	form.language = u.language
	form.isAdmin = u.is_admin
	form.status = u.status
	newPassword.value = ''
	requireChange.value = true
}

function closeEdit() {
	editTarget.value = null
}

const hasChanges = computed(() => {
	const u = editTarget.value
	if (!u) return false
	const profileChanged = u.profile_editable && (form.name !== u.name || form.email !== u.email || form.language !== u.language)
	return profileChanged || form.isAdmin !== u.is_admin || form.status !== u.status
})

async function save() {
	const u = editTarget.value
	if (!u) return
	saving.value = true
	try {
		if (u.profile_editable && (form.name !== u.name || form.email !== u.email || form.language !== u.language)) {
			await updateMutation.mutateAsync({id: u.id, name: form.name, email: form.email, language: form.language || undefined})
		}
		if (form.isAdmin !== u.is_admin) {
			await adminMutation.mutateAsync({id: u.id, is_admin: form.isAdmin})
		}
		if (form.status !== u.status) {
			await statusMutation.mutateAsync({id: u.id, status: form.status})
		}
		success({message: t('manage.users.saved', {username: u.username})})
		closeEdit()
	} catch (e) {
		error(e)
	} finally {
		saving.value = false
	}
}

async function setPassword() {
	const u = editTarget.value
	if (!u || !newPassword.value) return
	try {
		await passwordMutation.mutateAsync({id: u.id, new_password: newPassword.value, require_change: requireChange.value})
		success({message: t('manage.users.passwordSet', {username: u.username})})
		newPassword.value = ''
	} catch (e) {
		error(e)
	}
}

// ---- create --------------------------------------------------------------------------------
const createOpen = ref(false)
const createForm = reactive({username: '', email: '', name: '', password: '', requireChange: true, isAdmin: false})
const createMutation = useMutation(createManagedUserMutation())

function openCreate() {
	Object.assign(createForm, {username: '', email: '', name: '', password: '', requireChange: true, isAdmin: false})
	createOpen.value = true
}

async function create() {
	try {
		const created = await createMutation.mutateAsync({
			username: createForm.username,
			email: createForm.email,
			password: createForm.password,
			...(createForm.name ? {name: createForm.name} : {}),
			...(createForm.isAdmin ? {is_admin: true} : {}),
			skip_email_confirm: true,
			require_change: createForm.requireChange,
		})
		success({message: t('manage.users.created', {username: created.username})})
		createOpen.value = false
	} catch (e) {
		error(e)
	}
}

// ---- delete --------------------------------------------------------------------------------
const pendingDelete = ref<ManagedUser | null>(null)
const deleteMutation = useMutation(deleteManagedUserMutation())

async function remove(mode: 'now' | 'scheduled') {
	const u = pendingDelete.value
	if (!u) return
	try {
		await deleteMutation.mutateAsync({id: u.id, mode})
		success({message: t(mode === 'now' ? 'manage.users.deleted' : 'manage.users.deleteScheduledDone', {username: u.username})})
		pendingDelete.value = null
		closeEdit()
	} catch (e) {
		error(e)
	}
}

// ---- transfer projects ---------------------------------------------------------------------
const transferOpen = ref(false)
const newOwnerId = ref<number | null>(null)
const selectedProjects = reactive<Record<number, boolean>>({})
const transferMutation = useMutation(transferProjectsMutation())

const ownedProjects = useQuery(computed(() => ({
	...manageUserProjectsQuery(editTarget.value?.id ?? 0),
	enabled: transferOpen.value && !!editTarget.value,
})))

// Everything is selected by default: the usual case is handing a leaver's work to one person.
watch(() => ownedProjects.data.value, projects => {
	for (const key of Object.keys(selectedProjects)) delete selectedProjects[Number(key)]
	for (const p of projects ?? []) selectedProjects[p.id] = true
})

const chosenProjectIds = computed(() => Object.entries(selectedProjects).filter(([, on]) => on).map(([id]) => Number(id)))

// Only active people with a profile of their own can own projects.
const ownerOptions = computed(() => [
	{value: 0, label: t('manage.users.chooseOwner')},
	...users.value
		.filter(u => u.status === 0 && u.id !== editTarget.value?.id)
		.map(u => ({value: u.id, label: `${u.name || u.username} (${u.username})`})),
])

function openTransfer() {
	newOwnerId.value = 0
	transferOpen.value = true
}

async function transfer() {
	const u = editTarget.value
	if (!u || !newOwnerId.value) return
	try {
		const {transferred} = await transferMutation.mutateAsync({
			id: u.id,
			new_owner_id: newOwnerId.value,
			project_ids: chosenProjectIds.value,
		})
		success({message: t('manage.users.transferred', {count: transferred})})
		transferOpen.value = false
	} catch (e) {
		error(e)
	}
}
</script>

<style lang="scss" scoped>
.manage-users__toolbar {
	display: flex;
	gap: 0.5rem;
	margin-block-end: 1rem;
	flex-wrap: wrap;
}

.manage-users__projects {
	list-style: none;
	margin: 0 0 1rem;
	padding: 0;
	max-block-size: 14rem;
	overflow-y: auto;
}
</style>
