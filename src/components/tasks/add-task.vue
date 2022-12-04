<template>
	<div class="task-add">
		<div class="field is-grouped">
			<p class="control has-icons-left is-expanded">
				<textarea
					:disabled="loading || undefined"
					class="add-task-textarea input"
					:class="{'textarea-empty': newTaskTitle === ''}"
					:placeholder="$t('list.list.addPlaceholder')"
					rows="1"
					v-focus
					v-model="newTaskTitle"
					ref="newTaskInput"
					@keyup="resetEmptyTitleError"
					@keydown.enter="handleEnter"
				/>
				<span class="icon is-small is-left">
					<icon icon="tasks"/>
				</span>
			</p>
			<p class="control">
				<x-button
					class="add-task-button"
					:disabled="newTaskTitle === '' || loading || undefined"
					@click="addTask()"
					icon="plus"
					:loading="loading"
					:aria-label="$t('list.list.add')"
				>
					<span class="button-text">
						{{ $t('list.list.add') }}
					</span>
				</x-button>
			</p>
		</div>
		<p class="help is-danger" v-if="errorMessage !== ''">
			{{ errorMessage }}
		</p>
		<quick-add-magic v-else/>
	</div>
</template>

<script setup lang="ts">
import {computed, ref} from 'vue'
import {useI18n} from 'vue-i18n'

import QuickAddMagic from '@/components/tasks/partials/quick-add-magic.vue'
import type {ITask} from '@/modelTypes/ITask'
import {parseSubtasksViaIndention} from '@/helpers/parseSubtasksViaIndention'
import TaskRelationService from '@/services/taskRelation'
import TaskRelationModel from '@/models/taskRelation'
import {RELATION_KIND} from '@/types/IRelationKind'
import {useAuthStore} from '@/stores/auth'
import {useTaskStore} from '@/stores/tasks'
import {useAutoHeightTextarea} from '@/composables/useAutoHeightTextarea'
import {getLabelsFromPrefix} from '@/modules/parseTaskText'

const props = defineProps({
	defaultPosition: {
		type: Number,
		required: false,
	},
})

const emit = defineEmits(['taskAdded'])

const newTaskTitle = ref('')
const newTaskInput = useAutoHeightTextarea(newTaskTitle)

const {t} = useI18n({useScope: 'global'})
const authStore = useAuthStore()
const taskStore = useTaskStore()

const errorMessage = ref('')

function resetEmptyTitleError(e) {
	if (
		(e.which <= 90 && e.which >= 48 || e.which >= 96 && e.which <= 105)
		&& newTaskTitle.value !== ''
	) {
		errorMessage.value = ''
	}
}

const loading = computed(() => taskStore.isLoading)

async function addTask() {
	if (newTaskTitle.value === '') {
		errorMessage.value = t('list.create.addTitleRequired')
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
	const tasksToCreate = parseSubtasksViaIndention(newTaskTitle.value)

	// We ensure all labels exist prior to passing them down to the create task method
	// In the store it will only ever see one task at a time so there's no way to reliably 
	// check if a new label was created before (because everything happens async).
	const allLabels = tasksToCreate.map(({title}) => getLabelsFromPrefix(title) ?? [])
	await taskStore.ensureLabelsExist(allLabels.flat())

	const newTasks = tasksToCreate.map(async ({title, list}) => {
		if (title === '') {
			return
		}

		// If the task has a list specified, make sure to use it
		let listId = null
		if (list !== null) {
			listId = await taskStore.findListId({list, listId: 0})
		}

		const task = await taskStore.createNewTask({
			title,
			listId: listId || authStore.settings.defaultListId,
			position: props.defaultPosition,
		})
		createdTasks[title] = task
		return task
	})

	try {
		newTaskTitle.value = ''
		await Promise.all(newTasks)

		const taskRelationService = new TaskRelationService()
		const relations = tasksToCreate.map(async t => {
			const createdTask = createdTasks[t.title]
			if (typeof createdTask === 'undefined') {
				return
			}

			if (t.parent === null) {
				emit('taskAdded', createdTask)
				return
			}

			const createdParentTask = createdTasks[t.parent]
			if (typeof createdTask === 'undefined' || typeof createdParentTask === 'undefined') {
				return
			}

			const rel = await taskRelationService.create(new TaskRelationModel({
				taskId: createdTask.id,
				otherTaskId: createdParentTask.id,
				relationKind: RELATION_KIND.PARENTTASK,
			}))

			createdTask.relatedTasks[RELATION_KIND.PARENTTASK] = [createdParentTask]
			// we're only emitting here so that the relation shows up in the task list
			emit('taskAdded', createdTask)

			return rel
		})
		await Promise.all(relations)
	} catch (e: any) {
		newTaskTitle.value = taskTitleBackup
		if (e?.message === 'NO_LIST') {
			errorMessage.value = t('list.create.addListRequired')
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

	e.preventDefault()
	addTask()
}

function focusTaskInput() {
	newTaskInput.value?.focus()
}

defineExpose({
	focusTaskInput,
})
</script>

<style lang="scss" scoped>
.task-add {
	margin-bottom: 0;
}

.add-task-button {
	height: 100% !important;

	@media screen and (max-width: $mobile) {
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
</style>
