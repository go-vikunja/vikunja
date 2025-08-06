<template>
	<div class="task">
		<span>
			<span
				v-if="showProject && typeof project !== 'undefined'"
				v-tooltip="$t('task.detail.belongsToProject', {project: project.title})"
				class="task-project"
				:class="{'mie-2': task.hexColor !== ''}"
			>
				{{ project.title }}
			</span>

			<ColorBubble
				v-if="task.hexColor !== ''"
				:color="getHexColor(task.hexColor)"
				class="mie-1"
			/>

			<PriorityLabel
				:priority="task.priority"
				:done="task.done"
			/>

			<!-- Show any parent tasks to make it clear this task is a sub task of something -->
			<span
				v-if="typeof task.relatedTasks?.parenttask !== 'undefined'"
				class="parent-tasks"
			>
				<template v-for="(pt, i) in task.relatedTasks.parenttask">
					{{ pt.title }}<template v-if="(i + 1) < task.relatedTasks.parenttask.length">,&nbsp;</template>
				</template>
				&rsaquo;
			</span>
			{{ task.title }}
		</span>

		<Labels
			v-if="task.labels.length > 0"
			class="labels mis-2 mie-1"
			:labels="task.labels"
		/>

		<AssigneeList
			v-if="task.assignees.length > 0"
			:assignees="task.assignees"
			:avatar-size="20"
			class="mis-1"
			:inline="true"
		/>

		<span
			v-if="+new Date(task.dueDate) > 0"
			v-tooltip="formatDateLong(task.dueDate)"
			class="dueDate"
		>
			<time
				:datetime="formatISO(task.dueDate)"
				:class="{'overdue': task.dueDate <= new Date() && !task.done}"
				class="is-italic"
			>
				– {{ $t('task.detail.due', {at: formatDisplayDate(task.dueDate)}) }}
			</time>
		</span>

		<span>
			<span
				v-if="task.attachments.length > 0"
				class="project-task-icon"
			>
				<Icon icon="paperclip" />
			</span>
			<span
				v-if="task.description"
				class="project-task-icon"
			>
				<Icon icon="align-left" />
			</span>
			<span
				v-if="task.repeatAfter.amount > 0"
				class="project-task-icon"
			>
				<Icon icon="history" />
			</span>
		</span>

		<ChecklistSummary :task="task" />

		<progress
			v-if="task.percentDone > 0"
			class="progress is-small"
			:value="task.percentDone * 100"
			max="100"
		>
			{{ task.percentDone * 100 }}%
		</progress>
	</div>
</template>

<script setup lang="ts">
import {computed} from 'vue'

import {getHexColor} from '@/models/task'
import type {ITask} from '@/modelTypes/ITask'

import PriorityLabel from '@/components/tasks/partials/PriorityLabel.vue'
import Labels from '@/components/tasks/partials/Labels.vue'
import ChecklistSummary from '@/components/tasks/partials/ChecklistSummary.vue'

import ColorBubble from '@/components/misc/ColorBubble.vue'

import {formatDisplayDate, formatISO, formatDateLong} from '@/helpers/time/formatDate'

import {useProjectStore} from '@/stores/projects'
import AssigneeList from '@/components/tasks/partials/AssigneeList.vue'

const props = withDefaults(defineProps<{
	task: ITask,
	showProject?: boolean,
}>(), {
	showProject: false,
})

const projectStore = useProjectStore()

const project = computed(() => projectStore.projects[props.task.projectId])
</script>

<style lang="scss" scoped>
.task {
	display: flex;
	flex-wrap: wrap;
	transition: background-color $transition;
	align-items: center;
	cursor: pointer;
	border-radius: $radius;
	border: 2px solid transparent;

	text-overflow: ellipsis;
	word-wrap: break-word;
	word-break: break-word;
	//display: -webkit-box;
	hyphens: auto;
	-webkit-line-clamp: 4;
	-webkit-box-orient: vertical;
	overflow: hidden;

	//flex: 1 0 50%;

	.dueDate {
		display: inline-block;
		margin-inline-start: 5px;
	}

	.overdue {
		color: var(--danger);
	}

	.task-project {
		inline-size: auto;
		color: var(--grey-400);
		font-size: .9rem;
		white-space: nowrap;
	}

	.avatar {
		border-radius: 50%;
		vertical-align: bottom;
		margin-inline-start: .5rem;
		block-size: 21px;
		inline-size: 21px;
	}

	.project-task-icon {
		margin-inline-start: 6px;

		&:not(:first-of-type) {
			margin-inline-start: 8px;
		}

	}

	a {
		color: var(--text);
		transition: color ease $transition-duration;

		&:hover {
			color: var(--grey-900);
		}
	}

	.tasktext.done {
		text-decoration: line-through;
		color: var(--grey-500);
	}

	span.parent-tasks {
		color: var(--grey-500);
		inline-size: auto;
		margin-inline-start: .25rem;
	}
}
</style>
