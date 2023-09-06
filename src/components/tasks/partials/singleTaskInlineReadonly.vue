<template>
	<div class="task">
		<span>
			<span
				v-if="showProject && typeof project !== 'undefined'"
				class="task-project"
				:class="{'mr-2': task.hexColor !== ''}"
				v-tooltip="$t('task.detail.belongsToProject', {project: project.title})"
			>
				{{ project.title }}
			</span>

			<ColorBubble
				v-if="task.hexColor !== ''"
				:color="getHexColor(task.hexColor)"
				class="mr-1"
			/>

			<priority-label :priority="task.priority" :done="task.done"/>

			<!-- Show any parent tasks to make it clear this task is a sub task of something -->
			<span class="parent-tasks" v-if="typeof task.relatedTasks?.parenttask !== 'undefined'">
				<template v-for="(pt, i) in task.relatedTasks.parenttask">
					{{ pt.title }}<template v-if="(i + 1) < task.relatedTasks.parenttask.length">,&nbsp;</template>
				</template>
				&rsaquo;
			</span>
			{{ task.title }}
		</span>

		<labels
			v-if="task.labels.length > 0"
			class="labels ml-2 mr-1"
			:labels="task.labels"
		/>

		<assignee-list
			v-if="task.assignees.length > 0"
			:assignees="task.assignees"
			:avatar-size="20"
			class="ml-1"
			:inline="true"
		/>

		<span
			v-if="+new Date(task.dueDate) > 0"
			class="dueDate"
			v-tooltip="formatDateLong(task.dueDate)"
		>
			<time
				:datetime="formatISO(task.dueDate)"
				:class="{'overdue': task.dueDate <= new Date() && !task.done}"
				class="is-italic"
			>
				– {{ $t('task.detail.due', {at: formatDateSince(task.dueDate)}) }}
			</time>
		</span>

		<span>
			<span class="project-task-icon" v-if="task.attachments.length > 0">
				<icon icon="paperclip"/>
			</span>
			<span class="project-task-icon" v-if="task.description">
				<icon icon="align-left"/>
			</span>
			<span class="project-task-icon" v-if="task.repeatAfter.amount > 0">
				<icon icon="history"/>
			</span>
		</span>

		<checklist-summary :task="task"/>

		<progress
			class="progress is-small"
			v-if="task.percentDone > 0"
			:value="task.percentDone * 100" max="100"
		>
			{{ task.percentDone * 100 }}%
		</progress>
	</div>
</template>

<script setup lang="ts">
import {computed} from 'vue'

import {getHexColor} from '@/models/task'
import type {ITask} from '@/modelTypes/ITask'

import PriorityLabel from '@/components/tasks/partials/priorityLabel.vue'
import Labels from '@/components/tasks/partials//labels.vue'
import ChecklistSummary from '@/components/tasks/partials/checklist-summary.vue'

import ColorBubble from '@/components/misc/colorBubble.vue'

import {formatDateSince, formatISO, formatDateLong} from '@/helpers/time/formatDate'

import {useProjectStore} from '@/stores/projects'
import AssigneeList from '@/components/tasks/partials/assigneeList.vue'

const {
	task,
	showProject = false,
} = defineProps<{
	task: ITask,
	showProject?: boolean,
}>()

const projectStore = useProjectStore()

const project = computed(() => projectStore.projects[task.projectId])
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
		margin-left: 5px;
	}

	.overdue {
		color: var(--danger);
	}

	.task-project {
		width: auto;
		color: var(--grey-400);
		font-size: .9rem;
		white-space: nowrap;
	}

	.avatar {
		border-radius: 50%;
		vertical-align: bottom;
		margin-left: .5rem;
		height: 21px;
		width: 21px;
	}

	.project-task-icon {
		margin-left: 6px;

		&:not(:first-of-type) {
			margin-left: 8px;
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
		width: auto;
	}

	.progress {
		margin-bottom: 0;
	}
}
</style>
