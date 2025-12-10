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
						:group="{name: 'tasks', put: false}"
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
						@start="handleDragStart"
						@end="saveTaskPosition"
					>
						<template #item="{element: t, index}">
							<SingleTaskInProject
								:ref="(el) => setTaskRef(el, index)"
								:show-list-color="false"
								:disabled="!canDragTasks"
								:can-mark-as-done="canWrite || isPseudoProject"
								:the-task="t"
								:all-tasks="allTasks"
								@taskUpdated="updateTasks"
							/>
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


<script setup lang="ts">
import {ref, computed, nextTick, onMounted, onBeforeUnmount, watch, toRef} from 'vue'
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
import {useTaskDragToProject} from '@/composables/useTaskDragToProject'
import {shouldShowTaskInListView} from '@/composables/useTaskListFiltering'
import {PERMISSIONS as Permissions} from '@/constants/permissions'
import {calculateItemPosition} from '@/helpers/calculateItemPosition'
import type {ITask} from '@/modelTypes/ITask'
import {isSavedFilter, useSavedFilter} from '@/services/savedFilter'

import {useBaseStore} from '@/stores/base'
import {useTaskStore} from '@/stores/tasks'

import type {IProject} from '@/modelTypes/IProject'
import type {IProjectView} from '@/modelTypes/IProjectView'
import TaskPositionService from '@/services/taskPosition'
import TaskPositionModel from '@/models/taskPosition'

const props = defineProps<{
        isLoadingProject: boolean,
        projectId: IProject['id'],
        viewId: IProjectView['id'],
}>()

const projectId = toRef(props, 'projectId')

defineOptions({name: 'List'})

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
	() => projectId.value,
	() => props.viewId,
	{position: 'asc'},
	() => projectId.value === -1
		? ['comment_count', 'is_unread']
		: ['subtasks', 'comment_count', 'is_unread'],
)

const taskPositionService = ref(new TaskPositionService())

// Saved filter composable for accessing filter data
const _savedFilter = useSavedFilter(() => projectId.value < 0 ? projectId.value : undefined).filter

const tasks = ref<ITask[]>([])
watch(
	allTasks,
	() => {
		tasks.value = ([...allTasks.value]).filter(t => shouldShowTaskInListView(t, allTasks.value))
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
const taskStore = useTaskStore()
const {handleTaskDropToProject, clearDragState} = useTaskDragToProject()
const project = computed(() => baseStore.currentProject)

const canWrite = computed(() => {
	return project.value.maxPermission > Permissions.READ && project.value.id > 0
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
	if (projectId.value < 0) {
		// Reload tasks to keep saved filter results in sync
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

function handleDragStart(e: { item: HTMLElement }) {
	drag.value = true
	const taskId = parseInt(e.item.dataset.taskId ?? '', 10)
	const task = tasks.value.find(t => t.id === taskId)

	if (task) {
		taskStore.setDraggedTask(task)
	}
}

async function saveTaskPosition(e: { originalEvent?: MouseEvent, to: HTMLElement, from: HTMLElement, newIndex: number }) {
	drag.value = false

	// Check if the task was dropped over a sidebar project
	const {moved} = await handleTaskDropToProject(e, (task) => {
		// Remove from local list on successful move
		tasks.value = tasks.value.filter(t => t.id !== task.id)
	})

	if (moved) {
		return
	}

	// Clear drag state if not moved to a project
	clearDragState()

	// If the task was dropped outside this list (not on a project)
	if (e.to !== e.from) {
		return
	}

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

const taskRefs = ref<(InstanceType<typeof SingleTaskInProject> | null)[]>([])
const focusedIndex = ref(-1)

function setTaskRef(el: InstanceType<typeof SingleTaskInProject> | null, index: number) {
	if (el === null) {
		delete taskRefs.value[index]
	} else {
		taskRefs.value[index] = el
	}
}

function focusTask(index: number) {
	if (index < 0 || index >= tasks.value.length) {
		return
	}

	const taskRef = taskRefs.value[index]

	focusedIndex.value = index
	taskRef?.focus()
}

function handleListNavigation(e: KeyboardEvent) {
	if (e.target instanceof HTMLElement && (e.target.closest('input, textarea, select, [contenteditable="true"]'))) {
		return
	}

	if (e.key === 'j') {
		e.preventDefault()
		focusTask(Math.min(focusedIndex.value + 1, tasks.value.length - 1))
		return
	}

	if (e.key === 'k') {
		e.preventDefault()
		if (focusedIndex.value === -1) {
			focusTask(tasks.value.length - 1)
			return
		}

		if (focusedIndex.value === 0) {
			addTaskRef.value?.focusTaskInput()
			focusedIndex.value = -1
			return
		}

		focusTask(Math.max(focusedIndex.value - 1, 0))
		return
	}

	if (e.key === 'Enter') {
		if (e.isComposing) {
			return
		}
		e.preventDefault()
		taskRefs.value[focusedIndex.value]?.click(e)
	}
}

onMounted(() => {
	document.addEventListener('keydown', handleListNavigation)
})

onBeforeUnmount(() => {
	document.removeEventListener('keydown', handleListNavigation)
})
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

:deep(.single-task) {
	cursor: grab;

	&:active {
		cursor: grabbing;
	}
}

.list-view {
	padding-block-end: 1rem;

	:deep(.card) {
		margin-block-end: 0;
	}
}
</style>
