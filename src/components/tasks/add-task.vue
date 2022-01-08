<template>
	<div class="task-add">
		<div class="field is-grouped">
			<p class="control has-icons-left is-expanded">
				<textarea
					:disabled="taskService.loading || undefined"
					class="add-task-textarea input"
					:placeholder="$t('list.list.addPlaceholder')"
					rows="1"
					v-focus
					v-model="newTaskTitle"
					ref="newTaskInput"
					@keyup="errorMessage = ''"
					@keydown.enter="handleEnter"
				/>
				<span class="icon is-small is-left">
					<icon icon="tasks"/>
				</span>
			</p>
			<p class="control">
				<x-button
					class="add-task-button"
					:disabled="newTaskTitle === '' || taskService.loading || undefined"
					@click="addTask()"
					icon="plus"
					:loading="taskService.loading"
				>
					{{ $t('list.list.add') }}
				</x-button>
			</p>
		</div>
		<p class="help is-danger" v-if="errorMessage !== ''">
			{{ errorMessage }}
		</p>
		<quick-add-magic v-else />
	</div>
</template>

<script setup lang="ts">
import {ref, watch, unref, shallowReactive} from 'vue'
import {useI18n} from 'vue-i18n'
import {useStore} from 'vuex'
import { tryOnMounted, debouncedWatch, useWindowSize, MaybeRef } from '@vueuse/core'

import TaskService from '@/services/task'
import QuickAddMagic from '@/components/tasks/partials/quick-add-magic.vue'

function cleanupTitle(title: string) {
	return title.replace(/^((\* |\+ |- )(\[ \] )?)/g, '')
}

function useAutoHeightTextarea(value: MaybeRef<string>) {
	const textarea = ref<HTMLInputElement>()
	const minHeight = ref(0)

	// adapted from https://github.com/LeaVerou/stretchy/blob/47f5f065c733029acccb755cae793009645809e2/src/stretchy.js#L34
	function resize(textareaEl: HTMLInputElement|undefined) {
		if (!textareaEl) return

		let empty

		// the value here is the the attribute value
		if (!textareaEl.value && textareaEl.placeholder) {
			empty = true
			textareaEl.value = textareaEl.placeholder
		}

		const cs = getComputedStyle(textareaEl)

		textareaEl.style.minHeight = ''
		textareaEl.style.height = '0'
		const offset = textareaEl.offsetHeight - parseFloat(cs.paddingTop) - parseFloat(cs.paddingBottom)
		const height = textareaEl.scrollHeight + offset + 'px'

		textareaEl.style.height = height

		// calculate min-height for the first time
		if (!minHeight.value) {
			minHeight.value = parseFloat(height)
		}

		textareaEl.style.minHeight = minHeight.value.toString()


		if (empty) {
			textareaEl.value = ''
		}

	}

	tryOnMounted(() => {
		if (textarea.value) {
			// we don't want scrollbars
			textarea.value.style.overflowY = 'hidden'
		}
	})

	const { width: windowWidth } = useWindowSize()

	debouncedWatch(
		windowWidth,
		() => resize(textarea.value),
		{ debounce: 200 },
	)

	// It is not possible to get notified of a change of the value attribute of a textarea without workarounds (setTimeout) 
	// So instead we watch the value that we bound to it.
	watch(
		() => [textarea.value, unref(value)],
		() => resize(textarea.value),
		{
			immediate: true, // calculate initial size
			flush: 'post', // resize after value change is rendered to DOM
		},
	)

	return textarea
}

const props = defineProps({
	defaultPosition: {
		type: Number,
		required: false,
	},
})

const emit = defineEmits(['taskAdded'])

const newTaskTitle = ref('')
const newTaskInput = useAutoHeightTextarea(newTaskTitle)

const { t } = useI18n()
const store = useStore()

const taskService = shallowReactive(new TaskService())
const errorMessage = ref('')

async function addTask() {
	if (newTaskTitle.value === '') {
		errorMessage.value = t('list.create.addTitleRequired')
		return
	}
	errorMessage.value = ''

	if (taskService.loading) {
		return
	}

	const taskTitleBackup = newTaskTitle.value
	const newTasks = newTaskTitle.value.split(/[\r\n]+/).map(async uncleanedTitle => {
		const title = cleanupTitle(uncleanedTitle)
		if (title === '') {
			return
		}

		const task = await store.dispatch('tasks/createNewTask', {
			title,
			listId: store.state.auth.settings.defaultListId,
			position: props.defaultPosition,
		})
		emit('taskAdded', task)
		return task
	})

	try {
		newTaskTitle.value = ''
		await Promise.all(newTasks)
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
</script>

<style lang="scss" scoped>
.task-add {
	margin-bottom: 0;
}

.add-task-button {
	height: 2.5rem;
}
.add-task-textarea {
	transition: border-color $transition;
	resize: none;
}
</style>
