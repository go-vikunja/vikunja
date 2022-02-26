<template>
	<ListWrapper class="list-kanban" :list-id="listId" viewName="kanban">
		<template #header>
			<div class="filter-container" v-if="isSavedFilter">
				<div class="items">
					<filter-popup
						v-model="params"
						@update:modelValue="loadBuckets"
					/>
				</div>
			</div>
		</template>

		<template #default>
			<div class="kanban-view">
			<div
				:class="{ 'is-loading': loading && !oneTaskUpdating}"
				class="kanban kanban-bucket-container loader-container"
			>
			<draggable
				v-bind="dragOptions"
				:modelValue="buckets"
				@update:modelValue="updateBuckets"
				@end="updateBucketPosition"
				@start="() => dragBucket = true"
				group="buckets"
				:disabled="!canWrite"
				tag="transition-group"
				:item-key="({id}) => `bucket${id}`"
				:component-data="bucketDraggableComponentData"
			>
				<template #item="{element: bucket, index: bucketIndex }">
					<div
						class="bucket"
						:class="{'is-collapsed': collapsedBuckets[bucket.id]}"
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
								@keydown.enter.prevent.stop="$event.target.blur()"
								@keydown.esc.prevent.stop="$event.target.blur()"
								@blur="saveBucketTitle(bucket.id, $event.target.textContent)"
								@click="focusBucketTitle"
								class="title input"
								:contenteditable="(bucketTitleEditable && canWrite && !collapsedBuckets[bucket.id]) ? true : undefined"
								:spellcheck="false">{{ bucket.title }}</h2>
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
												@keyup.esc="() => showSetLimitInput = false"
												@keyup.enter="() => showSetLimitInput = false"
												:value="bucket.limit"
												@input="(event) => setBucketLimit(bucket.id, parseInt(event.target.value))"
												class="input"
												type="number"
												min="0"
												v-focus.always
											/>
										</div>
										<div class="control">
											<x-button
												:disabled="bucket.limit < 0"
												:icon="['far', 'save']"
												:shadow="false"
												v-cy="'setBucketLimit'"
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

						<draggable
							v-bind="dragOptions"
							:modelValue="bucket.tasks"
							@update:modelValue="(tasks) => updateTasks(bucket.id, tasks)"
							@start="() => dragstart(bucket)"
							@end="updateTaskPosition"
							:group="{name: 'tasks', put: shouldAcceptDrop(bucket) && !dragBucket}"
							:disabled="!canWrite"
							:data-bucket-index="bucketIndex"
							tag="transition-group"
							:item-key="(task) => `bucket${bucket.id}-task${task.id}`"
							:component-data="getTaskDraggableTaskComponentData(bucket)"
						>
							<template #footer>
								<div class="bucket-footer" v-if="canWrite">
									<div class="field" v-if="showNewTaskInput[bucket.id]">
										<div class="control" :class="{'is-loading': loading}">
											<input
												class="input"
												:disabled="loading || undefined"
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
										class="is-fullwidth has-text-centered"
										:shadow="false"
										v-else
										icon="plus"
										variant="secondary"
									>
										{{ bucket.tasks.length === 0 ? $t('list.kanban.addTask') : $t('list.kanban.addAnotherTask') }}
									</x-button>
								</div>
							</template>

							<template #item="{element: task}">
								<div class="task-item">
									<kanban-card class="kanban-card" :task="task"/>
								</div>
							</template>
						</draggable>
					</div>
				</template>
			</draggable>

			<div class="bucket new-bucket" v-if="canWrite && !loading && buckets.length > 0">
				<input
					:class="{'is-loading': loading}"
					:disabled="loading || null"
					@blur="() => showNewBucketInput = false"
					@keyup.enter="createNewBucket"
					@keyup.esc="$event.target.blur()"
					class="input"
					:placeholder="$t('list.kanban.addBucketPlaceholder')"
					type="text"
					v-focus.always
					v-if="showNewBucketInput"
					v-model="newBucketTitle"
				/>
				<x-button
					v-else
					@click="() => showNewBucketInput = true"
					:shadow="false"
					class="is-transparent is-fullwidth has-text-centered"
					variant="secondary"
					icon="plus"
				>
					{{ $t('list.kanban.addBucket') }}
				</x-button>
			</div>
		</div>

		<transition name="modal">
			<modal
				v-if="showBucketDeleteModal"
				@close="showBucketDeleteModal = false"
				@submit="deleteBucket()"
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
	</ListWrapper>
</template>

<script>
import draggable from 'vuedraggable'
import cloneDeep from 'lodash.clonedeep'

import BucketModel from '../../models/bucket'
import {mapState} from 'vuex'
import Rights from '../../models/constants/rights.json'
import {LOADING, LOADING_MODULE} from '@/store/mutation-types'
import ListWrapper from './ListWrapper'
import FilterPopup from '@/components/list/partials/filter-popup.vue'
import Dropdown from '@/components/misc/dropdown.vue'
import {getCollapsedBucketState, saveCollapsedBucketState} from '@/helpers/saveCollapsedBucketState'
import {calculateItemPosition} from '../../helpers/calculateItemPosition'
import KanbanCard from '@/components/tasks/partials/kanban-card'

const DRAG_OPTIONS = {
	// sortable options
	animation: 150,
	ghostClass: 'ghost',
	dragClass: 'task-dragging',
	delayOnTouchOnly: true,
	delay: 150,
}

const MIN_SCROLL_HEIGHT_PERCENT = 0.25

export default {
	name: 'Kanban',
	components: {
		ListWrapper,
		KanbanCard,
		Dropdown,
		FilterPopup,
		draggable,
	},

	props: {
		listId: {
			type: Number,
			required: true,
		},
	},

	data() {
		return {
			taskContainerRefs: {},

			dragOptions: DRAG_OPTIONS,

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
		}
	},

	watch: {
		loadBucketParameter: {
			handler: 'loadBuckets',
			immediate: true,
		},
	},
	computed: {
		getTaskDraggableTaskComponentData() {
			return (bucket) => ({
				ref: (el) => this.setTaskContainerRef(bucket.id, el),
				onScroll: (event) => this.handleTaskContainerScroll(bucket.id, bucket.listId, event.target),
				type: 'transition',
				tag: 'div',
				name: !this.drag ? 'move-card' : null,
				class: [
					'tasks',
					{'dragging-disabled': !this.canWrite},
				],
			})
		},

		isSavedFilter() {
			return this.list.isSavedFilter && !this.list.isSavedFilter()
		},
		loadBucketParameter() {
			return {
				listId: this.listId,
				params: this.params,
			}
		},
		bucketDraggableComponentData() {
			return {
				type: 'transition',
				tag: 'div',
				name: !this.dragBucket ? 'move-bucket' : null,
				class: [
					'kanban-bucket-container',
					{'dragging-disabled': !this.canWrite},
				],
			}
		},
		buckets() {
			return this.$store.state.kanban.buckets
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
			const {listId, params} = this.loadBucketParameter

			this.collapsedBuckets = getCollapsedBucketState(listId)

			console.debug(`Loading buckets, loadedListId = ${this.loadedListId}, $attrs = ${this.$attrs} $route.params =`, this.$route.params)

			this.$store.dispatch('kanban/loadBucketsForList', {listId, params})
		},

		setTaskContainerRef(id, el) {
			if (!el) return
			this.taskContainerRefs[id] = el
		},

		handleTaskContainerScroll(id, listId, el) {
			if (!el) {
				return
			}
			const scrollTopMax = el.scrollHeight - el.clientHeight
			const threshold = el.scrollTop + el.scrollTop * MIN_SCROLL_HEIGHT_PERCENT
			if (scrollTopMax > threshold) {
				return
			}

			this.$store.dispatch('kanban/loadNextTasksForBucket', {
				listId: listId,
				params: this.params,
				bucketId: id,
			})
		},

		updateTasks(bucketId, tasks) {
			const newBucket = {
				...this.$store.getters['kanban/getBucketById'](bucketId),
				tasks,
			}

			this.$store.commit('kanban/setBucketById', newBucket)
		},

		async updateTaskPosition(e) {
			this.drag = false

			// While we could just pass the bucket index in through the function call, this would not give us the 
			// new bucket id when a task has been moved between buckets, only the new bucket. Using the data-bucket-id
			// of the drop target works all the time.
			const bucketIndex = parseInt(e.to.dataset.bucketIndex)


			const newBucket = this.buckets[bucketIndex]

			// HACK:
			// this is a hacky workaround for a known problem of vue.draggable.next when using the footer slot
			// the problem: https://github.com/SortableJS/vue.draggable.next/issues/108
			// This hack doesn't remove the problem that the ghost item is still displayed below the footer
			// It just makes releasing the item possible.

			// The newIndex of the event doesn't count in the elements of the footer slot.
			// This is why in case the length of the tasks is identical with the newIndex
			// we have to remove 1 to get the correct index.
			const newTaskIndex = newBucket.tasks.length === e.newIndex
				? e.newIndex - 1
				: e.newIndex

			const task = newBucket.tasks[newTaskIndex]
			const taskBefore = newBucket.tasks[newTaskIndex - 1] ?? null
			const taskAfter = newBucket.tasks[newTaskIndex + 1] ?? null

			const newTask = cloneDeep(task) // cloning the task to avoid vuex store mutations
			newTask.bucketId = newBucket.id,
				newTask.kanbanPosition = calculateItemPosition(taskBefore !== null ? taskBefore.kanbanPosition : null, taskAfter !== null ? taskAfter.kanbanPosition : null)

			try {
				await this.$store.dispatch('tasks/update', newTask)
			} finally {
				this.taskUpdating[task.id] = false
				this.oneTaskUpdating = false
			}
		},

		toggleShowNewTaskInput(bucketId) {
			this.showNewTaskInput[bucketId] = !this.showNewTaskInput[bucketId]
		},

		async addTaskToBucket(bucketId) {
			if (this.newTaskText === '') {
				this.newTaskError[bucketId] = true
				return
			}
			this.newTaskError[bucketId] = false

			const task = await this.$store.dispatch('tasks/createNewTask', {
				title: this.newTaskText,
				bucketId,
				listId: this.listId,
			})
			this.newTaskText = ''
			this.$store.commit('kanban/addTaskToBucket', task)
			this.scrollTaskContainerToBottom(bucketId)
		},

		scrollTaskContainerToBottom(bucketId) {
			const bucketEl = this.taskContainerRefs[bucketId]
			if (!bucketEl) {
				return
			}
			bucketEl.scrollTop = bucketEl.scrollHeight
		},

		async createNewBucket() {
			if (this.newBucketTitle === '') {
				return
			}

			const newBucket = new BucketModel({
				title: this.newBucketTitle,
				listId: this.listId,
			})

			await this.$store.dispatch('kanban/createBucket', newBucket)
			this.newBucketTitle = ''
			this.showNewBucketInput = false
		},

		deleteBucketModal(bucketId) {
			if (this.buckets.length <= 1) {
				return
			}

			this.bucketToDelete = bucketId
			this.showBucketDeleteModal = true
		},

		async deleteBucket() {
			const bucket = new BucketModel({
				id: this.bucketToDelete,
				listId: this.listId,
			})

			try {
				await this.$store.dispatch('kanban/deleteBucket', {
					bucket,
					params: this.params,
				})
				this.$message.success({message: this.$t('list.kanban.deleteBucketSuccess')})
			} finally {
				this.showBucketDeleteModal = false
			}
		},

		focusBucketTitle(e) {
			// This little helper allows us to drag a bucket around at the title without focusing on it right away.
			this.bucketTitleEditable = true
			this.$nextTick(() => e.target.focus())
		},

		async saveBucketTitle(bucketId, bucketTitle) {
			const updatedBucketData = {
				id: bucketId,
				title: bucketTitle,
			}

			await this.$store.dispatch('kanban/updateBucketTitle', updatedBucketData)
			this.bucketTitleEditable = false
			this.$message.success({message: this.$t('list.kanban.bucketTitleSavedSuccess')})
		},

		updateBuckets(value) {
			// (1) buckets get updated in store and tasks positions get invalidated
			this.$store.commit('kanban/setBuckets', value)
		},

		updateBucketPosition(e) {
			// (2) bucket positon is changed
			this.dragBucket = false

			const bucket = this.buckets[e.newIndex]
			const bucketBefore = this.buckets[e.newIndex - 1] ?? null
			const bucketAfter = this.buckets[e.newIndex + 1] ?? null

			const updatedData = {
				id: bucket.id,
				position: calculateItemPosition(
					bucketBefore !== null ? bucketBefore.position : null,
					bucketAfter !== null ? bucketAfter.position : null,
				),
			}

			this.$store.dispatch('kanban/updateBucket', updatedData)
		},

		async setBucketLimit(bucketId, limit) {
			if (limit < 0) {
				return
			}

			const newBucket = {
				...this.$store.getters['kanban/getBucketById'](bucketId),
				limit,
			}

			await this.$store.dispatch('kanban/updateBucket', newBucket)
			this.$message.success({message: this.$t('list.kanban.bucketLimitSavedSuccess')})
		},

		shouldAcceptDrop(bucket) {
			return (
				// When dragging from a bucket who has its limit reached, dragging should still be possible
				bucket.id === this.sourceBucket ||
				// If there is no limit set, dragging & dropping should always work
				bucket.limit === 0 ||
				// Disallow dropping to buckets which have their limit reached
				bucket.tasks.length < bucket.limit
			)
		},

		dragstart(bucket) {
			this.drag = true
			this.sourceBucket = bucket.id
		},

		async toggleDoneBucket(bucket) {
			const newBucket = {
				...bucket,
				isDoneBucket: !bucket.isDoneBucket,
			}
			await this.$store.dispatch('kanban/updateBucket', newBucket)
			this.$message.success({message: this.$t('list.kanban.doneBucketSavedSuccess')})
		},

		collapseBucket(bucket) {
			this.collapsedBuckets[bucket.id] = true
			saveCollapsedBucketState(this.listId, this.collapsedBuckets)
		},
		unCollapseBucket(bucket) {
			if (!this.collapsedBuckets[bucket.id]) {
				return
			}

			this.collapsedBuckets[bucket.id] = false
			saveCollapsedBucketState(this.listId, this.collapsedBuckets)
		},
	},
}
</script>

<style lang="scss">
$ease-out: all .3s cubic-bezier(0.23, 1, 0.32, 1);
$bucket-width: 300px;
$bucket-header-height: 60px;
$bucket-right-margin: 1rem;

$crazy-height-calculation: '100vh - 4.5rem - 1.5rem - 1rem - 1.5rem - 11px';
$crazy-height-calculation-tasks: '#{$crazy-height-calculation} - 1rem - 2.5rem - 2rem - #{$button-height} - 1rem';
$filter-container-height: '1rem - #{$switch-view-height}';

// FIXME:
.app-content.list\.kanban {
	padding-bottom: 0 !important;
}

.kanban {
	overflow-x: auto;
	overflow-y: hidden;
	height: calc(#{$crazy-height-calculation});
	margin: 0 -1.5rem;
	padding: 0 1.5rem;

	@media screen and (max-width: $tablet) {
		height: calc(#{$crazy-height-calculation} - #{$filter-container-height});
	}

	&-bucket-container {
		display: flex;
	}

	.ghost {
		position: relative;

		* {
			opacity: 0;
		}
		&::after {
			content: '';
			position: absolute;
			display: block;
			top: 0.25rem;
			right: 0.5rem;
			bottom: 0.25rem;
			left: 0.5rem;
			border: 3px dashed var(--grey-300);
			border-radius: $radius;
		}
	}
	
	.sortable-chosen .task {
		transform: rotate(3deg);
	}

	.bucket {
		border-radius: $radius;
		position: relative;

		margin: 0 $bucket-right-margin 0 0;
		max-height: calc(100% - 1rem); // 1rem spacing to the bottom
		min-height: 20px;
		width: $bucket-width;
		display: flex;
		flex-direction: column;
		overflow: hidden; // Make sure the edges are always rounded		

		.tasks {
			overflow: hidden auto;
			height: 100%;
		}

		.task-item {
			background-color: var(--grey-100);
			padding: .25rem .5rem;

			&:first-of-type {
				padding-top: .5rem;
			}
			&:last-of-type {
				padding-bottom: .5rem;
			}	
		}

		.no-move {
			transition: transform 0s;
		}

		h2 {
			font-size: 1rem;
			margin: 0;
			font-weight: 600 !important;
		}

		&.new-bucket {
			// Because of reasons, this button ignores the margin we gave it to the right.
			// To make it still look like it has some, we modify the container to have a padding of 1rem,
			// which is the same as the margin it should have. Then we make the container itself bigger
			// to hide the fact we just made the button smaller.
			min-width: calc(#{$bucket-width} + 1rem);
			background: transparent;
			padding-right: 1rem;

			.button {
				background: var(--grey-100);
				width: 100%;
			}
		}

		a.dropdown-item {
			padding-right: 1rem;
		}

		&.is-collapsed {
			align-self: flex-start;
			transform: rotate(90deg) translateY(-100%);
			transform-origin: top left;
			// Using negative margins instead of translateY here to make all other buckets fill the empty space
			margin-right: calc((#{$bucket-width} - #{$bucket-header-height} - #{$bucket-right-margin}) * -1);
			cursor: pointer;

			.tasks, .bucket-footer {
				display: none;
			}
		}
	}

	.bucket-header {
		background-color: var(--grey-100);
		height: min-content;
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: .5rem;
		height: $bucket-header-height;

		.limit {
			padding: 0 .5rem;
			font-weight: bold;

			&.is-max {
				color: var(--danger);
			}
		}

		.title.input {
			height: auto;
			padding: .4rem .5rem;
			display: inline-block;
			cursor: pointer;
		}
	}

	:deep(.dropdown-trigger) {
		cursor: pointer;
		padding: .5rem;
	}

	.bucket-footer {
		position: sticky;
		bottom: 0;
		height: min-content;
		padding: .5rem;
		background-color: var(--grey-100);
		border-bottom-left-radius: $radius;
		border-bottom-right-radius: $radius;
		transform: none;

		.button {
			background-color: transparent;

			&:hover {
				background-color: var(--white);
			}
		}
	}
}

// FIXME: This does not seem to work
.task-dragging {
	transform: rotateZ(3deg);
	transition: transform 0.18s ease;
}

.move-card-move {
	transform: rotateZ(3deg);
	transition: transform $transition-duration;
}

.move-card-leave-from,
.move-card-leave-to,
.move-card-leave-active {
	display: none;
}

@include modal-transition();
</style>