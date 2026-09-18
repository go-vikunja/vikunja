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
					v-model="excludeInboxes"
					:label="$t('admin.projects.hideInboxes')"
				/>
			</div>

			<p v-if="loading && !hasLoaded">
				{{ $t('misc.loading') }}
			</p>
			<template v-else>
				<div
					class="loader-container"
					:class="{'is-loading': loading}"
					:aria-busy="loading"
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
import {ref, onMounted, watch} from 'vue'
import {useDebounceFn} from '@vueuse/core'
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
import FormInput from '@/components/input/FormInput.vue'
import FormCheckbox from '@/components/input/FormCheckbox.vue'
import Sort from '@/components/tasks/partials/Sort.vue'
import User from '@/components/misc/User.vue'
import ProjectSettingsDropdown from '@/components/project/ProjectSettingsDropdown.vue'
import DropdownItem from '@/components/misc/DropdownItem.vue'
import TimeDisplay from '@/components/misc/TimeDisplay.vue'
import {error, success} from '@/message'
import {useI18n} from 'vue-i18n'
import {useTableSort, type SortOrder, type TableSortState} from '@/composables/useTableSort'

const {t} = useI18n({useScope: 'global'})

const adminUserService = new AdminUserService()

type AdminProject = Project & Required<Pick<Project, 'id'>>
type SortField = 'id' | 'title' | 'owner' | 'created' | 'updated'

const projects = ref<AdminProject[]>([])
const loading = ref(false)
const hasLoaded = ref(false)
const currentPage = ref(1)
const totalPages = ref(1)

const searchTerm = ref('')
const ownerFilter = ref<IAdminUser | null>(null)
const excludeInboxes = ref(false)
const sortBy = ref<TableSortState<SortField>>({id: 'desc'})
const {sort, ariaSort} = useTableSort<SortField>(sortBy, reload)

const reassignTarget = ref<AdminProject | null>(null)
const userResults = ref<IAdminUser[]>([])
const userSearchLoading = ref(false)
const selectedUser = ref<IAdminUser | null>(null)

let requestId = 0

async function load() {
	const request = ++requestId
	loading.value = true
	try {
		const sortFields = Object.keys(sortBy.value) as SortField[]
		const {data} = await adminProjectsList({query: {
			page: currentPage.value,
			q: searchTerm.value || undefined,
			owner_id: ownerFilter.value?.id,
			exclude_inboxes: excludeInboxes.value || undefined,
			sort_by: sortFields,
			order_by: sortFields.map(field => sortBy.value[field] as SortOrder),
		}})
		if (request !== requestId) {
			return
		}
		projects.value = (data.items ?? []).filter((project): project is AdminProject => project.id !== undefined)
		totalPages.value = data.total_pages ?? 1
		hasLoaded.value = true
	} catch (e) {
		if (request === requestId) {
			error(e)
		}
	} finally {
		if (request === requestId) {
			loading.value = false
		}
	}
}

function goToPage(page: number) {
	currentPage.value = page
	load()
}

function reload() {
	// Reset to page 1 so a narrower filter doesn't strand the UI on an empty page.
	currentPage.value = 1
	load()
}

const onSearch = useDebounceFn(reload, 300)

watch([ownerFilter, excludeInboxes], reload)

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
