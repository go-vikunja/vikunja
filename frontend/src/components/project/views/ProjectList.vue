<template>
	<ProjectWrapper
		class="project-list"
		:is-loading-project="isLoadingProject"
		:project-id="projectId"
		:view-id
	>
		<template #header>
			<div class="filter-container">
				<SortPopup
					v-model="sortByParam"
				/>
				<FilterPopup
					v-if="!isSavedFilter(project)"
					v-model="params"
					:view-id="viewId"
					:project-id="projectId"
					@update:modelValue="loadTasks()"
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
						v-if="!project?.isArchived && canWrite"
						ref="addTaskRef"
						class="list-view__add-task d-print-none"
						:default-position="firstNewPosition"
						@taskAdded="updateTaskList"
					/>

					<Nothing v-if="ctaVisible && tasks.length === 0 && !loading">
						{{ $t('project.list.empty') }}
						<ButtonLink
							v-if="project?.id > 0 && canWrite"
							@click="focusNewTaskInput()"
						>
							{{ $t('project.list.newTaskCta') }}
						</ButtonLink>
					</Nothing>

					<TaskTreeDraggable
						v-if="tasks && tasks.length > 0"
						:tasks="tasks"
						:all-tasks="allTasksWithSubtasks"
						:hierarchical="true"
						:disabled="!canDragTasks"
						:can-mark-task-as-done="canMarkTaskAsDone"
						:show-drag-handle="canDragTasks"
						:dragging="drag"
						:task-ref-setter="setTaskTreeRef"
						:handle="dragHandle"
						:delay-on-touch-only="!isTouchDevice"
						:delay="isTouchDevice ? 0 : 1000"
						:transition-group="true"
						@dragStart="handleDragStart"
						@drop="saveTaskTreeDrop"
						@updateList="updateTaskTreeList"
						@taskUpdated="updateTasks"
					/>

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

import ProjectWrapper from '@/components/project/ProjectWrapper.vue'
import ButtonLink from '@/components/misc/ButtonLink.vue'
import AddTask from '@/components/tasks/AddTask.vue'
import SingleTaskInProject from '@/components/tasks/partials/SingleTaskInProject.vue'
import TaskTreeDraggable, {
	type TaskTreeDropEvent,
	type TaskTreeListUpdateEvent,
} from '@/components/tasks/partials/TaskTreeDraggable.vue'
import FilterPopup from '@/components/project/partials/FilterPopup.vue'
import Nothing from '@/components/misc/Nothing.vue'
import Pagination from '@/components/misc/Pagination.vue'
import SortPopup from '@/components/project/partials/SortPopup.vue'

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
import TaskRelationService from '@/services/taskRelation'
import TaskRelationModel from '@/models/taskRelation'
import {RELATION_KIND} from '@/types/IRelationKind'
import {error} from '@/message'

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
const taskRelationService = ref(new TaskRelationService())

// Saved filter composable for accessing filter data
const _savedFilter = useSavedFilter(() => isSavedFilter({id: projectId.value}) ? projectId.value : undefined).filter

const tasks = ref<ITask[]>([])
watch(
	allTasks,
	() => {
		const isFiltered = isSavedFilter({id: projectId.value})
		tasks.value = ([...allTasks.value]).filter(t => shouldShowTaskInListView(t, allTasks.value, isFiltered))
	},
)

const isPositionSorting = computed(() => 'position' in sortByParam.value)

const allTasksWithSubtasks = computed((): ITask[] => {
	const map = new Map<number, ITask>()
	allTasks.value.forEach(t => addEmbeddedSubtasks(t, map))
	allTasks.value.forEach(t => map.set(t.id, t))
	return [...map.values()]
})

function addEmbeddedSubtasks(task: ITask, map: Map<number, ITask>) {
	(task.relatedTasks?.subtask ?? []).forEach(subtask => {
		if (map.has(subtask.id)) {
			return
		}

		map.set(subtask.id, subtask)
		addEmbeddedSubtasks(subtask, map)
	})
}

const firstNewPosition = computed(() => {
	if (tasks.value.length === 0) {
		return 0
	}

	return calculateItemPosition(null, tasks.value[0].position)
})

const baseStore = useBaseStore()
const taskStore = useTaskStore()
const {handleTaskDropToProject} = useTaskDragToProject()
const project = computed(() => baseStore.currentProject)

const canWrite = computed(() => {
	return project.value?.maxPermission > Permissions.READ && project.value?.id > 0
})

const isPseudoProject = computed(() => (project.value && isSavedFilter(project.value)) || project.value?.id === -1)

function canMarkTaskAsDone() {
	return canWrite.value || Boolean(isPseudoProject.value)
}

onMounted(async () => {
	await nextTick()
	ctaVisible.value = true
})

const canDragTasks = computed(() => canWrite.value || isSavedFilter(project.value))

const isTouchDevice = ref(false)
if (typeof window !== 'undefined') {
	isTouchDevice.value = !window.matchMedia('(hover: hover) and (pointer: fine)').matches
}
const dragHandle = computed(() => isTouchDevice.value ? '.handle' : undefined)

const addTaskRef = ref<typeof AddTask | null>(null)

function focusNewTaskInput() {
	addTaskRef.value?.focusTaskInput()
}

function updateTaskList(task: ITask) {
	if (!isPositionSorting.value) {
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

	for (let t = 0; t < tasks.value.length; t++) {
		if (tasks.value[t].id === updatedTask.id) {
			tasks.value[t] = updatedTask
			break
		}
	}
}

function handleDragStart(e: { item: HTMLElement }) {
	drag.value = true
	const taskId = parseInt(e.item.dataset.taskId ?? '', 10)
	const task = findTaskById(taskId)

	if (task) {
		taskStore.setDraggedTask(task)
	} else {
		taskStore.setDraggedTask(null)
	}
}

function findTaskById(taskId: ITask['id'], taskList: ITask[] = allTasksWithSubtasks.value): ITask | undefined {
	for (const task of taskList) {
		if (task.id === taskId) {
			return task
		}

		const found = findTaskById(taskId, task.relatedTasks?.subtask ?? [])
		if (found) {
			return found
		}
	}
}

function updateTaskTreeList({parentTaskId, tasks: updatedTasks}: TaskTreeListUpdateEvent) {
	if (parentTaskId === null) {
		tasks.value = updatedTasks
		return
	}

	const parent = findTaskById(parentTaskId)
	if (parent) {
		parent.relatedTasks.subtask = updatedTasks
	}
}

function getTaskTreeSiblings(parentTaskId: ITask['id'] | null): ITask[] {
	if (parentTaskId === null) {
		return tasks.value
	}

	return findTaskById(parentTaskId)?.relatedTasks?.subtask ?? []
}

function removeTaskFromTree(task: ITask) {
	tasks.value = tasks.value.filter(t => t.id !== task.id)
	allTasksWithSubtasks.value.forEach(t => {
		if (typeof t.relatedTasks?.subtask !== 'undefined') {
			t.relatedTasks.subtask = t.relatedTasks.subtask.filter(subtask => subtask.id !== task.id)
		}
	})
}

async function updateTaskParent(task: ITask, oldParentTaskId: ITask['id'] | null, newParentTaskId: ITask['id'] | null) {
	if (oldParentTaskId === newParentTaskId) {
		return
	}

	if (oldParentTaskId !== null) {
		await taskRelationService.value.delete(new TaskRelationModel({
			taskId: oldParentTaskId,
			otherTaskId: task.id,
			relationKind: RELATION_KIND.SUBTASK,
		}))
	}

	if (newParentTaskId !== null) {
		await taskRelationService.value.create(new TaskRelationModel({
			taskId: newParentTaskId,
			otherTaskId: task.id,
			relationKind: RELATION_KIND.SUBTASK,
		}))
		task.relatedTasks.parenttask = [findTaskById(newParentTaskId)].filter((task): task is ITask => Boolean(task))
	} else {
		task.relatedTasks.parenttask = []
	}
}

async function updateTaskPosition(task: ITask, siblings: ITask[], index: number) {
	const taskBefore = siblings[index - 1] ?? null
	const taskAfter = siblings[index + 1] ?? null
	const position = calculateItemPosition(taskBefore?.position ?? null, taskAfter?.position ?? null)

	await taskPositionService.value.update(new TaskPositionModel({
		position,
		projectViewId: props.viewId,
		taskId: task.id,
	}))
	task.position = position
}

async function saveTaskTreeDrop(e: TaskTreeDropEvent) {
	drag.value = false

	// Check if dropped on a sidebar project
	const {moved} = await handleTaskDropToProject(e, removeTaskFromTree)

	if (moved) {
		return
	}

	const task = findTaskById(e.taskId)
	if (!task) {
		return
	}

	try {
		await updateTaskParent(task, e.oldParentTaskId, e.newParentTaskId)
		await updateTaskPosition(task, getTaskTreeSiblings(e.newParentTaskId), e.newIndex)

		if (!isPositionSorting.value) {
			sortByParam.value = {position: 'asc'}
		}
	} catch (e) {
		error(e)
		await loadTasks(false)
	}
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

function isSingleTaskComponent(el: unknown): el is InstanceType<typeof SingleTaskInProject> {
	return el !== null &&
		typeof el === 'object' &&
		'focus' in el &&
		'click' in el
}

function setTaskTreeRef(el: unknown, index: number) {
	setTaskRef(isSingleTaskComponent(el) ? el : null, index)
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

	if (e.code === 'KeyJ') {
		e.preventDefault()
		focusTask(Math.min(focusedIndex.value + 1, tasks.value.length - 1))
		return
	}

	if (e.code === 'KeyK') {
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

	if (e.code === 'Enter') {
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
.filter-container {
	display: flex;
	align-items: center;
	gap: .5rem;

	:deep(.popup) {
		inset-block-start: 3rem;
		inset-inline-end: 0;
		max-inline-size: 300px;
	}
}

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

:deep(.single-task .handle) {
	cursor: grab;
	margin-inline-end: .25rem;
	color: var(--grey-400);
}

@media (hover: hover) and (pointer: fine) {
	:deep(.single-task .handle) {
		display: none;
	}
}

:deep(.tasks:not(.dragging-disabled) .single-task) {
	cursor: grab;
	-webkit-touch-callout: none;
	user-select: none;
	touch-action: manipulation;

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
