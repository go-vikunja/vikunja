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
			</div>

			<p v-if="loading">
				{{ $t('misc.loading') }}
			</p>
			<p
				v-else-if="error"
				class="has-text-danger"
			>
				{{ error }}
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
						<th>{{ $t('admin.users.status') }}</th>
						<th>{{ $t('admin.users.admin') }}</th>
						<th>{{ $t('admin.users.actions') }}</th>
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
						<td>{{ statusLabel(u.status) }}</td>
						<td>
							<input
								type="checkbox"
								:checked="u.isAdmin"
								:disabled="busyId === u.id"
								@change="toggleAdmin(u)"
							>
						</td>
						<td class="admin-users__actions">
							<button
								class="button is-small"
								:disabled="busyId === u.id"
								@click="toggleStatus(u)"
							>
								{{ u.status === STATUS_DISABLED ? $t('admin.users.enable') : $t('admin.users.disable') }}
							</button>
							<button
								class="button is-small is-danger"
								:disabled="busyId === u.id || u.id === currentUserId"
								@click="confirmDelete(u)"
							>
								{{ $t('admin.users.delete') }}
							</button>
						</td>
					</tr>
				</tbody>
			</table>

			<Modal
				v-if="pendingDelete"
				@close="pendingDelete = null"
			>
				<template #header>
					<h3>{{ $t('admin.users.confirmDeleteTitle') }}</h3>
				</template>
				<template #text>
					<p>
						{{ $t('admin.users.confirmDeleteBody', {username: pendingDelete.username}) }}
					</p>
				</template>
				<template #footer>
					<XButton
						variant="tertiary"
						@click="pendingDelete = null"
					>
						{{ $t('misc.cancel') }}
					</XButton>
					<XButton
						variant="primary"
						@click="doDelete()"
					>
						{{ $t('admin.users.delete') }}
					</XButton>
				</template>
			</Modal>
		</div>
	</Card>
</template>

<script setup lang="ts">
import {ref, onMounted, computed} from 'vue'
import {useAuthStore} from '@/stores/auth'
import {listAdminUsers, setAdmin, setStatus, deleteUser, type AdminUser} from '@/services/admin/userService'
import Card from '@/components/misc/Card.vue'
import Modal from '@/components/misc/Modal.vue'
import XButton from '@/components/input/Button.vue'

const STATUS_ACTIVE = 0
const STATUS_DISABLED = 2

const authStore = useAuthStore()
const currentUserId = computed(() => authStore.info?.id)

const users = ref<AdminUser[]>([])
const loading = ref(false)
const error = ref('')
const searchTerm = ref('')
const busyId = ref<number | null>(null)
const pendingDelete = ref<AdminUser | null>(null)
let searchTimer: ReturnType<typeof setTimeout> | null = null

function statusLabel(status: number): string {
	switch (status) {
		case 0: return 'Active'
		case 1: return 'Email confirmation required'
		case 2: return 'Disabled'
		case 3: return 'Account locked'
		default: return String(status)
	}
}

async function load() {
	loading.value = true
	error.value = ''
	try {
		users.value = await listAdminUsers({s: searchTerm.value || undefined})
	} catch (e) {
		error.value = e instanceof Error ? e.message : String(e)
	} finally {
		loading.value = false
	}
}

function onSearch() {
	if (searchTimer) clearTimeout(searchTimer)
	searchTimer = setTimeout(load, 300)
}

async function toggleAdmin(u: AdminUser) {
	busyId.value = u.id
	try {
		const updated = await setAdmin(u.id, !u.isAdmin)
		const idx = users.value.findIndex(x => x.id === u.id)
		if (idx !== -1) users.value[idx] = updated
	} catch (e) {
		error.value = e instanceof Error ? e.message : String(e)
	} finally {
		busyId.value = null
	}
}

async function toggleStatus(u: AdminUser) {
	busyId.value = u.id
	try {
		const newStatus = u.status === STATUS_DISABLED ? STATUS_ACTIVE : STATUS_DISABLED
		const updated = await setStatus(u.id, newStatus)
		const idx = users.value.findIndex(x => x.id === u.id)
		if (idx !== -1) users.value[idx] = updated
	} catch (e) {
		error.value = e instanceof Error ? e.message : String(e)
	} finally {
		busyId.value = null
	}
}

function confirmDelete(u: AdminUser) {
	pendingDelete.value = u
}

async function doDelete() {
	if (!pendingDelete.value) return
	const target = pendingDelete.value
	pendingDelete.value = null
	busyId.value = target.id
	try {
		await deleteUser(target.id)
		users.value = users.value.filter(x => x.id !== target.id)
	} catch (e) {
		error.value = e instanceof Error ? e.message : String(e)
	} finally {
		busyId.value = null
	}
}

onMounted(load)
</script>

<style lang="scss" scoped>
.admin-users__toolbar {
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
}
</style>
