<template>
	<Card :title="$t('admin.projects.title')">
		<div class="admin-projects">
			<p v-if="loading">
				{{ $t('misc.loading') }}
			</p>
			<table
				v-else
				class="admin-projects__table"
			>
				<thead>
					<tr>
						<th>{{ $t('admin.projects.id') }}</th>
						<th>{{ $t('admin.projects.projectTitle') }}</th>
						<th>{{ $t('admin.projects.ownerLabel') }}</th>
						<th>{{ $t('admin.projects.createdLabel') }}</th>
						<th>{{ $t('admin.projects.updatedLabel') }}</th>
						<th>{{ $t('admin.projects.settings') }}</th>
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
						<td>
							<time :datetime="toISO(p.created)">{{ formatDisplayDate(p.created) }}</time>
						</td>
						<td>
							<time :datetime="toISO(p.updated)">{{ formatDisplayDate(p.updated) }}</time>
						</td>
						<td class="admin-projects__actions">
							<ProjectSettingsDropdown
								:project="p"
								:force-all-actions="true"
							>
								<template #extra>
									<DropdownItem
										icon="user-edit"
										@click="openReassign(p)"
									>
										{{ $t('admin.projects.reassignOwner') }}
									</DropdownItem>
								</template>
							</ProjectSettingsDropdown>
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
					<div class="field">
						<label class="label">
							{{ $t('admin.projects.newOwnerLabel') }}
						</label>
						<Multiselect
							v-model="selectedUser"
							:loading="userSearchLoading"
							:placeholder="$t('admin.projects.newOwnerPlaceholder')"
							:search-results="userResults"
							label="username"
							@search="searchUsers"
						>
							<template #searchResult="{option}">
								<User
									v-if="typeof option !== 'string'"
									:avatar-size="24"
									:show-username="true"
									:user="option"
								/>
							</template>
						</Multiselect>
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
						:disabled="!selectedUser"
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
import Multiselect from '@/components/input/Multiselect.vue'
import User from '@/components/misc/User.vue'
import ProjectSettingsDropdown from '@/components/project/ProjectSettingsDropdown.vue'
import DropdownItem from '@/components/misc/DropdownItem.vue'
import {formatDisplayDate, formatISO} from '@/helpers/time/formatDate'
import {error, success} from '@/message'
import {useI18n} from 'vue-i18n'

const {t} = useI18n({useScope: 'global'})

const projects = ref<IProject[]>([])
const loading = ref(false)

const reassignTarget = ref<IProject | null>(null)
const userResults = ref<AdminUser[]>([])
const userSearchLoading = ref(false)
const selectedUser = ref<AdminUser | null>(null)
let searchTimer: ReturnType<typeof setTimeout> | null = null

async function load() {
	loading.value = true
	try {
		projects.value = await listAdminProjects()
	} catch (e) {
		error(e)
	} finally {
		loading.value = false
	}
}

function openReassign(p: IProject) {
	reassignTarget.value = p
	userResults.value = []
	selectedUser.value = null
}

function searchUsers(query: string) {
	if (searchTimer) clearTimeout(searchTimer)
	if (!query || query.length < 2) {
		userResults.value = []
		return
	}
	userSearchLoading.value = true
	searchTimer = setTimeout(async () => {
		try {
			userResults.value = await listAdminUsers({s: query})
		} catch (e) {
			error(e)
		} finally {
			userSearchLoading.value = false
		}
	}, 200)
}

async function doReassign() {
	if (!reassignTarget.value || !selectedUser.value) return
	const target = reassignTarget.value
	const newOwnerId = selectedUser.value.id
	reassignTarget.value = null
	try {
		const updated = await reassignProjectOwner(target.id, newOwnerId)
		const idx = projects.value.findIndex(x => x.id === target.id)
		if (idx !== -1) projects.value[idx] = updated
		success({message: t('admin.projects.reassignedSuccess')})
	} catch (e) {
		error(e)
	}
}

function toISO(date: Date | string | null | undefined): string {
	return date ? formatISO(date) : ''
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

.admin-projects__actions {
	display: flex;
	gap: 0.5rem;
	align-items: center;
	justify-content: flex-end;
}
</style>
