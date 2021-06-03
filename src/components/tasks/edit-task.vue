<template>
	<form @submit.prevent="editTaskSubmit()">
		<div class="field">
			<label class="label" for="tasktext">Title</label>
			<div class="control">
				<input
					:class="{ disabled: taskService.loading }"
					:disabled="taskService.loading"
					@change="editTaskSubmit()"
					class="input"
					id="tasktext"
					placeholder="The task text is here..."
					type="text"
					v-focus
					v-model="taskEditTask.title"
				/>
			</div>
		</div>
		<div class="field">
			<label class="label" for="taskdescription">Description</label>
			<div class="control">
				<editor
					:preview-is-default="false"
					id="taskdescription"
					placeholder="The tasks description goes here..."
					v-if="editorActive"
					v-model="taskEditTask.description"
				/>
			</div>
		</div>

		<strong>Reminders</strong>
		<reminders
			@change="editTaskSubmit()"
			v-model="taskEditTask.reminderDates"
		/>

		<div class="field">
			<label class="label">Labels</label>
			<div class="control">
				<edit-labels
					:task-id="taskEditTask.id"
					v-model="taskEditTask.labels"
				/>
			</div>
		</div>

		<div class="field">
			<label class="label">Color</label>
			<div class="control">
				<color-picker v-model="taskEditTask.hexColor" />
			</div>
		</div>

		<x-button
			:loading="taskService.loading"
			class="is-fullwidth"
			@click="editTaskSubmit()"
		>
			Save
		</x-button>

		<router-link
			class="mt-2 has-text-centered is-block"
			:to="{name: 'task.detail', params: {id: taskEditTask.id}}"
		>
			Open task detail view
		</router-link>
	</form>
</template>

<script>
import ListService from '../../services/list'
import TaskService from '../../services/task'
import TaskModel from '../../models/task'
import priorities from '../../models/priorities'
import EditLabels from './partials/editLabels'
import Reminders from './partials/reminders'
import ColorPicker from '../input/colorPicker'
import LoadingComponent from '../misc/loading'
import ErrorComponent from '../misc/error'

export default {
	name: 'edit-task',
	data() {
		return {
			listId: this.$route.params.id,
			listService: ListService,
			taskService: TaskService,

			priorities: priorities,
			list: {},
			editorActive: false,
			newTask: TaskModel,
			isTaskEdit: false,
			taskEditTask: TaskModel,
		}
	},
	components: {
		ColorPicker,
		Reminders,
		EditLabels,
		editor: () => ({
			component: import(/* webpackChunkName: "editor" */ '../../components/input/editor'),
			loading: LoadingComponent,
			error: ErrorComponent,
			timeout: 60000,
		}),
	},
	props: {
		task: {
			type: TaskModel,
			required: true,
		},
	},
	watch: {
		task() {
			this.taskEditTask = this.task
			this.initTaskFields()
		},
	},
	created() {
		this.listService = new ListService()
		this.taskService = new TaskService()
		this.newTask = new TaskModel()
		this.taskEditTask = this.task
		this.initTaskFields()
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
		editTaskSubmit() {
			this.taskService
				.update(this.taskEditTask)
				.then((r) => {
					this.$set(this, 'taskEditTask', r)
					this.initTaskFields()
					this.success({message: 'The task has been saved successfully.'}, this)
				})
				.catch((e) => {
					this.error(e, this)
				})
		},
	},
}
</script>
