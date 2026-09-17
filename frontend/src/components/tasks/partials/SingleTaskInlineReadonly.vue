<template>
	<div
		class="task"
		:data-is-overdue="task.due_date <= new Date() && !task.done || undefined"
	>
		<span>
			<span
				v-if="showProject && typeof project !== 'undefined'"
				v-tooltip="$t('task.detail.belongsToProject', {project: project.title})"
				class="task-project"
				:class="{'mie-2': task.hex_color !== ''}"
			>
				{{ project.title }}
			</span>

			<ColorBubble
				v-if="task.hex_color !== ''"
				:color="getHexColor(task.hex_color)"
				class="mie-1"
			/>

			<PriorityLabel
				:priority="task.priority"
				:done="task.done"
			/>

			<!-- Show any parent tasks to make it clear this task is a sub task of something -->
			<span
				v-if="typeof task.related_tasks?.parenttask !== 'undefined'"
				class="parent-tasks"
			>
				<template v-for="(pt, i) in task.related_tasks.parenttask">
					{{ pt.title }}<template v-if="(i + 1) < task.related_tasks.parenttask.length">,&nbsp;</template>
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
			v-if="+new Date(task.due_date) > 0"
			v-tooltip="formatDateLong(task.due_date)"
			class="dueDate"
		>
			<time
				:datetime="formatISO(task.due_date)"
				class="is-italic"
			>
				– {{ $t('task.detail.due', {at: formatDisplayDate(task.due_date)}) }}
			</time>
		</span>

		<span>
			<span
				v-if="task.attachments.length > 0"
				class="project-task-icon"
				role="img"
				:aria-label="$t('task.attributes.attachment', task.attachments.length)"
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
				v-if="task.repeat_after.amount > 0"
				class="project-task-icon"
			>
				<Icon icon="history" />
			</span>
		</span>

		<ChecklistSummary :task="task" />

		<progress
			v-if="task.percent_done > 0"
			class="progress is-small"
			:value="task.percent_done * 100"
			max="100"
		>
			{{ task.percent_done * 100 }}%
		</progress>
	</div>
</template>

<script setup lang="ts">
import {computed} from 'vue'

import {getHexColor} from '@/helpers/task'
import type {Task as ITask} from '@/client/generated'

import PriorityLabel from '@/components/tasks/partials/PriorityLabel.vue'
import Labels from '@/components/tasks/partials/Labels.vue'
import ChecklistSummary from '@/components/tasks/partials/ChecklistSummary.vue'

import ColorBubble from '@/components/misc/ColorBubble.vue'

import {formatDisplayDate, formatISO, formatDateLong} from '@/helpers/time/formatDate'

import {useProjects} from '@/composables/useProjects'
import AssigneeList from '@/components/tasks/partials/AssigneeList.vue'

const props = withDefaults(defineProps<{
	task: ITask,
	showProject?: boolean,
}>(), {
	showProject: false,
})

const projectList = useProjects()

const project = computed(() => projectList.projects[props.task.project_id])
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

	.due_date {
		display: inline-block;
		margin-inline-start: 5px;
	}

	&[data-is-overdue] .due_date {
		color: var(--danger-text);
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
