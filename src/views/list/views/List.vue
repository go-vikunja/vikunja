<template>
	<div
		:class="{ 'is-loading': taskCollectionService.loading}"
		class="loader-container is-max-width-desktop list-view">
		<div class="filter-container" v-if="list.isSavedFilter && !list.isSavedFilter()">
			<div class="items">
				<div class="search">
					<div :class="{ 'hidden': !showTaskSearch }" class="field has-addons">
						<div class="control has-icons-left has-icons-right">
							<input
								@blur="hideSearchBar()"
								@keyup.enter="searchTasks"
								class="input"
								placeholder="Search"
								type="text"
								v-focus
								v-model="searchTerm"/>
							<span class="icon is-left">
								<icon icon="search"/>
							</span>
						</div>
						<div class="control">
							<x-button
								:loading="taskCollectionService.loading"
								@click="searchTasks"
								:shadow="false"
							>
								Search
							</x-button>
						</div>
					</div>
					<x-button
						@click="showTaskSearch = !showTaskSearch"
						icon="search"
						type="secondary"
						v-if="!showTaskSearch"
					/>
				</div>
				<x-button
					@click.prevent.stop="showTaskFilter = !showTaskFilter"
					type="secondary"
					icon="filter"
				>
					Filters
				</x-button>
			</div>
			<filter-popup
				@change="loadTasks(1)"
				:visible="showTaskFilter"
				v-model="params"
			/>
		</div>

		<card :padding="false" :has-content="false" class="has-overflow">
			<div class="field task-add" v-if="!list.isArchived && canWrite && list.id > 0">
				<div class="field is-grouped">
					<p :class="{ 'is-loading': taskService.loading}" class="control has-icons-left is-expanded">
						<input
							:class="{ 'disabled': taskService.loading}"
							@keyup.enter="addTask()"
							class="input"
							placeholder="Add a new task..."
							type="text"
							v-focus
							v-model="newTaskText"
							ref="newTaskInput"
						/>
						<span class="icon is-small is-left">
						<icon icon="tasks"/>
					</span>
					</p>
					<p class="control">
						<x-button
							:disabled="newTaskText.length === 0"
							@click="addTask()"
							icon="plus"
						>
							Add
						</x-button>
					</p>
				</div>
				<p class="help is-danger" v-if="showError && newTaskText === ''">
					Please specify a list title.
				</p>
			</div>

			<nothing v-if="ctaVisible && tasks.length === 0 && !taskCollectionService.loading">
				This list is currently empty.
				<a @click="$refs.newTaskInput.focus()">Create a new task.</a>
			</nothing>

			<div class="tasks-container">
				<div :class="{'short': isTaskEdit}" class="tasks mt-0" v-if="tasks && tasks.length > 0">
					<single-task-in-list
						:show-list-color="false"
						:disabled="!canWrite"
						:key="t.id"
						:the-task="t"
						@taskUpdated="updateTasks"
						task-detail-route="task.detail"
						v-for="t in tasks"
					>
						<div @click="editTask(t.id)" class="icon settings" v-if="!list.isArchived && canWrite">
							<icon icon="pencil-alt"/>
						</div>
					</single-task-in-list>
				</div>
				<card
					v-if="isTaskEdit"
					class="taskedit mt-0" title="Edit Task" :has-close="true" @close="() => isTaskEdit = false"
					:shadow="false">
					<edit-task :task="taskEditTask"/>
				</card>
			</div>

			<nav
				aria-label="pagination"
				class="pagination is-centered p-4"
				role="navigation"
				v-if="taskCollectionService.totalPages > 1">
				<router-link
					:disabled="currentPage === 1"
					:to="getRouteForPagination(currentPage - 1)"
					class="pagination-previous"
					tag="button">
					Previous
				</router-link>
				<router-link
					:disabled="currentPage === taskCollectionService.totalPages"
					:to="getRouteForPagination(currentPage + 1)"
					class="pagination-next"
					tag="button">
					Next page
				</router-link>
				<ul class="pagination-list">
					<template v-for="(p, i) in pages">
						<li :key="'page'+i" v-if="p.isEllipsis"><span class="pagination-ellipsis">&hellip;</span></li>
						<li :key="'page'+i" v-else>
							<router-link
								:aria-label="'Goto page ' + p.number"
								:class="{'is-current': p.number === currentPage}"
								:to="getRouteForPagination(p.number)"
								class="pagination-link">
								{{ p.number }}
							</router-link>
						</li>
					</template>
				</ul>
			</nav>
		</card>

		<!-- This router view is used to show the task popup while keeping the kanban board itself -->
		<transition name="modal">
			<router-view/>
		</transition>

	</div>
</template>

<script>
import TaskService from '../../../services/task'
import TaskModel from '../../../models/task'
import LabelTaskService from '../../../services/labelTask'
import LabelTask from '../../../models/labelTask'
import LabelModel from '../../../models/label'

import EditTask from '../../../components/tasks/edit-task'
import SingleTaskInList from '../../../components/tasks/partials/singleTaskInList'
import taskList from '../../../components/tasks/mixins/taskList'
import {saveListView} from '@/helpers/saveListView'
import Rights from '../../../models/rights.json'
import {mapState} from 'vuex'
import FilterPopup from '@/components/list/partials/filter-popup'
import {HAS_TASKS} from '@/store/mutation-types'
import Nothing from '@/components/misc/nothing'

export default {
	name: 'List',
	data() {
		return {
			taskService: TaskService,
			isTaskEdit: false,
			taskEditTask: TaskModel,
			newTaskText: '',

			showError: false,
			labelTaskService: LabelTaskService,

			ctaVisible: false,
		}
	},
	mixins: [
		taskList,
	],
	components: {
		Nothing,
		FilterPopup,
		SingleTaskInList,
		EditTask,
	},
	created() {
		this.taskService = new TaskService()
		this.labelTaskService = new LabelTaskService()

		// Save the current list view to local storage
		// We use local storage and not vuex here to make it persistent across reloads.
		saveListView(this.$route.params.listId, this.$route.name)
	},
	computed: mapState({
		canWrite: state => state.currentList.maxRight > Rights.READ,
		list: state => state.currentList,
	}),
	mounted() {
		this.$nextTick(() => this.ctaVisible = true)
	},
	methods: {
		// This function initializes the tasks page and loads the first page of tasks
		initTasks(page, search = '') {
			this.taskEditTask = null
			this.isTaskEdit = false
			this.loadTasks(page, search)
		},
		addTask() {
			if (this.newTaskText === '') {
				this.showError = true
				return
			}
			this.showError = false

			let task = new TaskModel({title: this.newTaskText, listId: this.$route.params.listId})
			this.taskService.create(task)
				.then(task => {
					this.tasks.push(task)
					this.sortTasks()
					this.newTaskText = ''

					// Unlike a proper programming language, Javascript only knows references to objects and does not
					// allow you to control what is a reference and what isnt. Because of this we can't just add
					// all labels to the task they belong to right after we found and added them to the task since
					// the task update method also ensures all data the api sees has the right format. That means
					// it processes labels. That processing changes the date format and the label color and makes
					// the label pretty much unusable for everything else. Normally, this is not a big deal, because
					// the labels on a task get thrown away anyway and replaced with the new models from the api
					// when we get the updated answer back. However, in this specific case because we're passing a
					// label we obtained from vuex that reference is kept and not thrown away. The task itself gets
					// a new label object - you won't notice the bad reference until you want to add the same label
					// again and notice it doesn't have a color anymore.
					// I think this is what happens: (or rather would happen without the hack I've put in)
					//   1. Query the store for a label which matches the name
					//   2. Find one - remember, we get only a *reference* to the label from the store, not a new label object.
					//      (Now there's *two* places with a reference to the same label object: in the store and in the
					//      variable which holds the label from the search in the store)
					//   3. .push the label to the task
					//   4. Update the task to remove the labels from the name
					//   4.1. The task update processes all labels belonging to that task, changing attributes of our
					//        label in the process. Because this is a reference, it is also "updated" in the store.
					//   5. Get an api response back. The service handler now creates a new label object for all labels
					//      returned from the api. It will throw away all references to the old label in the process.
					//   6. Now we have two objects with the same label data: The old one we originally obtained from
					//      the store and the one that was created when parsing the api response. The old one was
					//      modified before sending the api request and thus, our store which still holds a reference
					//      to the old label now contains old data.
					//      I guess this is the point where normally the GC would come in and collect the old label
					//      object if the store wouldn't still hold a reference to it.
					//
					// Now, as a workaround, I'm putting all new labels added to that task in this separate variable to
					// add them only after the task was updated to circumvent the task update service processing the
					// label before sending it. Feels more hacky than it probably is.
					const newLabels = []

					// Check if the task has words starting with ~ in the title and make them to labels
					const parts = task.title.split(' ~')
					// The first element will always contain the title, even if there is no occurrence of ~
					if (parts.length > 1) {

						// First, create an unresolved promise for each entry in the array to wait
						// until all labels are added to update the task title once again
						let labelAddings = []
						let labelAddsToWaitFor = []
						parts.forEach((p, index) => {
							if (index < 1) {
								return
							}

							labelAddsToWaitFor.push(new Promise((resolve, reject) => {
								labelAddings.push({resolve: resolve, reject: reject})
							}))
						})

						// Then do everything that is involved in finding, creating and adding the label to the task
						parts.forEach((p, index) => {
							if (index < 1) {
								return
							}

							// The part up until the next space
							const labelTitle = p.split(' ')[0]

							// Don't create an empty label
							if (labelTitle === '') {
								return
							}

							// Check if the label exists
							const label = Object.values(this.$store.state.labels.labels).find(l => {
								return l.title.toLowerCase() === labelTitle.toLowerCase()
							})

							// Label found, use it
							if (typeof label !== 'undefined') {
								const labelTask = new LabelTask({
									taskId: task.id,
									labelId: label.id,
								})
								this.labelTaskService.create(labelTask)
									.then(result => {
										newLabels.push(label)

										// Remove the label text from the task title
										task.title = task.title.replace(` ~${labelTitle}`, '')

										// Make the promise done (the one with the index 0 does not exist)
										labelAddings[index - 1].resolve(result)
									})
									.catch(e => {
										this.error(e)
									})
							} else {
								// label not found, create it
								const label = new LabelModel({title: labelTitle})
								this.$store.dispatch('labels/createLabel', label)
									.then(res => {
										const labelTask = new LabelTask({
											taskId: task.id,
											labelId: res.id,
										})
										this.labelTaskService.create(labelTask)
											.then(result => {
												newLabels.push(res)

												// Remove the label text from the task title
												task.title = task.title.replace(` ~${labelTitle}`, '')

												// Make the promise done (the one with the index 0 does not exist)
												labelAddings[index - 1].resolve(result)
											})
											.catch(e => {
												this.error(e)
											})
									})
									.catch(e => {
										this.error(e)
									})
							}
						})

						// This waits to update the task until all labels have been added and the title has
						// been modified to remove each label text
						Promise.all(labelAddsToWaitFor)
							.then(() => {
								this.taskService.update(task)
									.then(updatedTask => {
										updatedTask.labels = newLabels
										this.updateTasks(updatedTask)
										this.$store.commit(HAS_TASKS, true)
									})
									.catch(e => {
										this.error(e)
									})
							})
					}
				})
				.catch(e => {
					this.error(e)
				})
		},
		editTask(id) {
			// Find the selected task and set it to the current object
			let theTask = this.getTaskById(id) // Somehow this does not work if we directly assign this to this.taskEditTask
			this.taskEditTask = theTask
			this.isTaskEdit = true
		},
		getTaskById(id) {
			for (const t in this.tasks) {
				if (this.tasks[t].id === parseInt(id)) {
					return this.tasks[t]
				}
			}
			return {} // FIXME: This should probably throw something to make it clear to the user noting was found
		},
		updateTasks(updatedTask) {
			for (const t in this.tasks) {
				if (this.tasks[t].id === updatedTask.id) {
					this.$set(this.tasks, t, updatedTask)
					break
				}
			}
			this.sortTasks()
		},
	},
}
</script>