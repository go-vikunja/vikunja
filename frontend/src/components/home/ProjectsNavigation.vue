<template>
	<draggable
		v-model="availableProjects"
		animation="100"
		ghost-class="ghost"
		:group="{name: 'projects', put: ['tasks']}"
		handle=".handle"
		tag="menu"
		item-key="id"
		:disabled="!canEditOrder"
		filter=".drag-disabled"
		:component-data="{
			type: 'transition-group',
			name: !drag ? 'flip-list' : null,
			class: [
				'menu-list can-be-hidden',
				{ 'dragging-disabled': !canEditOrder }
			],
		}"
		@start="() => drag = true"
		@end="handleDrop"
	>
		<template #item="{element: project}">
			<ProjectsNavigationItem
				:class="{'drag-disabled': project.id < 0}"
				:project="project"
				:is-loading="projectUpdating[project.id]"
				:can-collapse="canCollapse"
				:can-edit-order="canEditOrder"
				:data-project-id="project.id"
			/>
		</template>
	</draggable>
</template>

<script lang="ts" setup>
import {ref, watch} from 'vue'
import draggable from 'zhyswan-vuedraggable'
import type {SortableEvent} from 'sortablejs'

import ProjectsNavigationItem from '@/components/home/ProjectsNavigationItem.vue'

import {calculateItemPosition} from '@/helpers/calculateItemPosition'
import type {IProject} from '@/modelTypes/IProject'

import {useProjectStore} from '@/stores/projects'
import {useTaskStore} from '@/stores/tasks'
import {useKanbanStore} from '@/stores/kanban'
import {PERMISSIONS} from '@/constants/permissions'
import {isSavedFilter} from '@/services/savedFilter'

const props = defineProps<{
	modelValue?: IProject[],
	canEditOrder: boolean,
	canCollapse?: boolean,
}>()
const emit = defineEmits<{
	(e: 'update:modelValue', projects: IProject[]): void
}>()

const drag = ref(false)

const projectStore = useProjectStore()
const taskStore = useTaskStore()
const kanbanStore = useKanbanStore()

// Vue draggable will modify the projects list as it changes their position which will not work on a prop.
// Hence, we'll clone the prop and work on the clone.
const availableProjects = ref<IProject[]>([])
watch(
	() => props.modelValue,
	projects => {
		availableProjects.value = projects || []
	},
	{immediate: true},
)

const projectUpdating = ref<{ [id: IProject['id']]: boolean }>({})

async function handleDrop(e: SortableEvent) {
	drag.value = false

	// Check if this is a task drop (from task views to sidebar)
	const taskIdStr = e.item.dataset.taskId
	if (taskIdStr) {
		await handleTaskDrop(e)
		return
	}

	// Otherwise handle as project reorder
	await saveProjectPosition(e)
}

async function handleTaskDrop(e: SortableEvent) {
	const draggedTask = taskStore.draggedTask
	if (!draggedTask) {
		// Clean up: remove the temporarily inserted DOM element
		e.item.remove()
		return
	}

	// Determine target project
	let targetProjectId: number | null = null
	const targetElement = e.to.children[e.newIndex ?? 0] as HTMLElement

	if (targetElement?.dataset?.projectId) {
		targetProjectId = parseInt(targetElement.dataset.projectId)
	} else if (e.newIndex !== undefined && e.newIndex >= 0) {
		const targetProject = availableProjects.value[e.newIndex]
		if (targetProject) {
			targetProjectId = targetProject.id
		}
	}

	// Clean up: remove the temporarily inserted DOM element
	e.item.remove()

	if (!targetProjectId || targetProjectId <= 0) {
		return
	}

	const targetProject = projectStore.projects[targetProjectId]
	if (!targetProject) {
		return
	}

	// Check permissions: reject pseudo projects and read-only projects
	if (isSavedFilter(targetProject) || targetProject.id < 0) {
		return
	}

	if (targetProject.maxPermission <= PERMISSIONS.READ) {
		return
	}

	// Don't move if it's already in the target project
	if (draggedTask.projectId === targetProjectId) {
		return
	}

	try {
		// Move the task to the new project
		await taskStore.update({
			...draggedTask,
			projectId: targetProjectId,
		})

		// Remove from kanban store if it was in a bucket
		kanbanStore.removeTaskInBucket(draggedTask)
	} catch (error) {
		console.error('Failed to move task to project:', error)
	}
}

async function saveProjectPosition(e: SortableEvent) {
	if (!e.newIndex && e.newIndex !== 0) return

	const projectsActive = availableProjects.value
	// If the project was dragged to the last position, Safari will report e.newIndex as the size of the projectsActive
	// array instead of using the position. Because the index is wrong in that case, dragging the project will fail.
	// To work around that we're explicitly checking that case here and decrease the index.
	const newIndex = e.newIndex === projectsActive.length ? e.newIndex - 1 : e.newIndex

	const projectIdStr = e.item.dataset.projectId
	if (!projectIdStr) return

	const projectId = parseInt(projectIdStr)
	const project = projectStore.projects[projectId]
	if (!project) return

	const parentNode = e.to.parentNode as HTMLElement | null
	const parentProjectId = parentNode?.dataset?.projectId ? parseInt(parentNode.dataset.projectId) : 0
	const projectBefore = projectsActive[newIndex - 1] ?? null
	const projectAfter = projectsActive[newIndex + 1] ?? null
	projectUpdating.value[project.id] = true

	const position = calculateItemPosition(
		projectBefore !== null ? projectBefore.position : null,
		projectAfter !== null ? projectAfter.position : null,
	)

	try {
		// create a copy of the project in order to not violate pinia manipulation
		await projectStore.updateProject({
			...project,
			position,
			parentProjectId,
		} as IProject)
		emit('update:modelValue', availableProjects.value)
	} finally {
		projectUpdating.value[project.id] = false
	}
}
</script>
