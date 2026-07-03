<template>
	<aside
		class="planner-sidebar"
		:class="{'is-drop-target': isDropTarget}"
		@dragover.prevent="isDropTarget = true"
		@dragleave="isDropTarget = false"
		@drop="onDrop"
	>
		<template v-if="overdueTasks.length > 0">
			<h3 class="sidebar-title is-overdue">
				{{ $t('planner.overdue') }}
				<span class="overdue-count">{{ overdueTasks.length }}</span>
			</h3>
			<ul class="task-list overdue-list">
				<li
					v-for="task in overdueTasks"
					:key="task.id"
					class="sidebar-task"
					:style="{'--task-color': taskColor(task)}"
					draggable="true"
					@dragstart="onDragStart($event, task)"
					@click="emit('openTask', task.id)"
				>
					<span class="task-title">{{ task.title }}</span>
					<span class="task-meta">
						<span class="task-overdue-date">{{ overdueDateLabel(task) }}</span>
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
					</span>
				</li>
			</ul>
		</template>

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
import {formatDate} from '@/helpers/time/formatDate'
import {PLANNER_SIDEBAR_SORTS, type PlannerSidebarSort} from './helpers/usePlannerTasks'
import FilterPopup from '@/components/project/partials/FilterPopup.vue'
import PriorityLabel from '@/components/tasks/partials/PriorityLabel.vue'
import {useProjectStore} from '@/stores/projects'
import {plannerTaskColor} from './helpers/taskColor'
import {overdueAnchor} from './helpers/overdue'

defineProps<{
	tasks: ITask[]
	overdueTasks: ITask[]
}>()

const emit = defineEmits<{
	openTask: [taskId: number]
	unschedule: [taskId: number]
}>()

const filter = defineModel<TaskFilterParams>('filter', {required: true})
const sort = defineModel<PlannerSidebarSort>('sort', {required: true})

const {t} = useI18n({useScope: 'global'})

// Typed Record so adding a sort to PLANNER_SIDEBAR_SORTS without a label (or
// vice versa) fails the type check instead of silently hiding the option.
const SORT_LABEL_KEYS: Record<PlannerSidebarSort, string> = {
	'none': 'planner.sortDefault',
	'priority:desc': 'sorting.options.priorityDesc',
	'priority:asc': 'sorting.options.priorityAsc',
	'title:asc': 'sorting.options.titleAsc',
	'title:desc': 'sorting.options.titleDesc',
	'created:desc': 'sorting.options.createdDesc',
	'created:asc': 'sorting.options.createdAsc',
	'percent_done:desc': 'sorting.options.percentDoneDesc',
	'percent_done:asc': 'sorting.options.percentDoneAsc',
	'random': 'planner.sortRandom',
}

const sortOptions = computed<{value: PlannerSidebarSort, label: string}[]>(
	() => PLANNER_SIDEBAR_SORTS.map(value => ({value, label: t(SORT_LABEL_KEYS[value])})),
)

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

function overdueDateLabel(task: ITask): string {
	const anchor = overdueAnchor(task)
	return anchor ? formatDate(anchor, 'll') : ''
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
	overflow-x: hidden;

	&.is-drop-target {
		border-color: var(--primary);
		background: var(--grey-100);
	}
}

.sidebar-title {
	font-size: .9rem;
	margin-block-end: .5rem;

	&.is-overdue {
		color: var(--danger);
	}
}

.overdue-count {
	display: inline-block;
	min-inline-size: 1.3em;
	padding: 0 .3em;
	border-radius: 1em;
	background: var(--danger);
	color: var(--white);
	font-size: .75rem;
	text-align: center;
}

.task-overdue-date {
	color: var(--danger);
	font-size: .75rem;
	white-space: nowrap;
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

.overdue-list {
	margin-block-end: 1.5rem;
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
	min-inline-size: 0;

	&:active {
		cursor: grabbing;
	}

	&:hover {
		background: var(--grey-100);
	}
}

.task-title {
	display: block;
	overflow-wrap: anywhere;
}

.task-meta {
	display: flex;
	align-items: center;
	flex-wrap: wrap;
	gap: .4rem;
	margin-block-start: .15rem;
	min-inline-size: 0;

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
