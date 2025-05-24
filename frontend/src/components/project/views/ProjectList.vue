<template>
	<ProjectWrapper
		class="project-list"
		:is-loading-project="isLoadingProject"
		:project-id="projectId"
		:view-id
	>
		<template #header>
			<div class="filter-container">
				<FilterPopup
					v-if="!isSavedFilter(project)"
					v-model="params"
					:view-id="viewId"
					:project-id="projectId"
					@update:modelValue="prepareFiltersAndLoadTasks()"
				/>
			</div>
		</template>

		<template #default>
			<div
				:class="{ 'is-loading': loading }"
				class="loader-container is-max-width-desktop list-view"
			>
				<Card
					:padding="false"
					:has-content="false"
					class="has-overflow"
				>
					<AddTask
						v-if="!project.isArchived && canWrite"
						ref="addTaskRef"
						class="list-view__add-task d-print-none"
						:default-position="firstNewPosition"
						@taskAdded="updateTaskList"
					/>

					<Nothing v-if="ctaVisible && tasks.length === 0 && !loading">
						{{ $t('project.list.empty') }}
						<ButtonLink
							v-if="project.id > 0 && canWrite"
							@click="focusNewTaskInput()"
						>
							{{ $t('project.list.newTaskCta') }}
						</ButtonLink>
					</Nothing>

					<draggable
						v-if="tasks && tasks.length > 0"
						v-model="tasks"
						group="tasks"
						handle=".handle"
						:disabled="!canDragTasks"
						item-key="id"
						tag="ul"
						:component-data="{
							class: {
								tasks: true,
								'dragging-disabled': !canDragTasks || isAlphabeticalSorting
							},
							type: 'transition-group'
						}"
						:animation="100"
						ghost-class="task-ghost"
						@start="() => drag = true"
						@end="saveTaskPosition"
					>
						<template #item="{element: t}">
							<SingleTaskInProject
								:show-list-color="false"
								:disabled="!canDragTasks"
								:can-mark-as-done="canWrite || isPseudoProject"
								:the-task="t"
								:all-tasks="allTasks"
								@taskUpdated="updateTasks"
							>
								<template v-if="canDragTasks">
									<span class="icon handle">
										<Icon icon="grip-lines" />
									</span>
								</template>
							</SingleTaskInProject>
						</template>
					</draggable>

					<Pagination
						:total-pages="totalPages"
						:current-page="currentPage"
					/>
				</Card>
			</div>
		</template>
	</ProjectWrapper>
</template>

<script lang="ts">
export default {name: 'List'}
</script>

<script setup lang="ts">
import {ref, computed, nextTick, onMounted, watch} from 'vue'
import draggable from 'zhyswan-vuedraggable'

import ProjectWrapper from '@/components/project/ProjectWrapper.vue'
import ButtonLink from '@/components/misc/ButtonLink.vue'
import AddTask from '@/components/tasks/AddTask.vue'
import SingleTaskInProject from '@/components/tasks/partials/SingleTaskInProject.vue'
import FilterPopup from '@/components/project/partials/FilterPopup.vue'
import Nothing from '@/components/misc/Nothing.vue'
import Pagination from '@/components/misc/Pagination.vue'
import {ALPHABETICAL_SORT} from '@/components/project/partials/Filters.vue'

import {useTaskList} from '@/composables/useTaskList'
import {RIGHTS as Rights} from '@/constants/rights'
import {calculateItemPosition} from '@/helpers/calculateItemPosition'
import type {ITask} from '@/modelTypes/ITask'
import {isSavedFilter} from '@/services/savedFilter'

import {useBaseStore} from '@/stores/base'

import type {IProject} from '@/modelTypes/IProject'
import type {IProjectView} from '@/modelTypes/IProjectView'
import TaskPositionService from '@/services/taskPosition'
import TaskPositionModel from '@/models/taskPosition'

const props = defineProps<{
	isLoadingProject: boolean,
	projectId: IProject['id'],
	viewId: IProjectView['id'],
}>()

const ctaVisible = ref(false)

const drag = ref(false)

const {
	tasks: allTasks,
	loading,
	totalPages,
	currentPage,
	loadTasks,
	params,
	sortByParam,
} = useTaskList(
	() => props.projectId,
	() => props.viewId,
	{position: 'asc'},
	() => props.projectId === -1
		? null
		: 'subtasks',
)

const taskPositionService = ref(new TaskPositionService())

const tasks = ref<ITask[]>([])
watch(
	allTasks,
	() => {
		tasks.value = [...allTasks.value]
		if (props.projectId < 0) {
			return
		}
		const tasksById = {}
		tasks.value.forEach(t => tasksById[t.id] = true)

		tasks.value = tasks.value.filter(t => {
			if (typeof t.relatedTasks?.parenttask === 'undefined') {
				return true
			}

			// If the task is a subtask, make sure the parent task is available in the current view as well
			for (const pt of t.relatedTasks.parenttask) {
				if (typeof tasksById[pt.id] === 'undefined') {
					return true
				}
			}

			return false
		})
	},
)

const isAlphabeticalSorting = computed(() => {
	return params.value.sort_by.find(sortBy => sortBy === ALPHABETICAL_SORT) !== undefined
})

const firstNewPosition = computed(() => {
	if (tasks.value.length === 0) {
		return 0
	}

	return calculateItemPosition(null, tasks.value[0].position)
})

const baseStore = useBaseStore()
const project = computed(() => baseStore.currentProject)

const canWrite = computed(() => {
	return project.value.maxRight > Rights.READ && project.value.id > 0
})

const isPseudoProject = computed(() => (project.value && isSavedFilter(project.value)) || project.value?.id === -1)

onMounted(async () => {
	await nextTick()
	ctaVisible.value = true
})

const canDragTasks = computed(() => canWrite.value || isSavedFilter(project.value))

const addTaskRef = ref<typeof AddTask | null>(null)

function focusNewTaskInput() {
	addTaskRef.value?.focusTaskInput()
}

function updateTaskList(task: ITask) {
	if (isAlphabeticalSorting.value) {
		// reload tasks with current filter and sorting
		loadTasks()
	} else {
		allTasks.value = [
			task,
			...allTasks.value,
		]
	}

	baseStore.setHasTasks(true)
}

function updateTasks(updatedTask: ITask) {
	if (props.projectId < 0) {
		// In the case of a filter, we'll reload the filter in the background to avoid tasks which do 
		// not match the filter show up here
		loadTasks(false)
		return
	}

	for (const t in tasks.value) {
		if (tasks.value[t].id === updatedTask.id) {
			tasks.value[t] = updatedTask
			break
		}
	}
}

async function saveTaskPosition(e) {
	drag.value = false

	const task = tasks.value[e.newIndex]
	const taskBefore = tasks.value[e.newIndex - 1] ?? null
	const taskAfter = tasks.value[e.newIndex + 1] ?? null

	const position = calculateItemPosition(taskBefore !== null ? taskBefore.position : null, taskAfter !== null ? taskAfter.position : null)

	await taskPositionService.value.update(new TaskPositionModel({
		position,
		projectViewId: props.viewId,
		taskId: task.id,
	}))
	tasks.value[e.newIndex] = {
		...task,
		position,
	}
}

function prepareFiltersAndLoadTasks() {
	if (isAlphabeticalSorting.value) {
		sortByParam.value = {}
		sortByParam.value[ALPHABETICAL_SORT] = 'asc'
	}

	loadTasks()
}
</script>

<style lang="scss" scoped>
.tasks {
	padding: .5rem;
}

.task-ghost {
	border-radius: $radius;
	background: var(--grey-100);
	border: 2px dashed var(--grey-300);

	* {
		opacity: 0;
	}
}

.list-view__add-task {
	padding: 1rem 1rem 0;
}

.link-share-view .card {
	border: none;
	box-shadow: none;
}

.control.has-icons-left .icon,
.control.has-icons-right .icon {
	transition: all $transition;
}

:deep(.single-task) {
	.handle {
		opacity: 1;
		transition: opacity $transition;
		margin-right: .25rem;
		cursor: grab;
	}

	@media(hover: hover) and (pointer: fine) {
		& .handle {
			opacity: 0;
		}

		&:hover .handle {
			opacity: 1;
		}
	}
}

.list-view {
	padding-bottom: 1rem;

	:deep(.card) {
		margin-bottom: 0;
	}
}
</style>
