<template>
	<ProjectWrapper class="project-list" :project-id="projectId" viewName="project">
		<template #header>
		<div
			class="filter-container"
			v-if="!isSavedFilter(project)"
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
					@update:modelValue="prepareFiltersAndLoadTasks()"
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
			<add-task
				v-if="!project.isArchived && canWrite"
				class="list-view__add-task d-print-none"
				ref="addTaskRef"
				:default-position="firstNewPosition"
				@taskAdded="updateTaskList"
			/>

			<nothing v-if="ctaVisible && tasks.length === 0 && !loading">
				{{ $t('project.list.empty') }}
				<ButtonLink @click="focusNewTaskInput()" v-if="project.id > 0">
					{{ $t('project.list.newTaskCta') }}
				</ButtonLink>
			</nothing>


			<draggable
				v-if="tasks && tasks.length > 0"
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
					class: {
						tasks: true,
						'dragging-disabled': !canWrite || isAlphabeticalSorting
					},
					type: 'transition-group'
				}"
			>
				<template #item="{element: t}">
					<single-task-in-project
						:show-list-color="false"
						:disabled="!canWrite"
						:can-mark-as-done="canWrite || isSavedFilter(project)"
						:the-task="t"
						@taskUpdated="updateTasks"
						:all-tasks="allTasks"
					>
						<template v-if="canWrite">
							<span class="icon handle">
								<icon icon="grip-lines"/>
							</span>
						</template>
					</single-task-in-project>
				</template>
			</draggable>

			<Pagination 
				:total-pages="totalPages"
				:current-page="currentPage"
			/>
		</card>
		</div>
		</template>
	</ProjectWrapper>
</template>

<script lang="ts">
export default { name: 'List' }
</script>

<script setup lang="ts">
import {ref, computed, nextTick, onMounted, watch} from 'vue'
import draggable from 'zhyswan-vuedraggable'
import {useRoute, useRouter} from 'vue-router'

import ProjectWrapper from '@/components/project/ProjectWrapper.vue'
import ButtonLink from '@/components/misc/ButtonLink.vue'
import AddTask from '@/components/tasks/add-task.vue'
import SingleTaskInProject from '@/components/tasks/partials/singleTaskInProject.vue'
import FilterPopup from '@/components/project/partials/filter-popup.vue'
import Nothing from '@/components/misc/nothing.vue'
import Pagination from '@/components/misc/pagination.vue'
import {ALPHABETICAL_SORT} from '@/components/project/partials/filters.vue'

import {useTaskList} from '@/composables/useTaskList'
import {RIGHTS as Rights} from '@/constants/rights'
import {calculateItemPosition} from '@/helpers/calculateItemPosition'
import type {ITask} from '@/modelTypes/ITask'
import {isSavedFilter} from '@/services/savedFilter'

import {useBaseStore} from '@/stores/base'
import {useTaskStore} from '@/stores/tasks'

import type {IProject} from '@/modelTypes/IProject'

const {
	projectId,
} = defineProps<{
	projectId: IProject['id'],
}>()

const ctaVisible = ref(false)
const showTaskSearch = ref(false)

const drag = ref(false)
const DRAG_OPTIONS = {
	animation: 100,
	ghostClass: 'task-ghost',
} as const

const {
	tasks: allTasks,
	loading,
	totalPages,
	currentPage,
	loadTasks,
	searchTerm,
	params,
	sortByParam,
} = useTaskList(() => projectId, {position: 'asc' })

const tasks = ref<ITask[]>([])
watch(
	allTasks,
	() => {
		tasks.value = [...allTasks.value]
		if (projectId < 0) {
			return
		}
		tasks.value = tasks.value.filter(t => typeof t.relatedTasks?.parenttask === 'undefined')
	},
)

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
const project = computed(() => baseStore.currentProject)

const canWrite = computed(() => {
	return project.value.maxRight > Rights.READ && project.value.id > 0
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
		name: 'project.list',
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
		loadTasks()
	}
	else {
		allTasks.value = [
			task,
			...allTasks.value,
		]
	}

	baseStore.setHasTasks(true)
}

function updateTasks(updatedTask: ITask) {
	for (const t in tasks.value) {
		if (tasks.value[t].id === updatedTask.id) {
			tasks.value[t] = updatedTask
			break
		}
	}
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

function prepareFiltersAndLoadTasks() {
	if(isAlphabeticalSorting.value) {
		sortByParam.value = {} 
		sortByParam.value[ALPHABETICAL_SORT] = 'asc'
	}
	
	loadTasks()
}
</script>

<style lang="scss" scoped>
.tasks {
	padding: .5rem;
}

.task-ghost {
	border-radius: $radius;
	background: var(--grey-100);
	border: 2px dashed var(--grey-300);
	
	* {
		opacity: 0;
	}
}

.list-view__add-task {
	padding: 1rem 1rem 0;
}

.link-share-view .card {
  border: none;
  box-shadow: none;
}

.control.has-icons-left .icon,
.control.has-icons-right .icon {
	transition: all $transition;
}
</style>