<template>
	<div :class="{'is-loading': taskService.loading}" class="task loader-container">
		<fancycheckbox
			:disabled="(isArchived || disabled) && !canMarkAsDone"
			@update:model-value="markAsDone"
			v-model="task.done"
		/>

		<ColorBubble
			v-if="showProjectColor && projectColor !== '' && currentProject.id !== task.projectId"
			:color="projectColor"
			class="mr-1"
		/>

		<router-link
			:to="taskDetailRoute"
			:class="{ 'done': task.done, 'show-project': showProject && project !== null}"
			class="tasktext"
		>
			<span>
				<router-link
					v-if="showProject && project !== null"
					:to="{ name: 'project.list', params: { projectId: task.projectId } }"
					class="task-project"
					:class="{'mr-2': task.hexColor !== ''}"
					v-tooltip="$t('task.detail.belongsToProject', {project: project.title})">
					{{ project.title }}
				</router-link>

				<ColorBubble
					v-if="task.hexColor !== ''"
					:color="getHexColor(task.hexColor)"
					class="mr-1"
				/>

				<!-- Show any parent tasks to make it clear this task is a sub task of something -->
				<span class="parent-tasks" v-if="typeof task.relatedTasks.parenttask !== 'undefined'">
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

			<User
				v-for="(a, i) in task.assignees"
				:avatar-size="27"
				:is-inline="true"
				:key="task.id + 'assignee' + a.id + i"
				:show-username="false"
				:user="a"
			/>

			<!-- FIXME: use popup -->
			<BaseButton
				v-if="+new Date(task.dueDate) > 0"
				class="dueDate"
				@click.prevent.stop="showDefer = !showDefer"
				v-tooltip="formatDateLong(task.dueDate)"
			>
				<time
					:datetime="formatISO(task.dueDate)"
					:class="{'overdue': task.dueDate <= new Date() && !task.done}"
					class="is-italic"
					:aria-expanded="showDefer ? 'true' : 'false'"
				>
					– {{ $t('task.detail.due', {at: formatDateSince(task.dueDate)}) }}
				</time>
			</BaseButton>
			<CustomTransition name="fade">
				<defer-task v-if="+new Date(task.dueDate) > 0 && showDefer" v-model="task" ref="deferDueDate"/>
			</CustomTransition>

			<priority-label :priority="task.priority" :done="task.done"/>

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
		</router-link>

		<progress
			class="progress is-small"
			v-if="task.percentDone > 0"
			:value="task.percentDone * 100" max="100"
		>
			{{ task.percentDone * 100 }}%
		</progress>

		<router-link
			v-if="!showProject && currentProject.id !== task.projectId && project !== null"
			:to="{ name: 'project.list', params: { projectId: task.projectId } }"
			class="task-project"
			v-tooltip="$t('task.detail.belongsToProject', {project: project.title})"
		>
			{{ project.title }}
		</router-link>

		<BaseButton
			:class="{'is-favorite': task.isFavorite}"
			@click="toggleFavorite"
			class="favorite"
		>
			<icon icon="star" v-if="task.isFavorite"/>
			<icon :icon="['far', 'star']" v-else/>
		</BaseButton>
		<slot />
	</div>
</template>

<script setup lang="ts">
import {ref, watch, shallowReactive, toRef, type PropType, onMounted, onBeforeUnmount, computed} from 'vue'
import {useI18n} from 'vue-i18n'

import TaskModel, { getHexColor } from '@/models/task'
import type {ITask} from '@/modelTypes/ITask'

import PriorityLabel from '@/components/tasks/partials/priorityLabel.vue'
import Labels from '@/components/tasks/partials//labels.vue'
import DeferTask from '@/components/tasks/partials//defer-task.vue'
import ChecklistSummary from '@/components/tasks/partials/checklist-summary.vue'

import User from '@/components/misc/user.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import Fancycheckbox from '@/components/input/fancycheckbox.vue'
import ColorBubble from '@/components/misc/colorBubble.vue'
import CustomTransition from '@/components/misc/CustomTransition.vue'

import TaskService from '@/services/task'

import {closeWhenClickedOutside} from '@/helpers/closeWhenClickedOutside'
import {formatDateSince, formatISO, formatDateLong} from '@/helpers/time/formatDate'
import {success} from '@/message'

import {useProjectStore} from '@/stores/projects'
import {useBaseStore} from '@/stores/base'
import {useTaskStore} from '@/stores/tasks'

const props = defineProps({
	theTask: {
		type: Object as PropType<ITask>,
		required: true,
	},
	isArchived: {
		type: Boolean,
		default: false,
	},
	showProject: {
		type: Boolean,
		default: false,
	},
	disabled: {
		type: Boolean,
		default: false,
	},
	showProjectColor: {
		type: Boolean,
		default: true,
	},
	canMarkAsDone: {
		type: Boolean,
		default: true,
	},
})

const emit = defineEmits(['task-updated'])

const {t} = useI18n({useScope: 'global'})

const taskService = shallowReactive(new TaskService())
const task = ref<ITask>(new TaskModel())
const showDefer = ref(false)

const theTask = toRef(props, 'theTask')

watch(
	theTask,
	newVal => {
		task.value = newVal
	},
)

onMounted(() => {
	task.value = theTask.value
	document.addEventListener('click', hideDeferDueDatePopup)
})

onBeforeUnmount(() => {
	document.removeEventListener('click', hideDeferDueDatePopup)
})

const baseStore = useBaseStore()
const projectStore = useProjectStore()
const taskStore = useTaskStore()

const project = computed(() => projectStore.getProjectById(task.value.projectId))
const projectColor = computed(() => project.value !== null ? project.value.hexColor : '')

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


async function markAsDone(checked: boolean) {
	const updateFunc = async () => {
		const newTask = await taskStore.update(task.value)
		task.value = newTask
		emit('task-updated', newTask)
		success({
			message: task.value.done ?
				t('task.doneSuccess') :
				t('task.undoneSuccess'),
		}, [{
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
	task.value.done = !task.value.done
	markAsDone(!checked)
}

async function toggleFavorite() {
	task.value.isFavorite = !task.value.isFavorite
	task.value = await taskService.update(task.value)
	await projectStore.loadProjects() // reloading the projects list so that the Favorites project shows up or is hidden when there are (or are not) favorite tasks
	emit('task-updated', task.value)
}

const deferDueDate = ref<typeof DeferTask | null>(null)
function hideDeferDueDatePopup(e) {
	if (!showDefer.value) {
		return
	}
	closeWhenClickedOutside(e, deferDueDate.value.$el, () => {
		showDefer.value = false
	})
}
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

	.tasktext,
	&.tasktext {
		white-space: nowrap;
		text-overflow: ellipsis;
		overflow: hidden;
		display: inline-block;
		flex: 1 0 50%;

		.dueDate {
			display: inline-block;
			margin-left: 5px;
		}

		.overdue {
			color: var(--danger);
		}
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

		&:hover {
			color: var(--warning);
		}

		&.is-favorite {
			opacity: 1;
			color: var(--warning);
		}
	}

	.handle {
		opacity: 1;
		transition: opacity $transition;
		margin-right: .25rem;
		cursor: grab;
	}

	@media(hover: hover) and (pointer: fine) {
		& .favorite,
		& .handle {
			opacity: 0;
		}

		&:hover .favorite,
		&:hover .handle {
			opacity: 1;
		}
	}


	:deep(.fancycheckbox) {
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

	input[type="checkbox"] {
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

	.progress {
		margin-bottom: 0;
	}
}
</style>
