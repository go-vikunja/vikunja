<template>
	<form @submit.prevent="editTaskSubmit()">
		<div class="field">
			<label class="label" for="tasktext">Task Text</label>
			<div class="control">
				<input v-focus :class="{ 'disabled': taskService.loading}" :disabled="taskService.loading" class="input"
					type="text" id="tasktext" placeholder="The task text is here..." v-model="taskEditTask.title" @change="editTaskSubmit()">
			</div>
		</div>
		<div class="field">
			<label class="label" for="taskdescription">Description</label>
			<div class="control">
				<textarea :class="{ 'disabled': taskService.loading}" :disabled="taskService.loading" class="textarea"
					placeholder="The tasks description goes here..." id="taskdescription"
					v-model="taskEditTask.description" @change="editTaskSubmit()">
				</textarea>
			</div>
		</div>

		<b>Reminder Dates</b>
		<reminders v-model="taskEditTask.reminderDates" @change="editTaskSubmit()"/>

		<div class="field">
			<label class="label" for="taskduedate">Due Date</label>
			<div class="control">
				<flat-pickr
						:class="{ 'disabled': taskService.loading}"
						class="input"
						:disabled="taskService.loading"
						v-model="taskEditTask.dueDate"
						:config="flatPickerConfig"
						@on-close="editTaskSubmit()"
						id="taskduedate"
						placeholder="The tasks due date is here...">
				</flat-pickr>
			</div>
		</div>

		<div class="field">
			<label class="label" for="">Duration</label>
			<div class="control columns">
				<div class="column">
					<flat-pickr
							:class="{ 'disabled': taskService.loading}"
							class="input"
							:disabled="taskService.loading"
							v-model="taskEditTask.startDate"
							:config="flatPickerConfig"
							@on-close="editTaskSubmit()"
							id="taskduedate"
							placeholder="Start date">
					</flat-pickr>
				</div>
				<div class="column">
					<flat-pickr
							:class="{ 'disabled': taskService.loading}"
							class="input"
							:disabled="taskService.loading"
							v-model="taskEditTask.endDate"
							:config="flatPickerConfig"
							@on-close="editTaskSubmit()"
							id="taskduedate"
							placeholder="End date">
					</flat-pickr>
				</div>
			</div>
		</div>

		<div class="field">
			<label class="label" for="">Repeat after</label>
			<repeat-after v-model="taskEditTask.repeatAfter" @change="editTaskSubmit()"/>
		</div>

		<div class="field">
			<label class="label" for="">Priority</label>
			<div class="control priority-select">
				<priority-select v-model="taskEditTask.priority" @change="editTaskSubmit()"/>
			</div>
		</div>

		<div class="field">
			<label class="label">Percent Done</label>
			<div class="control">
				<percent-done-select v-model="taskEditTask.percentDone" @change="editTaskSubmit()"/>
			</div>
		</div>

		<div class="field">
			<label class="label">Color</label>
			<div class="control">
				<color-picker v-model="taskEditTask.hexColor"/>
			</div>
		</div>

		<div class="field">
			<label class="label" for="">Assignees</label>
			<ul class="assingees">
				<li v-for="(a, index) in taskEditTask.assignees" :key="a.id">
					{{a.username}}
					<a @click="deleteAssigneeByIndex(index)">
						<icon icon="times"/>
					</a>
				</li>
			</ul>
		</div>

		<div class="field has-addons">
			<div class="control is-expanded">
				<edit-assignees :task-id="taskEditTask.id" :list-id="taskEditTask.listId" :initial-assignees="taskEditTask.assignees"/>
			</div>
		</div>

		<div class="field">
			<label class="label">Labels</label>
			<div class="control">
				<edit-labels :task-id="taskEditTask.id" v-model="taskEditTask.labels"/>
			</div>
		</div>

		<related-tasks
			class="is-narrow"
			:task-id="task.id"
			:list-id="task.listId"
			:initial-related-tasks="task.relatedTasks"
		/>

		<button type="submit" class="button is-success is-fullwidth" :class="{ 'is-loading': taskService.loading}">
			Save
		</button>

	</form>
</template>

<script>
	import flatPickr from 'vue-flatpickr-component'
	import 'flatpickr/dist/flatpickr.css'

	import ListService from '../../services/list'
	import TaskService from '../../services/task'
	import TaskModel from '../../models/task'
	import priorities from '../../models/priorities'
	import PrioritySelect from './partials/prioritySelect'
	import PercentDoneSelect from './partials/percentDoneSelect'
	import EditLabels from './partials/editLabels'
	import EditAssignees from './partials/editAssignees'
	import RelatedTasks from './partials/relatedTasks'
	import RepeatAfter from './partials/repeatAfter'
	import Reminders from './partials/reminders'
	import ColorPicker from '../input/colorPicker'

	export default {
		name: 'edit-task',
		data() {
			return {
				listId: this.$route.params.id,
				listService: ListService,
				taskService: TaskService,

				priorities: priorities,
				list: {},
				newTask: TaskModel,
				isTaskEdit: false,
				taskEditTask: TaskModel,
				flatPickerConfig: {
					altFormat: 'j M Y H:i',
					altInput: true,
					dateFormat: 'Y-m-d H:i',
					enableTime: true,
					onOpen: this.updateLastReminderDate,
					onClose: this.addReminderDate,
				},
			}
		},
		components: {
			ColorPicker,
			Reminders,
			RepeatAfter,
			RelatedTasks,
			EditAssignees,
			EditLabels,
			PercentDoneSelect,
			PrioritySelect,
			flatPickr,
		},
		props: {
			task: {
				type: TaskModel,
				required: true,
			}
		},
		watch: {
			task() {
				this.taskEditTask = this.task
				this.initTaskFields()
			}
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
				this.taskEditTask.dueDate = +new Date(this.task.dueDate) === 0 ? null : this.task.dueDate
				this.taskEditTask.startDate = +new Date(this.task.startDate) === 0 ? null : this.task.startDate
				this.taskEditTask.endDate = +new Date(this.task.endDate) === 0 ? null : this.task.endDate
			},
			editTaskSubmit() {
				this.taskService.update(this.taskEditTask)
					.then(r => {
						this.$set(this, 'taskEditTask', r)
						this.initTaskFields()
					})
					.catch(e => {
						this.error(e, this)
					})
			},
		},
	}
</script>

<style scoped>
	form {
		margin-bottom: 1em;
	}
</style>