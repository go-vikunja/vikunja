<template>
	<Card :title="$t('admin.projects.title')">
		<div class="admin-projects">
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
				class="admin-projects__table"
			>
				<thead>
					<tr>
						<th>{{ $t('admin.projects.id') }}</th>
						<th>{{ $t('admin.projects.projectTitle') }}</th>
						<th>{{ $t('admin.projects.owner') }}</th>
						<th>{{ $t('admin.projects.created') }}</th>
						<th />
					</tr>
				</thead>
				<tbody>
					<tr
						v-for="p in projects"
						:key="p.id"
					>
						<td>{{ p.id }}</td>
						<td>{{ p.title }}</td>
						<td>{{ p.owner?.username ?? p.owner?.id }}</td>
						<td>{{ p.created }}</td>
						<td>
							<button
								class="button is-small"
								@click="openReassign(p)"
							>
								{{ $t('admin.projects.reassignOwner') }}
							</button>
						</td>
					</tr>
				</tbody>
			</table>

			<Modal
				v-if="reassignTarget"
				@close="reassignTarget = null"
			>
				<template #header>
					<h3>{{ $t('admin.projects.reassignTitle', {title: reassignTarget.title}) }}</h3>
				</template>
				<template #text>
					<div>
						<label for="admin-reassign-input">{{ $t('admin.projects.newOwnerLabel') }}</label>
						<input
							id="admin-reassign-input"
							v-model="newOwnerQuery"
							class="input"
							type="text"
							:placeholder="$t('admin.projects.newOwnerPlaceholder')"
							@input="searchUsers"
						>
						<ul
							v-if="userResults.length"
							class="admin-projects__results"
						>
							<li
								v-for="u in userResults"
								:key="u.id"
								:class="{selected: selectedUserId === u.id}"
								@click="selectedUserId = u.id"
							>
								{{ u.username }} ({{ u.email }})
							</li>
						</ul>
					</div>
				</template>
				<template #footer>
					<XButton
						variant="tertiary"
						@click="reassignTarget = null"
					>
						{{ $t('misc.cancel') }}
					</XButton>
					<XButton
						variant="primary"
						:disabled="!selectedUserId"
						@click="doReassign()"
					>
						{{ $t('admin.projects.reassignOwner') }}
					</XButton>
				</template>
			</Modal>
		</div>
	</Card>
</template>

<script setup lang="ts">
import {ref, onMounted} from 'vue'
import type {IProject} from '@/modelTypes/IProject'
import {listAdminProjects, reassignProjectOwner} from '@/services/admin/projectService'
import {listAdminUsers, type AdminUser} from '@/services/admin/userService'
import Card from '@/components/misc/Card.vue'
import Modal from '@/components/misc/Modal.vue'
import XButton from '@/components/input/Button.vue'

const projects = ref<IProject[]>([])
const loading = ref(false)
const error = ref('')

const reassignTarget = ref<IProject | null>(null)
const newOwnerQuery = ref('')
const userResults = ref<AdminUser[]>([])
const selectedUserId = ref<number | null>(null)
let searchTimer: ReturnType<typeof setTimeout> | null = null

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

function openReassign(p: IProject) {
	reassignTarget.value = p
	newOwnerQuery.value = ''
	userResults.value = []
	selectedUserId.value = null
}

function searchUsers() {
	if (searchTimer) clearTimeout(searchTimer)
	const q = newOwnerQuery.value
	if (!q || q.length < 2) {
		userResults.value = []
		return
	}
	searchTimer = setTimeout(async () => {
		try {
			userResults.value = await listAdminUsers({s: q})
		} catch (e) {
			error.value = e instanceof Error ? e.message : String(e)
		}
	}, 200)
}

async function doReassign() {
	if (!reassignTarget.value || !selectedUserId.value) return
	const target = reassignTarget.value
	const newOwnerId = selectedUserId.value
	reassignTarget.value = null
	try {
		const updated = await reassignProjectOwner(target.id, newOwnerId)
		const idx = projects.value.findIndex(x => x.id === target.id)
		if (idx !== -1) projects.value[idx] = updated
	} catch (e) {
		error.value = e instanceof Error ? e.message : String(e)
	}
}

onMounted(load)
</script>

<style lang="scss" scoped>
.admin-projects__table {
	inline-size: 100%;
	border-collapse: collapse;

	th, td {
		padding: 0.5rem 0.75rem;
		text-align: start;
		border-block-end: 1px solid var(--grey-200);
	}
}

.admin-projects__results {
	list-style: none;
	padding: 0;
	margin: 0.5rem 0 0;
	border: 1px solid var(--grey-200);
	border-radius: 4px;
	max-block-size: 200px;
	overflow-y: auto;

	li {
		padding: 0.5rem 0.75rem;
		cursor: pointer;

		&:hover,
		&.selected {
			background: var(--grey-100);
		}
	}
}
</style>
