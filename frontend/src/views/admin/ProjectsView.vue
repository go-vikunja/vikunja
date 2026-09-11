<template>
	<Card>
		<div class="admin-projects">
			<p v-if="loading">
				{{ $t('misc.loading') }}
			</p>
			<template v-else>
				<table class="table has-actions is-striped is-hoverable is-fullwidth">
					<thead>
						<tr>
							<th>{{ $t('misc.id') }}</th>
							<th>{{ $t('project.title') }}</th>
							<th>{{ $t('admin.projects.ownerLabel') }}</th>
							<th>{{ $t('task.attributes.created') }}</th>
							<th>{{ $t('task.attributes.updated') }}</th>
							<th>{{ $t('navigation.settings') }}</th>
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
								<TimeDisplay :date="p.created" />
							</td>
							<td>
								<TimeDisplay :date="p.updated" />
							</td>
							<td class="actions">
								<ProjectSettingsDropdown
									:project="p"
									:force-all-actions="true"
								>
									<template #before-delete>
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
				<PaginationEmit
					v-if="totalPages > 1"
					:total-pages="totalPages"
					:current-page="currentPage"
					@pageChanged="goToPage"
				/>
			</template>

			<Modal
				v-if="reassignTarget"
				variant="hint-modal"
				@close="reassignTarget = null"
			>
				<Card
					class="has-no-shadow"
					:title="$t('admin.projects.reassignTitle', {title: reassignTarget.title})"
				>
					<FormField :label="$t('admin.projects.newOwnerLabel')">
						<Multiselect
							v-model="selectedUser"
							:loading="userSearchLoading"
							:placeholder="$t('admin.searchUsersPlaceholder')"
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
					</FormField>

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
				</Card>
			</Modal>
		</div>
	</Card>
</template>

<script setup lang="ts">
import {ref, onMounted} from 'vue'
import {
	adminProjectsList,
	adminProjectsPatchOwner,
} from '@/client/generated'
import type {Project} from '@/client/generated'
import type {IAdminUser} from '@/modelTypes/IAdminUser'
import AdminUserService from '@/services/admin/userService'
import AdminUserModel from '@/models/adminUser'
import Card from '@/components/misc/Card.vue'
import Modal from '@/components/misc/Modal.vue'
import PaginationEmit from '@/components/misc/PaginationEmit.vue'
import XButton from '@/components/input/Button.vue'
import FormField from '@/components/input/FormField.vue'
import Multiselect from '@/components/input/Multiselect.vue'
import User from '@/components/misc/User.vue'
import ProjectSettingsDropdown from '@/components/project/ProjectSettingsDropdown.vue'
import DropdownItem from '@/components/misc/DropdownItem.vue'
import TimeDisplay from '@/components/misc/TimeDisplay.vue'
import {error, success} from '@/message'
import {useI18n} from 'vue-i18n'

const {t} = useI18n({useScope: 'global'})

const adminUserService = new AdminUserService()

type AdminProject = Project & Required<Pick<Project, 'id'>>

const projects = ref<AdminProject[]>([])
const loading = ref(false)
const currentPage = ref(1)
const totalPages = ref(1)

const reassignTarget = ref<AdminProject | null>(null)
const userResults = ref<IAdminUser[]>([])
const userSearchLoading = ref(false)
const selectedUser = ref<IAdminUser | null>(null)

async function load() {
	loading.value = true
	try {
		const {data} = await adminProjectsList({query: {page: currentPage.value}})
		projects.value = (data.items ?? []).filter((project): project is AdminProject => project.id !== undefined)
		totalPages.value = data.total_pages ?? 1
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

function openReassign(p: AdminProject) {
	reassignTarget.value = p
	userResults.value = []
	selectedUser.value = null
}

async function searchUsers(query: string) {
	if (!query || query.length < 2) {
		userResults.value = []
		return
	}
	userSearchLoading.value = true
	try {
		userResults.value = await adminUserService.getAll(new AdminUserModel(), {s: query})
	} catch (e) {
		error(e)
	} finally {
		userSearchLoading.value = false
	}
}

async function doReassign() {
	if (!reassignTarget.value || !selectedUser.value) return
	const target = reassignTarget.value
	const newOwnerId = selectedUser.value.id
	reassignTarget.value = null
	try {
		const {data: updated} = await adminProjectsPatchOwner({
			path: {id: target.id},
			body: {owner_id: newOwnerId},
		})
		const idx = projects.value.findIndex(x => x.id === target.id)
		if (idx !== -1) projects.value[idx] = {...projects.value[idx], ...updated}
		success({message: t('admin.projects.reassignedSuccess')})
	} catch (e) {
		error(e)
	}
}

onMounted(load)
</script>

<style lang="scss" scoped>
// `.table.has-actions` sets overflow: hidden which clips the dropdown menu.
.admin-projects :deep(.table.has-actions) {
	overflow: visible;
}
</style>
