<template>
	<div
		class="task loader-container draggable"
		:class="{
			'is-loading': loadingInternal || loading,
			'draggable': !(loadingInternal || loading),
			'has-light-text': color !== TASK_DEFAULT_COLOR && !colorIsDark(color),
		}"
		:style="{'background-color': color !== TASK_DEFAULT_COLOR ? color : undefined}"
		@click.exact="openTaskDetail()"
		@click.ctrl="() => toggleTaskDone(task)"
		@click.meta="() => toggleTaskDone(task)"
	>
		<img
			v-if="coverImageBlobUrl"
			:src="coverImageBlobUrl"
			alt=""
			class="cover-image"
		/>
		<div class="p-2">
			<span class="task-id">
				<Done class="kanban-card__done" :is-done="task.done" variant="small"/>
				<template v-if="task.identifier === ''">
					#{{ task.index }}
				</template>
				<template v-else>
					{{ task.identifier }}
				</template>
			</span>
			<span
				:class="{'overdue': task.dueDate <= new Date() && !task.done}"
				class="due-date"
				v-if="task.dueDate > 0"
				v-tooltip="formatDateLong(task.dueDate)">
				<span class="icon">
					<icon :icon="['far', 'calendar-alt']"/>
				</span>
				<time :datetime="formatISO(task.dueDate)">
					{{ formatDateSince(task.dueDate) }}
				</time>
			</span>
			<h3>{{ task.title }}</h3>
			<progress
				class="progress is-small"
				v-if="task.percentDone > 0"
				:value="task.percentDone * 100" max="100">
				{{ task.percentDone * 100 }}%
			</progress>
			<div class="footer">
				<labels :labels="task.labels"/>
				<priority-label :priority="task.priority" :done="task.done" class="is-inline-flex is-align-items-center"/>
				<assignee-list
					v-if="task.assignees.length > 0"
					:assignees="task.assignees"
					:avatar-size="24"
					class="ml-1"
					:inline="true"
				/>
				<checklist-summary :task="task"/>
				<span class="icon" v-if="task.attachments.length > 0">
					<icon icon="paperclip"/>	
				</span>
				<span v-if="task.description" class="icon">
					<icon icon="align-left"/>
				</span>
				<span class="icon" v-if="task.repeatAfter.amount > 0">
					<icon icon="history"/>
				</span>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import {ref, computed, watch} from 'vue'
import {useRouter} from 'vue-router'

import PriorityLabel from '@/components/tasks/partials/priorityLabel.vue'
import Done from '@/components/misc/Done.vue'
import Labels from '@/components/tasks/partials/labels.vue'
import ChecklistSummary from './checklist-summary.vue'

import {TASK_DEFAULT_COLOR, getHexColor} from '@/models/task'
import type {ITask} from '@/modelTypes/ITask'
import {SUPPORTED_IMAGE_SUFFIX} from '@/models/attachment'
import AttachmentService from '@/services/attachment'

import {formatDateLong, formatISO, formatDateSince} from '@/helpers/time/formatDate'
import {colorIsDark} from '@/helpers/color/colorIsDark'
import {useTaskStore} from '@/stores/tasks'
import AssigneeList from '@/components/tasks/partials/assigneeList.vue'
import {useAuthStore} from '@/stores/auth'
import {playPopSound} from '@/helpers/playPop'

const router = useRouter()

const loadingInternal = ref(false)

const {
	task,
	loading = false,
} = defineProps<{
	task: ITask,
	loading: boolean,
}>()

const color = computed(() => getHexColor(task.hexColor))

async function toggleTaskDone(task: ITask) {
	loadingInternal.value = true
	try {
		const updatedTask = await useTaskStore().update({
			...task,
			done: !task.done,
		})
		
		if (updatedTask.done && useAuthStore().settings.frontendSettings.playSoundWhenDone) {
			playPopSound()
		}
	} finally {
		loadingInternal.value = false
	}
}

function openTaskDetail() {
	router.push({
		name: 'task.detail',
		params: {id: task.id},
		state: {backdropView: router.currentRoute.value.fullPath},
	})
}

const coverImageBlobUrl = ref<string | null>(null)

async function maybeDownloadCoverImage() {
	if (!task.coverImageAttachmentId) {
		coverImageBlobUrl.value = null
		return
	}

	const attachment = task.attachments.find(a => a.id === task.coverImageAttachmentId)
	if (!attachment || !SUPPORTED_IMAGE_SUFFIX.some((suffix) => attachment.file.name.endsWith(suffix))) {
		return
	}

	const attachmentService = new AttachmentService()
	coverImageBlobUrl.value = await attachmentService.getBlobUrl(attachment)
}

watch(
	() => task.coverImageAttachmentId,
	maybeDownloadCoverImage,
	{immediate: true},
)
</script>

<style lang="scss" scoped>
$task-background: var(--white);

.task {
	-webkit-touch-callout: none; // iOS Safari
	user-select: none;
	cursor: pointer;
	box-shadow: var(--shadow-xs);
	display: block;

	font-size: .9rem;
	border-radius: $radius;
	background: $task-background;
	overflow: hidden;

	&.loader-container.is-loading::after {
		width: 1.5rem;
		height: 1.5rem;
		top: calc(50% - .75rem);
		left: calc(50% - .75rem);
		border-width: 2px;
	}

	h3 {
		font-family: $family-sans-serif;
		font-size: .85rem;
		word-break: break-word;
	}

	.progress {
		margin: 8px 0 0 0;
		width: 100%;
		height: 0.5rem;
	}

	.due-date {
		float: right;
		display: flex;
		align-items: center;

		.icon {
			margin-right: .25rem;
		}

		&.overdue {
			color: var(--danger);
		}
	}

	.label-wrapper .tag {
		margin: .5rem .5rem 0 0;
	}

	.footer {
		background: transparent;
		padding: 0;
		display: flex;
		flex-wrap: wrap;
		align-items: center;

		:deep(.tag),
		.assignees,
		.icon,
		.priority-label {
			margin-top: .25rem;
			margin-right: .25rem;
		}

		.assignees {
			display: flex;

			.user {
				display: inline;
				margin: 0;

				img {
					margin: 0;
				}
			}
		}

		// FIXME: should be in labels.vue
		:deep(.tag) {
			margin-left: 0;
		}

		.priority-label {
			font-size: .75rem;
			padding: 0 .5rem 0 .25rem;

			.icon {
				height: 1rem;
				padding: 0 .25rem;
				margin-top: 0;
			}
		}
	}

	.footer .icon,
	.due-date,
	.priority-label {
		background: var(--grey-100);
		border-radius: $radius;
		padding: 0 .5rem;
	}

	.due-date {
		padding: 0 .25rem;
	}

	.task-id {
		color: var(--grey-500);
		font-size: .8rem;
		margin-bottom: .25rem;
		display: flex;
	}

	&.is-moving {
		opacity: .5;
	}

	span {
		width: auto;
	}

	&.has-light-text {
		color: var(--white);

		.task-id {
			color: var(--grey-200);
		}

		.footer .icon,
		.due-date,
		.priority-label {
			background: var(--grey-800);
		}

		.footer {
			.icon svg {
				fill: var(--white);
			}
		}
	}
}

.kanban-card__done {
	margin-right: .25rem;
}
</style>