<template>
	<div class="kanban-view">
		<div class="filter-container" v-if="list.isSavedFilter && !list.isSavedFilter()">
			<div class="items">
				<button @click.prevent.stop="showFilters = !showFilters" class="button">
					<span class="icon is-small">
						<icon icon="filter"/>
					</span>
					Filters
				</button>
			</div>
			<filter-popup
				@change="() => {filtersChanged = true; loadBuckets()}"
				:visible="showFilters"
				v-model="params"
			/>
		</div>
		<div :class="{ 'is-loading': loading && !oneTaskUpdating}" class="kanban loader-container">
			<div :key="`bucket${bucket.id}`" class="bucket" v-for="bucket in buckets">
				<div class="bucket-header">
					<h2
						:ref="`bucket${bucket.id}title`"
						@focusout="() => saveBucketTitle(bucket.id)"
						@keydown.enter.prevent.stop="() => saveBucketTitle(bucket.id)"
						class="title input"
						contenteditable="true"
						spellcheck="false">{{ bucket.title }}</h2>
					<span
						:class="{'is-max': bucket.tasks.length >= bucket.limit}"
						class="limit"
						v-if="bucket.limit > 0">
						{{ bucket.tasks.length }}/{{ bucket.limit }}
					</span>
					<div
						:class="{ 'is-active': bucketOptionsDropDownActive[bucket.id] }"
						class="dropdown is-right options"
						v-if="canWrite"
					>
						<div @click.stop="toggleBucketDropdown(bucket.id)" class="dropdown-trigger">
						<span class="icon">
							<icon icon="ellipsis-v"/>
						</span>
						</div>
						<div class="dropdown-menu" role="menu">
							<div class="dropdown-content">
								<a
									@click.stop="showSetLimitInput = true"
									class="dropdown-item"
								>
									<div class="field has-addons" v-if="showSetLimitInput">
										<div class="control">
											<input
												@change="() => updateBucket(bucket)"
												@keyup.enter="() => updateBucket(bucket)"
												class="input"
												type="number"
												v-focus.always
												v-model="bucket.limit"
											/>
										</div>
										<div class="control">
											<a class="button is-primary has-no-shadow">
											<span class="icon">
												<icon :icon="['far', 'save']"/>
											</span>
											</a>
										</div>
									</div>
									<template v-else>
										Limit: {{ bucket.limit > 0 ? bucket.limit : 'Not set' }}
									</template>
								</a>
								<a
									:class="{'is-disabled': buckets.length <= 1}"
									@click="() => deleteBucketModal(bucket.id)"
									class="dropdown-item has-text-danger"
									v-tooltip="buckets.length <= 1 ? 'You cannot remove the last bucket.' : ''"
								>
									<span class="icon is-small"><icon icon="trash-alt"/></span>
									Delete
								</a>
							</div>
						</div>
					</div>
				</div>
				<div :ref="`tasks-container${bucket.id}`" class="tasks">
					<!-- Make the component either a div or a draggable component based on the user rights -->
					<component
						:animation-duration="150"
						:drop-placeholder="dropPlaceholderOptions"
						:get-child-payload="getTaskPayload(bucket.id)"
						:is="canWrite ? 'Container' : 'div'"
						:should-accept-drop="() => shouldAcceptDrop(bucket)"
						@drop="e => onDrop(bucket.id, e)"
						drag-class="ghost-task"
						drag-class-drop="ghost-task-drop"
						drag-handle-selector=".task.draggable"
						group-name="buckets"
					>
						<!-- Make the component either a div or a draggable component based on the user rights -->
						<component
							:is="canWrite ? 'Draggable' : 'div'"
							:key="`bucket${bucket.id}-task${task.id}`"
							v-for="task in bucket.tasks"
						>
							<div
								:class="{
							'is-loading': (taskService.loading || taskLoading) && taskUpdating[task.id],
							'draggable': !(taskService.loading || taskLoading) || !taskUpdating[task.id],
							'has-light-text': !colorIsDark(task.hexColor) && task.hexColor !== `#${task.defaultColor}` && task.hexColor !== task.defaultColor,
						}"
								:style="{'background-color': task.hexColor !== '#' && task.hexColor !== `#${task.defaultColor}` ? task.hexColor : false}"
								@click.ctrl="() => markTaskAsDone(task)"
								@click.exact="() => $router.push({ name: 'task.kanban.detail', params: { id: task.id } })"
								@click.meta="() => markTaskAsDone(task)"
								class="task loader-container draggable"
							>
							<span class="task-id">
								<span class="is-done" v-if="task.done">Done</span>
								<template v-if="task.identifier === ''">
									#{{ task.index }}
								</template>
								<template v-else>
									{{ task.identifier }}
								</template>
							</span>
								<span
									:class="{'overdue': task.dueDate <= new Date() && !task.done}"
									class="due-date"
									v-if="task.dueDate > 0"
									v-tooltip="formatDate(task.dueDate)">
									<span class="icon">
										<icon :icon="['far', 'calendar-alt']"/>
									</span>
									<span>
										{{ formatDateSince(task.dueDate) }}
									</span>
								</span>
								<h3>{{ task.title }}</h3>
								<progress
									class="progress is-small"
									v-if="task.percentDone > 0"
									:value="task.percentDone * 100" max="100">
									{{ task.percentDone * 100 }}%
								</progress>
								<div class="footer">
									<span
										:key="label.id"
										:style="{'background': label.hexColor, 'color': label.textColor}"
										class="tag"
										v-for="label in task.labels">
										<span>{{ label.title }}</span>
									</span>
									<priority-label :priority="task.priority"/>
									<div class="assignees" v-if="task.assignees.length > 0">
										<user
											:avatar-size="24"
											:key="task.id + 'assignee' + u.id"
											:show-username="false"
											:user="u"
											v-for="u in task.assignees"
										/>
									</div>
									<span class="icon" v-if="task.attachments.length > 0">
										<icon icon="paperclip"/>	
									</span>
									<span v-if="task.description" class="icon">
										<icon icon="align-left"/>
									</span>
								</div>
							</div>
						</component>
					</component>
				</div>
				<div class="bucket-footer" v-if="canWrite">
					<div class="field" v-if="showNewTaskInput[bucket.id]">
						<div class="control" :class="{'is-loading': taskService.loading || loading}">
							<input
								class="input"
								:disabled="taskService.loading || loading"
								@focusout="toggleShowNewTaskInput(bucket.id)"
								@keyup.enter="addTaskToBucket(bucket.id)"
								@keyup.esc="toggleShowNewTaskInput(bucket.id)"
								placeholder="Enter the new task text..."
								type="text"
								v-focus.always
								v-model="newTaskText"
							/>
						</div>
						<p class="help is-danger" v-if="newTaskError[bucket.id] && newTaskText === ''">
							Please specify a title.
						</p>
					</div>
					<a
						@click="toggleShowNewTaskInput(bucket.id)"
						class="button noshadow is-transparent is-fullwidth has-text-centered"
						v-if="!showNewTaskInput[bucket.id]">
						<span class="icon is-small">
							<icon icon="plus"/>
						</span>
						<span v-if="bucket.tasks.length === 0">
						Add a task
						</span>
						<span v-else>
						Add another task
					</span>
					</a>
				</div>
			</div>

			<div class="bucket new-bucket" v-if="canWrite && !loading && buckets.length > 0">
				<input
					:class="{'is-loading': loading}"
					:disabled="loading"
					@focusout="() => showNewBucketInput = false"
					@keyup.enter="createNewBucket"
					@keyup.esc="() => showNewBucketInput = false"
					class="input"
					placeholder="Enter the new bucket title..."
					type="text"
					v-focus.always
					v-if="showNewBucketInput"
					v-model="newBucketTitle"
				/>
				<a
					@click="() => showNewBucketInput = true"
					class="button noshadow is-transparent is-fullwidth has-text-centered" v-if="!showNewBucketInput">
					<span class="icon is-small">
						<icon icon="plus"/>
					</span>
					<span>
						Create a new bucket
					</span>
				</a>
			</div>
		</div>

		<!-- This router view is used to show the task popup while keeping the kanban board itself -->
		<transition name="modal">
			<router-view/>
		</transition>

		<modal
			@close="showBucketDeleteModal = false"
			@submit="deleteBucket()"
			v-if="showBucketDeleteModal">
			<span slot="header">Delete the bucket</span>
			<p slot="text">
				Are you sure you want to delete this bucket?<br/>
				This will not delete any tasks but move them into the default bucket.
			</p>
		</modal>

	</div>
</template>

<script>
import TaskService from '../../../services/task'
import TaskModel from '../../../models/task'
import BucketModel from '../../../models/bucket'

import {Container, Draggable} from 'vue-smooth-dnd'
import PriorityLabel from '../../../components/tasks/partials/priorityLabel'
import User from '../../../components/misc/user'
import Labels from '../../../components/tasks/partials/labels'

import {filterObject} from '@/helpers/filterObject'
import {applyDrag} from '@/helpers/applyDrag'
import {mapState} from 'vuex'
import {saveListView} from '@/helpers/saveListView'
import Rights from '../../../models/rights.json'
import { LOADING, LOADING_MODULE } from '../../../store/mutation-types'
import FilterPopup from '@/components/list/partials/filter-popup'

export default {
	name: 'Kanban',
	components: {
		FilterPopup,
		Container,
		Draggable,
		Labels,
		User,
		PriorityLabel,
	},
	data() {
		return {
			taskService: TaskService,

			dropPlaceholderOptions: {
				className: 'drop-preview',
				animationDuration: 150,
				showOnTop: true,
			},
			sourceBucket: 0,
			bucketOptionsDropDownActive: {},

			showBucketDeleteModal: false,
			bucketToDelete: 0,

			newTaskText: '',
			showNewTaskInput: {},
			newBucketTitle: '',
			showNewBucketInput: false,
			newTaskError: {},
			showSetLimitInput: false,

			// We're using this to show the loading animation only at the task when updating it
			taskUpdating: {},
			oneTaskUpdating: false,

			params: {
				filter_by: [],
				filter_value: [],
				filter_comparator: [],
				filter_concat: 'and',
			},
			showFilters: false,
			filtersChanged: false, // To trigger a reload of the board
		}
	},
	created() {
		this.taskService = new TaskService()
		this.loadBuckets()
		this.$nextTick(() => document.addEventListener('click', this.closeBucketDropdowns))

		// Save the current list view to local storage
		// We use local storage and not vuex here to make it persistent across reloads.
		saveListView(this.$route.params.listId, this.$route.name)
	},
	watch: {
		'$route.params.listId': 'loadBuckets',
	},
	computed: mapState({
		buckets: state => state.kanban.buckets,
		loadedListId: state => state.kanban.listId,
		loading: state => state[LOADING] && state[LOADING_MODULE] === 'kanban',
		taskLoading: state => state[LOADING] && state[LOADING_MODULE] === 'tasks',
		canWrite: state => state.currentList.maxRight > Rights.READ,
		list: state => state.currentList,
	}),
	methods: {
		loadBuckets() {

			// Prevent trying to load buckets if the task popup view is active
			if (this.$route.name !== 'list.kanban') {
				return
			}

			// Only load buckets if we don't already loaded them
			if (
				!this.filtersChanged && (
				this.loadedListId === this.$route.params.listId ||
				this.loadedListId === parseInt(this.$route.params.listId))
			) {
				return
			}

			console.debug(`Loading buckets, loadedListId = ${this.loadedListId}, $route.params =`, this.$route.params)
			this.filtersChanged = false

			this.$store.dispatch('kanban/loadBucketsForList', {listId: this.$route.params.listId, params: this.params})
				.catch(e => {
					this.error(e, this)
				})
		},
		onDrop(bucketId, dropResult) {

			// Note: A lot of this example comes from the excellent kanban example on https://github.com/kutlugsahin/vue-smooth-dnd/blob/master/demo/src/pages/cards.vue

			const bucketIndex = filterObject(this.buckets, b => b.id === bucketId)

			if (dropResult.removedIndex !== null || dropResult.addedIndex !== null) {

				// FIXME: This is probably not the best solution and more of a naive brute-force approach

				// Duplicate the buckets to avoid stuff moving around without noticing
				const buckets = Object.assign({}, this.buckets)
				// Get the index of the bucket and the bucket itself
				const bucket = buckets[bucketIndex]

				// Rebuild the tasks from the bucket, removing/adding the moved task
				bucket.tasks = applyDrag(bucket.tasks, dropResult)
				// Update the bucket in the list of all buckets
				delete buckets[bucketIndex]
				buckets[bucketIndex] = bucket
				// Set the buckets, triggering a state update in vue
				// FIXME: This seems to set some task attributes (like due date) wrong. Commented out, but seems to still work?
				//   Not sure what to do about this.
				// this.$store.commit('kanban/setBuckets', buckets)
			}

			if (dropResult.addedIndex !== null) {

				const taskIndex = dropResult.addedIndex
				const taskBefore = typeof this.buckets[bucketIndex].tasks[taskIndex - 1] === 'undefined' ? null : this.buckets[bucketIndex].tasks[taskIndex - 1]
				const taskAfter = typeof this.buckets[bucketIndex].tasks[taskIndex + 1] === 'undefined' ? null : this.buckets[bucketIndex].tasks[taskIndex + 1]
				const task = this.buckets[bucketIndex].tasks[taskIndex]
				this.$set(this.taskUpdating, task.id, true)
				this.oneTaskUpdating = true

				// If there is no task before, our task is the first task in which case we let it have half of the position of the task after it
				if (taskBefore === null && taskAfter !== null) {
					task.position = taskAfter.position / 2
				}
				// If there is no task after it, we just add 2^16 to the last position
				if (taskBefore !== null && taskAfter === null) {
					task.position = taskBefore.position + Math.pow(2, 16)
				}
				// If we have both a task before and after it, we acually calculate the position
				if (taskAfter !== null && taskBefore !== null) {
					task.position = taskBefore.position + (taskAfter.position - taskBefore.position) / 2
				}

				task.bucketId = bucketId

				this.$store.dispatch('tasks/update', task)
					.catch(e => {
						this.error(e, this)
					})
					.finally(() => {
						this.$set(this.taskUpdating, task.id, false)
						this.oneTaskUpdating = false
					})
			}
		},
		markTaskAsDone(task) {
			this.oneTaskUpdating = true
			this.$set(this.taskUpdating, task.id, true)
			task.done = !task.done
			this.$store.dispatch('tasks/update', task)
				.catch(e => {
					this.error(e, this)
				})
				.finally(() => {
					this.$set(this.taskUpdating, task.id, false)
					this.oneTaskUpdating = false
				})
		},
		getTaskPayload(bucketId) {
			return index => {
				const bucket = this.buckets[filterObject(this.buckets, b => b.id === bucketId)]
				this.sourceBucket = bucket.id
				return bucket.tasks[index]
			}
		},
		toggleShowNewTaskInput(bucket) {
			this.$set(this.showNewTaskInput, bucket, !this.showNewTaskInput[bucket])
		},
		toggleBucketDropdown(bucketId) {
			this.closeBucketDropdowns() // Close all eventually open dropdowns
			this.$set(this.bucketOptionsDropDownActive, bucketId, !this.bucketOptionsDropDownActive[bucketId])
		},
		closeBucketDropdowns() {
			this.showSetLimitInput = false
			for (const bucketId in this.bucketOptionsDropDownActive) {
				this.bucketOptionsDropDownActive[bucketId] = false
			}
		},
		addTaskToBucket(bucketId) {

			if (this.newTaskText === '') {
				this.$set(this.newTaskError, bucketId, true)
				return
			}
			this.$set(this.newTaskError, bucketId, false)

			// We need the actual bucket index so we put that in a seperate function
			const bucketIndex = () => {
				for (const t in this.buckets) {
					if (this.buckets[t].id === bucketId) {
						return t
					}
				}
			}

			const bi = bucketIndex()

			const task = new TaskModel({
				title: this.newTaskText,
				bucketId: this.buckets[bi].id,
				listId: this.$route.params.listId,
			})

			this.taskService.create(task)
				.then(r => {
					this.newTaskText = ''
					this.$store.commit('kanban/addTaskToBucket', r)
				})
				.catch(e => {
					this.error(e, this)
				})
				.finally(() => {
					if (!this.$refs[`tasks-container${task.bucketId}`][0]) {
						return
					}
					this.$refs[`tasks-container${task.bucketId}`][0].scrollTop = this.$refs[`tasks-container${task.bucketId}`][0].scrollHeight
				})
		},
		createNewBucket() {
			if (this.newBucketTitle === '') {
				return
			}

			const newBucket = new BucketModel({
				title: this.newBucketTitle,
				listId: parseInt(this.$route.params.listId),
			})

			this.$store.dispatch('kanban/createBucket', newBucket)
				.then(() => {
					this.newBucketTitle = ''
					this.showNewBucketInput = false
				})
				.catch(e => {
					this.error(e, this)
				})
		},
		deleteBucketModal(bucketId) {
			if (this.buckets.length <= 1) {
				return
			}

			this.bucketToDelete = bucketId
			this.showBucketDeleteModal = true
		},
		deleteBucket() {
			const bucket = new BucketModel({
				id: this.bucketToDelete,
				listId: this.$route.params.listId,
			})

			this.$store.dispatch('kanban/deleteBucket', bucket)
				.then(() => {
					this.success({message: 'The bucket has been deleted successfully.'}, this)
				})
				.catch(e => {
					this.error(e, this)
				})
				.finally(() => {
					this.showBucketDeleteModal = false
				})
		},
		saveBucketTitle(bucketId) {
			const bucketTitleElement = this.$refs[`bucket${bucketId}title`][0]
			const bucketTitle = bucketTitleElement.textContent
			const bucket = new BucketModel({
				id: bucketId,
				title: bucketTitle,
				listId: Number(this.$route.params.listId),
			})

			// Because the contenteditable does not have a change event,
			// we're building it ourselves here and only updating the bucket
			// if the title changed.
			const realBucket = this.buckets[filterObject(this.buckets, b => b.id === bucketId)]
			if (realBucket.title === bucketTitle) {
				return
			}

			this.$store.dispatch('kanban/updateBucket', bucket)
				.then(r => {
					realBucket.title = r.title
					bucketTitleElement.blur()
					this.success({message: 'The bucket title has been saved successfully.'}, this)
				})
				.catch(e => {
					this.error(e, this)
				})
		},
		updateBucket(bucket) {
			bucket.limit = parseInt(bucket.limit)
			this.$store.dispatch('kanban/updateBucket', bucket)
				.then(() => {
					this.success({message: 'The bucket limit been saved successfully.'}, this)
				})
				.catch(e => {
					this.error(e, this)
				})
		},
		shouldAcceptDrop(bucket) {
			return bucket.id === this.sourceBucket || // When dragging from a bucket who has its limit reached, dragging should still be possible
				bucket.limit === 0 || // If there is no limit set, dragging & dropping should always work
				bucket.tasks.length < bucket.limit // Disallow dropping to buckets which have their limit reached
		},
	},
}
</script>
