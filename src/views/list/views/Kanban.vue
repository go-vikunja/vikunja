<template>
	<div class="kanban-view">
		<div class="filter-container" v-if="list.isSavedFilter && !list.isSavedFilter()">
			<div class="items">
				<x-button
					@click.prevent.stop="showFilters = !showFilters"
					icon="filter"
					type="secondary"
				>
					{{ $t('filters.title') }}
				</x-button>
			</div>
			<filter-popup
				@change="() => {filtersChanged = true; loadBuckets()}"
				:visible="showFilters"
				v-model="params"
			/>
		</div>
		<div
			:class="{ 'is-loading': loading && !oneTaskUpdating}"
			class="kanban kanban-bucket-container loader-container">
			<draggable
				v-bind="dragOptions"
				v-model="buckets"
				@start="() => dragBucket = true"
				@end="updateBucketPosition"
				group="buckets"
				:disabled="!canWrite"
				:class="{'dragging-disabled': !canWrite}"
			>
				<transition-group
					type="transition"
					:name="!dragBucket ? 'move-bucket': null"
					tag="div"
					class="kanban-bucket-container">
					<div
						:key="`bucket${bucket.id}`"
						class="bucket"
						:class="{'is-collapsed': collapsedBuckets[bucket.id]}"
						v-for="(bucket, k) in buckets"
					>
						<div class="bucket-header" @click="() => unCollapseBucket(bucket)">
							<span
								v-if="bucket.isDoneBucket"
								class="icon is-small has-text-success mr-2"
								v-tooltip="$t('list.kanban.doneBucketHint')"
							>
								<icon icon="check-double"/>
							</span>
							<h2
								:ref="`bucket${bucket.id}title`"
								@focusout="() => saveBucketTitle(bucket.id)"
								@keydown.enter.prevent.stop="() => saveBucketTitle(bucket.id)"
								@click="focusBucketTitle"
								class="title input"
								:contenteditable="bucketTitleEditable && canWrite && !collapsedBuckets[bucket.id]"
								spellcheck="false">{{ bucket.title }}</h2>
							<span
								:class="{'is-max': bucket.tasks.length >= bucket.limit}"
								class="limit"
								v-if="bucket.limit > 0">
								{{ bucket.tasks.length }}/{{ bucket.limit }}
							</span>
							<dropdown
								class="is-right options"
								v-if="canWrite && !collapsedBuckets[bucket.id]"
								trigger-icon="ellipsis-v"
								@close="() => showSetLimitInput = false"
							>
								<a
									@click.stop="showSetLimitInput = true"
									class="dropdown-item"
								>
									<div class="field has-addons" v-if="showSetLimitInput">
										<div class="control">
											<input
												@change="() => setBucketLimit(bucket)"
												@keyup.enter="() => setBucketLimit(bucket)"
												@keyup.esc="() => showSetLimitInput = false"
												class="input"
												type="number"
												min="0"
												v-focus.always
												v-model="bucket.limit"
											/>
										</div>
										<div class="control">
											<x-button
												:disabled="bucket.limit < 0"
												:icon="['far', 'save']"
												:shadow="false"
											/>
										</div>
									</div>
									<template v-else>
										{{
											$t('list.kanban.limit', {limit: bucket.limit > 0 ? bucket.limit : $t('list.kanban.noLimit')})
										}}
									</template>
								</a>
								<a
									@click.stop="toggleDoneBucket(bucket)"
									class="dropdown-item"
									v-tooltip="$t('list.kanban.doneBucketHintExtended')"
								>
									<span class="icon is-small" :class="{'has-text-success': bucket.isDoneBucket}">
										<icon icon="check-double"/>
									</span>
									{{ $t('list.kanban.doneBucket') }}
								</a>
								<a
									class="dropdown-item"
									@click.stop="() => collapseBucket(bucket)"
								>
									{{ $t('list.kanban.collapse') }}
								</a>
								<a
									:class="{'is-disabled': buckets.length <= 1}"
									@click.stop="() => deleteBucketModal(bucket.id)"
									class="dropdown-item has-text-danger"
									v-tooltip="buckets.length <= 1 ? $t('list.kanban.deleteLast') : ''"
								>
									<span class="icon is-small">
										<icon icon="trash-alt"/>
									</span>
									{{ $t('misc.delete') }}
								</a>
							</dropdown>
						</div>
						<div :ref="`tasks-container${bucket.id}`" class="tasks">
							<draggable
								v-bind="dragOptions"
								v-model="bucket.tasks"
								@start="() => dragstart(bucket)"
								@end="updateTaskPosition"
								:group="{name: 'tasks', put: shouldAcceptDrop(bucket) && !dragBucket}"
								:disabled="!canWrite"
								:class="{'dragging-disabled': !canWrite}"
								:data-bucket-index="k"
								class="dropper"
							>
								<transition-group type="transition" :name="!drag ? 'move-card': null" tag="div">
									<kanban-card
										:key="`bucket${bucket.id}-task${task.id}`"
										v-for="task in bucket.tasks"
										:task="task"
									/>
								</transition-group>
							</draggable>
						</div>
						<div class="bucket-footer" v-if="canWrite">
							<div class="field" v-if="showNewTaskInput[bucket.id]">
								<div class="control" :class="{'is-loading': loading}">
									<input
										class="input"
										:disabled="loading || null"
										@focusout="toggleShowNewTaskInput(bucket.id)"
										@keyup.enter="addTaskToBucket(bucket.id)"
										@keyup.esc="toggleShowNewTaskInput(bucket.id)"
										:placeholder="$t('list.kanban.addTaskPlaceholder')"
										type="text"
										v-focus.always
										v-model="newTaskText"
									/>
								</div>
								<p class="help is-danger" v-if="newTaskError[bucket.id] && newTaskText === ''">
									{{ $t('list.create.addTitleRequired') }}
								</p>
							</div>
							<x-button
								@click="toggleShowNewTaskInput(bucket.id)"
								class="is-transparent is-fullwidth has-text-centered"
								:shadow="false"
								v-if="!showNewTaskInput[bucket.id]"
								icon="plus"
								type="secondary"
							>
								{{
									bucket.tasks.length === 0 ? $t('list.kanban.addTask') : $t('list.kanban.addAnotherTask')
								}}
							</x-button>
						</div>
					</div>
				</transition-group>
			</draggable>

			<div class="bucket new-bucket" v-if="canWrite && !loading && buckets.length > 0">
				<input
					:class="{'is-loading': loading}"
					:disabled="loading || null"
					@focusout="() => showNewBucketInput = false"
					@keyup.enter="createNewBucket"
					@keyup.esc="() => showNewBucketInput = false"
					class="input"
					:placeholder="$t('list.kanban.addBucketPlaceholder')"
					type="text"
					v-focus.always
					v-if="showNewBucketInput"
					v-model="newBucketTitle"
				/>
				<x-button
					@click="() => showNewBucketInput = true"
					:shadow="false"
					class="is-transparent is-fullwidth has-text-centered"
					v-if="!showNewBucketInput"
					type="secondary"
					icon="plus"
				>
					{{ $t('list.kanban.addBucket') }}
				</x-button>
			</div>
		</div>

		<!-- This router view is used to show the task popup while keeping the kanban board itself -->
		<transition name="modal">
			<router-view/>
		</transition>

		<transition name="modal">
			<modal
				@close="showBucketDeleteModal = false"
				@submit="deleteBucket()"
				v-if="showBucketDeleteModal"
			>
				<template #header><span>{{ $t('list.kanban.deleteHeaderBucket') }}</span></template>
		
				<template #text>
					<p>{{ $t('list.kanban.deleteBucketText1') }}<br/>
					{{ $t('list.kanban.deleteBucketText2') }}</p>
				</template>
			</modal>
		</transition>
	</div>
</template>

<script>
import draggable from 'vuedraggable'

import BucketModel from '../../../models/bucket'
import {filterObject} from '@/helpers/filterObject'
import {mapState} from 'vuex'
import {saveListView} from '@/helpers/saveListView'
import Rights from '../../../models/constants/rights.json'
import {LOADING, LOADING_MODULE} from '@/store/mutation-types'
import FilterPopup from '@/components/list/partials/filter-popup.vue'
import Dropdown from '@/components/misc/dropdown.vue'
import createTask from '../../../components/tasks/mixins/createTask'
import {getCollapsedBucketState, saveCollapsedBucketState} from '@/helpers/saveCollapsedBucketState'
import {calculateItemPosition} from '../../../helpers/calculateItemPosition'
import KanbanCard from '../../../components/tasks/partials/kanban-card'

export default {
	name: 'Kanban',
	components: {
		KanbanCard,
		Dropdown,
		FilterPopup,
		draggable,
	},
	data() {
		return {
			drag: false,
			dragBucket: false,
			sourceBucket: 0,

			showBucketDeleteModal: false,
			bucketToDelete: 0,
			bucketTitleEditable: false,

			newTaskText: '',
			showNewTaskInput: {},
			newBucketTitle: '',
			showNewBucketInput: false,
			newTaskError: {},
			showSetLimitInput: false,
			collapsedBuckets: {},

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
	mixins: [
		createTask,
	],
	created() {
		this.loadBuckets()

		// Save the current list view to local storage
		// We use local storage and not vuex here to make it persistent across reloads.
		saveListView(this.$route.params.listId, this.$route.name)
	},
	watch: {
		'$route.params.listId': 'loadBuckets',
	},
	computed: {
		buckets: {
			get() {
				return this.$store.state.kanban.buckets
			},
			set(value) {
				this.$store.commit('kanban/setBuckets', value)
			},
		},
		dragOptions() {
			const options = {
				animation: 150,
				ghostClass: 'ghost',
				dragClass: 'task-dragging',
				delay: 150,
				delayOnTouchOnly: true,
			}

			return options
		},
		...mapState({
			loadedListId: state => state.kanban.listId,
			loading: state => state[LOADING] && state[LOADING_MODULE] === 'kanban',
			taskLoading: state => state[LOADING] && state[LOADING_MODULE] === 'tasks',
			canWrite: state => state.currentList.maxRight > Rights.READ,
			list: state => state.currentList,
		}),
	},
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

			this.collapsedBuckets = getCollapsedBucketState(this.$route.params.listId)

			console.debug(`Loading buckets, loadedListId = ${this.loadedListId}, $route.params =`, this.$route.params)
			this.filtersChanged = false

			const minScrollHeightPercent = 0.25

			this.$store.dispatch('kanban/loadBucketsForList', {listId: this.$route.params.listId, params: this.params})
				.then(bs => {
					bs.forEach(b => {
						const e = this.$refs[`tasks-container${b.id}`][0]
						e.addEventListener('scroll', () => {
							const scrollTopMax = e.scrollHeight - e.clientHeight
							if (scrollTopMax <= e.scrollTop + e.scrollTop * minScrollHeightPercent) {
								this.$store.dispatch('kanban/loadNextTasksForBucket', {
									listId: this.$route.params.listId,
									params: this.params,
									bucketId: b.id,
								})
									.catch(e => {
										this.$message.error(e)
									})
							}
						})
					})
				})
				.catch(e => {
					this.$message.error(e)
				})
		},
		updateTaskPosition(e) {
			this.drag = false

			// While we could just pass the bucket index in through the function call, this would not give us the 
			// new bucket id when a task has been moved between buckets, only the new bucket. Using the data-bucket-id
			// of the drop target works all the time.
			const bucketIndex = parseInt(e.to.parentNode.dataset.bucketIndex)

			const newBucket = this.buckets[bucketIndex]
			const task = newBucket.tasks[e.newIndex]
			const taskBefore = newBucket.tasks[e.newIndex - 1] ?? null
			const taskAfter = newBucket.tasks[e.newIndex + 1] ?? null

			task.kanbanPosition = calculateItemPosition(taskBefore !== null ? taskBefore.kanbanPosition : null, taskAfter !== null ? taskAfter.kanbanPosition : null)
			task.bucketId = newBucket.id

			this.$store.dispatch('tasks/update', task)
				.catch(e => {
					this.$message.error(e)
				})
				.finally(() => {
					this.taskUpdating[task.id] = false
					this.oneTaskUpdating = false
				})
		},
		toggleShowNewTaskInput(bucket) {
			this.showNewTaskInput[bucket] = !this.showNewTaskInput[bucket]
		},
		addTaskToBucket(bucketId) {

			if (this.newTaskText === '') {
				this.newTaskError[bucketId] = true
				return
			}
			this.newTaskError[bucketId] = false

			this.createNewTask(this.newTaskText, bucketId)
				.then(r => {
					this.newTaskText = ''
					this.$store.commit('kanban/addTaskToBucket', r)
				})
				.catch(e => {
					this.$message.error(e)
				})
				.finally(() => {
					if (!this.$refs[`tasks-container${bucketId}`][0]) {
						return
					}
					this.$refs[`tasks-container${bucketId}`][0].scrollTop = this.$refs[`tasks-container${bucketId}`][0].scrollHeight
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
					this.$message.error(e)
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

			this.$store.dispatch('kanban/deleteBucket', {bucket: bucket, params: this.params})
				.then(() => {
					this.$message.success({message: this.$t('list.kanban.deleteBucketSuccess')})
				})
				.catch(e => {
					this.$message.error(e)
				})
				.finally(() => {
					this.showBucketDeleteModal = false
				})
		},
		focusBucketTitle(e) {
			// This little helper allows us to drag a bucket around at the title without focusing on it right away.
			this.bucketTitleEditable = true
			this.$nextTick(() => e.target.focus())
		},
		saveBucketTitle(bucketId) {
			this.bucketTitleEditable = false
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
					this.$message.success({message: this.$t('list.kanban.bucketTitleSavedSuccess')})
				})
				.catch(e => {
					this.$message.error(e)
				})
		},
		updateBucket(bucket) {
			bucket.limit = parseInt(bucket.limit)
			this.$store.dispatch('kanban/updateBucket', bucket)
				.then(() => {
					this.$message.success({message: this.$t('list.kanban.bucketLimitSavedSuccess')})
				})
				.catch(e => {
					this.$message.error(e)
				})
		},
		updateBucketPosition(e) {
			this.dragBucket = false

			const bucket = this.buckets[e.newIndex]
			const bucketBefore = this.buckets[e.newIndex - 1] ?? null
			const bucketAfter = this.buckets[e.newIndex + 1] ?? null

			bucket.position = calculateItemPosition(bucketBefore !== null ? bucketBefore.position : null, bucketAfter !== null ? bucketAfter.position : null)

			this.$store.dispatch('kanban/updateBucket', bucket)
				.catch(e => {
					this.$message.error(e)
				})
		},
		setBucketLimit(bucket) {
			if (bucket.limit < 0) {
				return
			}

			this.updateBucket(bucket)
		},
		shouldAcceptDrop(bucket) {
			return bucket.id === this.sourceBucket || // When dragging from a bucket who has its limit reached, dragging should still be possible
				bucket.limit === 0 || // If there is no limit set, dragging & dropping should always work
				bucket.tasks.length < bucket.limit // Disallow dropping to buckets which have their limit reached
		},
		dragstart(bucket) {
			this.drag = true
			this.sourceBucket = bucket.id
		},
		toggleDoneBucket(bucket) {
			bucket.isDoneBucket = !bucket.isDoneBucket
			this.$store.dispatch('kanban/updateBucket', bucket)
				.then(() => {
					this.$message.success({message: this.$t('list.kanban.doneBucketSavedSuccess')})
				})
				.catch(e => {
					this.$message.error(e)
					bucket.isDoneBucket = !bucket.isDoneBucket
				})
		},
		collapseBucket(bucket) {
			this.collapsedBuckets[bucket.id] = true
			saveCollapsedBucketState(this.$route.params.listId, this.collapsedBuckets)
		},
		unCollapseBucket(bucket) {
			if (!this.collapsedBuckets[bucket.id]) {
				return
			}

			this.collapsedBuckets[bucket.id] = false
			saveCollapsedBucketState(this.$route.params.listId, this.collapsedBuckets)
		},
	},
}
</script>
