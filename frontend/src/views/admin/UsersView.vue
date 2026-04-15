<template>
	<div class="admin-users">
		<div class="admin-users__toolbar">
			<input
				v-model="searchTerm"
				class="input"
				type="text"
				:placeholder="$t('admin.users.searchPlaceholder')"
				@input="onSearch"
			/>
		</div>

		<p v-if="loading">{{ $t('misc.loading') }}</p>
		<p v-else-if="error" class="has-text-danger">{{ error }}</p>
		<table v-else class="admin-users__table">
			<thead>
				<tr>
					<th>{{ $t('admin.users.id') }}</th>
					<th>{{ $t('admin.users.username') }}</th>
					<th>{{ $t('admin.users.email') }}</th>
					<th>{{ $t('admin.users.status') }}</th>
					<th>{{ $t('admin.users.admin') }}</th>
					<th></th>
				</tr>
			</thead>
			<tbody>
				<tr v-for="u in users" :key="u.id">
					<td>{{ u.id }}</td>
					<td>{{ u.username }}</td>
					<td>{{ u.email }}</td>
					<td>{{ u.status }}</td>
					<td>
						<input
							type="checkbox"
							:checked="u.isAdmin"
							:disabled="togglingId === u.id"
							@change="toggleAdmin(u)"
						/>
					</td>
					<td></td>
				</tr>
			</tbody>
		</table>
	</div>
</template>

<script setup lang="ts">
import {ref, onMounted} from 'vue'
import {listAdminUsers, setAdmin, type AdminUser} from '@/services/admin/userService'

const users = ref<AdminUser[]>([])
const loading = ref(false)
const error = ref('')
const searchTerm = ref('')
const togglingId = ref<number | null>(null)
let searchTimer: ReturnType<typeof setTimeout> | null = null

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
	togglingId.value = u.id
	try {
		const updated = await setAdmin(u.id, !u.isAdmin)
		const idx = users.value.findIndex(x => x.id === u.id)
		if (idx !== -1) users.value[idx] = updated
	} catch (e) {
		error.value = e instanceof Error ? e.message : String(e)
	} finally {
		togglingId.value = null
	}
}

onMounted(load)
</script>

<style lang="scss" scoped>
.admin-users__toolbar {
	margin-bottom: 1rem;
}

.admin-users__table {
	width: 100%;
	border-collapse: collapse;

	th, td {
		padding: 0.5rem 0.75rem;
		text-align: start;
		border-bottom: 1px solid var(--grey-200);
	}

	th {
		font-weight: 600;
		font-size: 0.85rem;
		text-transform: uppercase;
		color: var(--grey-600);
	}
}
</style>
