<template>
	<div class="kanban loader-container" :class="{ 'is-loading': loading}">
		<div v-for="bucket in buckets" :key="`bucket${bucket.id}`" class="bucket">
			<div class="bucket-header">
				<h2
						class="title input"
						contenteditable="true"
						@focusout="() => saveBucketTitle(bucket.id)"
						:ref="`bucket${bucket.id}title`"
						@keyup.ctrl.enter="() => saveBucketTitle(bucket.id)">{{ bucket.title }}</h2>
				<div class="dropdown is-right options" :class="{ 'is-active': bucketOptionsDropDownActive[bucket.id] }">
					<div class="dropdown-trigger" @click.stop="toggleBucketDropdown(bucket.id)">
						<span class="icon">
							<icon icon="ellipsis-v"/>
						</span>
					</div>
					<div class="dropdown-menu" role="menu">
						<div class="dropdown-content">
							<a
									class="dropdown-item has-text-danger"
									@click="() => deleteBucketModal(bucket.id)"
									:class="{'is-disabled': buckets.length <= 1}"
									v-tooltip="buckets.length <= 1 ? 'You cannot remove the last bucket.' : ''"
							>
								<span class="icon is-small"><icon icon="trash-alt"/></span>
								Delete
							</a>
						</div>
					</div>
				</div>
			</div>
			<div class="tasks">
				<Container
						@drop="e => onDrop(bucket.id, e)"
						group-name="buckets"
						:get-child-payload="getTaskPayload(bucket.id)"
						:drop-placeholder="dropPlaceholderOptions"
						:animation-duration="150"
						drag-class="ghost-task"
						drag-class-drop="ghost-task-drop"
						drag-handle-selector=".task.draggable"
				>
					<Draggable v-for="task in bucket.tasks" :key="`bucket${bucket.id}-task${task.id}`">
						<router-link
								:to="{ name: 'task.kanban.detail', params: { id: task.id } }"
								class="task loader-container draggable"
								tag="div"
								:class="{
							'is-loading': taskService.loading && taskUpdating[task.id],
							'draggable': !taskService.loading || !taskUpdating[task.id],
							'has-light-text': !colorIsDark(task.hexColor) && task.hexColor !== `#${task.defaultColor}`,
						}"
								:style="{'background-color': task.hexColor !== '#' && task.hexColor !== `#${task.defaultColor}` ? task.hexColor : false}"
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
									v-if="task.dueDate > 0"
									class="due-date"
									:class="{'overdue': task.dueDate <= new Date() && !task.done}"
									v-tooltip="formatDate(task.dueDate)">
								<span class="icon">
									<icon :icon="['far', 'calendar-alt']"/>
								</span>
								<span>
									{{ formatDateSince(task.dueDate) }}
								</span>
							</span>
							<h3>{{ task.title }}</h3>
							<labels :labels="task.labels"/>
							<div class="footer">
								<div class="items">
									<priority-label :priority="task.priority" class="priority-label"/>
									<div class="assignees" v-if="task.assignees.length > 0">
										<user
												v-for="u in task.assignees"
												:key="task.id + 'assignee' + u.id"
												:user="u"
												:show-username="false"
												:avatar-size="24"
										/>
									</div>
								</div>
								<div>
									<span class="icon" v-if="task.attachments.length > 0">
										<svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
											<rect fill="none" rx="0" ry="0"></rect>
											<path
													fill-rule="evenodd"
													clip-rule="evenodd"
													d="M19.86 8.29994C19.8823 8.27664 19.9026 8.25201 19.9207 8.22634C20.5666 7.53541 20.93 6.63567 20.93 5.68001C20.93 4.69001 20.55 3.76001 19.85 3.06001C18.45 1.66001 16.02 1.66001 14.62 3.06001L9.88002 7.80001C9.86705 7.81355 9.85481 7.82753 9.8433 7.8419L4.58 13.1C3.6 14.09 3.06 15.39 3.06 16.78C3.06 18.17 3.6 19.48 4.58 20.46C5.6 21.47 6.93 21.98 8.26 21.98C9.59 21.98 10.92 21.47 11.94 20.46L17.74 14.66C17.97 14.42 17.98 14.04 17.74 13.81C17.5 13.58 17.12 13.58 16.89 13.81L11.09 19.61C10.33 20.36 9.33 20.78 8.26 20.78C7.19 20.78 6.19 20.37 5.43 19.61C4.68 18.85 4.26 17.85 4.26 16.78C4.26 15.72 4.68 14.71 5.43 13.96L15.47 3.91996C15.4962 3.89262 15.5195 3.86346 15.54 3.83292C16.4992 2.95103 18.0927 2.98269 19.01 3.90001C19.48 4.37001 19.74 5.00001 19.74 5.67001C19.74 6.34001 19.48 6.97001 19.01 7.44001L14.27 12.18C14.2571 12.1935 14.2448 12.2075 14.2334 12.2218L8.96 17.4899C8.59 17.8699 7.93 17.8699 7.55 17.4899C7.36 17.2999 7.26 17.0399 7.26 16.7799C7.26 16.5199 7.36 16.2699 7.55 16.0699L15.47 8.14994C15.7 7.90994 15.71 7.52994 15.47 7.29994C15.23 7.06994 14.85 7.06994 14.62 7.29994L6.7 15.2199C6.29 15.6399 6.06 16.1899 6.06 16.7799C6.06 17.3699 6.29 17.9199 6.7 18.3399C7.12 18.7499 7.67 18.9799 8.26 18.9799C8.85 18.9799 9.4 18.7599 9.82 18.3399L19.86 8.29994Z"></path>
										</svg>
									</span>
								</div>
							</div>
						</router-link>
					</Draggable>
				</Container>
			</div>
			<div class="bucket-footer">
				<div class="field" v-if="showNewTaskInput[bucket.id]">
					<div class="control">
						<input

								class="input"
								type="text"
								placeholder="Enter the new task text..."
								v-focus
								@focusout="toggleShowNewTaskInput(bucket.id)"
								@keyup.esc="toggleShowNewTaskInput(bucket.id)"
								@keyup.enter="addTaskToBucket(bucket.id)"
								v-model="newTaskText"
								:disabled="taskService.loading"
								:class="{'is-loading': taskService.loading}"
						/>
					</div>
					<p class="help is-danger" v-if="newTaskError[bucket.id] && newTaskText === ''">
						Please specify a title.
					</p>
				</div>
				<a
						class="button noshadow is-transparent is-fullwidth has-text-centered"
						@click="toggleShowNewTaskInput(bucket.id)"
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

		<div class="bucket new-bucket" v-if="!loading">
			<input
					v-if="showNewBucketInput"
					class="input"
					type="text"
					placeholder="Enter the new bucket title..."
					v-focus
					@focusout="() => showNewBucketInput = false"
					@keyup.esc="() => showNewBucketInput = false"
					@keyup.enter="createNewBucket"
					v-model="newBucketTitle"
					:disabled="loading"
					:class="{'is-loading': loading}"
			/>
			<a
					class="button noshadow is-transparent is-fullwidth has-text-centered"
					@click="() => showNewBucketInput = true" v-if="!showNewBucketInput">
				<span class="icon is-small">
					<icon icon="plus"/>
				</span>
				<span>
					Create a new bucket
				</span>
			</a>
		</div>

		<!-- This router view is used to show the task popup while keeping the kanban board itself -->
		<transition name="modal">
			<router-view/>
		</transition>

		<modal
				v-if="showBucketDeleteModal"
				@close="showBucketDeleteModal = false"
				@submit="deleteBucket()">
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

	import {filterObject} from '../../../helpers/filterObject'
	import {applyDrag} from '../../../helpers/applyDrag'
	import {mapState} from 'vuex'
	import {LOADING} from '../../../store/mutation-types'
	import {saveListView} from '../../../helpers/saveListView'

	export default {
		name: 'Kanban',
		components: {
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
				bucketOptionsDropDownActive: {},

				showBucketDeleteModal: false,
				bucketToDelete: 0,

				newTaskText: '',
				showNewTaskInput: {},
				newBucketTitle: '',
				showNewBucketInput: false,
				newTaskError: {},

				// We're using this to show the loading animation only at the task when updating it
				taskUpdating: {},
			}
		},
		created() {
			this.taskService = new TaskService()
			this.loadBuckets()
			setTimeout(() => document.addEventListener('click', this.closeBucketDropdowns), 0)

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
			loading: LOADING,
		}),
		methods: {
			loadBuckets() {

				// Prevent trying to load buckets if the task popup view is active
				if (this.$route.name !== 'list.kanban') {
					return
				}

				// Only load buckets if we don't already loaded them
				if (this.loadedListId === this.$route.params.listId) {
					return
				}

				this.$store.dispatch('kanban/loadBucketsForList', this.$route.params.listId)
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
						})
				}
			},
			getTaskPayload(bucketId) {
				return index => {
					const bucket = this.buckets[filterObject(this.buckets, b => b.id === bucketId)]
					return bucket.tasks[index]
				}
			},
			toggleShowNewTaskInput(bucket) {
				this.$set(this.showNewTaskInput, bucket, !this.showNewTaskInput[bucket])
			},
			toggleBucketDropdown(bucketId) {
				this.$set(this.bucketOptionsDropDownActive, bucketId, !this.bucketOptionsDropDownActive[bucketId])
			},
			closeBucketDropdowns() {
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
					listId: this.$route.params.listId
				})

				this.taskService.create(task)
					.then(r => {
						this.newTaskText = ''
						this.$store.commit('kanban/addTaskToBucket', r)
					})
					.catch(e => {
						this.error(e, this)
					})
			},
			createNewBucket() {
				if (this.newBucketTitle === '') {
					return
				}

				const newBucket = new BucketModel({
					title: this.newBucketTitle,
					listId: parseInt(this.$route.params.listId)
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
					.catch(e => {
						this.error(e, this)
					})
					.finally(() => {
						this.showBucketDeleteModal = false
					})
			},
			saveBucketTitle(bucketId) {
				const bucketTitle = this.$refs[`bucket${bucketId}title`][0].textContent
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
					})
					.catch(e => {
						this.error(e, this)
					})
			},
		},
	}
</script>
