<template>
	<Card :title="$t('admin.users.title')">
		<div class="admin-users">
			<div class="admin-users__toolbar">
				<input
					v-model="searchTerm"
					class="input"
					type="text"
					:placeholder="$t('admin.users.searchPlaceholder')"
					@input="onSearch"
				>
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
			<table
				v-else
				class="admin-users__table"
			>
				<thead>
					<tr>
						<th>{{ $t('admin.users.id') }}</th>
						<th>{{ $t('admin.users.username') }}</th>
						<th>{{ $t('admin.users.email') }}</th>
						<th>{{ $t('admin.users.issuer') }}</th>
						<th>{{ $t('admin.users.status') }}</th>
						<th>{{ $t('admin.users.created') }}</th>
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
						<td>{{ issuerLabel(u.issuer) }}</td>
						<td>{{ statusLabel(u.status) }}</td>
						<td>
							<time :datetime="formatISO(u.created)">{{ formatDisplayDate(u.created) }}</time>
						</td>
						<td class="admin-users__actions">
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
						<dt>{{ $t('admin.users.id') }}</dt>
						<dd>{{ detailTarget.id }}</dd>
						<dt>{{ $t('admin.users.emailLabel') }}</dt>
						<dd>{{ detailTarget.email }}</dd>
						<dt>{{ $t('admin.users.issuer') }}</dt>
						<dd>{{ issuerLabel(detailTarget.issuer) }}</dd>
						<dt>{{ $t('admin.users.createdLabel') }}</dt>
						<dd>
							<time :datetime="formatISO(detailTarget.created)">{{ formatDisplayDate(detailTarget.created) }}</time>
						</dd>
						<dt>{{ $t('admin.users.updatedLabel') }}</dt>
						<dd>
							<time :datetime="formatISO(detailTarget.updated)">{{ formatDisplayDate(detailTarget.updated) }}</time>
						</dd>
					</dl>

					<div class="field">
						<label class="checkbox">
							<input
								v-model="editable.isAdmin"
								type="checkbox"
							>
							{{ $t('admin.users.isAdminLabel') }}
						</label>
					</div>

					<div class="field">
						<label
							class="label"
							for="admin-user-status"
						>{{ $t('admin.users.statusLabel') }}</label>
						<div class="select">
							<select
								id="admin-user-status"
								v-model.number="editable.status"
							>
								<option :value="0">
									{{ $t('admin.users.statusActive') }}
								</option>
								<option :value="1">
									{{ $t('admin.users.statusEmailConfirmation') }}
								</option>
								<option :value="2">
									{{ $t('admin.users.statusDisabled') }}
								</option>
								<option :value="3">
									{{ $t('admin.users.statusLocked') }}
								</option>
							</select>
						</div>
					</div>

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
							{{ $t('admin.users.delete') }}
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
					<div class="field">
						<label
							class="label"
							for="admin-create-username"
						>{{ $t('admin.users.usernameLabel') }}</label>
						<input
							id="admin-create-username"
							v-model="createForm.username"
							class="input"
							type="text"
							required
						>
					</div>
					<div class="field">
						<label
							class="label"
							for="admin-create-email"
						>{{ $t('admin.users.emailLabel') }}</label>
						<input
							id="admin-create-email"
							v-model="createForm.email"
							class="input"
							type="email"
							required
						>
					</div>
					<div class="field">
						<label
							class="label"
							for="admin-create-name"
						>{{ $t('admin.users.nameLabel') }}</label>
						<input
							id="admin-create-name"
							v-model="createForm.name"
							class="input"
							type="text"
						>
					</div>
					<div class="field">
						<label
							class="label"
							for="admin-create-password"
						>{{ $t('admin.users.passwordLabel') }}</label>
						<input
							id="admin-create-password"
							v-model="createForm.password"
							class="input"
							type="password"
							autocomplete="new-password"
						>
						<p class="help">
							{{ $t('admin.users.passwordHelp') }}
						</p>
					</div>
					<div class="field">
						<label
							class="label"
							for="admin-create-language"
						>{{ $t('admin.users.languageLabel') }}</label>
						<input
							id="admin-create-language"
							v-model="createForm.language"
							class="input"
							type="text"
						>
					</div>
					<div class="field">
						<label class="checkbox">
							<input
								v-model="createForm.isAdmin"
								type="checkbox"
							>
							{{ $t('admin.users.isAdminLabel') }}
						</label>
					</div>
					<div class="field">
						<label class="checkbox">
							<input
								v-model="createForm.skipEmailConfirm"
								type="checkbox"
							>
							{{ $t('admin.users.skipEmailConfirm') }}
						</label>
					</div>

					<template #footer>
						<XButton
							variant="tertiary"
							@click="closeCreate"
						>
							{{ $t('misc.cancel') }}
						</XButton>
						<XButton
							variant="primary"
							:disabled="creating || !createForm.username || !createForm.email"
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
				@close="pendingDelete = null"
			>
				<Card
					class="has-no-shadow"
					:title="$t('admin.users.confirmDeleteTitle')"
				>
					<p>
						{{ $t('admin.users.confirmDeleteBody', {username: pendingDelete.username}) }}
					</p>

					<template #footer>
						<XButton
							variant="tertiary"
							@click="pendingDelete = null"
						>
							{{ $t('misc.cancel') }}
						</XButton>
						<XButton
							variant="primary"
							:danger="true"
							:loading="deleting"
							@click="doDelete()"
						>
							{{ $t('admin.users.delete') }}
						</XButton>
					</template>
				</Card>
			</Modal>
		</div>
	</Card>
</template>

<script setup lang="ts">
import {ref, computed, onMounted, reactive, watch} from 'vue'
import {useI18n} from 'vue-i18n'
import {useAuthStore} from '@/stores/auth'
import {listAdminUsers, setAdmin, setStatus, deleteUser, createAdminUser, type AdminUser, type CreateAdminUserBody} from '@/services/admin/userService'
import {error, success} from '@/message'
import {formatDisplayDate, formatISO} from '@/helpers/time/formatDate'
import Card from '@/components/misc/Card.vue'
import Modal from '@/components/misc/Modal.vue'
import XButton from '@/components/input/Button.vue'

const {t} = useI18n({useScope: 'global'})
const authStore = useAuthStore()
const currentUserId = computed(() => authStore.info?.id)

const users = ref<AdminUser[]>([])
const loading = ref(false)
const searchTerm = ref('')
const detailTarget = ref<AdminUser | null>(null)
const pendingDelete = ref<AdminUser | null>(null)
const saving = ref(false)
const deleting = ref(false)
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
let searchTimer: ReturnType<typeof setTimeout> | null = null

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

function issuerLabel(issuer: string): string {
	if (!issuer || issuer === 'local') return t('admin.users.issuerLocal')
	return issuer
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

async function load() {
	loading.value = true
	try {
		users.value = await listAdminUsers({s: searchTerm.value || undefined})
	} catch (e) {
		error(e)
	} finally {
		loading.value = false
	}
}

function onSearch() {
	if (searchTimer) clearTimeout(searchTimer)
	searchTimer = setTimeout(load, 300)
}

function openDetails(u: AdminUser) {
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
		}
		if (createForm.name) body.name = createForm.name
		if (createForm.password) body.password = createForm.password
		if (createForm.language) body.language = createForm.language
		if (createForm.isAdmin) body.isAdmin = true
		if (createForm.skipEmailConfirm) body.skipEmailConfirm = true
		const created = await createAdminUser(body)
		users.value = [created, ...users.value]
		success(t('admin.users.createdSuccess', {username: created.username}))
		createOpen.value = false
	} catch (e) {
		error(e)
	} finally {
		creating.value = false
	}
}

function replaceUser(updated: AdminUser) {
	const idx = users.value.findIndex(x => x.id === updated.id)
	if (idx !== -1) users.value[idx] = updated
}

async function saveChanges() {
	if (!detailTarget.value) return
	const target = detailTarget.value
	saving.value = true
	try {
		let latest: AdminUser = target
		if (editable.isAdmin !== !!target.isAdmin) {
			latest = await setAdmin(target.id, editable.isAdmin)
		}
		if (editable.status !== target.status) {
			latest = await setStatus(target.id, editable.status)
		}
		replaceUser(latest)
		success(t('admin.users.updatedSuccess', {username: latest.username}))
		detailTarget.value = null
	} catch (e) {
		error(e)
	} finally {
		saving.value = false
	}
}

async function doDelete() {
	if (!pendingDelete.value) return
	const target = pendingDelete.value
	deleting.value = true
	try {
		await deleteUser(target.id)
		users.value = users.value.filter(x => x.id !== target.id)
		success(t('admin.users.deletedSuccess', {username: target.username}))
		pendingDelete.value = null
		detailTarget.value = null
	} catch (e) {
		error(e)
	} finally {
		deleting.value = false
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

.admin-users__table {
	inline-size: 100%;
	border-collapse: collapse;

	th, td {
		padding: 0.5rem 0.75rem;
		text-align: start;
		border-block-end: 1px solid var(--grey-200);
	}

	th {
		font-weight: 600;
		font-size: 0.85rem;
		text-transform: uppercase;
		color: var(--grey-600);
	}
}

.admin-users__actions {
	display: flex;
	gap: 0.5rem;
	justify-content: flex-end;
}

.admin-users__detail {
	text-align: start;
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
</style>
