<template>
	<div>
		<div
			ref="taskRoot"
			:class="{'is-loading': taskService.loading}"
			class="task loader-container single-task"
			tabindex="-1"
			:data-is-overdue="isOverdue || undefined"
			@click="openTaskDetail"
			@keyup.enter="openTaskDetail"
		>
			<FancyCheckbox
				v-model="task.done"
				:disabled="(isArchived || disabled) && !canMarkAsDone"
				@update:modelValue="markAsDone"
				@click.stop
			/>

			<ColorBubble
				v-if="!showProjectSeparately && projectColor !== '' && currentProject?.id !== task.projectId"
				:color="projectColor"
				class="mr-1"
			/>

			<div
				:class="{ 'done': task.done, 'show-project': showProject && project}"
				class="tasktext"
			>
				<span>
					<RouterLink
						v-if="showProject && typeof project !== 'undefined'"
						v-tooltip="$t('task.detail.belongsToProject', {project: project.title})"
						:to="{ name: 'project.index', params: { projectId: task.projectId } }"
						class="task-project mr-1"
						:class="{'mr-2': task.hexColor !== ''}"
						@click.stop
					>
						{{ project.title }}
					</RouterLink>

					<ColorBubble
						v-if="task.hexColor !== ''"
						:color="getHexColor(task.hexColor)"
						class="mr-1"
					/>
	
					<PriorityLabel
						:priority="task.priority"
						:done="task.done"
						class="pr-2"
					/>

					<RouterLink
						ref="taskLinkRef"
						:to="taskDetailRoute"
						class="task-link"
						tabindex="-1"
					>
						{{ task.title }}
					</RouterLink>
				</span>

				<Labels
					v-if="task.labels.length > 0"
					class="labels ml-2 mr-1"
					:labels="task.labels"
				/>

				<AssigneeList
					v-if="task.assignees.length > 0"
					:assignees="task.assignees"
					:avatar-size="25"
					class="ml-1"
					:inline="true"
				/>

				<Popup
					v-if="+new Date(task.dueDate) > 0"
				>
					<template #trigger="{toggle, isOpen}">
						<BaseButton
							v-tooltip="formatDateLong(task.dueDate)"
							class="dueDate"
							@click.prevent.stop="toggle()"
						>	
							<time
								:datetime="formatISO(task.dueDate)"
								class="is-italic"
								:aria-expanded="isOpen ? 'true' : 'false'"
							>
								– {{ $t('task.detail.due', {at: dueDateFormatted}) }}
							</time>
						</BaseButton>
					</template>
					<template #content="{isOpen}">
						<DeferTask
							v-if="isOpen"
							v-model="task"
							@update:modelValue="deferTaskUpdate"
						/>
					</template>
				</Popup>

				<span>
					<span
						v-if="task.attachments.length > 0"
						class="project-task-icon"
					>
						<Icon icon="paperclip" />
					</span>
					<span
						v-if="!isEditorContentEmpty(task.description)"
						class="project-task-icon"
					>
						<Icon icon="align-left" />
					</span>
					<span
						v-if="isRepeating"
						class="project-task-icon"
					>
						<Icon icon="history" />
					</span>
				</span>

				<ChecklistSummary :task="task" />
			</div>

			<ProgressBar
				v-if="task.percentDone > 0"
				:value="task.percentDone * 100"
				is-small
			/>

			<ColorBubble
				v-if="showProjectSeparately && projectColor !== '' && currentProject?.id !== task.projectId"
				:color="projectColor"
				class="mr-1"
			/>

			<RouterLink
				v-if="showProjectSeparately"
				v-tooltip="$t('task.detail.belongsToProject', {project: project.title})"
				:to="{ name: 'project.index', params: { projectId: task.projectId } }"
				class="task-project"
				@click.stop
			>
				{{ project.title }}
			</RouterLink>

			<BaseButton
				:class="{'is-favorite': task.isFavorite}"
				class="favorite"
				@click.stop="toggleFavorite"
			>
				<span class="tw-sr-only">{{ task.isFavorite ? $t('task.detail.actions.unfavorite') : $t('task.detail.actions.favorite') }}</span>
				<Icon
					v-if="task.isFavorite"
					icon="star"
				/>
				<Icon
					v-else
					:icon="['far', 'star']"
				/>
			</BaseButton>
			<slot />
		</div>
		<template v-if="typeof task.relatedTasks?.subtask !== 'undefined'">
			<template v-for="subtask in task.relatedTasks.subtask">
				<template v-if="getTaskById(subtask.id)">
					<single-task-in-project
						:key="subtask.id"
						:the-task="getTaskById(subtask.id)"
						:disabled="disabled"
						:can-mark-as-done="canMarkAsDone"
						:all-tasks="allTasks"
						class="subtask-nested"
					/>
				</template>
			</template>
		</template>
	</div>
</template>

<script setup lang="ts">
import {ref, watch, shallowReactive, onMounted, computed} from 'vue'
import {useI18n} from 'vue-i18n'

import TaskModel, {getHexColor} from '@/models/task'
import type {ITask} from '@/modelTypes/ITask'

import PriorityLabel from '@/components/tasks/partials/PriorityLabel.vue'
import Labels from '@/components/tasks/partials/Labels.vue'
import DeferTask from '@/components/tasks/partials/DeferTask.vue'
import ChecklistSummary from '@/components/tasks/partials/ChecklistSummary.vue'

import ProgressBar from '@/components/misc/ProgressBar.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import ColorBubble from '@/components/misc/ColorBubble.vue'
import Popup from '@/components/misc/Popup.vue'

import TaskService from '@/services/task'

import {formatDateSince, formatISO, formatDateLong} from '@/helpers/time/formatDate'
import {success} from '@/message'

import {useProjectStore} from '@/stores/projects'
import {useBaseStore} from '@/stores/base'
import {useTaskStore} from '@/stores/tasks'
import AssigneeList from '@/components/tasks/partials/AssigneeList.vue'
import {useIntervalFn} from '@vueuse/core'
import {playPopSound} from '@/helpers/playPop'
import {isEditorContentEmpty} from '@/helpers/editorContentEmpty'
import {TASK_REPEAT_MODES} from '@/types/IRepeatMode'
import {useGlobalNow} from '@/composables/useGlobalNow'

const props = withDefaults(defineProps<{
	theTask: ITask,
	isArchived?: boolean,
	showProject?: boolean,
	disabled?: boolean,
	canMarkAsDone?: boolean,
	allTasks?: ITask[],
}>(), {
	isArchived: false,
	showProject: false,
	disabled: false,
	canMarkAsDone: true,
	allTasks: () => [],
})

const emit = defineEmits<{
	'taskUpdated': [task: ITask],
}>()

function getTaskById(taskId: number): ITask | undefined {
	if (typeof props.allTasks === 'undefined' || props.allTasks.length === 0) {
		return null
	}

	return props.allTasks.find(t => t.id === taskId)
}

const {t} = useI18n({useScope: 'global'})

const taskService = shallowReactive(new TaskService())
const task = ref<ITask>(new TaskModel())

const isRepeating = computed(() => task.value.repeatAfter.amount > 0 || (task.value.repeatAfter.amount === 0 && task.value.repeatMode === TASK_REPEAT_MODES.REPEAT_MODE_MONTH))

watch(
	() => props.theTask,
	newVal => {
		task.value = newVal
	},
	{
		immediate: true,
		deep: true,
	},
)

const baseStore = useBaseStore()
const projectStore = useProjectStore()
const taskStore = useTaskStore()

const project = computed(() => projectStore.projects[task.value.projectId])
const projectColor = computed(() => project.value ? project.value?.hexColor : '')

const showProjectSeparately = computed(() => !props.showProject && currentProject.value?.id !== task.value.projectId && project.value)

const currentProject = computed(() => {
	return typeof baseStore.currentProject === 'undefined' ? {
		id: 0,
		title: '',
	} : baseStore.currentProject
})

const taskDetailRoute = computed(() => ({
	name: 'task.detail',
	params: {id: task.value.id},
	// TODO: re-enable opening task detail in modal
	// state: { backdropView: router.currentRoute.value.fullPath },
}))

function updateDueDate() {
	if (!task.value.dueDate) {
		return
	}

	dueDateFormatted.value = formatDateSince(task.value.dueDate)
}

const dueDateFormatted = ref('')
useIntervalFn(updateDueDate, 60_000, {
	immediateCallback: true,
})
onMounted(updateDueDate)

watch(() => task.value.dueDate, updateDueDate)

const {now} = useGlobalNow()
const isOverdue = computed(() => (
       !task.value.done &&
       task.value.dueDate !== null &&
       task.value.dueDate.getTime() > 0 &&
       task.value.dueDate.getTime() <= now.value.getTime()
))

let oldTask

async function markAsDone(checked: boolean, wasReverted: boolean = false) {
	const updateFunc = async () => {
		oldTask = {...task.value}
		const newTask = await taskStore.update(task.value)
		task.value = newTask

		updateDueDate()

		if (wasReverted) {
			return
		}

		if (checked) {
			playPopSound()
		}
		emit('taskUpdated', newTask)

		let message = t('task.doneSuccess')
		if (!task.value.done && !isRepeating.value) {
			message = t('task.undoneSuccess')
		}

		success({message}, [{
			title: t('task.undo'),
			callback: () => undoDone(checked),
		}])
	}

	if (checked) {
		setTimeout(updateFunc, 300) // Delay it to show the animation when marking a task as done
	} else {
		await updateFunc() // Don't delay it when un-marking it as it doesn't have an animation the other way around
	}
}

function undoDone(checked: boolean) {
	if (isRepeating.value) {
		task.value = {...oldTask}
	}
	task.value.done = !task.value.done
	markAsDone(!checked, true)
}

async function toggleFavorite() {
	task.value = await taskStore.toggleFavorite(task.value)
	emit('taskUpdated', task.value)
}

const taskRoot = ref<HTMLElement | null>(null)
const taskLinkRef = ref<HTMLElement | null>(null)

function hasTextSelected() {
	const isTextSelected = window.getSelection().toString()
	return !(typeof isTextSelected === 'undefined' || isTextSelected === '' || isTextSelected === '\n')
}

function openTaskDetail(event: MouseEvent | KeyboardEvent) {
	if (event.target instanceof HTMLElement) {
		const isInteractiveElement = event.target.closest('a, button, .favorite, [role="button"]')
		if (isInteractiveElement || hasTextSelected()) {
			return
		}
	}

	taskLinkRef.value?.$el.click()
}

defineExpose({
	focus: () => taskRoot.value?.focus(),
	click: (e: MouseEvent | KeyboardEvent) => openTaskDetail(e),
})
</script>

<style lang="scss" scoped>
.task {
	display: flex;
	flex-wrap: wrap;
	padding: .4rem;
	transition: background-color $transition;
	align-items: center;
	cursor: pointer;
	border-radius: $radius;
	border: 2px solid transparent;

	&:hover {
		background-color: var(--grey-100);
	}

	&:has(*:focus-visible), &:focus {
		box-shadow: 0 0 0 2px hsla(var(--primary-hsl), 0.5);

		a.task-link {
			box-shadow: none;
		}
	}

	@supports not selector(:focus-within) {
		:focus {
			box-shadow: 0 0 0 2px hsla(var(--primary-hsl), 0.5);

			a.task-link {
				box-shadow: none;
			}
		}
	}

	.tasktext,
	&.tasktext {
		text-overflow: ellipsis;
		word-wrap: break-word;
		word-break: break-word;
		display: -webkit-box;
		hyphens: auto;
		-webkit-line-clamp: 4;
		-webkit-box-orient: vertical;
		overflow: hidden;

		flex: 1 0 50%;

	}

	.dueDate {
		display: inline-block;
		margin-left: 5px;

		&:focus-visible {
			box-shadow: none;

			time {
				box-shadow: 0 0 0 1px hsla(var(--primary-hsl), 0.5);
				border-radius: 3px;
			}
		}
	}

	&[data-is-overdue] .dueDate {
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
		margin-left: 5px;
		height: 27px;
		width: 27px;
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

	.favorite {
		opacity: 1;
		text-align: center;
		width: 27px;
		transition: opacity $transition, color $transition;
		border-radius: $radius;

		&:hover {
			color: var(--warning);
		}

		&.is-favorite {
			opacity: 1;
			color: var(--warning);
		}
	}

	@media(hover: hover) and (pointer: fine) {
		& .favorite {
			opacity: 0;
		}

		&:hover .favorite {
			opacity: 1;
		}
	}

	.favorite:focus {
		opacity: 1;
	}

	:deep(.fancy-checkbox) {
		height: 18px;
		padding-top: 0;
		padding-right: .5rem;

		span {
			display: none;
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

	.show-project .parent-tasks {
		padding-left: .25rem;
	}

	.remove {
		color: var(--danger);
	}

	input[type='checkbox'] {
		vertical-align: middle;
	}

	.settings {
		float: right;
		width: 24px;
		cursor: pointer;
	}

	&.loader-container.is-loading:after {
		top: calc(50% - 1rem);
		left: calc(50% - 1rem);
		width: 2rem;
		height: 2rem;
		border-left-color: var(--grey-300);
		border-bottom-color: var(--grey-300);
	}
}

.subtask-nested {
	margin-left: 1.75rem;
}

:deep(.popup) {
	border-radius: $radius;
	background-color: var(--white);
	box-shadow: var(--shadow-lg);
	color: var(--text);
	top: unset;
	
	&.is-open {
		padding: 1rem;
		border: 1px solid var(--grey-200);
	}
}
</style>
