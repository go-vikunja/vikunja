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
							<td>{{ u.authProvider || $t('admin.users.issuerLocal') }}</td>
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
							{{ detailTarget.authProvider || $t('admin.users.issuerLocal') }}
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
						v-model="editable.isAdmin"
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
						v-model="createForm.isAdmin"
						:label="$t('admin.users.isAdminLabel')"
					/>
					<FormCheckbox
						v-model="createForm.skipEmailConfirm"
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
import {ref, computed, onMounted, reactive, watch} from 'vue'
import {useDebounceFn} from '@vueuse/core'
import {useI18n} from 'vue-i18n'
import {useAuthStore} from '@/stores/auth'
import AdminUserService, {type CreateAdminUserBody, type DeleteUserMode} from '@/services/admin/userService'
import AdminUserModel from '@/models/adminUser'
import type {IAdminUser} from '@/modelTypes/IAdminUser'
import {error, success} from '@/message'
import Card from '@/components/misc/Card.vue'
import Modal from '@/components/misc/Modal.vue'
import PaginationEmit from '@/components/misc/PaginationEmit.vue'
import XButton from '@/components/input/Button.vue'
import FormField from '@/components/input/FormField.vue'
import FormInput from '@/components/input/FormInput.vue'
import FormSelect from '@/components/input/FormSelect.vue'
import FormCheckbox from '@/components/input/FormCheckbox.vue'
import TimeDisplay from '@/components/misc/TimeDisplay.vue'

const {t} = useI18n({useScope: 'global'})
const authStore = useAuthStore()
const currentUserId = computed(() => authStore.info?.id)

const adminUserService = new AdminUserService()

const users = ref<IAdminUser[]>([])
const loading = ref(false)
const searchTerm = ref('')
const currentPage = ref(1)
const totalPages = ref(1)
const detailTarget = ref<IAdminUser | null>(null)
const pendingDelete = ref<IAdminUser | null>(null)
const saving = ref(false)
const deleting = ref(false)
const deleteMode = ref<DeleteUserMode | null>(null)
const createOpen = ref(false)
const creating = ref(false)
const editable = reactive({isAdmin: false, status: 0})

function emptyCreateForm(): Required<Pick<CreateAdminUserBody, 'username' | 'email'>> & CreateAdminUserBody {
	return {
		username: '',
		email: '',
		name: '',
		password: '',
		language: '',
		isAdmin: false,
		skipEmailConfirm: false,
	}
}

const createForm = reactive(emptyCreateForm())

const hasChanges = computed(() => {
	if (!detailTarget.value) return false
	return editable.isAdmin !== !!detailTarget.value.isAdmin
		|| editable.status !== detailTarget.value.status
})

watch(detailTarget, (u) => {
	if (!u) return
	editable.isAdmin = !!u.isAdmin
	editable.status = u.status
})

function statusLabel(status: number): string {
	switch (status) {
		case 0: return t('admin.users.statusActive')
		case 1: return t('admin.users.statusEmailConfirmation')
		case 2: return t('admin.users.statusDisabled')
		case 3: return t('admin.users.statusLocked')
		default: return String(status)
	}
}

const statusOptions = computed(() => [
	{value: 0, label: t('admin.users.statusActive')},
	{value: 1, label: t('admin.users.statusEmailConfirmation')},
	{value: 2, label: t('admin.users.statusDisabled')},
	{value: 3, label: t('admin.users.statusLocked')},
])

async function load() {
	loading.value = true
	try {
		const params = searchTerm.value ? {s: searchTerm.value} : {}
		users.value = await adminUserService.getAll(new AdminUserModel(), params, currentPage.value)
		totalPages.value = adminUserService.totalPages || 1
	} catch (e) {
		error(e)
	} finally {
		loading.value = false
	}
}

function goToPage(page: number) {
	currentPage.value = page
	load()
}

const onSearch = useDebounceFn(() => {
	// Reset to page 1 so a narrower search doesn't strand the UI on an empty page.
	currentPage.value = 1
	load()
}, 300)

function openDetails(u: IAdminUser) {
	detailTarget.value = u
}

function closeDetail() {
	detailTarget.value = null
}

function openCreate() {
	Object.assign(createForm, emptyCreateForm())
	createOpen.value = true
}

function closeCreate() {
	createOpen.value = false
}

async function submitCreate() {
	creating.value = true
	try {
		const body: CreateAdminUserBody = {
			username: createForm.username,
			email: createForm.email,
			password: createForm.password,
		}
		if (createForm.name) body.name = createForm.name
		if (createForm.language) body.language = createForm.language
		if (createForm.isAdmin) body.isAdmin = true
		if (createForm.skipEmailConfirm) body.skipEmailConfirm = true
		const created = await adminUserService.createUser(body)
		users.value = [created, ...users.value]
		success({message: t('admin.users.createdSuccess', {username: created.username})})
		createOpen.value = false
	} catch (e) {
		error(e)
	} finally {
		creating.value = false
	}
}

function replaceUser(updated: IAdminUser) {
	const idx = users.value.findIndex(x => x.id === updated.id)
	if (idx !== -1) users.value[idx] = updated
}

async function saveChanges() {
	if (!detailTarget.value) return
	const target = detailTarget.value
	saving.value = true
	try {
		let latest: IAdminUser = target
		if (editable.isAdmin !== !!target.isAdmin) {
			latest = await adminUserService.setAdmin(target.id, editable.isAdmin)
		}
		if (editable.status !== target.status) {
			latest = await adminUserService.setStatus(target.id, editable.status)
		}
		replaceUser(latest)
		success({message: t('admin.users.updatedSuccess', {username: latest.username})})
		detailTarget.value = null
	} catch (e) {
		error(e)
	} finally {
		saving.value = false
	}
}

function cancelDelete() {
	if (deleting.value) return
	pendingDelete.value = null
	deleteMode.value = null
}

async function doDelete(mode: DeleteUserMode) {
	if (!pendingDelete.value || deleting.value) return
	const target = pendingDelete.value
	deleting.value = true
	deleteMode.value = mode
	try {
		await adminUserService.deleteUser(target.id, mode)
		if (mode === 'now') {
			users.value = users.value.filter(x => x.id !== target.id)
			success({message: t('admin.users.deletedSuccess', {username: target.username})})
		} else {
			success({message: t('admin.users.deleteScheduledSuccess', {username: target.username})})
		}
		pendingDelete.value = null
		detailTarget.value = null
	} catch (e) {
		error(e)
	} finally {
		deleting.value = false
		deleteMode.value = null
	}
}

onMounted(load)
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
