<template>
	<form @submit.prevent="editTaskSubmit()">
		<div class="field">
			<label class="label" for="tasktext">Task Text</label>
			<div class="control">
				<input
					:class="{ 'disabled': taskService.loading}"
					:disabled="taskService.loading"
					@change="editTaskSubmit()"
					class="input"
					id="tasktext"
					placeholder="The task text is here..."
					type="text"
					v-focus
					v-model="taskEditTask.title"/>
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

		<b>Reminder Dates</b>
		<reminders @change="editTaskSubmit()" v-model="taskEditTask.reminderDates"/>

		<div class="field">
			<label class="label" for="taskduedate">Due Date</label>
			<div class="control">
				<flat-pickr
					:class="{ 'disabled': taskService.loading}"
					:config="flatPickerConfig"
					:disabled="taskService.loading"
					@on-close="editTaskSubmit()"
					class="input"
					id="taskduedate"
					placeholder="The tasks due date is here..."
					v-model="taskEditTask.dueDate">
				</flat-pickr>
			</div>
		</div>

		<div class="field">
			<label class="label" for="">Duration</label>
			<div class="control columns">
				<div class="column">
					<flat-pickr
						:class="{ 'disabled': taskService.loading}"
						:config="flatPickerConfig"
						:disabled="taskService.loading"
						@on-close="editTaskSubmit()"
						class="input"
						id="taskduedate"
						placeholder="Start date"
						v-model="taskEditTask.startDate">
					</flat-pickr>
				</div>
				<div class="column">
					<flat-pickr
						:class="{ 'disabled': taskService.loading}"
						:config="flatPickerConfig"
						:disabled="taskService.loading"
						@on-close="editTaskSubmit()"
						class="input"
						id="taskduedate"
						placeholder="End date"
						v-model="taskEditTask.endDate">
					</flat-pickr>
				</div>
			</div>
		</div>

		<div class="field">
			<label class="label" for="">Repeat after</label>
			<repeat-after @change="editTaskSubmit()" v-model="taskEditTask.repeatAfter"/>
		</div>

		<div class="field">
			<label class="label" for="">Priority</label>
			<div class="control priority-select">
				<priority-select @change="editTaskSubmit()" v-model="taskEditTask.priority"/>
			</div>
		</div>

		<div class="field">
			<label class="label">Percent Done</label>
			<div class="control">
				<percent-done-select @change="editTaskSubmit()" v-model="taskEditTask.percentDone"/>
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
				<li :key="a.id" v-for="(a, index) in taskEditTask.assignees">
					{{ a.getDisplayName() }}
					<a @click="deleteAssigneeByIndex(index)">
						<icon icon="times"/>
					</a>
				</li>
			</ul>
		</div>

		<div class="field has-addons">
			<div class="control is-expanded">
				<edit-assignees
					:initial-assignees="taskEditTask.assignees"
					:list-id="taskEditTask.listId"
					:task-id="taskEditTask.id"/>
			</div>
		</div>

		<div class="field">
			<label class="label">Labels</label>
			<div class="control">
				<edit-labels :task-id="taskEditTask.id" v-model="taskEditTask.labels"/>
			</div>
		</div>

		<related-tasks
			:initial-related-tasks="task.relatedTasks"
			:list-id="task.listId"
			:task-id="task.id"
			class="is-narrow"
		/>

		<button :class="{ 'is-loading': taskService.loading}" class="button is-primary is-fullwidth" type="submit">
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
			this.taskEditTask.dueDate = +new Date(this.task.dueDate) === 0 ? null : this.task.dueDate
			this.taskEditTask.startDate = +new Date(this.task.startDate) === 0 ? null : this.task.startDate
			this.taskEditTask.endDate = +new Date(this.task.endDate) === 0 ? null : this.task.endDate
			// This makes the editor trigger its mounted function again which makes it forget every input
			// it currently has in its textarea. This is a counter-hack to a hack inside of vue-easymde
			// which made it impossible to detect change from the outside. Therefore the component would
			// not update if new content from the outside was made available.
			// See https://github.com/NikulinIlya/vue-easymde/issues/3
			this.editorActive = false
			this.$nextTick(() => this.editorActive = true)
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