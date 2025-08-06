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
			class="tw-w-full"
		>
		<div class="p-2">
			<div class="task-id tw-flex tw-justify-between">
				<div>
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
					<span
						v-if="showTaskPosition"
						class="tw-text-red-600 tw-ps-2"
					>
						{{ task.position }}
					</span>
				</div>
				<div v-if="projectTitle">
					{{ projectTitle }}
				</div>
			</div>
			<span
				v-if="task.dueDate > 0"
				v-tooltip="formatDateLong(task.dueDate)"
				:class="{'overdue': isOverdue}"
				class="due-date"
			>
				<span class="icon">
					<Icon :icon="['far', 'calendar-alt']" />
				</span>
				<time :datetime="formatISO(task.dueDate)">
					{{ formatDisplayDate(task.dueDate) }}
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
					class="mie-1"
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
import {computed, ref, watch} from 'vue'
import {useRouter} from 'vue-router'

import {useGlobalNow} from '@/composables/useGlobalNow'

import PriorityLabel from '@/components/tasks/partials/PriorityLabel.vue'
import ProgressBar from '@/components/misc/ProgressBar.vue'
import Done from '@/components/misc/Done.vue'
import Labels from '@/components/tasks/partials/Labels.vue'
import ChecklistSummary from './ChecklistSummary.vue'

import {getHexColor} from '@/models/task'
import type {ITask} from '@/modelTypes/ITask'
import type {IProject} from '@/modelTypes/IProject'
import {SUPPORTED_IMAGE_SUFFIX} from '@/models/attachment'
import AttachmentService, {PREVIEW_SIZE} from '@/services/attachment'

import {formatDateLong, formatDisplayDate, formatISO} from '@/helpers/time/formatDate'
import {colorIsDark} from '@/helpers/color/colorIsDark'
import {useTaskStore} from '@/stores/tasks'
import AssigneeList from '@/components/tasks/partials/AssigneeList.vue'
import {playPopSound} from '@/helpers/playPop'
import {isEditorContentEmpty} from '@/helpers/editorContentEmpty'
import {useProjectStore} from '@/stores/projects'

const props = withDefaults(defineProps<{
	task: ITask,
	projectId: IProject['id'],
	loading?: boolean,
}>(), {
	loading: false,
})

const router = useRouter()

const loadingInternal = ref(false)

const color = computed(() => getHexColor(props.task.hexColor))

const projectStore = useProjectStore()

const projectTitle = computed(() => {
	if (props.projectId === props.task.projectId) {
		return
	}
	
	const project = projectStore.projects[props.task.projectId]
	return project?.title
})

const showTaskPosition = computed(() => window.DEBUG_TASK_POSITION)

const {now} = useGlobalNow()
const isOverdue = computed(() => (
	!props.task.done &&
	props.task.dueDate !== null &&
	props.task.dueDate.getTime() > 0 &&
	props.task.dueDate.getTime() <= now.value.getTime()
))

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
	coverImageBlobUrl.value = await attachmentService.getBlobUrl(attachment, PREVIEW_SIZE.LG)
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
		inline-size: 1.5rem;
		block-size: 1.5rem;
		inset-block-start: calc(50% - .75rem);
		inset-inline-start: calc(50% - .75rem);
		border-width: 2px;
	}

	h3 {
		font-family: $family-sans-serif;
		font-size: .85rem;
		word-break: break-word;
	}


	.due-date {
		float: inline-end;
		display: flex;
		align-items: center;
		padding: 0 .25rem;

		.icon {
			margin-inline-end: .25rem;
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
		margin-block-start: .25rem;

		:deep(.tag),
		:deep(.checklist-summary),
		.assignees,
		.icon,
		.priority-label {
			margin-inline-end: .25rem;
		}

		:deep(.checklist-summary) {
			padding-inline-start: 0;
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
			margin-inline-start: 0;
		}

		.priority-label {
			font-size: .75rem;
			padding: 0 .5rem 0 .25rem;

			.icon {
				block-size: 1rem;
				padding: 0 .25rem;
				margin-block-start: 0;
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

	.task-id {
		color: var(--grey-500);
		font-size: .8rem;
		margin-block-end: .25rem;
		display: flex;
	}

	&.is-moving {
		opacity: .5;
	}

	span {
		inline-size: auto;
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
	margin-inline-end: .25rem;
}

.task-progress {
	margin: 8px 0 0;
	inline-size: 100%;
	block-size: 0.5rem;
}
</style>
