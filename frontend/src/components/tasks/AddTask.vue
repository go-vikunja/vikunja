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
import {parseSubtasksViaIndention, type TaskWithParent} from '@/helpers/parseSubtasksViaIndention'
import TaskRelationService from '@/services/taskRelation'
import TaskRelationModel from '@/models/taskRelation'
import {getLabelsFromPrefix} from '@/modules/quickAddMagic'
import {runWrites} from '@/helpers/runWrites'
import {error} from '@/message'

import {useAuthStore} from '@/stores/auth'
import {useConfigStore} from '@/stores/config'
import {useTaskStore} from '@/stores/tasks'

import {useAutoHeightTextarea} from '@/composables/useAutoHeightTextarea'

const emit = defineEmits<{
	tasksAdded: [tasks: ITask[]],
}>()

const textareaId = computed(() => `task-add-textarea-${Math.random().toString(36).substr(2, 9)}`)

const newTaskTitle = ref('')
const {textarea: newTaskInput} = useAutoHeightTextarea(newTaskTitle)

const {t} = useI18n({useScope: 'global'})
const authStore = useAuthStore()
const configStore = useConfigStore()
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
	// Keyed by the title the task had before quick add magic parsed it. A Map,
	// because a user-entered `__proto__` would corrupt plain-object lookups.
	const createdTasks = new Map<ITask['title'], ITask>()
	const tasksToCreate = parseSubtasksViaIndention(newTaskTitle.value, authStore.settings.frontendSettings.quickAddMagicMode)

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

	let currentProjectId = authStore.settings.defaultProjectId
	if (typeof router.currentRoute.value.params.projectId !== 'undefined') {
		currentProjectId = Number(router.currentRoute.value.params.projectId)
	}

	try {
		newTaskTitle.value = ''

		const entries = await Promise.all(tasksToCreate
			.filter(({title}) => title !== '')
			.map(async ({title, project}) => ({
				title,
				projectId: (project !== null
					? await taskStore.findProjectId({project, projectId: 0})
					: currentProjectId) || authStore.settings.defaultProjectId || 0,
			})))

		// Input like a lone bullet passes the empty check but parses to nothing.
		if (entries.length === 0) {
			newTaskTitle.value = taskTitleBackup
			errorMessage.value = t('project.create.addTitleRequired')
			return
		}

		// Separate from createdTasks: duplicate lines create two tasks but collapse
		// into a single map entry.
		const allCreated: ITask[] = []

		const bulk = await taskStore.createNewTasksBulk(entries)
		entries.forEach(({title}, index) => {
			const task = bulk.tasks[index]
			if (task === null) {
				return
			}
			createdTasks.set(title, task)
			allCreated.push(task)
		})

		const taskRelationService = new TaskRelationService()
		const allParentTasks = tasksToCreate.filter(t => t.parent !== null).map(t => t.parent)
		const createRelation = async (t: TaskWithParent) => {
			const createdTask = createdTasks.get(t.title)
			if (typeof createdTask === 'undefined') {
				return
			}

			const isParent = allParentTasks.includes(t.title)
			if (t.parent === null && !isParent) {
				return
			}

			const createdParentTask = createdTasks.get(t.parent ?? '')
			if (typeof createdTask === 'undefined' || typeof createdParentTask === 'undefined') {
				return
			}

			const rel = await taskRelationService.create(new TaskRelationModel({
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

			return rel
		}

		try {
			await runWrites(tasksToCreate, createRelation, configStore.concurrentWrites)
		} catch (e) {
			// The tasks themselves exist by now — reporting and moving on beats
			// restoring the input and letting the user duplicate all of them.
			error(e)
		}

		if (allCreated.length > 0) {
			emit('tasksAdded', allCreated)
		}

		if (bulk.error !== null) {
			newTaskTitle.value = taskTitleBackup
			error(bulk.error)
		}
	} catch (e) {
		newTaskTitle.value = taskTitleBackup
		if (e?.message === 'NO_PROJECT') {
			errorMessage.value = t('project.create.addProjectRequired')
			return
		}
		throw e
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
