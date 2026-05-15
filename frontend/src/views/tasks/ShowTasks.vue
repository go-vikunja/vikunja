<template>
	<div
		v-cy="'showTasks'"
		class="is-max-width-desktop has-text-start"
	>
		<h3 class="mbe-2 title">
			{{ pageTitle }}
		</h3>
		<Message
			v-if="filteredLabels.length > 0"
			class="label-filter-info mbe-2"
		>
			<i18n-t
				keypath="task.show.filterByLabel"
				tag="span"
				class="filter-label-text"
			>
				<template #label>
					<XLabel
						v-for="label in filteredLabels"
						:key="label.id"
						:label="label"
					/>
				</template>
			</i18n-t>
			<BaseButton
				v-tooltip="$t('task.show.clearLabelFilter')"
				class="clear-filter-button"
				@click="clearLabelFilter"
			>
				<Icon icon="times" />
			</BaseButton>
		</Message>
		<Message
			v-if="savedFilterIgnored"
			class="mbe-2"
		>
			{{ $t('task.show.savedFilterIgnored') }}
		</Message>
		<p
			v-if="!showAll"
			class="show-tasks-options"
		>
			<DatepickerWithRange @update:modelValue="setDate">
				<template #trigger="{toggle}">
					<XButton
						variant="primary"
						:shadow="false"
						class="mbe-2"
						@click.prevent.stop="toggle()"
					>
						{{ $t('task.show.select') }}
					</XButton>
				</template>
			</DatepickerWithRange>
			<FancyCheckbox
				:model-value="showNulls"
				class="mie-2"
				@update:modelValue="setShowNulls"
			>
				{{ $t('task.show.noDates') }}
			</FancyCheckbox>
			<FancyCheckbox
				:model-value="showOverdue"
				@update:modelValue="setShowOverdue"
			>
				{{ $t('task.show.overdue') }}
			</FancyCheckbox>
		</p>
		<p class="show-tasks-options">
			<FancyCheckbox
				v-model="hierarchical"
				@update:modelValue="saveHierarchical"
			>
				{{ $t('task.show.hierarchical') }}
			</FancyCheckbox>
		</p>
		<template v-if="!loading && (!tasks || tasks.length === 0) && showNothingToDo">
			<h3 class="has-text-centered mbs-6">
				{{ $t('task.show.noTasks') }}
			</h3>
			<LlamaCool class="llama-cool" />
		</template>

		<Card
			v-if="hasTasks"
			:padding="false"
			class="has-overflow"
			:has-content="false"
			:loading="loading"
		>
			<TaskTreeDraggable
				:tasks="displayedTasks"
				:all-tasks="allTasksWithSubtasks"
				:hierarchical="hierarchical"
				:can-mark-task-as-done="canMarkTaskAsDone"
				:disabled="!canDragTasks"
				:show-project="true"
				:dragging="drag"
				@taskUpdated="updateTasks"
				@dragStart="handleDragStart"
				@drop="handleTaskTreeDrop"
				@updateList="updateTaskTreeList"
			/>
		</Card>
		<div
			v-else
			:class="{ 'is-loading': loading}"
			class="spinner"
		/>
	</div>
</template>

<script setup lang="ts">
import {computed, ref, watch, watchEffect} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {useI18n} from 'vue-i18n'

import {formatDate} from '@/helpers/time/formatDate'
import {setTitle} from '@/helpers/setTitle'

import BaseButton from '@/components/base/BaseButton.vue'
import Icon from '@/components/misc/Icon'
import Message from '@/components/misc/Message.vue'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import DatepickerWithRange from '@/components/date/DatepickerWithRange.vue'
import XLabel from '@/components/tasks/partials/Label.vue'
import TaskTreeDraggable, {
	type TaskTreeDropEvent,
	type TaskTreeListUpdateEvent,
} from '@/components/tasks/partials/TaskTreeDraggable.vue'
import {DATE_RANGES} from '@/components/date/dateRanges'
import LlamaCool from '@/assets/llama-cool.svg?component'
import type {ITask} from '@/modelTypes/ITask'
import {PROJECT_VIEW_KINDS} from '@/modelTypes/IProjectView'
import {useAuthStore} from '@/stores/auth'
import {useTaskStore} from '@/stores/tasks'
import {useProjectStore} from '@/stores/projects'
import {useLabelStore} from '@/stores/labels'
import type {TaskFilterParams} from '@/services/taskCollection'
import TaskCollectionService from '@/services/taskCollection'
import {PERMISSIONS} from '@/constants/permissions'
import {useTaskDragToProject} from '@/composables/useTaskDragToProject'
import {calculateItemPosition} from '@/helpers/calculateItemPosition'
import TaskPositionService from '@/services/taskPosition'
import TaskPositionModel from '@/models/taskPosition'
import TaskRelationService from '@/services/taskRelation'
import TaskRelationModel from '@/models/taskRelation'
import {RELATION_KIND} from '@/types/IRelationKind'
import {error} from '@/message'

const props = withDefaults(defineProps<{
	dateFrom?: Date | string,
	dateTo?: Date | string,
	showNulls?: boolean,
	showOverdue?: boolean,
	labelIds?: string[],
}>(), {
	showNulls: false,
	showOverdue: false,
	dateFrom: undefined,
	dateTo: undefined,
	labelIds: undefined,
})

const emit = defineEmits<{
	'tasksLoaded': true,
	'clearLabelFilter': void,
}>()

const authStore = useAuthStore()
const taskStore = useTaskStore()
const projectStore = useProjectStore()
const labelStore = useLabelStore()
const {handleTaskDropToProject} = useTaskDragToProject()

const route = useRoute()
const router = useRouter()
const {t} = useI18n({useScope: 'global'})

const tasks = ref<ITask[]>([])
const showNothingToDo = ref<boolean>(false)
const taskCollectionService = ref(new TaskCollectionService())
const taskPositionService = ref(new TaskPositionService())
const taskRelationService = ref(new TaskRelationService())
const hierarchical = ref(localStorage.getItem('showTasksHierarchical') === 'true')
const drag = ref(false)

setTimeout(() => showNothingToDo.value = true, 100)

const showAll = computed(() => typeof props.dateFrom === 'undefined' || typeof props.dateTo === 'undefined')

const filteredLabels = computed(() => {
	if (!props.labelIds || props.labelIds.length === 0) {
		return []
	}
	return props.labelIds
		.map(id => labelStore.getLabelById(Number(id)))
		.filter(label => label !== null && label !== undefined)
})

const savedFilterIgnored = computed(() => {
	return filteredLabels.value.length > 0
		&& filterIdUsedOnOverview.value
		&& typeof projectStore.projects[filterIdUsedOnOverview.value] !== 'undefined'
})

const pageTitle = computed(() => {
	// We need to define "key" because it is the first parameter in the array and we need the second
	const predefinedRange = Object.entries(DATE_RANGES)
		.find(([, value]) => props.dateFrom === value[0] && props.dateTo === value[1])
		?.[0]
	if (typeof predefinedRange !== 'undefined') {
		return t(`input.datepickerRange.ranges.${predefinedRange}`)
	}

	return showAll.value
		? t('task.show.titleCurrent')
		: t('task.show.fromuntil', {
			from: formatDate(props.dateFrom, 'LL'),
			until: formatDate(props.dateTo, 'LL'),
		})
})
const hasTasks = computed(() => tasks.value && tasks.value.length > 0)

// Build a flat list including embedded subtasks for hierarchical rendering.
// Tasks from tasks.value take priority over embedded stubs (full data wins).
const allTasksWithSubtasks = computed((): ITask[] => {
	if (!hierarchical.value) return tasks.value
	const map = new Map<number, ITask>()
	tasks.value.forEach(t => addEmbeddedSubtasks(t, map))
	tasks.value.forEach(t => map.set(t.id, t))
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

// Top-level tasks: a task is hidden if any of its parents exists in the full task tree.
// relatedTasks.parenttask is always populated by the API for subtasks.
// We check against allTasksWithSubtasks (not just tasks.value) to cover all levels.
const displayedTasks = computed((): ITask[] => {
	if (!hierarchical.value) return tasks.value
	const allTaskIds = new Set(allTasksWithSubtasks.value.map(t => t.id))
	return tasks.value.filter(t => {
		const parentIds = (t.relatedTasks?.parenttask ?? []).map((p: ITask) => p.id)
		return parentIds.length === 0 || !parentIds.some(pid => allTaskIds.has(pid))
	})
})
const userAuthenticated = computed(() => authStore.authenticated)
const loading = computed(() => taskStore.isLoading || taskCollectionService.value.loading)
const filterIdUsedOnOverview = computed(() => authStore.settings?.frontendSettings?.filterIdUsedOnOverview)
const canDragTasks = computed(() => tasks.value.some(canMarkTaskAsDone))

function canMarkTaskAsDone(task: ITask) {
	return (projectStore.projects[task.projectId]?.maxPermission ?? 0) > PERMISSIONS.READ
}

interface dateStrings {
	dateFrom: string,
	dateTo: string,
}

function setDate(dates: dateStrings) {
	router.push({
		name: route.name as string,
		query: {
			from: dates.dateFrom ?? props.dateFrom,
			to: dates.dateTo ?? props.dateTo,
			showOverdue: props.showOverdue ? 'true' : 'false',
			showNulls: props.showNulls ? 'true' : 'false',
		},
	})
}

function setShowOverdue(show: boolean) {
	router.push({
		name: route.name as string,
		query: {
			...route.query,
			showOverdue: show ? 'true' : 'false',
		},
	})
}

function setShowNulls(show: boolean) {
	router.push({
		name: route.name as string,
		query: {
			...route.query,
			showNulls: show ? 'true' : 'false',
		},
	})
}

function saveHierarchical(value: boolean) {
	localStorage.setItem('showTasksHierarchical', String(value))
}

function clearLabelFilter() {
	emit('clearLabelFilter')
}

function handleDragStart(e: { item: HTMLElement }) {
	drag.value = true
	const taskId = parseInt(e.item.dataset.taskId ?? '', 10)
	const task = allTasksWithSubtasks.value.find(t => t.id === taskId)

	if (task) {
		taskStore.setDraggedTask(task)
	} else {
		taskStore.setDraggedTask(null)
		drag.value = false
	}
}

function removeTaskFromOverview(task: ITask) {
	tasks.value = tasks.value.filter(t => t.id !== task.id)
	tasks.value.forEach(t => {
		if (typeof t.relatedTasks?.subtask !== 'undefined') {
			t.relatedTasks.subtask = t.relatedTasks.subtask.filter(subtask => subtask.id !== task.id)
		}
	})
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
		const updatedTaskIds = new Set(updatedTasks.map(({id}) => id))
		tasks.value = [
			...updatedTasks,
			...tasks.value.filter(({id}) => !updatedTaskIds.has(id)),
		]
		return
	}

	const parent = findTaskById(parentTaskId)
	if (parent) {
		parent.relatedTasks.subtask = updatedTasks
	}
}

function getTaskTreeSiblings(parentTaskId: ITask['id'] | null): ITask[] {
	if (parentTaskId === null) {
		return displayedTasks.value
	}

	return findTaskById(parentTaskId)?.relatedTasks?.subtask ?? []
}

function resolveListViewId(task: ITask) {
	return projectStore.projects[task.projectId]?.views.find(({viewKind}) => viewKind === PROJECT_VIEW_KINDS.LIST)?.id
}

async function updateTaskPosition(task: ITask, siblings: ITask[], index: number) {
	const projectViewId = resolveListViewId(task)
	if (typeof projectViewId === 'undefined') {
		return
	}

	const taskBefore = siblings[index - 1] ?? null
	const taskAfter = siblings[index + 1] ?? null
	const position = calculateItemPosition(taskBefore?.position ?? null, taskAfter?.position ?? null)

	await taskPositionService.value.update(new TaskPositionModel({
		position,
		projectViewId,
		taskId: task.id,
	}))
	task.position = position
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

async function handleTaskTreeDrop(e: TaskTreeDropEvent) {
	drag.value = false
	const {moved} = await handleTaskDropToProject(e, removeTaskFromOverview)
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
	} catch (e) {
		error(e)
		await loadPendingTasks(props.dateFrom, props.dateTo, filterIdUsedOnOverview.value)
	}
}

async function loadPendingTasks(from: Date|string, to: Date|string, filterId: number | null | undefined) {
	// FIXME: HACK! This should never happen.
	// Since this route is authentication only, users would get an error message if they access the page unauthenticated.
	// Since this component is mounted as the home page before unauthenticated users get redirected
	// to the login page, they will almost always see the error message.
	if (!userAuthenticated.value) {
		return
	}

	const params: TaskFilterParams = {
		sort_by: ['due_date', 'id'],
		order_by: ['asc', 'desc'],
		filter: 'done = false',
		filter_include_nulls: props.showNulls,
		s: '',
		expand: ['subtasks', 'comment_count', 'is_unread'],
	}

	if (!showAll.value) {

		params.filter += ` && due_date < '${to instanceof Date ? to.toISOString() : to}'`

		// NOTE: Ideally we could also show tasks with a start or end date in the specified range, but the api
		//       is not capable (yet) of combining multiple filters with 'and' and 'or'.

		if (!props.showOverdue) {
			params.filter += ` && due_date > '${from instanceof Date ? from.toISOString() : from}'`
		}
	}

	// Add label filtering
	if (props.labelIds && props.labelIds.length > 0) {
		const labelFilter = `labels in ${props.labelIds.join(', ')}`
		params.filter += params.filter ? ` && ${labelFilter}` : labelFilter
	}

	let projectId = null
	if (showAll.value && filterId && typeof projectStore.projects[filterId] !== 'undefined'
		&& (!props.labelIds || props.labelIds.length === 0)) {
		projectId = filterId
	}

	tasks.value = await taskStore.loadTasks(params, projectId)
	emit('tasksLoaded', true)
}

// FIXME: this modification should happen in the store
function updateTasks(updatedTask: ITask) {
	for (let t = 0; t < tasks.value.length; t++) {
		if (tasks.value[t].id === updatedTask.id) {
			tasks.value[t] = updatedTask
			// Move the task to the end of the done tasks if it is now done
			if (updatedTask.done) {
				tasks.value.splice(t, 1)
				tasks.value.push(updatedTask)
			}
			break
		}
	}
}

// Use watch instead of watchEffect to prevent reloading tasks when unrelated settings change.
// watchEffect would track all reactive dependencies accessed inside loadPendingTasks,
// which includes the entire settings object. When sidebarWidth changes, the settings
// object is replaced, triggering the watchEffect even though filterIdUsedOnOverview
// hasn't changed. Using watch with explicit dependencies and immediate:true gives us
// the same behavior but only triggers when these specific values actually change.
watch(
	[() => props.dateFrom, () => props.dateTo, filterIdUsedOnOverview],
	([from, to, filterId]) => loadPendingTasks(from, to, filterId),
	{immediate: true},
)
watchEffect(() => setTitle(pageTitle.value))
</script>

<style lang="scss" scoped>
.show-tasks-options {
	display: flex;
	flex-direction: column;
}

.llama-cool {
	margin: 3rem auto 0;
	display: block;
}

.label-filter-info {
	margin-block-end: 1rem;
	
	.clear-filter-button {
		margin-inline-start: auto;
		padding: 0.25rem 0.5rem;
		
		&:hover {
			color: var(--danger);
		}
	}

	:deep(.message.info) {
		inline-size: 100%;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
	}
}
</style>
