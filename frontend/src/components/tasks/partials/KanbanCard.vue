<template>
	<div
		class="task loader-container draggable"
		:class="{
			'is-loading': loadingInternal || loading,
			'draggable': !(loadingInternal || loading),
			'has-light-text': !colorIsDark(color),
			'has-custom-background-color': color ?? undefined,
		}"
		:style="{'background-color': color ?? undefined}"
		@click.exact="openTaskDetail()"
		@click.ctrl="() => toggleTaskDone(task)"
		@click.meta="() => toggleTaskDone(task)"
	>
		<img
			v-if="coverImageBlobUrl"
			:src="coverImageBlobUrl"
			alt=""
			class="cover-image"
		>
		<div class="p-2">
			<span class="task-id">
				<Done
					class="kanban-card__done"
					:is-done="task.done"
					variant="small"
				/>
				<template v-if="task.identifier === ''">
					#{{ task.index }}
				</template>
				<template v-else>
					{{ task.identifier }}
				</template>
			</span>
			<span
				v-if="task.dueDate > 0"
				v-tooltip="formatDateLong(task.dueDate)"
				:class="{'overdue': task.dueDate <= new Date() && !task.done}"
				class="due-date"
			>
				<span class="icon">
					<Icon :icon="['far', 'calendar-alt']" />
				</span>
				<time :datetime="formatISO(task.dueDate)">
					{{ formatDateSince(task.dueDate) }}
				</time>
			</span>
			<h3>{{ task.title }}</h3>

			<ProgressBar
				v-if="task.percentDone > 0"
				class="task-progress"
				:value="task.percentDone * 100"
			/>
			<div class="footer">
				<Labels :labels="task.labels" />
				<PriorityLabel
					:priority="task.priority"
					:done="task.done"
					class="is-inline-flex is-align-items-center"
				/>
				<AssigneeList
					v-if="task.assignees.length > 0"
					:assignees="task.assignees"
					:avatar-size="24"
					class="mr-1"
				/>
				<ChecklistSummary
					:task="task"
					class="checklist"
				/>
				<span
					v-if="task.attachments.length > 0"
					class="icon"
				>
					<Icon icon="paperclip" />	
				</span>
				<span
					v-if="!isEditorContentEmpty(task.description)"
					class="icon"
				>
					<Icon icon="align-left" />
				</span>
				<span
					v-if="task.repeatAfter.amount > 0"
					class="icon"
				>
					<Icon icon="history" />
				</span>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import {ref, computed, watch} from 'vue'
import {useRouter} from 'vue-router'

import PriorityLabel from '@/components/tasks/partials/PriorityLabel.vue'
import ProgressBar from '@/components/misc/ProgressBar.vue'
import Done from '@/components/misc/Done.vue'
import Labels from '@/components/tasks/partials/Labels.vue'
import ChecklistSummary from './ChecklistSummary.vue'

import {getHexColor} from '@/models/task'
import type {ITask} from '@/modelTypes/ITask'
import {SUPPORTED_IMAGE_SUFFIX} from '@/models/attachment'
import AttachmentService from '@/services/attachment'

import {formatDateLong, formatISO, formatDateSince} from '@/helpers/time/formatDate'
import {colorIsDark} from '@/helpers/color/colorIsDark'
import {useTaskStore} from '@/stores/tasks'
import AssigneeList from '@/components/tasks/partials/AssigneeList.vue'
import {playPopSound} from '@/helpers/playPop'
import {isEditorContentEmpty} from '@/helpers/editorContentEmpty'

const props = withDefaults(defineProps<{
	task: ITask,
	loading: boolean,
}>(), {
	loading: false,
})

const router = useRouter()

const loadingInternal = ref(false)

const color = computed(() => getHexColor(props.task.hexColor))

async function toggleTaskDone(task: ITask) {
	loadingInternal.value = true
	try {
		const updatedTask = await useTaskStore().update({
			...task,
			done: !task.done,
		})

		if (updatedTask.done) {
			playPopSound()
		}
	} finally {
		loadingInternal.value = false
	}
}

function openTaskDetail() {
	router.push({
		name: 'task.detail',
		params: {id: props.task.id},
		state: {backdropView: router.currentRoute.value.fullPath},
	})
}

const coverImageBlobUrl = ref<string | null>(null)

async function maybeDownloadCoverImage() {
	if (!props.task.coverImageAttachmentId) {
		coverImageBlobUrl.value = null
		return
	}

	const attachment = props.task.attachments.find(a => a.id === props.task.coverImageAttachmentId)
	if (!attachment || !SUPPORTED_IMAGE_SUFFIX.some((suffix) => attachment.file.name.toLowerCase().endsWith(suffix))) {
		return
	}

	const attachmentService = new AttachmentService()
	coverImageBlobUrl.value = await attachmentService.getBlobUrl(attachment)
}

watch(
	() => props.task.coverImageAttachmentId,
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
		margin-top: .25rem;

		:deep(.tag),
		:deep(.checklist-summary),
		.assignees,
		.icon,
		.priority-label {
			margin-right: .25rem;
		}

		:deep(.checklist-summary) {
			padding-left: 0;
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

		// FIXME: should be in Labels.vue
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

	&.has-custom-background-color {
		color: hsl(215, 27.9%, 16.9%); // copied from grey-800 to avoid different values in dark mode

		.footer .icon,
		.due-date,
		.priority-label {
			background: hsl(220, 13%, 91%);
		}

		.footer :deep(.checklist-summary) {
			color: hsl(216.9, 19.1%, 26.7%); // grey-700
		}
	}

	&.has-light-text {
		--white: hsla(var(--white-h), var(--white-s), var(--white-l), var(--white-a)) !important;
		color: var(--white);

		.task-id {
			color: hsl(220, 13%, 91%); // grey-200;
		}

		.footer .icon,
		.due-date,
		.priority-label {
			background: hsl(215, 27.9%, 16.9%); // grey-800
		}

		.footer {
			.icon svg {
				fill: var(--white);
			}

			:deep(.checklist-summary) {
				color: hsl(220, 13%, 91%); // grey-200
			}
		}
	}
}

.kanban-card__done {
	margin-right: .25rem;
}

.task-progress {
	margin: 8px 0 0 0;
	width: 100%;
	height: 0.5rem;
}
</style>