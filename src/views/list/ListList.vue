<template>
	<ListWrapper class="list-list" :list-id="listId" viewName="list">
		<template #header>
		<div
			class="filter-container"
			v-if="!isSavedFilter(list)"
		>
			<div class="items">
				<div class="search">
					<div :class="{ hidden: !showTaskSearch }" class="field has-addons">
						<div class="control has-icons-left has-icons-right">
							<input
								@blur="hideSearchBar()"
								@keyup.enter="searchTasks"
								class="input"
								:placeholder="$t('misc.search')"
								type="text"
								v-focus
								v-model="searchTerm"
							/>
							<span class="icon is-left">
								<icon icon="search"/>
							</span>
						</div>
						<div class="control">
							<x-button
								:loading="loading"
								@click="searchTasks"
								:shadow="false"
							>
								{{ $t('misc.search') }}
							</x-button>
						</div>
					</div>
					<x-button
						@click="showTaskSearch = !showTaskSearch"
						icon="search"
						variant="secondary"
						v-if="!showTaskSearch"
					/>
				</div>
				<filter-popup
					v-model="params"
					@update:modelValue="loadTasks()"
				/>
			</div>
		</div>
		</template>

		<template #default>
		<div
			:class="{ 'is-loading': loading }"
			class="loader-container is-max-width-desktop list-view"
		>
		<card :padding="false" :has-content="false" class="has-overflow">
			<template
				v-if="!list.isArchived && canWrite"
			>
				<add-task
					@taskAdded="updateTaskList"
					ref="addTaskRef"
					:default-position="firstNewPosition"
				/>
			</template>

			<nothing v-if="ctaVisible && tasks.length === 0 && !loading">
				{{ $t('list.list.empty') }}
				<ButtonLink @click="focusNewTaskInput()">
					{{ $t('list.list.newTaskCta') }}
				</ButtonLink>
			</nothing>

			<div class="tasks-container" :class="{ 'has-task-edit-open': isTaskEdit }">
				<div
					class="tasks mt-0"
					v-if="tasks && tasks.length > 0"
				>
					<draggable
						v-bind="DRAG_OPTIONS"
						v-model="tasks"
						group="tasks"
						@start="() => drag = true"
						@end="saveTaskPosition"
						handle=".handle"
						:disabled="!canWrite"
						item-key="id"
						tag="ul"
						:component-data="{
							class: { 'dragging-disabled': !canWrite || isAlphabeticalSorting },
							type: 'transition-group'
						}"
					>
						<template #item="{element: t}">
							<single-task-in-list
								:show-list-color="false"
								:disabled="!canWrite"
								:can-mark-as-done="canWrite || isSavedFilter(list)"
								:the-task="t"
								@taskUpdated="updateTasks"
							>
								<template v-if="canWrite">
									<span class="icon handle">
										<icon icon="grip-lines"/>
									</span>
									<BaseButton
										@click="editTask(t.id)"
										class="icon settings"
										v-if="!list.isArchived"
									>
										<icon icon="pencil-alt"/>
									</BaseButton>
								</template>
							</single-task-in-list>
						</template>
					</draggable>
				</div>
				<EditTask
					v-if="isTaskEdit"
					class="taskedit mt-0"
					:title="$t('list.list.editTask')"
					@close="closeTaskEditPane()"
					:shadow="false"
					:task="taskEditTask"
				/>
			</div>

			<Pagination 
				:total-pages="totalPages"
				:current-page="currentPage"
			/>
		</card>
		</div>
		</template>
	</ListWrapper>
</template>

<script lang="ts">
export default { name: 'List' }
</script>

<script setup lang="ts">
import {ref, computed, toRef, nextTick, onMounted, type PropType, watch} from 'vue'
import draggable from 'zhyswan-vuedraggable'
import {useRoute, useRouter} from 'vue-router'

import ListWrapper from './ListWrapper.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import ButtonLink from '@/components/misc/ButtonLink.vue'
import EditTask from '@/components/tasks/edit-task.vue'
import AddTask from '@/components/tasks/add-task.vue'
import SingleTaskInList from '@/components/tasks/partials/singleTaskInList.vue'
import FilterPopup from '@/components/list/partials/filter-popup.vue'
import Nothing from '@/components/misc/nothing.vue'
import Pagination from '@/components/misc/pagination.vue'
import {ALPHABETICAL_SORT} from '@/components/list/partials/filters.vue'

import {useTaskList} from '@/composables/useTaskList'
import {RIGHTS as Rights} from '@/constants/rights'
import {calculateItemPosition} from '@/helpers/calculateItemPosition'
import type {ITask} from '@/modelTypes/ITask'
import {isSavedFilter} from '@/services/savedFilter'

import {useBaseStore} from '@/stores/base'
import {useTaskStore} from '@/stores/tasks'

import type {IList} from '@/modelTypes/IList'

function sortTasks(tasks: ITask[]) {
	if (tasks === null || Array.isArray(tasks) && tasks.length === 0) {
		return
	}
	return tasks.sort((a, b) => {
		if (a.done < b.done)
			return -1
		if (a.done > b.done)
			return 1

		if (a.position < b.position)
			return -1
		if (a.position > b.position)
			return 1
		return 0
	})
}

const props = defineProps({
	listId: {
		type: Number as PropType<IList['id']>,
		required: true,
	},
})

const ctaVisible = ref(false)
const showTaskSearch = ref(false)

const drag = ref(false)
const DRAG_OPTIONS = {
	animation: 100,
	ghostClass: 'ghost',
} as const


const taskEditTask = ref<ITask | null>(null)
const isTaskEdit = ref(false)

function closeTaskEditPane() {
	isTaskEdit.value = false
	taskEditTask.value = null
}

watch(
	() => props.listId,
	closeTaskEditPane,
)

const {
	tasks,
	loading,
	totalPages,
	currentPage,
	loadTasks,
	searchTerm,
	params,
	// sortByParam,
} = useTaskList(toRef(props, 'listId'), {position: 'asc' })


const isAlphabeticalSorting = computed(() => {
	return params.value.sort_by.find(sortBy => sortBy === ALPHABETICAL_SORT) !== undefined
})

const firstNewPosition = computed(() => {
	if (tasks.value.length === 0) {
		return 0
	}

	return calculateItemPosition(null, tasks.value[0].position)
})

const taskStore = useTaskStore()
const baseStore = useBaseStore()
const list = computed(() => baseStore.currentList)

const canWrite = computed(() => {
	return list.value.maxRight > Rights.READ && list.value.id > 0
})

onMounted(async () => {
	await nextTick()
	ctaVisible.value = true
})

const route = useRoute()
const router = useRouter()

function searchTasks() {
	// Only search if the search term changed
	if (route.query as unknown as string === searchTerm.value) {
		return
	}

	router.push({
		name: 'list.list',
		query: {search: searchTerm.value},
	})
}

function hideSearchBar() {
	// This is a workaround.
	// When clicking on the search button, @blur from the input is fired. If we
	// would then directly hide the whole search bar directly, no click event
	// from the button gets fired. To prevent this, we wait 200ms until we hide
	// everything so the button has a chance of firing the search event.
	setTimeout(() => {
		showTaskSearch.value = false
	}, 200)
}

const addTaskRef = ref<typeof AddTask | null>(null)
function focusNewTaskInput() {
	addTaskRef.value?.focusTaskInput()
}

function updateTaskList(task: ITask) {
	if (isAlphabeticalSorting.value ) {
		// reload tasks with current filter and sorting
		loadTasks(1, undefined, undefined, true)
	}
	else {
		tasks.value = [
			task,
			...tasks.value,
		]
	}

	baseStore.setHasTasks(true)
}

function editTask(id: ITask['id']) {
	taskEditTask.value = {...tasks.value.find(t => t.id === Number(id))}
	isTaskEdit.value = true
}

function updateTasks(updatedTask: ITask) {
	for (const t in tasks.value) {
		if (tasks.value[t].id === updatedTask.id) {
			tasks.value[t] = updatedTask
			break
		}
	}
	// FIXME: Use computed
	sortTasks(tasks.value)
}

async function saveTaskPosition(e) {
	drag.value = false

	const task = tasks.value[e.newIndex]
	const taskBefore = tasks.value[e.newIndex - 1] ?? null
	const taskAfter = tasks.value[e.newIndex + 1] ?? null

	const newTask = {
		...task,
		position: calculateItemPosition(taskBefore !== null ? taskBefore.position : null, taskAfter !== null ? taskAfter.position : null),
	}

	const updatedTask = await taskStore.update(newTask)
	tasks.value[e.newIndex] = updatedTask
}
</script>

<style lang="scss" scoped>
.tasks-container {
	display: flex;

	&.has-task-edit-open {
		flex-direction: column;

		@media screen and (min-width: $tablet) {
			flex-direction: row;

			.tasks {
				width: 66%;
			}
		}
	}

	.tasks {
		width: 100%;
		padding: .5rem;

		.ghost {
			border-radius: $radius;
			background: var(--grey-100);
			border: 2px dashed var(--grey-300);

			* {
				opacity: 0;
			}
		}
	}

	.taskedit {
		width: 33%;
		margin-right: 1rem;
		margin-left: .5rem;
		min-height: calc(100% - 1rem);

		@media screen and (max-width: $tablet) {
			width: 100%;
			border-radius: 0;
			margin: 0;
			border-left: 0;
			border-right: 0;
			border-bottom: 0;
		}
	}
}

.list-view .task-add {
	padding: 1rem 1rem 0;
}

.link-share-view .card {
  border: none;
  box-shadow: none;
}
</style>