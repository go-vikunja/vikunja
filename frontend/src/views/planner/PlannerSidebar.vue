<template>
	<aside
		class="planner-sidebar"
		:class="{'is-drop-target': isDropTarget}"
		@dragover.prevent="isDropTarget = true"
		@dragleave="isDropTarget = false"
		@drop="onDrop"
	>
		<h3 class="sidebar-title">
			{{ $t('planner.unscheduled') }}
		</h3>

		<div class="sidebar-controls">
			<FilterPopup v-model="filter" />
			<div class="select is-small sort-select">
				<select
					v-model="sort"
					:aria-label="$t('misc.sortBy')"
				>
					<option
						v-for="o in sortOptions"
						:key="o.value"
						:value="o.value"
					>
						{{ o.label }}
					</option>
				</select>
			</div>
		</div>

		<p
			v-if="!tasks.length"
			class="no-tasks"
		>
			{{ $t('planner.noUnscheduled') }}
		</p>

		<ul class="task-list">
			<li
				v-for="task in tasks"
				:key="task.id"
				class="sidebar-task"
				:style="{'--task-color': taskColor(task)}"
				draggable="true"
				@dragstart="onDragStart($event, task)"
				@click="emit('openTask', task.id)"
			>
				<span class="task-title">{{ task.title }}</span>
				<span class="task-meta">
					<span
						v-if="projectName(task)"
						class="task-project"
					>
						{{ projectName(task) }}
					</span>
					<PriorityLabel
						:priority="task.priority"
						:done="task.done"
					/>
					<span
						v-if="task.percentDone > 0"
						class="task-percent"
					>{{ Math.round(task.percentDone * 100) }}%</span>
				</span>
			</li>
		</ul>
	</aside>
</template>

<script setup lang="ts">
import {computed, ref} from 'vue'

import {useI18n} from 'vue-i18n'

import type {ITask} from '@/modelTypes/ITask'
import type {TaskFilterParams} from '@/services/taskCollection'
import type {PlannerSidebarSort} from './helpers/usePlannerTasks'
import FilterPopup from '@/components/project/partials/FilterPopup.vue'
import PriorityLabel from '@/components/tasks/partials/PriorityLabel.vue'
import {useProjectStore} from '@/stores/projects'
import {plannerTaskColor} from './helpers/taskColor'

defineProps<{
	tasks: ITask[]
}>()

const emit = defineEmits<{
	openTask: [taskId: number]
	unschedule: [taskId: number]
}>()

const filter = defineModel<TaskFilterParams>('filter', {required: true})
const sort = defineModel<PlannerSidebarSort>('sort', {required: true})

const {t} = useI18n({useScope: 'global'})

// Curated for unscheduled tasks: no date sorts (those tasks live in the grid),
// no manual/position (errors cross-project), plus a client-side random shuffle.
const sortOptions = computed<{value: PlannerSidebarSort, label: string}[]>(() => [
	{value: 'none', label: t('planner.sortDefault')},
	{value: 'priority:desc', label: t('sorting.options.priorityDesc')},
	{value: 'priority:asc', label: t('sorting.options.priorityAsc')},
	{value: 'title:asc', label: t('sorting.options.titleAsc')},
	{value: 'title:desc', label: t('sorting.options.titleDesc')},
	{value: 'created:desc', label: t('sorting.options.createdDesc')},
	{value: 'created:asc', label: t('sorting.options.createdAsc')},
	{value: 'percent_done:desc', label: t('sorting.options.percentDoneDesc')},
	{value: 'percent_done:asc', label: t('sorting.options.percentDoneAsc')},
	{value: 'random', label: t('planner.sortRandom')},
])

const isDropTarget = ref(false)

function onDrop(event: DragEvent) {
	isDropTarget.value = false
	const taskId = Number(event.dataTransfer?.getData('text/plain'))
	if (taskId) {
		emit('unschedule', taskId)
	}
}

const projectStore = useProjectStore()

function taskColor(task: ITask): string {
	return plannerTaskColor(task.hexColor, projectStore.projects[task.projectId]?.hexColor)
}

function projectName(task: ITask): string {
	return projectStore.projects[task.projectId]?.title ?? ''
}

function onDragStart(event: DragEvent, task: ITask) {
	event.dataTransfer?.setData('text/plain', String(task.id))
	if (event.dataTransfer) {
		event.dataTransfer.effectAllowed = 'move'
	}
}
</script>

<style lang="scss" scoped>
.planner-sidebar {
	flex: 0 0 16rem;
	inline-size: 16rem;
	display: flex;
	flex-direction: column;
	min-block-size: 0;
	border: 1px solid var(--grey-200);
	border-radius: $radius;
	background: var(--white);
	padding: .75rem;
	overflow-y: auto;

	&.is-drop-target {
		border-color: var(--primary);
		background: var(--grey-100);
	}
}

.sidebar-title {
	font-size: .9rem;
	margin-block-end: .5rem;
}

.sidebar-controls {
	display: flex;
	align-items: center;
	gap: .35rem;
	margin-block-end: .75rem;
}

.sort-select {
	flex: 1 1 auto;
	min-inline-size: 0;

	select {
		inline-size: 100%;
	}
}

.no-tasks {
	font-size: .8rem;
	color: var(--grey-500);
}

.task-list {
	list-style: none;
	margin: 0;
	display: flex;
	flex-direction: column;
	gap: .35rem;
}

.sidebar-task {
	cursor: grab;
	border: 1px solid var(--grey-200);
	border-inline-start: 3px solid var(--task-color);
	border-radius: 4px;
	padding: .4rem .5rem;
	background: var(--white);
	font-size: .95rem;

	&:active {
		cursor: grabbing;
	}

	&:hover {
		background: var(--grey-100);
	}
}

.task-title {
	display: block;
}

.task-meta {
	display: flex;
	align-items: center;
	gap: .4rem;
	margin-block-start: .15rem;

	// Tame PriorityLabel's oversized Bulma .icon box and centre it inline.
	:deep(.priority-label) {
		display: inline-flex;
		align-items: center;
		font-size: .8rem;
		line-height: 1.4;

		.icon {
			block-size: auto;
			margin: 0;
			padding: 0 .15rem 0 0;
		}

		svg {
			block-size: .85em;
			inline-size: auto;
			vertical-align: middle;
		}
	}
}

.task-project {
	font-size: .8rem;
	color: var(--grey-500);
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.task-percent {
	flex: 0 0 auto;
	font-size: .8rem;
	color: var(--grey-500);
	font-variant-numeric: tabular-nums;
}
</style>
