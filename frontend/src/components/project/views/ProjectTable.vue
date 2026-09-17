<template>
	<ProjectWrapper
		class="project-table"
		:is-loading-project="isLoadingProject"
		:project-id="projectId"
		:view-id
	>
		<template #header>
			<div class="filter-container">
				<Popup
					placement="bottom-start"
					:anchor="columnsTriggerEl"
				>
					<template #trigger="{toggle}">
						<XButton
							ref="columnsTrigger"
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
							<FancyCheckbox v-model="activeColumns.project">
								{{ $t('task.attributes.project') }}
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
							<FancyCheckbox v-model="activeColumns.comment_count">
								{{ $t('task.attributes.commentCount') }}
							</FancyCheckbox>
							<FancyCheckbox v-model="activeColumns.due_date">
								{{ $t('task.attributes.dueDate') }}
							</FancyCheckbox>
							<FancyCheckbox v-model="activeColumns.start_date">
								{{ $t('task.attributes.startDate') }}
							</FancyCheckbox>
							<FancyCheckbox v-model="activeColumns.end_date">
								{{ $t('task.attributes.endDate') }}
							</FancyCheckbox>
							<FancyCheckbox v-model="activeColumns.percent_done">
								{{ $t('task.attributes.percentDone') }}
							</FancyCheckbox>
							<FancyCheckbox v-model="activeColumns.done_at">
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
					v-if="!isSavedFilterProject({id: projectId})"
					v-model="params"
					:view-id="viewId"
					:project-id="projectId"
					@update:modelValue="taskList.loadTasks()"
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
									<th
										v-if="activeColumns.index"
										:aria-sort="ariaSort(sortBy.index)"
									>
										#
										<Sort
											:order="sortBy.index"
											label="#"
											@click="sort('index', $event)"
										/>
									</th>
									<th
										v-if="activeColumns.done"
										:aria-sort="ariaSort(sortBy.done)"
									>
										{{ $t('task.attributes.done') }}
										<Sort
											:order="sortBy.done"
											:label="$t('task.attributes.done')"
											@click="sort('done', $event)"
										/>
									</th>
									<th v-if="activeColumns.project">
										{{ $t('task.attributes.project') }}
									</th>
									<th
										v-if="activeColumns.title"
										:aria-sort="ariaSort(sortBy.title)"
									>
										{{ $t('task.attributes.title') }}
										<Sort
											:order="sortBy.title"
											:label="$t('task.attributes.title')"
											@click="sort('title', $event)"
										/>
									</th>
									<th
										v-if="activeColumns.priority"
										:aria-sort="ariaSort(sortBy.priority)"
									>
										{{ $t('task.attributes.priority') }}
										<Sort
											:order="sortBy.priority"
											:label="$t('task.attributes.priority')"
											@click="sort('priority', $event)"
										/>
									</th>
									<th v-if="activeColumns.labels">
										{{ $t('task.attributes.labels') }}
									</th>
									<th v-if="activeColumns.assignees">
										{{ $t('task.attributes.assignees') }}
									</th>
									<th
										v-if="activeColumns.due_date"
										:aria-sort="ariaSort(sortBy.due_date)"
									>
										{{ $t('task.attributes.dueDate') }}
										<Sort
											:order="sortBy.due_date"
											:label="$t('task.attributes.dueDate')"
											@click="sort('due_date', $event)"
										/>
									</th>
									<th v-if="activeColumns.comment_count">
										{{ $t('task.attributes.commentCount') }}
									</th>
									<th
										v-if="activeColumns.start_date"
										:aria-sort="ariaSort(sortBy.start_date)"
									>
										{{ $t('task.attributes.startDate') }}
										<Sort
											:order="sortBy.start_date"
											:label="$t('task.attributes.startDate')"
											@click="sort('start_date', $event)"
										/>
									</th>
									<th
										v-if="activeColumns.end_date"
										:aria-sort="ariaSort(sortBy.end_date)"
									>
										{{ $t('task.attributes.endDate') }}
										<Sort
											:order="sortBy.end_date"
											:label="$t('task.attributes.endDate')"
											@click="sort('end_date', $event)"
										/>
									</th>
									<th
										v-if="activeColumns.percent_done"
										:aria-sort="ariaSort(sortBy.percent_done)"
									>
										{{ $t('task.attributes.percentDone') }}
										<Sort
											:order="sortBy.percent_done"
											:label="$t('task.attributes.percentDone')"
											@click="sort('percent_done', $event)"
										/>
									</th>
									<th
										v-if="activeColumns.done_at"
										:aria-sort="ariaSort(sortBy.done_at)"
									>
										{{ $t('task.attributes.doneAt') }}
										<Sort
											:order="sortBy.done_at"
											:label="$t('task.attributes.doneAt')"
											@click="sort('done_at', $event)"
										/>
									</th>
									<th
										v-if="activeColumns.created"
										:aria-sort="ariaSort(sortBy.created)"
									>
										{{ $t('task.attributes.created') }}
										<Sort
											:order="sortBy.created"
											:label="$t('task.attributes.created')"
											@click="sort('created', $event)"
										/>
									</th>
									<th
										v-if="activeColumns.updated"
										:aria-sort="ariaSort(sortBy.updated)"
									>
										{{ $t('task.attributes.updated') }}
										<Sort
											:order="sortBy.updated"
											:label="$t('task.attributes.updated')"
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
											{{ getTaskIdentifier(t) }}
										</RouterLink>
									</td>
									<td v-if="activeColumns.done">
										<Done
											:is-done="t.done"
											variant="small"
										/>
									</td>
									<td v-if="activeColumns.project">
										<RouterLink
											v-if="projectList.projects[t.project_id]"
											:to="{ name: 'project.index', params: { projectId: t.project_id } }"
										>
											{{ projectList.projects[t.project_id].title }}
										</RouterLink>
									</td>
									<td v-if="activeColumns.title">
										<TaskGlanceTooltip :task="t">
											<RouterLink :to="taskDetailRoutes[t.id]">
												{{ t.title }}
											</RouterLink>
										</TaskGlanceTooltip>
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
										v-if="activeColumns.due_date"
										:date="t.due_date"
									/>
									<td v-if="activeColumns.comment_count">
										<CommentCount :task="t" />
									</td>
									<DateTableCell
										v-if="activeColumns.start_date"
										:date="t.start_date"
									/>
									<DateTableCell
										v-if="activeColumns.end_date"
										:date="t.end_date"
									/>
									<td v-if="activeColumns.percent_done">
										{{ t.percent_done * 100 }}%
									</td>
									<DateTableCell
										v-if="activeColumns.done_at"
										:date="t.done_at"
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
import {computed, ref, type ComponentPublicInstance, type Ref, watch} from 'vue'

import {useStorage} from '@vueuse/core'

import ProjectWrapper from '@/components/project/ProjectWrapper.vue'
import Done from '@/components/misc/Done.vue'
import User from '@/components/misc/User.vue'
import PriorityLabel from '@/components/tasks/partials/PriorityLabel.vue'
import Labels from '@/components/tasks/partials/Labels.vue'
import TaskGlanceTooltip from '@/components/tasks/partials/TaskGlanceTooltip.vue'
import DateTableCell from '@/components/tasks/partials/DateTableCell.vue'
import CommentCount from '@/components/tasks/partials/CommentCount.vue'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import Sort from '@/components/tasks/partials/Sort.vue'
import FilterPopup from '@/components/project/partials/FilterPopup.vue'
import Pagination from '@/components/misc/Pagination.vue'
import Popup from '@/components/misc/Popup.vue'

import type {SortBy} from '@/composables/useTaskList'
import {useTaskList} from '@/composables/useTaskList'
import type {Task as ITask} from '@/client/generated'
import AssigneeList from '@/components/tasks/partials/AssigneeList.vue'
import {getTaskIdentifier} from '@/helpers/task'
import { camelCase } from 'change-case'
import {isSavedFilterProject} from '@/client/queries/projects'
import {useProjects} from '@/composables/useProjects'

const props = defineProps<{
	isLoadingProject: boolean,
	projectId: number,
	viewId: number,
}>()

const projectList = useProjects()

const columnsTrigger = ref<ComponentPublicInstance | null>(null)
const columnsTriggerEl = computed<HTMLElement | null>(() => (columnsTrigger.value?.$el as HTMLElement) ?? null)

const ACTIVE_COLUMNS_DEFAULT = {
	index: true,
	done: true,
	project: false,
	title: true,
	priority: false,
	labels: true,
	assignees: true,
	due_date: true,
	start_date: false,
	end_date: false,
	percent_done: false,
	created: false,
	updated: false,
	createdBy: false,
	done_at: false,
	comment_count: false,
}

const SORT_BY_DEFAULT: SortBy = {
	index: 'desc',
}

const activeColumns = useStorage('tableViewColumns', {...ACTIVE_COLUMNS_DEFAULT})
const sortBy = useStorage<SortBy>('tableViewSortBy', {...SORT_BY_DEFAULT})

const taskList = useTaskList(
	() => props.projectId, 
	() => props.viewId, 
	sortBy.value,
	() => ['comment_count', 'is_unread'],
)

const {
	loading,
	params,
	totalPages,
	currentPage,
	sortByParam,
} = taskList
const tasks: Ref<ITask[]> = taskList.tasks

watch(
	() => activeColumns.value,
	() => setActiveColumnsSortParam(),
	{deep: true},
)

function ariaSort(order: 'asc' | 'desc' | 'none' | undefined): 'ascending' | 'descending' | undefined {
	if (order === 'asc') {
		return 'ascending'
	}
	if (order === 'desc') {
		return 'descending'
	}
	return undefined
}

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

}

.link-share-view .card {
	border: none;
	box-shadow: none;
}
</style>
