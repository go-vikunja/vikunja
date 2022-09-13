<template>
	<card
		class="taskedit"
		:title="$t('list.list.editTask')"
		@close="$emit('close')"
		:has-close="true"
	>
	<form @submit.prevent="editTaskSubmit()">
		<div class="field">
			<label class="label" for="tasktext">{{ $t('task.attributes.title') }}</label>
			<div class="control">
				<input
					:class="{ disabled: taskService.loading }"
					:disabled="taskService.loading || undefined"
					@change="editTaskSubmit()"
					class="input"
					id="tasktext"
					type="text"
					v-focus
					v-model="taskEditTask.title"
				/>
			</div>
		</div>
		<div class="field">
			<label class="label" for="taskdescription">{{ $t('task.attributes.description') }}</label>
			<div class="control">
				<editor
					:preview-is-default="false"
					id="taskdescription"
					:placeholder="$t('task.description.placeholder')"
					v-if="editorActive"
					v-model="taskEditTask.description"
				/>
			</div>
		</div>

		<strong>{{ $t('task.attributes.reminders') }}</strong>
		<reminders
			v-model="taskEditTask.reminderDates"
			@update:model-value="editTaskSubmit()"
		/>

		<div class="field">
			<label class="label">{{ $t('task.attributes.labels') }}</label>
			<div class="control">
				<edit-labels
					:task-id="taskEditTask.id"
					v-model="taskEditTask.labels"
				/>
			</div>
		</div>

		<div class="field">
			<label class="label">{{ $t('task.attributes.color') }}</label>
			<div class="control">
				<color-picker v-model="taskEditTask.hexColor" />
			</div>
		</div>

		<x-button
			:loading="taskService.loading"
			class="is-fullwidth"
			@click="editTaskSubmit()"
		>
			{{ $t('misc.save') }}
		</x-button>

		<router-link
			class="mt-2 has-text-centered is-block"
			:to="taskDetailRoute"
		>
			{{ $t('task.openDetail') }}
		</router-link>
	</form>
	</card>
</template>

<script setup lang="ts">
import {ref, reactive, computed, shallowReactive, watch, nextTick, type PropType} from 'vue'
import {useRouter} from 'vue-router'
import {useI18n} from 'vue-i18n'

import Editor from '@/components/input/AsyncEditor'

import TaskService from '@/services/task'
import TaskModel from '@/models/task'
import type {ITask} from '@/modelTypes/ITask'
import EditLabels from './partials/editLabels.vue'
import Reminders from './partials/reminders.vue'
import ColorPicker from '../input/colorPicker.vue'

import {success} from '@/message'

const {t} = useI18n({useScope: 'global'})
const router = useRouter()

const props = defineProps({
	task: {
		type: Object as PropType<ITask | null>,
	},
})

const taskService = shallowReactive(new TaskService())

const editorActive = ref(false)
let taskEditTask: ITask | undefined


// FIXME: this initialization should not be necessary here 
function initTaskFields() {
	taskEditTask.dueDate =
		+new Date(props.task.dueDate) === 0 ? null : props.task.dueDate
	taskEditTask.startDate =
		+new Date(props.task.startDate) === 0
			? null
			: props.task.startDate
	taskEditTask.endDate =
		+new Date(props.task.endDate) === 0 ? null : props.task.endDate
	// This makes the editor trigger its mounted function again which makes it forget every input
	// it currently has in its textarea. This is a counter-hack to a hack inside of vue-easymde
	// which made it impossible to detect change from the outside. Therefore the component would
	// not update if new content from the outside was made available.
	// See https://github.com/NikulinIlya/vue-easymde/issues/3
	editorActive.value = false
	nextTick(() => (editorActive.value = true))
}

watch(
	() => props.task,
	() => {
		if (!taskEditTask) {
			taskEditTask = reactive(props.task)
		} else {
			Object.assign(taskEditTask, new TaskModel(props.task))
		}
		initTaskFields()
	},
	{immediate: true },
)
const taskDetailRoute = computed(() => {
	return {
		name: 'task.detail',
		params: { id: taskEditTask.id },
		state: { backdropView: router.currentRoute.value.fullPath },
	}
})

async function editTaskSubmit() {
	const newTask = await taskService.update(taskEditTask)
	Object.assign(taskEditTask, newTask)
	initTaskFields()
	success({message: t('task.detail.updateSuccess')})
}
</script>

<style lang="scss" scoped>
.priority-select {
	.select,
	select {
		width: 100%;
	}
}

ul.assingees {
	list-style: none;
	margin: 0;

	li {
		padding: 0.5rem 0.5rem 0;

		a {
			float: right;
			color: var(--danger);
			transition: all $transition;
		}
	}
}

.tag {
	margin-right: 0.5rem;
	margin-bottom: 0.5rem;

	&:last-child {
		margin-right: 0;
	}
}
</style>