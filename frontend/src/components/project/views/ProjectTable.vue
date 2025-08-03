<template>
	<ProjectWrapper
		class="project-table"
		:is-loading-project="isLoadingProject"
		:project-id="projectId"
		:view-id
	>
		<template #header>
			<div class="filter-container">
				<Popup>
					<template #trigger="{toggle}">
						<XButton
							icon="th"
							variant="secondary"
							class="mie-2"
							@click.prevent.stop="toggle()"
						>
							{{ $t('project.table.columns') }}
						</XButton>
					</template>
					<template #content="{isOpen}">
						<Card
							class="columns-filter"
							:class="{'is-open': isOpen}"
						>
							<FancyCheckbox v-model="activeColumns.index">
								#
							</FancyCheckbox>
							<FancyCheckbox v-model="activeColumns.done">
								{{ $t('task.attributes.done') }}
							</FancyCheckbox>
							<FancyCheckbox v-model="activeColumns.title">
								{{ $t('task.attributes.title') }}
							</FancyCheckbox>
							<FancyCheckbox v-model="activeColumns.priority">
								{{ $t('task.attributes.priority') }}
							</FancyCheckbox>
							<FancyCheckbox v-model="activeColumns.labels">
								{{ $t('task.attributes.labels') }}
							</FancyCheckbox>
							<FancyCheckbox v-model="activeColumns.assignees">
								{{ $t('task.attributes.assignees') }}
							</FancyCheckbox>
							<FancyCheckbox v-model="activeColumns.dueDate">
								{{ $t('task.attributes.dueDate') }}
							</FancyCheckbox>
							<FancyCheckbox v-model="activeColumns.startDate">
								{{ $t('task.attributes.startDate') }}
							</FancyCheckbox>
							<FancyCheckbox v-model="activeColumns.endDate">
								{{ $t('task.attributes.endDate') }}
							</FancyCheckbox>
							<FancyCheckbox v-model="activeColumns.percentDone">
								{{ $t('task.attributes.percentDone') }}
							</FancyCheckbox>
							<FancyCheckbox v-model="activeColumns.doneAt">
								{{ $t('task.attributes.doneAt') }}
							</FancyCheckbox>
							<FancyCheckbox v-model="activeColumns.created">
								{{ $t('task.attributes.created') }}
							</FancyCheckbox>
							<FancyCheckbox v-model="activeColumns.updated">
								{{ $t('task.attributes.updated') }}
							</FancyCheckbox>
							<FancyCheckbox v-model="activeColumns.createdBy">
								{{ $t('task.attributes.createdBy') }}
							</FancyCheckbox>
						</Card>
					</template>
				</Popup>
				<FilterPopup
					v-if="!isSavedFilter({id: projectId})"
					v-model="params"
					:view-id="viewId"
					:project-id="projectId"
				/>
			</div>
		</template>

		<template #default>
			<div
				:class="{'is-loading': loading}"
				class="loader-container"
			>
				<Card
					:padding="false"
					:has-content="false"
				>
					<div class="has-horizontal-overflow">
						<table class="table has-actions is-hoverable is-fullwidth mbe-0">
							<thead>
								<tr>
									<th v-if="activeColumns.index">
										#
										<Sort
											:order="sortBy.index"
											@click="sort('index', $event)"
										/>
									</th>
									<th v-if="activeColumns.done">
										{{ $t('task.attributes.done') }}
										<Sort
											:order="sortBy.done"
											@click="sort('done', $event)"
										/>
									</th>
									<th v-if="activeColumns.title">
										{{ $t('task.attributes.title') }}
										<Sort
											:order="sortBy.title"
											@click="sort('title', $event)"
										/>
									</th>
									<th v-if="activeColumns.priority">
										{{ $t('task.attributes.priority') }}
										<Sort
											:order="sortBy.priority"
											@click="sort('priority', $event)"
										/>
									</th>
									<th v-if="activeColumns.labels">
										{{ $t('task.attributes.labels') }}
									</th>
									<th v-if="activeColumns.assignees">
										{{ $t('task.attributes.assignees') }}
									</th>
									<th v-if="activeColumns.dueDate">
										{{ $t('task.attributes.dueDate') }}
										<Sort
											:order="sortBy.due_date"
											@click="sort('due_date', $event)"
										/>
									</th>
									<th v-if="activeColumns.startDate">
										{{ $t('task.attributes.startDate') }}
										<Sort
											:order="sortBy.start_date"
											@click="sort('start_date', $event)"
										/>
									</th>
									<th v-if="activeColumns.endDate">
										{{ $t('task.attributes.endDate') }}
										<Sort
											:order="sortBy.end_date"
											@click="sort('end_date', $event)"
										/>
									</th>
									<th v-if="activeColumns.percentDone">
										{{ $t('task.attributes.percentDone') }}
										<Sort
											:order="sortBy.percent_done"
											@click="sort('percent_done', $event)"
										/>
									</th>
									<th v-if="activeColumns.doneAt">
										{{ $t('task.attributes.doneAt') }}
										<Sort
											:order="sortBy.done_at"
											@click="sort('done_at', $event)"
										/>
									</th>
									<th v-if="activeColumns.created">
										{{ $t('task.attributes.created') }}
										<Sort
											:order="sortBy.created"
											@click="sort('created', $event)"
										/>
									</th>
									<th v-if="activeColumns.updated">
										{{ $t('task.attributes.updated') }}
										<Sort
											:order="sortBy.updated"
											@click="sort('updated', $event)"
										/>
									</th>
									<th v-if="activeColumns.createdBy">
										{{ $t('task.attributes.createdBy') }}
									</th>
								</tr>
							</thead>
							<tbody>
								<tr
									v-for="t in tasks"
									:key="t.id"
								>
									<td v-if="activeColumns.index">
										<RouterLink :to="taskDetailRoutes[t.id]">
											<template v-if="t.identifier === ''">
												#{{ t.index }}
											</template>
											<template v-else>
												{{ t.identifier }}
											</template>
										</RouterLink>
									</td>
									<td v-if="activeColumns.done">
										<Done
											:is-done="t.done"
											variant="small"
										/>
									</td>
									<td v-if="activeColumns.title">
										<RouterLink :to="taskDetailRoutes[t.id]">
											{{ t.title }}
										</RouterLink>
									</td>
									<td v-if="activeColumns.priority">
										<PriorityLabel
											:priority="t.priority"
											:done="t.done"
											:show-all="true"
										/>
									</td>
									<td v-if="activeColumns.labels">
										<Labels :labels="t.labels" />
									</td>
									<td v-if="activeColumns.assignees">
										<AssigneeList
											v-if="t.assignees.length > 0"
											:assignees="t.assignees"
											:avatar-size="28"
											class="mis-1"
											:inline="true"
										/>
									</td>
									<DateTableCell
										v-if="activeColumns.dueDate"
										:date="t.dueDate"
									/>
									<DateTableCell
										v-if="activeColumns.startDate"
										:date="t.startDate"
									/>
									<DateTableCell
										v-if="activeColumns.endDate"
										:date="t.endDate"
									/>
									<td v-if="activeColumns.percentDone">
										{{ t.percentDone * 100 }}%
									</td>
									<DateTableCell
										v-if="activeColumns.doneAt"
										:date="t.doneAt"
									/>
									<DateTableCell
										v-if="activeColumns.created"
										:date="t.created"
									/>
									<DateTableCell
										v-if="activeColumns.updated"
										:date="t.updated"
									/>
									<td v-if="activeColumns.createdBy">
										<User
											:avatar-size="27"
											:show-username="false"
											:user="t.createdBy"
										/>
									</td>
								</tr>
							</tbody>
						</table>
					</div>

					<Pagination
						:total-pages="totalPages"
						:current-page="currentPage"
					/>
				</Card>
			</div>
		</template>
	</ProjectWrapper>
</template>

<script setup lang="ts">
import {computed, type Ref, watch} from 'vue'

import {useStorage} from '@vueuse/core'

import ProjectWrapper from '@/components/project/ProjectWrapper.vue'
import Done from '@/components/misc/Done.vue'
import User from '@/components/misc/User.vue'
import PriorityLabel from '@/components/tasks/partials/PriorityLabel.vue'
import Labels from '@/components/tasks/partials/Labels.vue'
import DateTableCell from '@/components/tasks/partials/DateTableCell.vue'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import Sort from '@/components/tasks/partials/Sort.vue'
import FilterPopup from '@/components/project/partials/FilterPopup.vue'
import Pagination from '@/components/misc/Pagination.vue'
import Popup from '@/components/misc/Popup.vue'

import type {SortBy} from '@/composables/useTaskList'
import {useTaskList} from '@/composables/useTaskList'
import type {ITask} from '@/modelTypes/ITask'
import type {IProject} from '@/modelTypes/IProject'
import AssigneeList from '@/components/tasks/partials/AssigneeList.vue'
import type {IProjectView} from '@/modelTypes/IProjectView'
import { camelCase } from 'change-case'
import {isSavedFilter} from '@/services/savedFilter'

const props = defineProps<{
	isLoadingProject: boolean,
	projectId: IProject['id'],
	viewId: IProjectView['id'],
}>()

const ACTIVE_COLUMNS_DEFAULT = {
	index: true,
	done: true,
	title: true,
	priority: false,
	labels: true,
	assignees: true,
	dueDate: true,
	startDate: false,
	endDate: false,
	percentDone: false,
	created: false,
	updated: false,
	createdBy: false,
	doneAt: false,
}

const SORT_BY_DEFAULT: SortBy = {
	index: 'desc',
}

const activeColumns = useStorage('tableViewColumns', {...ACTIVE_COLUMNS_DEFAULT})
const sortBy = useStorage<SortBy>('tableViewSortBy', {...SORT_BY_DEFAULT})

const taskList = useTaskList(() => props.projectId, () => props.viewId, sortBy.value)

const {
	loading,
	params,
	totalPages,
	currentPage,
	sortByParam,
} = taskList
const tasks: Ref<ITask[]> = taskList.tasks

Object.assign(params.value, {
	filter: '',
})

watch(
	() => activeColumns.value,
	() => setActiveColumnsSortParam(),
	{deep: true},
)

// Allow sorting by multiple columns only when ctrl is pressed
function sort(property: keyof SortBy, event?: MouseEvent) {
	const ctrlPressed = event?.ctrlKey || event?.metaKey

	const currentOrder = sortBy.value[property]
	let newOrder: 'asc' | 'desc' | 'none' | undefined = undefined
	if (typeof currentOrder === 'undefined' || currentOrder === 'none') {
		newOrder = 'desc'
	} else if (currentOrder === 'desc') {
		newOrder = 'asc'
	}

	if (!ctrlPressed) {
		sortBy.value = {} as SortBy
	}

	if (newOrder) {
		sortBy.value[property] = newOrder
	} else {
		delete sortBy.value[property]
	}

	setActiveColumnsSortParam()
}

function setActiveColumnsSortParam() {
	sortByParam.value = Object.keys(sortBy.value)
		.filter(prop => activeColumns.value[camelCase(prop)])
		.reduce((obj, key) => {
			obj[key] = sortBy.value[key]
			return obj
		}, {})
}

// TODO: re-enable opening task detail in modal
// const router = useRouter()
const taskDetailRoutes = computed(() => Object.fromEntries(
	tasks.value.map(({id}) => ([
		id,
		{
			name: 'task.detail',
			params: {id},
			// state: { backdropView: router.currentRoute.value.fullPath },
		},
	])),
))
</script>

<style lang="scss" scoped>
.table {
	background: transparent;
	overflow-x: auto;
	overflow-y: hidden;

	th {
		white-space: nowrap;
	}

	.user {
		margin: 0;
	}
}

.columns-filter {
	margin: 0;

	:deep(.card-content .content) {
		display: flex;
		flex-direction: column;
	}

	&.is-open {
		margin: 2rem 0 1rem;
	}
}

.link-share-view .card {
	border: none;
	box-shadow: none;
}

.filter-container :deep(.popup) {
	inset-block-start: 7rem;
}
</style>
