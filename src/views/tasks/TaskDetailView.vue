<template>
	<div :class="{ 'is-loading': taskService.loading}" class="loader-container task-view-container">
		<div class="task-view">
			<div class="heading">
				<h1 class="title task-id" v-if="task.identifier === ''">
					#{{ task.index }}
				</h1>
				<h1 class="title task-id" v-else>
					{{ task.identifier }}
				</h1>
				<div class="is-done" v-if="task.done">Done</div>
				<h1 @focusout="saveTaskOnChange()" @keyup.ctrl.enter="saveTaskOnChange()" class="title input" contenteditable="true"
					ref="taskTitle">{{ task.title }}</h1>
			</div>
			<h6 class="subtitle" v-if="parent && parent.namespace && parent.list">
				{{ parent.namespace.title }} >
				<router-link :to="{ name: listViewName, params: { listId: parent.list.id } }">
					{{ parent.list.title }}
				</router-link>
			</h6>

			<!-- Content and buttons -->
			<div class="columns">
				<!-- Content -->
				<div :class="{'is-two-thirds': canWrite}" class="column">
					<div class="columns details">
						<div class="column assignees" v-if="activeFields.assignees">
							<!-- Assignees -->
							<div class="detail-title">
								<icon icon="users"/>
								Assignees
							</div>
							<edit-assignees
								:disabled="!canWrite"
								:list-id="task.listId"
								:task-id="task.id"
								ref="assignees"
								v-model="task.assignees"
							/>
						</div>
						<div class="column" v-if="activeFields.priority">
							<!-- Priority -->
							<div class="detail-title">
								<icon :icon="['far', 'star']"/>
								Priority
							</div>
							<priority-select
								:disabled="!canWrite"
								@change="saveTask"
								ref="priority"
								v-model="task.priority"/>
						</div>
						<div class="column" v-if="activeFields.dueDate">
							<!-- Due Date -->
							<div class="detail-title">
								<icon icon="calendar"/>
								Due Date
							</div>
							<div class="date-input">
								<flat-pickr
									:class="{ 'disabled': taskService.loading}"
									:config="flatPickerConfig"
									:disabled="taskService.loading || !canWrite"
									@on-close="saveTask"
									class="input"
									placeholder="Click here to set a due date"
									ref="dueDate"
									v-model="dueDate"
								>
								</flat-pickr>
								<a @click="() => {dueDate = task.dueDate = null;saveTask()}" v-if="dueDate && canWrite">
									<span class="icon is-small">
										<icon icon="times"></icon>
									</span>
								</a>
							</div>
						</div>
						<div class="column" v-if="activeFields.percentDone">
							<!-- Percent Done -->
							<div class="detail-title">
								<icon icon="percent"/>
								Percent Done
							</div>
							<percent-done-select
								:disabled="!canWrite"
								@change="saveTask"
								ref="percentDone"
								v-model="task.percentDone"/>
						</div>
						<div class="column" v-if="activeFields.startDate">
							<!-- Start Date -->
							<div class="detail-title">
								<icon icon="calendar-week"/>
								Start Date
							</div>
							<div class="date-input">
								<flat-pickr
									:class="{ 'disabled': taskService.loading}"
									:config="flatPickerConfig"
									:disabled="taskService.loading || !canWrite"
									@on-close="saveTask"
									class="input"
									placeholder="Click here to set a start date"
									ref="startDate"
									v-model="task.startDate"
								>
								</flat-pickr>
								<a @click="() => {task.startDate = null;saveTask()}" v-if="task.startDate && canWrite">
									<span class="icon is-small">
										<icon icon="times"></icon>
									</span>
								</a>
							</div>
						</div>
						<div class="column" v-if="activeFields.endDate">
							<!-- End Date -->
							<div class="detail-title">
								<icon icon="calendar-week"/>
								End Date
							</div>
							<div class="date-input">
								<flat-pickr
									:class="{ 'disabled': taskService.loading}"
									:config="flatPickerConfig"
									:disabled="taskService.loading || !canWrite"
									@on-close="saveTask"
									class="input"
									placeholder="Click here to set an end date"
									ref="endDate"
									v-model="task.endDate"
								>
								</flat-pickr>
								<a @click="() => {task.endDate = null;saveTask()}" v-if="task.endDate && canWrite">
									<span class="icon is-small">
										<icon icon="times"></icon>
									</span>
								</a>
							</div>
						</div>
						<div class="column" v-if="activeFields.reminders">
							<!-- Reminders -->
							<div class="detail-title">
								<icon icon="history"/>
								Reminders
							</div>
							<reminders
								:disabled="!canWrite"
								@change="saveTask"
								ref="reminders"
								v-model="task.reminderDates"/>
						</div>
						<div class="column" v-if="activeFields.repeatAfter">
							<!-- Repeat after -->
							<div class="detail-title">
								<icon :icon="['far', 'clock']"/>
								Repeat
							</div>
							<repeat-after
								:disabled="!canWrite"
								@change="saveTask"
								ref="repeatAfter"
								v-model="task"/>
						</div>
						<div class="column" v-if="activeFields.color">
							<!-- Color -->
							<div class="detail-title">
								<icon icon="fill-drip"/>
								Color
							</div>
							<color-picker
								@change="saveTask"
								menu-position="bottom"
								ref="color"
								v-model="taskColor"/>
						</div>
					</div>

					<!-- Labels -->
					<div class="labels-list details" v-if="activeFields.labels">
						<div class="detail-title">
							<span class="icon is-grey">
								<icon icon="tags"/>
							</span>
							Labels
						</div>
						<edit-labels :disabled="!canWrite" :task-id="taskId" ref="labels" v-model="task.labels"/>
					</div>

					<!-- Description -->
					<div :class="{ 'has-top-border': activeFields.labels }" class="details content description">
						<h3>
							<span class="icon is-grey">
								<icon icon="align-left"/>
							</span>
							Description
						</h3>
						<editor
							:is-edit-enabled="canWrite"
							:upload-callback="attachmentUpload"
							:upload-enabled="true"
							@change="saveTask"
							placeholder="Click here to enter a description..."
							v-model="task.description"/>
					</div>

					<!-- Attachments -->
					<div class="content attachments has-top-border" v-if="activeFields.attachments">
						<attachments
							:edit-enabled="canWrite"
							:task-id="taskId"
							ref="attachments"
						/>
					</div>

					<!-- Related Tasks -->
					<div class="content details has-top-border" v-if="activeFields.relatedTasks">
						<h3>
							<span class="icon is-grey">
								<icon icon="tasks"/>
							</span>
							Related Tasks
						</h3>
						<related-tasks
							:edit-enabled="canWrite"
							:initial-related-tasks="task.relatedTasks"
							:list-id="task.listId"
							:show-no-relations-notice="true"
							:task-id="taskId"
							ref="relatedTasks"
						/>
					</div>

					<!-- Move Task -->
					<div class="content details has-top-border" v-if="activeFields.moveList">
						<h3>
							<span class="icon is-grey">
								<icon icon="list"/>
							</span>
							Move task to a different list
						</h3>
						<div class="field has-addons">
							<div class="control is-expanded">
								<list-search @selected="changeList" ref="moveList"/>
							</div>
						</div>
					</div>

					<!-- Comments -->
					<comments :can-write="canWrite" :task-id="taskId"/>
				</div>
				<div class="column is-one-third action-buttons" v-if="canWrite">
					<a
						:class="{'is-success': !task.done}"
						@click="toggleTaskDone()"
						class="button is-outlined noshadow has-no-border">
						<span class="icon is-small"><icon icon="check-double"/></span>
						<template v-if="task.done">
							Mark as undone
						</template>
						<template v-else>
							Done!
						</template>
					</a>
					<a
						@click="setFieldActive('assignees')"
						@shortkey="setFieldActive('assignees')"
						class="button"
						v-shortkey="['ctrl', 'shift', 'a']">
						<span class="icon is-small"><icon icon="users"/></span>
						Assign this task to a user
					</a>
					<a
						@click="setFieldActive('labels')"
						@shortkey="setFieldActive('labels')"
						class="button"
						v-shortkey="['ctrl', 'shift', 'l']">
						<span class="icon is-small"><icon icon="tags"/></span>
						Add labels
					</a>
					<a @click="setFieldActive('reminders')" class="button">
						<span class="icon is-small"><icon icon="history"/></span>
						Set Reminders
					</a>
					<a
						@click="setFieldActive('dueDate')"
						@shortkey="setFieldActive('dueDate')"
						class="button"
						v-shortkey="['ctrl', 'shift', 'd']">
						<span class="icon is-small"><icon icon="calendar"/></span>
						Set Due Date
					</a>
					<a @click="setFieldActive('startDate')" class="button">
						<span class="icon is-small"><icon icon="calendar-week"/></span>
						Set a Start Date
					</a>
					<a @click="setFieldActive('endDate')" class="button">
						<span class="icon is-small"><icon icon="calendar-week"/></span>
						Set an End Date
					</a>
					<a @click="setFieldActive('repeatAfter')" class="button">
						<span class="icon is-small"><icon :icon="['far', 'clock']"/></span>
						Set a repeating interval
					</a>
					<a @click="setFieldActive('priority')" class="button">
						<span class="icon is-small"><icon :icon="['far', 'star']"/></span>
						Set Priority
					</a>
					<a @click="setFieldActive('percentDone')" class="button">
						<span class="icon is-small"><icon icon="percent"/></span>
						Set Percent Done
					</a>
					<a
						@click="setFieldActive('attachments')"
						@shortkey="setFieldActive('attachments')"
						class="button"
						v-shortkey="['ctrl', 'shift', 'f']">
						<span class="icon is-small"><icon icon="paperclip"/></span>
						Add attachments
					</a>
					<a
						@click="setFieldActive('relatedTasks')"
						@shortkey="setFieldActive('relatedTasks')"
						class="button"
						v-shortkey="['ctrl', 'shift', 'r']">
						<span class="icon is-small"><icon icon="tasks"/></span>
						Add task relations
					</a>
					<a @click="setFieldActive('moveList')" class="button">
						<span class="icon is-small"><icon icon="list"/></span>
						Move task
					</a>
					<a @click="setFieldActive('color')" class="button">
						<span class="icon is-small"><icon icon="fill-drip"/></span>
						Set task color
					</a>
					<a @click="showDeleteModal = true" class="button is-danger is-outlined noshadow has-no-border">
						<span class="icon is-small"><icon icon="trash-alt"/></span>
						Delete task
					</a>
				</div>
			</div>

			<!-- Created / Updated [by] -->
		</div>

		<modal
			@close="showDeleteModal = false"
			@submit="deleteTask()"
			v-if="showDeleteModal">
			<span slot="header">Delete this task</span>
			<p slot="text">
				Are you sure you want to remove this task? <br/>
				This will also remove all attachments, reminders and relations associated with this task and
				<b>cannot be undone!</b>
			</p>
		</modal>
	</div>
</template>

<script>
import TaskService from '../../services/task'
import TaskModel from '../../models/task'
import relationKinds from '../../models/relationKinds.json'

import priorites from '../../models/priorities.json'
import rights from '../../models/rights.json'

import flatPickr from 'vue-flatpickr-component'
import 'flatpickr/dist/flatpickr.css'

import PrioritySelect from '../../components/tasks/partials/prioritySelect'
import PercentDoneSelect from '../../components/tasks/partials/percentDoneSelect'
import EditLabels from '../../components/tasks/partials/editLabels'
import EditAssignees from '../../components/tasks/partials/editAssignees'
import Attachments from '../../components/tasks/partials/attachments'
import RelatedTasks from '../../components/tasks/partials/relatedTasks'
import RepeatAfter from '../../components/tasks/partials/repeatAfter'
import Reminders from '../../components/tasks/partials/reminders'
import Comments from '../../components/tasks/partials/comments'
import router from '../../router'
import ListSearch from '../../components/tasks/partials/listSearch'
import ColorPicker from '../../components/input/colorPicker'
import attachmentUpload from '../../components/tasks/mixins/attachmentUpload'
import LoadingComponent from '../../components/misc/loading'
import ErrorComponent from '../../components/misc/error'

export default {
	name: 'TaskDetailView',
	components: {
		ColorPicker,
		ListSearch,
		Reminders,
		RepeatAfter,
		RelatedTasks,
		Attachments,
		EditAssignees,
		EditLabels,
		PercentDoneSelect,
		PrioritySelect,
		Comments,
		flatPickr,
		editor: () => ({
			component: import(/* webpackPrefetch: true *//* webpackChunkName: "editor" */ '../../components/input/editor'),
			loading: LoadingComponent,
			error: ErrorComponent,
			timeout: 60000,
		}),
	},
	mixins: [
		attachmentUpload,
	],
	data() {
		return {
			taskId: Number(this.$route.params.id),
			taskService: TaskService,
			task: TaskModel,
			relationKinds: relationKinds,
			// The due date is a seperate property in the task to prevent flatpickr from modifying the task model
			// in store right after updating it from the api resulting in the wrong due date format being saved in the task.
			dueDate: null,
			// We doubled the task color property here because verte does not have a real change property, leading
			// to the color property change being triggered when the # is removed from it, leading to an update,
			// which leads in turn to a change... This creates an infinite loop in which the task is updated, changed,
			// updated, changed, updated and so on.
			// To prevent this, we put the task color property in a seperate value which is set to the task color
			// when it is saved and loaded.
			taskColor: '',

			showDeleteModal: false,
			taskTitle: '',
			descriptionChanged: false,
			listViewName: 'list.list',

			priorities: priorites,
			flatPickerConfig: {
				altFormat: 'j M Y H:i',
				altInput: true,
				dateFormat: 'Y-m-d H:i',
				enableTime: true,
				time_24hr: true,
			},
			activeFields: {
				assignees: false,
				priority: false,
				dueDate: false,
				percentDone: false,
				startDate: false,
				endDate: false,
				reminders: false,
				repeatAfter: false,
				labels: false,
				attachments: false,
				relatedTasks: false,
				moveList: false,
				color: false,
			},
		}
	},
	watch: {
		'$route': 'loadTask',
	},
	created() {
		this.taskService = new TaskService()
		this.task = new TaskModel()
	},
	mounted() {

		// Build the list path from the task detail name to send the user to the view they came from.
		const parts = this.$route.name.split('.')
		if (parts.length > 2 && parts[2] === 'detail') {
			this.listViewName = `list.${parts[1]}`
		}

		this.loadTask()
	},
	computed: {
		parent() {
			if (!this.task.listId) {
				return {
					namespace: null,
					list: null,
				}
			}

			if (!this.$store.getters['namespaces/getListAndNamespaceById']) {
				return null
			}

			return this.$store.getters['namespaces/getListAndNamespaceById'](this.task.listId)
		},
		canWrite() {
			return this.task && this.task.maxRight && this.task.maxRight > rights.READ
		},
	},
	methods: {
		loadTask() {
			this.taskId = Number(this.$route.params.id)
			this.taskService.get({id: this.taskId})
				.then(r => {
					this.$set(this, 'task', r)
					this.$store.commit('attachments/set', r.attachments)
					this.taskTitle = this.task.title
					this.taskColor = this.task.hexColor
					this.setActiveFields()
					this.setTitle(this.task.title)
				})
				.catch(e => {
					this.error(e, this)
				})
		},
		setActiveFields() {

			this.dueDate = this.task.dueDate ? this.task.dueDate : null
			this.task.startDate = this.task.startDate ? this.task.startDate : null
			this.task.endDate = this.task.endDate ? this.task.endDate : null

			// Set all active fields based on values in the model
			this.activeFields.assignees = this.task.assignees.length > 0
			this.activeFields.priority = this.task.priority !== priorites.UNSET
			this.activeFields.dueDate = this.task.dueDate !== null
			this.activeFields.percentDone = this.task.percentDone > 0
			this.activeFields.startDate = this.task.startDate !== null
			this.activeFields.endDate = this.task.endDate !== null
			// On chrome, reminderDates.length holds the actual number of reminders that are not null.
			// Unlike on desktop where it holds all reminders, including the ones which are null.
			// This causes the reminders to dissapear entierly when only one is set and the user is on mobile.
			this.activeFields.reminders = this.task.reminderDates.length > 1 || (window.innerWidth < 769 && this.task.reminderDates.length > 0)
			this.activeFields.repeatAfter = this.task.repeatAfter.amount > 0
			this.activeFields.labels = this.task.labels.length > 0
			this.activeFields.attachments = this.task.attachments.length > 0
			this.activeFields.relatedTasks = Object.keys(this.task.relatedTasks).length > 0
		},
		saveTaskOnChange() {
			this.$refs.taskTitle.spellcheck = false

			// Pull the task title from the contenteditable
			let taskTitle = this.$refs.taskTitle.textContent
			this.task.title = taskTitle

			// We only want to save if the title was actually change.
			// Because the contenteditable does not have a change event,
			// we're building it ourselves and only calling saveTask()
			// if the task title changed.
			if (this.task.title !== this.taskTitle) {
				this.saveTask()
				this.taskTitle = taskTitle
			}
		},
		saveTask(undoCallback = null) {

			if (!this.canWrite) {
				return
			}

			this.task.dueDate = this.dueDate
			this.task.hexColor = this.taskColor

			// If no end date is being set, but a start date and due date,
			// use the due date as the end date
			if (this.task.endDate === null && this.task.startDate !== null && this.task.dueDate !== null) {
				this.task.endDate = this.task.dueDate
			}

			this.$store.dispatch('tasks/update', this.task)
				.then(r => {
					this.$set(this, 'task', r)
					let actions = []
					if (undoCallback !== null) {
						actions = [{
							title: 'Undo',
							callback: undoCallback,
						}]
						this.success({message: 'The task was saved successfully.'}, this, actions)
					}
					this.dueDate = this.task.dueDate
					this.setActiveFields()
				})
				.catch(e => {
					this.error(e, this)
				})
		},
		setFieldActive(fieldName) {
			this.activeFields[fieldName] = true
			this.$nextTick(() => {
				if (this.$refs[fieldName]) {
					this.$refs[fieldName].$el.focus()
				}
			})
		},
		deleteTask() {
			this.$store.dispatch('tasks/delete', this.task)
				.then(() => {
					this.success({message: 'The task been deleted successfully.'}, this)
					router.back()
				})
				.catch(e => {
					this.error(e, this)
				})
		},
		toggleTaskDone() {
			this.task.done = !this.task.done
			this.saveTask(() => this.toggleTaskDone())
		},
		setDescriptionChanged(e) {
			if (e.key === 'Enter' || e.key === 'Control') {
				return
			}
			this.descriptionChanged = true
		},
		saveTaskIfDescriptionChanged() {
			// We want to only save the description if it was changed.
			// Since we can either trigger this with ctrl+enter or @change, it would be possible to save a task first
			// with ctrl+enter and then with @change although nothing changed since the last save when @change gets fired.
			// To only save one time we added this method.
			if (this.descriptionChanged) {
				this.descriptionChanged = false
				this.saveTask()
			}
		},
		changeList(list) {
			this.task.listId = list.id
			this.saveTask()
		},
	},
}
</script>
