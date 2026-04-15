<template>
	<div class="admin-projects">
		<p v-if="loading">{{ $t('misc.loading') }}</p>
		<p v-else-if="error" class="has-text-danger">{{ error }}</p>
		<table v-else class="admin-projects__table">
			<thead>
				<tr>
					<th>{{ $t('admin.projects.id') }}</th>
					<th>{{ $t('admin.projects.title') }}</th>
					<th>{{ $t('admin.projects.owner') }}</th>
					<th>{{ $t('admin.projects.created') }}</th>
					<th></th>
				</tr>
			</thead>
			<tbody>
				<tr v-for="p in projects" :key="p.id">
					<td>{{ p.id }}</td>
					<td>{{ p.title }}</td>
					<td>{{ p.owner?.username ?? p.owner?.id }}</td>
					<td>{{ p.created }}</td>
					<td>
						<button class="button is-small" @click="promptReassign(p)">
							{{ $t('admin.projects.reassignOwner') }}
						</button>
					</td>
				</tr>
			</tbody>
		</table>
	</div>
</template>

<script setup lang="ts">
import {ref, onMounted} from 'vue'
import type {IProject} from '@/modelTypes/IProject'
import {listAdminProjects, reassignProjectOwner} from '@/services/admin/projectService'

const projects = ref<IProject[]>([])
const loading = ref(false)
const error = ref('')

async function load() {
	loading.value = true
	error.value = ''
	try {
		projects.value = await listAdminProjects()
	} catch (e) {
		error.value = e instanceof Error ? e.message : String(e)
	} finally {
		loading.value = false
	}
}

async function promptReassign(p: IProject) {
	const input = window.prompt(`New owner ID for "${p.title}"?`)
	if (!input) return
	const newId = parseInt(input, 10)
	if (!Number.isFinite(newId) || newId < 1) return
	try {
		const updated = await reassignProjectOwner(p.id, newId)
		const idx = projects.value.findIndex(x => x.id === p.id)
		if (idx !== -1) projects.value[idx] = updated
	} catch (e) {
		error.value = e instanceof Error ? e.message : String(e)
	}
}

onMounted(load)
</script>

<style lang="scss" scoped>
.admin-projects__table {
	width: 100%;
	border-collapse: collapse;

	th, td {
		padding: 0.5rem 0.75rem;
		text-align: start;
		border-bottom: 1px solid var(--grey-200);
	}
}
</style>
