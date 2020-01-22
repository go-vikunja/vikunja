<template>
	<div class="loader-container" :class="{ 'is-loading': taskService.loading}">
		<div class="task-view">
			<div class="heading">
				<h1 class="title task-id">
					#{{ task.id }}
				</h1>
				<div class="is-done" v-if="task.done">Done</div>
				<input type="text" v-model="task.text" @change="saveTask()" class="input title" @keyup.enter="saveTask()"/>
			</div>
			<h6 class="subtitle">
				{{ namespace.name }} >
				<router-link :to="{ name: 'showList', params: { id: list.id } }">
					{{ list.title }}
				</router-link>
			</h6>

			<!-- Content and buttons -->
			<div class="columns">
				<!-- Content -->
				<div class="column">
					<div class="columns details">
						<div class="column assignees" v-if="activeFields.assignees">
							<!-- Assignees -->
							<div class="detail-title">
								<icon icon="users"/>
								Assignees
							</div>
							<edit-assignees
									:task-i-d="task.id"
									:list-i-d="task.listID"
									:initial-assignees="task.assignees"
									ref="assignees"
							/>
						</div>
						<div class="column" v-if="activeFields.priority">
							<!-- Priority -->
							<div class="detail-title">
								<icon :icon="['far', 'star']"/>
								Priority
							</div>
							<priority-select v-model="task.priority" @change="saveTask" ref="priority"/>
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
									class="input"
									:disabled="taskService.loading"
									v-model="task.dueDate"
									:config="flatPickerConfig"
									@on-close="saveTask"
									placeholder="Click here to set a due date"
									ref="dueDate"
								>
								</flat-pickr>
								<a v-if="task.dueDate" @click="() => {task.dueDate = null;saveTask()}">
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
							<percent-done-select v-model="task.percentDone" @change="saveTask" ref="percentDone"/>
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
									class="input"
									:disabled="taskService.loading"
									v-model="task.startDate"
									:config="flatPickerConfig"
									@on-close="saveTask"
									placeholder="Click here to set a start date"
									ref="startDate"
								>
								</flat-pickr>
								<a v-if="task.startDate" @click="() => {task.startDate = null;saveTask()}">
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
									class="input"
									:disabled="taskService.loading"
									v-model="task.endDate"
									:config="flatPickerConfig"
									@on-close="saveTask"
									placeholder="Click here to set an end date"
									ref="endDate"
								>
								</flat-pickr>
								<a v-if="task.endDate" @click="() => {task.endDate = null;saveTask()}">
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
							<reminders v-model="task.reminderDates" @change="saveTask" ref="reminders"/>
						</div>
						<div class="column" v-if="activeFields.repeatAfter">
							<!-- Repeat after -->
							<div class="detail-title">
								<icon :icon="['far', 'clock']"/>
								Repeat
							</div>
							<repeat-after v-model="task.repeatAfter" @change="saveTask" ref="repeatAfter"/>
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
						<edit-labels :task-i-d="taskID" :start-labels="task.labels" ref="labels"/>
					</div>

					<!-- Description -->
					<div class="details content" :class="{ 'has-top-border': activeFields.labels }">
						<h3>
							<span class="icon is-grey">
								<icon icon="align-left"/>
							</span>
							Description
						</h3>
						<!-- We're using a normal textarea until the problem with the icons is resolved in easymde -->
						<!-- <easymde v-model="task.description" @change="saveTask"/>-->
						<textarea class="textarea" v-model="task.description" @change="saveTask" rows="6" placeholder="Click here to enter a description..."></textarea>
					</div>

					<!-- Attachments -->
					<div class="content attachments has-top-border" v-if="activeFields.attachments">
						<attachments
								:task-i-d="taskID"
								:initial-attachments="task.attachments"
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
								:task-i-d="taskID"
								:initial-related-tasks="task.related_tasks"
								:show-no-relations-notice="true"
								ref="relatedTasks"
						/>
					</div>
				</div>
				<div class="column is-one-fifth action-buttons">
					<a class="button" @click="setFieldActive('assignees')">
						<span class="icon is-small"><icon icon="users"/></span>
						Assign this task to a user
					</a>
					<a class="button" @click="setFieldActive('labels')">
						<span class="icon is-small"><icon icon="tags"/></span>
						Add labels
					</a>
					<a class="button" @click="setFieldActive('attachments')">
						<span class="icon is-small"><icon icon="paperclip"/></span>
						Add attachments
					</a>
					<a class="button" @click="setFieldActive('relatedTasks')">
						<span class="icon is-small"><icon icon="tasks"/></span>
						Add task relations
					</a>
					<a class="button" @click="setFieldActive('priority')">
						<span class="icon is-small"><icon :icon="['far', 'star']"/></span>
						Set Priority
					</a>
					<a class="button" @click="setFieldActive('dueDate')">
						<span class="icon is-small"><icon icon="calendar"/></span>
						Set Due Date
					</a>
					<a class="button" @click="setFieldActive('percentDone')">
						<span class="icon is-small"><icon icon="percent"/></span>
						Set Percent Done
					</a>
					<a class="button" @click="setFieldActive('startDate')">
						<span class="icon is-small"><icon icon="calendar-week"/></span>
						Set a Start Date
					</a>
					<a class="button" @click="setFieldActive('endDate')">
						<span class="icon is-small"><icon icon="calendar-week"/></span>
						Set an End Date
					</a>
					<a class="button" @click="setFieldActive('reminders')">
						<span class="icon is-small"><icon icon="history"/></span>
						Set Reminders
					</a>
					<a class="button" @click="setFieldActive('repeatAfter')">
						<span class="icon is-small"><icon :icon="['far', 'clock']"/></span>
						Set a repeating interval
					</a>
					<a class="button is-danger is-outlined noshadow has-no-border" @click="showDeleteModal = true">
						<span class="icon is-small"><icon icon="trash-alt"/></span>
						Delete task
					</a>
				</div>
			</div>

			<!-- Created / Updated [by] -->
		</div>

		<modal
				v-if="showDeleteModal"
				@close="showDeleteModal = false"
				@submit="deleteTask()">
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
	import message from '../../message'
	import TaskService from '../../services/task'
	import TaskModel from '../../models/task'
	import relationKinds from '../../models/relationKinds'
	import ListModel from '../../models/list'
	import NamespaceModel from '../../models/namespace'

	import priorites from '../../models/priorities'

	import flatPickr from 'vue-flatpickr-component'
	import 'flatpickr/dist/flatpickr.css'
	import PrioritySelect from './reusable/prioritySelect'
	import PercentDoneSelect from './reusable/percentDoneSelect'
	import EditLabels from './reusable/editLabels'
	import EditAssignees from './reusable/editAssignees'
	import Attachments from './reusable/attachments'
	import RelatedTasks from './reusable/relatedTasks'
	import RepeatAfter from './reusable/repeatAfter'
	import Reminders from './reusable/reminders'
	import router from '../../router'

	export default {
		name: 'TaskDetailView',
		components: {
			Reminders,
			RepeatAfter,
			RelatedTasks,
			Attachments,
			EditAssignees,
			EditLabels,
			PercentDoneSelect,
			PrioritySelect,
			flatPickr,
		},
		data() {
			return {
				taskID: Number(this.$route.params.id),
				taskService: TaskService,
				task: TaskModel,
				relationKinds: relationKinds,

				list: ListModel,
				namespace: NamespaceModel,
				showDeleteModal: false,

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
				},
			}
		},
		watch: {
			'$route': 'loadTask'
		},
		created() {
			this.taskService = new TaskService()
			this.task = new TaskModel()
		},
		mounted() {
			this.loadTask()
		},
		methods: {
			loadTask() {
				this.taskID = Number(this.$route.params.id)
				this.taskService.get({id: this.taskID})
					.then(r => {
						this.$set(this, 'task', r)
						this.setListAndNamespaceTitleFromParent()

						// Set all active fields based on values in the model
						this.activeFields.assignees = this.task.assignees.length > 0
						this.activeFields.priority = this.task.priority !== priorites.UNSET
						this.activeFields.dueDate = this.task.dueDate !== null
						this.activeFields.percentDone = this.task.percentDone > 0
						this.activeFields.startDate = this.task.startDate !== null
						this.activeFields.endDate = this.task.endDate !== null
						this.activeFields.reminders = this.task.reminderDates.length > 1
						this.activeFields.repeatAfter = this.task.repeatAfter.amount > 0
						this.activeFields.labels = this.task.labels.length > 0
						this.activeFields.attachments = this.task.attachments.length > 0
						this.activeFields.relatedTasks = Object.keys(this.task.related_tasks).length > 0
					})
					.catch(e => {
						message.error(e, this)
					})
			},
			saveTask() {
				this.taskService.update(this.task)
					.then(r => {
						this.$set(this, 'task', r)
						message.success({message: 'The task was saved successfully.'}, this)
					})
					.catch(e => {
						message.error(e, this)
					})
			},
			setListAndNamespaceTitleFromParent() {
				// FIXME: Throw this away once we have vuex
				this.$parent.namespaces.forEach(n => {
					n.lists.forEach(l => {
						if (l.id === this.task.listID) {
							this.list = l
							this.namespace = n
							return
						}
					})
				})
			},
			setFieldActive(fieldName) {
				this.activeFields[fieldName] = true
				this.$nextTick(() => this.$refs[fieldName].$el.focus())
			},
			deleteTask() {
				this.taskService.delete(this.task)
					.then(() => {
						message.success({message: 'The task been deleted successfully.'}, this)
						router.push({name: 'showList', params: {id: this.list.id}})
					})
					.catch(e => {
						message.error(e, this)
					})
			},
		},
	}
</script>
