<template>
	<div :class="{ 'is-loading': taskService.loading, 'visible': visible}" class="loader-container task-view-container">
		<div class="task-view">
			<heading v-model="task" :can-write="canWrite" ref="heading"/>
			<h6 class="subtitle" v-if="parent && parent.namespace && parent.list">
				{{ getNamespaceTitle(parent.namespace) }} >
				<router-link :to="{ name: 'list.index', params: { listId: parent.list.id } }">
					{{ getListTitle(parent.list) }}
				</router-link>
			</h6>

			<checklist-summary :task="task"/>

			<!-- Content and buttons -->
			<div class="columns mt-2">
				<!-- Content -->
				<div :class="{'is-two-thirds': canWrite}" class="column">
					<div class="columns details">
						<div class="column assignees" v-if="activeFields.assignees">
							<!-- Assignees -->
							<div class="detail-title">
								<icon icon="users"/>
								{{ $t('task.attributes.assignees') }}
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
									<icon icon="exclamation"/>
									{{ $t('task.attributes.priority') }}
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
									{{ $t('task.attributes.dueDate') }}
								</div>
								<div class="date-input">
									<datepicker
										v-model="task.dueDate"
										@close-on-change="() => saveTask()"
										:choose-date-label="$t('task.detail.chooseDueDate')"
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
									{{ $t('task.attributes.percentDone') }}
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
									<icon icon="play"/>
									{{ $t('task.attributes.startDate') }}
								</div>
								<div class="date-input">
									<datepicker
										v-model="task.startDate"
										@close-on-change="() => saveTask()"
										:choose-date-label="$t('task.detail.chooseStartDate')"
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
									<icon icon="stop"/>
									{{ $t('task.attributes.endDate') }}
								</div>
								<div class="date-input">
									<datepicker
										v-model="task.endDate"
										@close-on-change="() => saveTask()"
										:choose-date-label="$t('task.detail.chooseEndDate')"
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
									<icon :icon="['far', 'clock']"/>
									{{ $t('task.attributes.reminders') }}
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
									<icon icon="history"/>
									{{ $t('task.attributes.repeat') }}
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
									{{ $t('task.attributes.color') }}
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
							{{ $t('task.attributes.labels') }}
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
								<icon icon="sitemap"/>
							</span>
							{{ $t('task.attributes.relatedTasks') }}
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
							{{ $t('task.detail.move') }}
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
				<div class="column is-one-third action-buttons">
					<a @click="$router.back()" class="is-fullwidth is-block has-text-centered mb-4" v-if="shouldShowClosePopup">
						<icon icon="arrow-left"/>
						{{ $t('task.detail.closePopup') }}
					</a>
					<template v-if="canWrite">
						<x-button
							:class="{'is-success': !task.done}"
							:shadow="task.done"
							@click="toggleTaskDone()"
							class="is-outlined has-no-border"
							icon="check-double"
							type="secondary"
						>
							{{ task.done ? $t('task.detail.undone') : $t('task.detail.done') }}
						</x-button>
						<task-subscription
							entity="task"
							:entity-id="task.id"
							:subscription="task.subscription"
							@change="sub => task.subscription = sub"
						/>
						<x-button
							@click="setFieldActive('assignees')"
							type="secondary"
							v-shortcut="'a'"
							v-cy="'taskDetail.assign'"
						>
							<span class="icon is-small"><icon icon="users"/></span>
							{{ $t('task.detail.actions.assign') }}
						</x-button>
						<x-button
							@click="setFieldActive('labels')"
							type="secondary"
							icon="tags"
							v-shortcut="'l'"
						>
							{{ $t('task.detail.actions.label') }}
						</x-button>
						<x-button
							@click="setFieldActive('priority')"
							type="secondary"
							icon="exclamation"
						>
							{{ $t('task.detail.actions.priority') }}
						</x-button>
						<x-button
							@click="setFieldActive('dueDate')"
							type="secondary"
							icon="calendar"
							v-shortcut="'d'"
						>
							{{ $t('task.detail.actions.dueDate') }}
						</x-button>
						<x-button
							@click="setFieldActive('startDate')"
							type="secondary"
							icon="play"
						>
							{{ $t('task.detail.actions.startDate') }}
						</x-button>
						<x-button
							@click="setFieldActive('endDate')"
							type="secondary"
							icon="stop"
						>
							{{ $t('task.detail.actions.endDate') }}
						</x-button>
						<x-button
							@click="setFieldActive('reminders')"
							type="secondary"
							:icon="['far', 'clock']"
						>
							{{ $t('task.detail.actions.reminders') }}
						</x-button>
						<x-button
							@click="setFieldActive('repeatAfter')"
							type="secondary"
							icon="history"
						>
							{{ $t('task.detail.actions.repeatAfter') }}
						</x-button>
						<x-button
							@click="setFieldActive('percentDone')"
							type="secondary"
							icon="percent"
						>
							{{ $t('task.detail.actions.percentDone') }}
						</x-button>
						<x-button
							@click="setFieldActive('attachments')"
							type="secondary"
							icon="paperclip"
							v-shortcut="'f'"
						>
							{{ $t('task.detail.actions.attachments') }}
						</x-button>
						<x-button
							@click="setFieldActive('relatedTasks')"
							type="secondary"
							icon="sitemap"
							v-shortcut="'r'"
						>
							{{ $t('task.detail.actions.relatedTasks') }}
						</x-button>
						<x-button
							@click="setFieldActive('moveList')"
							type="secondary"
							icon="list"
						>
							{{ $t('task.detail.actions.moveList') }}
						</x-button>
						<x-button
							@click="setFieldActive('color')"
							type="secondary"
							icon="fill-drip"
						>
							{{ $t('task.detail.actions.color') }}
						</x-button>
						<x-button
							@click="toggleFavorite"
							type="secondary"
							:icon="task.isFavorite ? 'star' : ['far', 'star']"
						>
							{{
								task.isFavorite ? $t('task.detail.actions.unfavorite') : $t('task.detail.actions.favorite')
							}}
						</x-button>
						<x-button
							@click="showDeleteModal = true"
							icon="trash-alt"
							:shadow="false"
							class="is-danger is-outlined has-no-border"
						>
							{{ $t('task.detail.actions.delete') }}
						</x-button>
					</template>

					<!-- Created / Updated [by] -->
					<p class="created">
						<i18n-t keypath="task.detail.created">
							<span v-tooltip="formatDate(task.created)">{{ formatDateSince(task.created) }}</span>
							{{ task.createdBy.getDisplayName() }}
						</i18n-t>
						<template v-if="+new Date(task.created) !== +new Date(task.updated)">
							<br/>
							<!-- Computed properties to show the actual date every time it gets updated -->
							<i18n-t keypath="task.detail.updated">
								<span v-tooltip="updatedFormatted">{{ updatedSince }}</span>
							</i18n-t>
						</template>
						<template v-if="task.done">
							<br/>
							<i18n-t keypath="task.detail.doneAt">
								<span v-tooltip="doneFormatted">{{ doneSince }}</span>
							</i18n-t>
						</template>
					</p>
				</div>
			</div>
		</div>

		<transition name="modal">
			<modal
				@close="showDeleteModal = false"
				@submit="deleteTask()"
				v-if="showDeleteModal"
			>
				<template #header><span>{{ $t('task.detail.delete.header') }}</span></template>

				<template #text>
					<p>{{ $t('task.detail.delete.text1') }}<br/>
						{{ $t('task.detail.delete.text2') }}</p>
				</template>
			</modal>
		</transition>
	</div>
</template>

<script>
import TaskService from '../../services/task'
import TaskModel from '../../models/task'

import priorites from '../../models/constants/priorities.json'
import rights from '../../models/constants/rights.json'

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
import description from '@/components/tasks/partials/description.vue'
import ColorPicker from '../../components/input/colorPicker'
import heading from '@/components/tasks/partials/heading.vue'
import Datepicker from '@/components/input/datepicker.vue'
import {playPop} from '@/helpers/playPop'
import TaskSubscription from '@/components/misc/subscription.vue'
import {CURRENT_LIST} from '@/store/mutation-types'

import {uploadFile} from '@/helpers/attachments'
import ChecklistSummary from '../../components/tasks/partials/checklist-summary'

export default {
	name: 'TaskDetailView',
	components: {
		ChecklistSummary,
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
	data() {
		return {
			taskService: new TaskService(),
			task: new TaskModel(),
			// We doubled the task color property here because verte does not have a real change property, leading
			// to the color property change being triggered when the # is removed from it, leading to an update,
			// which leads in turn to a change... This creates an infinite loop in which the task is updated, changed,
			// updated, changed, updated and so on.
			// To prevent this, we put the task color property in a seperate value which is set to the task color
			// when it is saved and loaded.
			taskColor: '',

			showDeleteModal: false,
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
		taskId: {
			handler: 'loadTask',
			immediate: true,
		},
		parent: {
			handler(parent) {
				const parentList = parent !== null ? parent.list : null
				if (parentList !== null) {
					this.$store.commit(CURRENT_LIST, parentList)
				}
			},
			immediate: true,
		},
	},
	computed: {
		taskId() {
			const {id} = this.$route.params
			return id === undefined ? id : Number(id)
		},
		currentList() {
			return this.$store.state[CURRENT_LIST]
		},
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
			return typeof this.task !== 'undefined' && typeof this.task.maxRight !== 'undefined' && this.task.maxRight > rights.READ
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
		shouldShowClosePopup() {
			return this.$route.name.includes('kanban')
		},
	},
	methods: {
		attachmentUpload(...args) {
			return uploadFile(this.taskId, ...args)
		},

		async loadTask(taskId) {
			if (taskId === undefined) {
				return
			}

			try {
				this.task = await this.taskService.get({id: taskId})
				this.$store.commit('attachments/set', this.task.attachments)
				this.taskColor = this.task.hexColor
				this.setActiveFields()
				this.setTitle(this.task.title)
			} finally {
				this.scrollToHeading()
				await this.$nextTick()
				this.visible = true
			}
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
		async saveTask(showNotification = true, undoCallback = null) {
			if (!this.canWrite) {
				return
			}

			// We're doing the whole update in a nextTick because sometimes race conditions can occur when
			// setting the due date on mobile which leads to no due date change being saved.
			await this.$nextTick()

			this.task.hexColor = this.taskColor

			// If no end date is being set, but a start date and due date,
			// use the due date as the end date
			if (this.task.endDate === null && this.task.startDate !== null && this.task.dueDate !== null) {
				this.task.endDate = this.task.dueDate
			}

			this.task = await this.$store.dispatch('tasks/update', this.task)
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
			this.$message.success({message: this.$t('task.detail.updateSuccess')}, actions)
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
							inline: 'nearest',
						})
				}
			})
		},

		async deleteTask() {
			await this.$store.dispatch('tasks/delete', this.task)
			this.$message.success({message: this.$t('task.detail.deleteSuccess')})
			this.$router.push({name: 'list.index', params: {listId: this.task.listId}})
		},

		toggleTaskDone() {
			this.task.done = !this.task.done

			if (this.task.done) {
				playPop()
			}

			this.saveTask(true, this.toggleTaskDone)
		},

		async changeList(list) {
			this.$store.commit('kanban/removeTaskInBucket', this.task)
			this.task.listId = list.id
			await this.saveTask()
		},

		async toggleFavorite() {
			this.task.isFavorite = !this.task.isFavorite
			this.task = await this.taskService.update(this.task)
			this.$store.dispatch('namespaces/loadNamespacesIfFavoritesDontExist')
		},
	},
}
</script>

<style lang="scss" scoped>
$flash-background-duration: 750ms;

.task-view {
  // This is a workaround to hide the llama background from the top on the task detail page
  margin-top: -1.5rem;
  padding: 1rem;
  background-color: var(--site-background);

  @media screen and (max-width: $desktop) {
    padding-bottom: 0;
  }

  .subtitle {
    color: var(--grey-500);
	margin-bottom: 1rem;

    a {
      color: var(--grey-800);
    }
  }

  h3 .button {
    vertical-align: middle;
  }

  .icon.is-grey {
    color: var(--grey-400);
  }

  :deep(.heading) {
    display: flex;
    justify-content: space-between;
    text-transform: none;
    align-items: center;

    @media screen and (max-width: $tablet) {
      flex-direction: column;
      align-items: start;
    }

    .title {
      margin-bottom: 0;
	}

   .title.input {
		// 1.8rem is the font-size, 1.125 is the line-height, .3rem padding everywhere, 1px border around the whole thing.
		min-height: calc(1.8rem * 1.125 + .6rem + 2px);

		@media screen and (max-width: $tablet) {
			margin: 0 -.3rem .5rem -.3rem; // the title has 0.3rem padding - this make the text inside of it align with the rest
		}
    }

    .title.task-id {
      color: var(--grey-400);
      white-space: nowrap;
    }

  }

  .date-input {
    display: flex;
    align-items: center;

    a.remove {
      color: var(--danger);
      vertical-align: middle;
      padding-left: .5rem;
      line-height: 1;
    }
  }

  :deep(.datepicker) {
    width: 100%;

    a.show {
      color: var(--text);
      padding: .25rem .5rem;
      transition: background-color $transition;
      border-radius: $radius;
      display: block;
      margin: .1rem 0;

      &:hover {
        background: var(--white);
      }
    }

    &.disabled a.show:hover {
      background: transparent;
    }
  }

  .details {
    padding-bottom: 0.75rem;
    flex-flow: row wrap;
    margin-bottom: 0;

    .detail-title {
      display: block;
      color: var(--grey-400);
    }

    .none {
      font-style: italic;
    }

    // Break after the 2nd element
    .column:nth-child(2n) {
      page-break-after: always; // CSS 2.1 syntax
      break-after: always; // New syntax
    }

    &.labels-list,
	.assignees {
      :deep(.multiselect) {
        .input-wrapper {
          &:not(:focus-within):not(:hover) {
            background: transparent !important;
            border-color: transparent !important;
          }
        }
      }
    }
  }

  :deep(.details),
  :deep(.heading) {
    .input:not(.has-defaults),
    .textarea,
    .select:not(.has-defaults) select {
      border-color: transparent;
      background: transparent;
      cursor: pointer;
      transition: all $transition-duration;

      &::placeholder {
        color: var(--text-light);
        opacity: 1;
        font-style: italic;
      }

      &:not(:disabled) {
        &:hover,
        &:active,
        &:focus {
          background: var(--scheme-main);
          border-color: var(--border);
          cursor: text;
        }

        &:hover,
        &:active {
          cursor: text;
          border-color: var(--link)
        }
      }
    }

    .select:not(.has-defaults):after {
      opacity: 0;
    }

    .select:not(.has-defaults):hover:after {
      opacity: 1;
    }
  }

  .attachments {
    margin-bottom: 0;

    table tr:last-child td {
      border-bottom: none;
    }
  }

  .action-buttons {
    a.button {
      width: 100%;
      margin-bottom: .5rem;
      justify-content: left;
    }
  }

  .created {
    font-size: .75rem;
    color: var(--grey-500);
    text-align: right;
  }

  .checklist-summary {
    margin-left: .25rem;
  }
}

.task-view-container {
  padding-bottom: 1rem;

  @media screen and (max-width: $desktop) {
	padding-bottom: 0;
  }

  .task-view * {
    opacity: 0;
    transition: opacity 50ms ease;
  }

  &.is-loading {
    opacity: 1;

    .task-view * {
      opacity: 0;
    }
  }

  &.visible:not(.is-loading) .task-view * {
    opacity: 1;
  }
}

.task-view-container {
  // simulate sass lighten($primary, 30) by increasing lightness 30% to 73%
  --primary-light: hsla(var(--primary-h), var(--primary-s), 73%, var(--primary-a));
}

.flash-background-enter-from,
.flash-background-enter-active  {
  animation: flash-background $flash-background-duration ease 1;
}

@keyframes flash-background {
  0% {
    background: var(--primary-light);
  }
  100% {
    background: transparent;
  }
}
</style>