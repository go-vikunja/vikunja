<template>
	<div
		ref="taskAdd"
		class="task-add"
	>
		<div class="add-task__field field">
			<p class="control task-input-wrapper">
				<label
					class="is-sr-only"
					:for="textareaId"
				>
					{{ $t('project.list.addPlaceholder') }}
				</label>
				<span class="icon is-small task-icon">
					<Icon icon="tasks" />
				</span>
				<textarea
					:id="textareaId"
					ref="newTaskInput"
					v-model="newTaskTitle"
					v-focus
					class="add-task-textarea input"
					:class="{'textarea-empty': newTaskTitle === ''}"
					:placeholder="$t('project.list.addPlaceholder')"
					rows="1"
					@keydown="resetEmptyTitleError"
					@keydown.enter="handleEnter"
					@keydown.esc="blurTaskInput"
				/>
				<QuickAddMagic
					:highlight-hint-icon="taskAddHovered"
				/>
			</p>
			<p class="control">
				<XButton
					class="add-task-button"
					:disabled="newTaskTitle === '' || loading || undefined"
					icon="plus"
					:loading="loading"
					:aria-label="$t('project.list.add')"
					@click="addTask()"
				>
					<span class="button-text">
						{{ $t('project.list.add') }}
					</span>
				</XButton>
			</p>
		</div>
		<Expandable :open="errorMessage !== ''">
			<p
				v-if="errorMessage !== ''"
				class="pbs-3 mbs-0 help is-danger"
			>
				{{ errorMessage }}
			</p>
		</Expandable>
	</div>
</template>

<script setup lang="ts">
import {computed, ref} from 'vue'
import {useI18n} from 'vue-i18n'
import {useElementHover} from '@vueuse/core'
import {useRouter} from 'vue-router'

import {RELATION_KIND} from '@/types/IRelationKind'
import type {ITask} from '@/modelTypes/ITask'

import Expandable from '@/components/base/Expandable.vue'
import QuickAddMagic from '@/components/tasks/partials/QuickAddMagic.vue'
import {parseSubtasksViaIndention} from '@/helpers/parseSubtasksViaIndention'
import TaskRelationService from '@/services/taskRelation'
import TaskRelationModel from '@/models/taskRelation'
import {getLabelsFromPrefix} from '@/modules/quickAddMagic'
import {calculateItemPosition, calculateItemPositions, POSITION_SPACING} from '@/helpers/calculateItemPosition'
import {error} from '@/message'

import {useAuthStore} from '@/stores/auth'
import {useTaskStore} from '@/stores/tasks'

import {useAutoHeightTextarea} from '@/composables/useAutoHeightTextarea'
import TaskService from '@/services/task'
import TaskModel from '@/models/task'

const props = withDefaults(defineProps<{
	// Position of the task the new ones are inserted before, if the caller shows a
	// sorted list. A whole batch is spread out below it so it keeps the entered order.
	positionAfter?: number | null,
}>(), {
	positionAfter: null,
})

const emit = defineEmits(['taskAdded'])

const textareaId = computed(() => `task-add-textarea-${Math.random().toString(36).substr(2, 9)}`)

const newTaskTitle = ref('')
const {textarea: newTaskInput} = useAutoHeightTextarea(newTaskTitle)

const {t} = useI18n({useScope: 'global'})
const authStore = useAuthStore()
const taskStore = useTaskStore()
const router = useRouter()

// enable only if we don't have a modal
// onStartTyping(() => {
// 	if (newTaskInput.value === null || document.activeElement === newTaskInput.value) {
// 		return
// 	}
// 	newTaskInput.value.focus()
// })

const taskAdd = ref<HTMLElement | null>(null)
const taskAddHovered = useElementHover(taskAdd)

const errorMessage = ref('')

function resetEmptyTitleError() {
	if (!newTaskTitle.value) {
		errorMessage.value = ''
	}
}

const loading = computed(() => taskStore.isLoading)

async function addTask() {
	if (newTaskTitle.value === '') {
		errorMessage.value = t('project.create.addTitleRequired')
		return
	}
	errorMessage.value = ''

	if (loading.value) {
		return
	}

	const taskTitleBackup = newTaskTitle.value
	// This allows us to find the tasks with the title they had before being parsed
	// by quick add magic.
	const createdTasks: { [key: ITask['title']]: ITask } = {}
	const tasksToCreate = parseSubtasksViaIndention(newTaskTitle.value, authStore.settings.frontendSettings.quickAddMagicMode)
	// Raw input lines, filtered the same way parseSubtasksViaIndention does so the indices
	// line up. Used to put only the not-yet-created lines back if creating one fails.
	const inputLines = newTaskTitle.value
		.split(/[\r\n]+/)
		.filter(t => t.replace(/\s/g, '').length > 0)

	// We ensure all labels exist prior to passing them down to the create task method
	// In the store it will only ever see one task at a time so there's no way to reliably 
	// check if a new label was created before (because everything happens async).
	const allLabels = tasksToCreate.map(({title}) => getLabelsFromPrefix(title, authStore.settings.frontendSettings.quickAddMagicMode) ?? [])
	const requestedLabels = [...new Set(allLabels.flat())]
	const resolvedLabels = await taskStore.ensureLabelsExist(requestedLabels)

	// Skipped labels (e.g. link shares may not create them) don't block task creation; just tell the user.
	const resolvedTitles = new Set(resolvedLabels.map(l => l.title.toLowerCase()))
	const failedLabels = requestedLabels.filter(title => !resolvedTitles.has(title.toLowerCase()))
	if (failedLabels.length > 0) {
		error({message: t('task.label.createFailed', {labels: failedLabels.join(', ')})})
	}

	const taskCollectionService = new TaskService()
	const projectIndices = new Map<number, number>()

	let currentProjectId = authStore.settings.defaultProjectId
	if (typeof router.currentRoute.value.params.projectId !== 'undefined') {
		currentProjectId = Number(router.currentRoute.value.params.projectId)
	}

	const isBatch = tasksToCreate.length > 1

	// Create a map of project indices before creating tasks
	if (isBatch) {
		for (const {project} of tasksToCreate) {
			const projectId = project !== null
				? await taskStore.findProjectId({project, projectId: 0})
				: currentProjectId

			if (!projectIndices.has(projectId)) {
				const newestTask = await taskCollectionService.getAll(new TaskModel({}), {
					sort_by: ['id'],
					order_by: ['desc'],
					per_page: 1,
					filter: `project_id = ${projectId}`,
				})
				projectIndices.set(projectId, newestTask[0]?.index || 0)
			}
		}
	}

	// One position for the whole batch would make every task collide, leaving the final
	// order up to the api's conflict repair. Spread them out instead.
	const batchPositions = props.positionAfter !== null
		? calculateItemPositions(tasksToCreate.length, null, props.positionAfter)
		: []

	// Index of the task currently being created, -1 once all of them exist.
	let failedIndex = -1

	try {
		newTaskTitle.value = ''

		// Created one after the other on purpose: the api derives a task's index and
		// fallback position from what already exists, so parallel creates race each
		// other and the tasks end up in an arbitrary order.
		for (let index = 0; index < tasksToCreate.length; index++) {
			failedIndex = index
			const {title, project} = tasksToCreate[index]
			if (title === '') {
				continue
			}

			// If the task has a project specified, make sure to use it
			const projectId = project !== null
				? await taskStore.findProjectId({project, projectId: 0})
				: currentProjectId

			let taskIndex: number | undefined
			let position: number | undefined = calculateItemPosition(null, props.positionAfter)
			if (isBatch) {
				// Calculate new index for this task per project
				taskIndex = (projectIndices.get(projectId) ?? 0) + index + 1
				// Without a task to insert before, fall back to the formula the api uses
				// for a fresh view, which appends to the end.
				position = batchPositions[index] ?? taskIndex * POSITION_SPACING
			}

			createdTasks[title] = await taskStore.createNewTask({
				title,
				projectId: projectId || authStore.settings.defaultProjectId,
				position,
				index: taskIndex,
			})
		}

		failedIndex = -1

		const taskRelationService = new TaskRelationService()
		const allParentTasks = tasksToCreate.filter(t => t.parent !== null).map(t => t.parent)
		// Sequential as well: subtasks are rendered in the order their relations were
		// created, so racing them scrambles the nesting.
		for (const t of tasksToCreate) {
			const createdTask = createdTasks[t.title]
			if (typeof createdTask === 'undefined') {
				continue
			}

			const isParent = allParentTasks.includes(t.title)
			if (t.parent === null && !isParent) {
				continue
			}

			const createdParentTask = createdTasks[t.parent]
			if (typeof createdParentTask === 'undefined') {
				continue
			}

			await taskRelationService.create(new TaskRelationModel({
				taskId: createdTask.id,
				otherTaskId: createdParentTask.id,
				relationKind: RELATION_KIND.PARENTTASK,
			}))

			if (typeof createdTask.relatedTasks === 'undefined') {
				createdTask.relatedTasks = {}
			}
			if (typeof createdTask.relatedTasks[RELATION_KIND.PARENTTASK] === 'undefined') {
				createdTask.relatedTasks[RELATION_KIND.PARENTTASK] = []
			}
			createdTask.relatedTasks[RELATION_KIND.PARENTTASK].push({
				...createdParentTask,
				relatedTasks: {}, // To avoid endless references
			})

			if (typeof createdParentTask.relatedTasks === 'undefined') {
				createdParentTask.relatedTasks = {}
			}
			if (typeof createdParentTask.relatedTasks[RELATION_KIND.SUBTASK] === 'undefined') {
				createdParentTask.relatedTasks[RELATION_KIND.SUBTASK] = []
			}
			createdParentTask.relatedTasks[RELATION_KIND.SUBTASK].push({
				...createdTask,
				relatedTasks: {}, // To avoid endless references
			})
		}
	} catch (e) {
		// Tasks before the failing one are already persisted, so only put the lines back
		// which did not make it.
		if (failedIndex === 0) {
			newTaskTitle.value = taskTitleBackup
		} else if (failedIndex > 0) {
			newTaskTitle.value = inputLines.slice(failedIndex).join('\n')
		}

		if (e?.message === 'NO_PROJECT') {
			errorMessage.value = t('project.create.addProjectRequired')
			return
		}
		throw e
	} finally {
		// We're emitting all tasks at once at the end to avoid the same task showing up
		// multiple times. Also runs on failure so already created tasks show up in the list.
		Object.values(createdTasks).forEach(task => {
			emit('taskAdded', task)
		})
	}
}

function handleEnter(e: KeyboardEvent) {
	// when pressing shift + enter we want to continue as we normally would. Otherwise, we want to create 
	// the new task(s). The vue event modifier don't allow this, hence this method.
	if (e.shiftKey) {
		return
	}

	if (e.isComposing) {
		return
	}

	e.preventDefault()
	addTask()
}

function focusTaskInput() {
	newTaskInput.value?.focus()
}

function blurTaskInput() {
	newTaskInput.value?.blur()
}

defineExpose({
	focusTaskInput,
})
</script>

<style lang="scss" scoped>
.task-add,
	// overwrite bulma styles
.task-add .add-task__field {
	margin-block-end: 0;
}

.task-add .add-task__field {
	display: flex;
	justify-content: flex-start;
	gap: .75rem;

	.control {
		flex-shrink: 0;
		margin-block-end: 0;
	}
}

.task-input-wrapper {
	position: relative;
	flex-shrink: 1;
	flex-grow: 1;

	textarea {
		padding-inline: 2.5rem;
	}

	.icon {
		color: var(--grey-300);
	}

	.task-icon, 
	:deep(.quick-add-magic-trigger-btn) {
		position: absolute;
		inset-block-start: .75rem;
	}

	:deep(.quick-add-magic-trigger-btn) {
		inset-inline-end: .75rem;
	}

	.task-icon {
		inset-inline-start: 1rem;
	}
}

.add-task-button {
	block-size: 100% !important;

	@media screen and (max-width: $tablet) {
		.button-text {
			display: none;
		}

		:deep(.icon) {
			margin: 0 !important;
		}
	}
}

.add-task-textarea {
	transition: border-color $transition;
	resize: none;
}

// Adding this class when the textarea has no text prevents the textarea from wrapping the placeholder.
.textarea-empty {
	white-space: nowrap;
	text-overflow: ellipsis;
}

.control .icon {
	transition: all $transition;
	z-index: 4;
}
</style>

<style>
button.show-helper-text {
	inset-inline-end: 0;
}
</style>
