<template>
	<div :class="{ 'is-loading': taskService.loading, 'visible': visible}" class="loader-container task-view-container">
		<div class="task-view">
			<heading v-model="task" :can-write="canWrite" ref="heading"/>
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
						<transition name="flash-background" appear>
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
						</transition>
						<transition name="flash-background" appear>
							<div class="column" v-if="activeFields.dueDate">
								<!-- Due Date -->
								<div class="detail-title">
									<icon icon="calendar"/>
									Due Date
								</div>
								<div class="date-input">
									<datepicker
										v-model="task.dueDate"
										@close-on-change="() => saveTask()"
										choose-date-label="Click here to set a due date"
										:disabled="taskService.loading || !canWrite"
										ref="dueDate"
									/>
									<a
										@click="() => {task.dueDate = null;saveTask()}"
										v-if="task.dueDate && canWrite"
										class="remove">
										<span class="icon is-small">
											<icon icon="times"></icon>
										</span>
									</a>
								</div>
							</div>
						</transition>
						<transition name="flash-background" appear>
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
						</transition>
						<transition name="flash-background" appear>
							<div class="column" v-if="activeFields.startDate">
								<!-- Start Date -->
								<div class="detail-title">
									<icon icon="calendar-week"/>
									Start Date
								</div>
								<div class="date-input">
									<datepicker
										v-model="task.startDate"
										@close-on-change="() => saveTask()"
										choose-date-label="Click here to set a start date"
										:disabled="taskService.loading || !canWrite"
										ref="startDate"
									/>
									<a
										@click="() => {task.startDate = null;saveTask()}"
										v-if="task.startDate && canWrite"
										class="remove"
									>
										<span class="icon is-small">
											<icon icon="times"></icon>
										</span>
									</a>
								</div>
							</div>
						</transition>
						<transition name="flash-background" appear>
							<div class="column" v-if="activeFields.endDate">
								<!-- End Date -->
								<div class="detail-title">
									<icon icon="calendar-week"/>
									End Date
								</div>
								<div class="date-input">
									<datepicker
										v-model="task.endDate"
										@close-on-change="() => saveTask()"
										choose-date-label="Click here to set an end date"
										:disabled="taskService.loading || !canWrite"
										ref="endDate"
									/>
									<a
										@click="() => {task.endDate = null;saveTask()}"
										v-if="task.endDate && canWrite"
										class="remove">
										<span class="icon is-small">
											<icon icon="times"></icon>
										</span>
									</a>
								</div>
							</div>
						</transition>
						<transition name="flash-background" appear>
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
						</transition>
						<transition name="flash-background" appear>
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
						</transition>
						<transition name="flash-background" appear>
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
						</transition>
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
					<div class="details content description">
						<description
							v-model="task"
							:can-write="canWrite"
							:attachment-upload="attachmentUpload"
						/>
					</div>

					<!-- Attachments -->
					<div class="content attachments" v-if="activeFields.attachments || hasAttachments">
						<attachments
							:edit-enabled="canWrite"
							:task-id="taskId"
							ref="attachments"
						/>
					</div>

					<!-- Related Tasks -->
					<div class="content details mb-0" v-if="activeFields.relatedTasks">
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
					<div class="content details" v-if="activeFields.moveList">
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
					<x-button
						:class="{'is-success': !task.done}"
						:shadow="task.done"
						@click="toggleTaskDone()"
						class="is-outlined has-no-border"
						icon="check-double"
						type="secondary"
					>
						{{ task.done ? 'Mark as undone' : 'Done!' }}
					</x-button>
					<task-subscription
						entity="task"
						:entity-id="task.id"
						:subscription="task.subscription"
						@change="sub => task.subscription = sub"
					/>
					<x-button
						@click="setFieldActive('assignees')"
						@shortkey.native="setFieldActive('assignees')"
						type="secondary"
						v-shortkey="['a']">
						<span class="icon is-small"><icon icon="users"/></span>
						Assign this task to a user
					</x-button>
					<x-button
						@click="setFieldActive('labels')"
						@shortkey.native="setFieldActive('labels')"
						type="secondary"
						v-shortkey="['l']"
						icon="tags"
					>
						Add labels
					</x-button>
					<x-button
						@click="setFieldActive('priority')"
						type="secondary"
						:icon="['far', 'star']"
					>
						Set Priority
					</x-button>
					<x-button
						@click="setFieldActive('dueDate')"
						@shortkey.native="setFieldActive('dueDate')"
						type="secondary"
						v-shortkey="['d']"
						icon="calendar"
					>
						Set Due Date
					</x-button>
					<x-button
						@click="setFieldActive('startDate')"
						type="secondary"
						icon="calendar-week"
					>
						Set a Start Date
					</x-button>
					<x-button
						@click="setFieldActive('endDate')"
						type="secondary"
						icon="calendar-week"
					>
						Set an End Date
					</x-button>
					<x-button
						@click="setFieldActive('reminders')"
						type="secondary"
						icon="history"
					>
						Set Reminders
					</x-button>
					<x-button
						@click="setFieldActive('repeatAfter')"
						type="secondary"
						:icon="['far', 'clock']"
					>
						Set a repeating interval
					</x-button>
					<x-button
						@click="setFieldActive('percentDone')"
						type="secondary"
						icon="percent"
					>
						Set Percent Done
					</x-button>
					<x-button
						@click="setFieldActive('attachments')"
						@shortkey.native="setFieldActive('attachments')"
						type="secondary"
						v-shortkey="['f']"
						icon="paperclip"
					>
						Add attachments
					</x-button>
					<x-button
						@click="setFieldActive('relatedTasks')"
						@shortkey.native="setFieldActive('relatedTasks')"
						type="secondary"
						v-shortkey="['r']"
						icon="tasks"
					>
						Add task relations
					</x-button>
					<x-button
						@click="setFieldActive('moveList')"
						type="secondary"
						icon="list"
					>
						Move task
					</x-button>
					<x-button
						@click="setFieldActive('color')"
						type="secondary"
						icon="fill-drip"
					>
						Set task color
					</x-button>
					<x-button
						@click="showDeleteModal = true"
						icon="trash-alt"
						:shadow="false"
						class="is-danger is-outlined has-no-border"
					>
						Delete task
					</x-button>

					<!-- Created / Updated [by] -->
					<p class="created">
						Created <span v-tooltip="formatDate(task.created)">{{ formatDateSince(task.created) }}</span>
						by {{ task.createdBy.getDisplayName() }}
						<template v-if="+new Date(task.created) !== +new Date(task.updated)">
							<br/>
							<!-- Computed properties to show the actual date every time it gets updated -->
							Updated <span v-tooltip="updatedFormatted">{{ updatedSince }}</span>
						</template>
						<template v-if="task.done">
							<br/>
							Done <span v-tooltip="doneFormatted">{{ doneSince }}</span>
						</template>
					</p>
				</div>
			</div>
		</div>

		<transition name="modal">
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
		</transition>
	</div>
</template>

<script>
import TaskService from '../../services/task'
import TaskModel from '../../models/task'
import relationKinds from '../../models/relationKinds.json'

import priorites from '../../models/priorities.json'
import rights from '../../models/rights.json'

import PrioritySelect from '../../components/tasks/partials/prioritySelect'
import PercentDoneSelect from '../../components/tasks/partials/percentDoneSelect'
import EditLabels from '../../components/tasks/partials/editLabels'
import EditAssignees from '../../components/tasks/partials/editAssignees'
import Attachments from '../../components/tasks/partials/attachments'
import RelatedTasks from '../../components/tasks/partials/relatedTasks'
import RepeatAfter from '../../components/tasks/partials/repeatAfter'
import Reminders from '../../components/tasks/partials/reminders'
import Comments from '../../components/tasks/partials/comments'
import ListSearch from '../../components/tasks/partials/listSearch'
import description from '@/components/tasks/partials/description'
import ColorPicker from '../../components/input/colorPicker'
import attachmentUpload from '../../components/tasks/mixins/attachmentUpload'
import heading from '@/components/tasks/partials/heading'
import Datepicker from '@/components/input/datepicker'
import {playPop} from '@/helpers/playPop'
import TaskSubscription from '@/components/misc/subscription'

export default {
	name: 'TaskDetailView',
	components: {
		TaskSubscription,
		Datepicker,
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
		description,
		heading,
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
			// We doubled the task color property here because verte does not have a real change property, leading
			// to the color property change being triggered when the # is removed from it, leading to an update,
			// which leads in turn to a change... This creates an infinite loop in which the task is updated, changed,
			// updated, changed, updated and so on.
			// To prevent this, we put the task color property in a seperate value which is set to the task color
			// when it is saved and loaded.
			taskColor: '',

			showDeleteModal: false,
			descriptionChanged: false,
			listViewName: 'list.list',
			// Used to avoid flashing of empty elements if the task content is not yet loaded.
			visible: false,

			priorities: priorites,
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
		updatedSince() {
			return this.formatDateSince(this.task.updated)
		},
		updatedFormatted() {
			return this.formatDate(this.task.updated)
		},
		doneSince() {
			return this.formatDateSince(this.task.doneAt)
		},
		doneFormatted() {
			return this.formatDate(this.task.doneAt)
		},
		hasAttachments() {
			return this.$store.state.attachments.attachments.length > 0
		},
	},
	methods: {
		loadTask() {
			this.taskId = Number(this.$route.params.id)
			this.taskService.get({id: this.taskId})
				.then(r => {
					this.$set(this, 'task', r)
					this.$store.commit('attachments/set', r.attachments)
					this.taskColor = this.task.hexColor
					this.setActiveFields()
					this.setTitle(this.task.title)
				})
				.catch(e => {
					this.error(e, this)
				})
				.finally(() => {
					this.$nextTick(() => this.visible = true)
					this.scrollToHeading()
				})
		},
		scrollToHeading() {
			this.$refs.heading.$el.scrollIntoView({block: 'center'})
		},
		setActiveFields() {

			this.task.startDate = this.task.startDate ? this.task.startDate : null
			this.task.endDate = this.task.endDate ? this.task.endDate : null

			// Set all active fields based on values in the model
			this.activeFields.assignees = this.task.assignees.length > 0
			this.activeFields.priority = this.task.priority !== priorites.UNSET
			this.activeFields.dueDate = this.task.dueDate !== null
			this.activeFields.percentDone = this.task.percentDone > 0
			this.activeFields.startDate = this.task.startDate !== null
			this.activeFields.endDate = this.task.endDate !== null
			this.activeFields.reminders = this.task.reminderDates.length > 0
			this.activeFields.repeatAfter = this.task.repeatAfter.amount > 0
			this.activeFields.labels = this.task.labels.length > 0
			this.activeFields.attachments = this.task.attachments.length > 0
			this.activeFields.relatedTasks = Object.keys(this.task.relatedTasks).length > 0
		},
		saveTask(showNotification = true, undoCallback = null) {

			if (!this.canWrite) {
				return
			}

			// We're doing the whole update in a nextTick because sometimes race conditions can occur when
			// setting the due date on mobile which leads to no due date change being saved.
			this.$nextTick(() => {
				this.task.hexColor = this.taskColor

				// If no end date is being set, but a start date and due date,
				// use the due date as the end date
				if (this.task.endDate === null && this.task.startDate !== null && this.task.dueDate !== null) {
					this.task.endDate = this.task.dueDate
				}

				this.$store.dispatch('tasks/update', this.task)
					.then(r => {
						this.$set(this, 'task', r)
						this.setActiveFields()

						if (!showNotification) {
							return
						}

						let actions = []
						if (undoCallback !== null) {
							actions = [{
								title: 'Undo',
								callback: undoCallback,
							}]
						}
						this.success({message: 'The task was saved successfully.'}, this, actions)
					})
					.catch(e => {
						this.error(e, this)
					})
			})
		},
		setFieldActive(fieldName) {
			this.activeFields[fieldName] = true
			this.$nextTick(() => {
				if (this.$refs[fieldName]) {
					this.$refs[fieldName].$el.focus()

					// scroll the field to the center of the screen if not in viewport already
					const boundingRect = this.$refs[fieldName].$el.getBoundingClientRect()

					if (boundingRect.top > (window.scrollY + window.innerHeight) || boundingRect.top < window.scrollY)
						this.$refs[fieldName].$el.scrollIntoView({
							behavior: 'smooth',
							block: 'center',
							inline: 'nearest'
						})
				}
			})
		},
		deleteTask() {
			this.$store.dispatch('tasks/delete', this.task)
				.then(() => {
					this.success({message: 'The task has been deleted successfully.'}, this)
					this.$router.push({name: 'list.index', params: {listId: this.task.listId}})
				})
				.catch(e => {
					this.error(e, this)
				})
		},
		toggleTaskDone() {
			this.task.done = !this.task.done

			if (this.task.done) {
				playPop()
			}

			this.saveTask(true, () => this.toggleTaskDone())
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
			this.$store.commit('kanban/removeTaskInBucket', this.task)
		},
	},
}
</script>
