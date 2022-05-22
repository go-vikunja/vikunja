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
			@change="editTaskSubmit()"
			v-model="taskEditTask.reminderDates"
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

<script lang="ts">
import {defineComponent} from 'vue'

import AsyncEditor from '@/components/input/AsyncEditor'

import TaskService from '../../services/task'
import TaskModel from '../../models/task'
import priorities from '../../models/constants/priorities'
import EditLabels from './partials/editLabels'
import Reminders from './partials/reminders'
import ColorPicker from '../input/colorPicker'

export default defineComponent({
	name: 'edit-task',
	data() {
		return {
			taskService: new TaskService(),

			priorities: priorities,
			editorActive: false,
			isTaskEdit: false,
			taskEditTask: TaskModel,
		}
	},
	computed: {
		taskDetailRoute() {
			return {
				name: 'task.detail',
				params: { id: this.taskEditTask.id },
				state: { backdropView: this.$router.currentRoute.value.fullPath },
			}
		},
	},
	components: {
		ColorPicker,
		Reminders,
		EditLabels,
		editor: AsyncEditor,
	},
	props: {
		task: {
			type: TaskModel,
			required: true,
		},
	},
	watch: {
		task: {
			handler() {
				this.taskEditTask = this.task
				this.initTaskFields()
			},
			immediate: true,
		},
	},
	methods: {
		initTaskFields() {
			this.taskEditTask.dueDate =
				+new Date(this.task.dueDate) === 0 ? null : this.task.dueDate
			this.taskEditTask.startDate =
				+new Date(this.task.startDate) === 0
					? null
					: this.task.startDate
			this.taskEditTask.endDate =
				+new Date(this.task.endDate) === 0 ? null : this.task.endDate
			// This makes the editor trigger its mounted function again which makes it forget every input
			// it currently has in its textarea. This is a counter-hack to a hack inside of vue-easymde
			// which made it impossible to detect change from the outside. Therefore the component would
			// not update if new content from the outside was made available.
			// See https://github.com/NikulinIlya/vue-easymde/issues/3
			this.editorActive = false
			this.$nextTick(() => (this.editorActive = true))
		},
		async editTaskSubmit() {
			this.taskEditTask = await this.taskService.update(this.taskEditTask)
			this.initTaskFields()
			this.$message.success({message: this.$t('task.detail.updateSuccess')})
		},
	},
})
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