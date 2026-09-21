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
							<td>
								<User
									v-if="p.owner"
									:user="p.owner"
									:avatar-size="24"
								/>
							</td>
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
				@close="closeReassign"
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
							@click="closeReassign"
						>
							{{ $t('misc.cancel') }}
						</XButton>
						<XButton
							variant="primary"
							:disabled="!selectedUser || reassigning"
							:loading="reassigning"
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
import {ref, computed} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import type {AdminUser, Project} from '@/client/generated'
import {adminProjectsQuery, adminUserSearchQuery, useReassignAdminProjectMutation} from '@/client/queries/admin'
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
const currentPage = ref(1)
const {data, isPending: loading} = useQuery(computed(() => adminProjectsQuery(currentPage.value)))
type AdminProject = Project & {id: number}
const projects = computed(() => (data.value?.items ?? []).filter((p): p is AdminProject => p.id !== undefined))
const totalPages = computed(() => data.value?.total_pages ?? 1)
const reassignTarget = ref<AdminProject | null>(null)
const selectedUser = ref<AdminUser | null>(null)
const search = ref('')
const {data: searchData, isFetching: userSearchLoading} = useQuery(computed(() => ({...adminUserSearchQuery(search.value), enabled: !!reassignTarget.value && search.value.length >= 2})))
const userResults = computed(() => search.value.length >= 2 ? searchData.value ?? [] : [])
const reassignMutation = useReassignAdminProjectMutation()
const {isPending: reassigning} = reassignMutation
function goToPage(page: number) { currentPage.value = page }
function openReassign(p: AdminProject) { reassignTarget.value = p; selectedUser.value = null; search.value = '' }
function closeReassign() { if (!reassigning.value) reassignTarget.value = null }
function searchUsers(query: string) { search.value = query }
async function doReassign() {
	const target = reassignTarget.value
	const ownerId = selectedUser.value?.id
	if (!target || !ownerId || reassigning.value) return
	try {
		await reassignMutation.mutateAsync({id: target.id, ownerId})
		if (reassignTarget.value?.id === target.id) reassignTarget.value = null
	} catch { /* Mutation reports the error. */ }
}
</script>

<style lang="scss" scoped>
// `.table.has-actions` sets overflow: hidden which clips the dropdown menu.
.admin-projects :deep(.table.has-actions) {
	overflow: visible;
}
</style>
