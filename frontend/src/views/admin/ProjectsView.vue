<template>
	<Card>
		<div class="admin-projects">
			<div class="admin-projects__toolbar">
				<div class="admin-projects__search">
					<FormInput
						v-model="searchTerm"
						type="text"
						:placeholder="$t('admin.projects.searchPlaceholder')"
						:aria-label="$t('admin.projects.searchPlaceholder')"
						@input="onSearch"
					/>
				</div>
				<Multiselect
					v-model="ownerFilter"
					class="admin-projects__owner-filter"
					:loading="userSearchLoading"
					:placeholder="$t('admin.projects.filterByOwner')"
					:aria-label="$t('admin.projects.filterByOwner')"
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
				<FormCheckbox
					v-model="excludeDefaultProjects"
					:label="$t('admin.projects.hideDefaultProjects')"
				/>
			</div>

			<p v-if="loading">
				{{ $t('misc.loading') }}
			</p>
			<template v-else>
				<div
					class="loader-container"
					:class="{'is-loading': showReloading}"
					:aria-busy="reloading"
				>
					<table class="table has-actions is-striped is-hoverable is-fullwidth">
						<thead>
							<tr>
								<th :aria-sort="ariaSort('id')">
									{{ $t('misc.id') }}
									<Sort
										:order="sortBy.id"
										:label="$t('misc.id')"
										@click="sort('id', $event)"
									/>
								</th>
								<th :aria-sort="ariaSort('title')">
									{{ $t('project.title') }}
									<Sort
										:order="sortBy.title"
										:label="$t('project.title')"
										@click="sort('title', $event)"
									/>
								</th>
								<th :aria-sort="ariaSort('owner')">
									{{ $t('admin.projects.ownerLabel') }}
									<Sort
										:order="sortBy.owner"
										:label="$t('admin.projects.ownerLabel')"
										@click="sort('owner', $event)"
									/>
								</th>
								<th :aria-sort="ariaSort('created')">
									{{ $t('task.attributes.created') }}
									<Sort
										:order="sortBy.created"
										:label="$t('task.attributes.created')"
										@click="sort('created', $event)"
									/>
								</th>
								<th :aria-sort="ariaSort('updated')">
									{{ $t('task.attributes.updated') }}
									<Sort
										:order="sortBy.updated"
										:label="$t('task.attributes.updated')"
										@click="sort('updated', $event)"
									/>
								</th>
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
				</div>
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
							:disabled="!selectedUser || showReassigning"
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
import {ref, computed, watch} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import {useDebounceFn} from '@vueuse/core'
import type {AdminUser, Project} from '@/client/generated'
import {adminProjectsQuery, adminUserSearchQuery, useReassignAdminProjectMutation} from '@/client/queries/admin'
import Card from '@/components/misc/Card.vue'
import Modal from '@/components/misc/Modal.vue'
import PaginationEmit from '@/components/misc/PaginationEmit.vue'
import XButton from '@/components/input/Button.vue'
import FormField from '@/components/input/FormField.vue'
import Multiselect from '@/components/input/Multiselect.vue'
import FormInput from '@/components/input/FormInput.vue'
import FormCheckbox from '@/components/input/FormCheckbox.vue'
import Sort from '@/components/tasks/partials/Sort.vue'
import User from '@/components/misc/User.vue'
import ProjectSettingsDropdown from '@/components/project/ProjectSettingsDropdown.vue'
import DropdownItem from '@/components/misc/DropdownItem.vue'
import TimeDisplay from '@/components/misc/TimeDisplay.vue'
import {useTableSort, type SortOrder, type TableSortState} from '@/composables/useTableSort'
import {useDelayedLoading} from '@/composables/useDelayedLoading'
type SortField = 'id' | 'title' | 'owner' | 'created' | 'updated'
const currentPage = ref(1)
const searchTerm = ref('')
const search = ref('')
const ownerFilter = ref<AdminUser | null>(null)
const excludeDefaultProjects = ref(false)
const sortBy = ref<TableSortState<SortField>>({id: 'desc'})
const {data, isPending: loading, isFetching: reloading} = useQuery(computed(() => {
	const sortFields = Object.keys(sortBy.value) as SortField[]
	return adminProjectsQuery({
		page: currentPage.value,
		q: search.value || undefined,
		owner_id: ownerFilter.value?.id,
		exclude_default_projects: excludeDefaultProjects.value || undefined,
		sort_by: sortFields,
		order_by: sortFields.map(field => sortBy.value[field] as SortOrder),
	})
}))
const showReloading = useDelayedLoading(reloading)
type AdminProject = Project & {id: number}
const projects = computed(() => (data.value?.items ?? []).filter((p): p is AdminProject => p.id !== undefined))
const totalPages = computed(() => data.value?.total_pages ?? 1)
const reassignTarget = ref<AdminProject | null>(null)
const selectedUser = ref<AdminUser | null>(null)
const userSearch = ref('')
const {data: searchData, isFetching: userSearchLoading} = useQuery(computed(() => ({
	...adminUserSearchQuery(userSearch.value),
	enabled: userSearch.value.length >= 2,
})))
const userResults = computed(() => userSearch.value.length >= 2 ? searchData.value?.items ?? [] : [])
const reassignMutation = useReassignAdminProjectMutation()
const {isPending: reassigning} = reassignMutation
const showReassigning = useDelayedLoading(reassigning)
function goToPage(page: number) { currentPage.value = page }
// Reset to page 1 so a narrower filter doesn't strand the UI on an empty page.
function resetPage() { currentPage.value = 1 }
const onSearch = useDebounceFn(() => {
	resetPage()
	search.value = searchTerm.value
}, 300)
watch([ownerFilter, excludeDefaultProjects], resetPage)
const {sort, ariaSort} = useTableSort<SortField>(sortBy, resetPage)
function openReassign(p: AdminProject) {
	reassignTarget.value = p
	selectedUser.value = null
	userSearch.value = ''
}
function closeReassign() { if (!reassigning.value) reassignTarget.value = null }
function searchUsers(query: string) { userSearch.value = query }
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
.admin-projects__toolbar {
	display: flex;
	flex-wrap: wrap;
	align-items: center;
	gap: 0.5rem;
	margin-block-end: 1rem;
}

.admin-projects__search {
	flex: 1 1 15rem;
}

.admin-projects__owner-filter {
	flex: 0 1 15rem;
}

// `.table.has-actions` sets overflow: hidden which clips the dropdown menu.
.admin-projects :deep(.table.has-actions) {
	overflow: visible;
}
</style>
