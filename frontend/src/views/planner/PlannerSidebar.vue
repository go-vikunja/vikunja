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

		<FilterPopup
			v-model="filter"
			class="sidebar-filter"
		/>

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
				<span
					v-if="task.dueDate"
					class="task-due"
				>
					<Icon :icon="['far', 'calendar-alt']" />
					{{ formatDueDate(task.dueDate) }}
				</span>
			</li>
		</ul>
	</aside>
</template>

<script setup lang="ts">
import {ref} from 'vue'
import dayjs from 'dayjs'

import type {ITask} from '@/modelTypes/ITask'
import type {TaskFilterParams} from '@/services/taskCollection'
import FilterPopup from '@/components/project/partials/FilterPopup.vue'
import {useProjectStore} from '@/stores/projects'

defineProps<{
	tasks: ITask[]
}>()

const emit = defineEmits<{
	openTask: [taskId: number]
	unschedule: [taskId: number]
}>()

const filter = defineModel<TaskFilterParams>('filter', {required: true})

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
	const hex = projectStore.projects[task.projectId]?.hexColor || task.hexColor
	if (!hex) {
		return 'var(--primary)'
	}
	return hex.startsWith('#') ? hex : `#${hex}`
}

function onDragStart(event: DragEvent, task: ITask) {
	event.dataTransfer?.setData('text/plain', String(task.id))
	if (event.dataTransfer) {
		event.dataTransfer.effectAllowed = 'move'
	}
}

function formatDueDate(date: Date): string {
	return dayjs(date).format('ll')
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

.sidebar-filter {
	margin-block-end: .75rem;
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
	font-size: .8rem;

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

.task-due {
	display: block;
	margin-block-start: .15rem;
	font-size: .7rem;
	color: var(--grey-500);
}
</style>
