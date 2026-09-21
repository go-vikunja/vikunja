<template>
	<draggable
		v-model="availableProjects"
		animation="100"
		ghost-class="ghost"
		group="projects"
		handle=".handle"
		tag="menu"
		item-key="id"
		:disabled="!canEditOrder"
		filter=".drag-disabled"
		:component-data="{
			type: 'transition-group',
			name: !isDraggingProject ? 'flip-list' : null,
			class: [
				'menu-list can-be-hidden',
				{ 'dragging-disabled': !canEditOrder }
			],
		}"
		@start="() => isDraggingProject = true"
		@end="saveProjectPosition"
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
import {useUpdateProjectMutation, type ProjectResponse} from '@/client/queries/projects'

import {useProjects} from '@/composables/useProjects'
import {useProjectDragState} from '@/composables/useProjectDragState'

const props = defineProps<{
	modelValue?: ProjectResponse[],
	canEditOrder: boolean,
	canCollapse?: boolean,
}>()
const emit = defineEmits<{
	(e: 'update:modelValue', projects: ProjectResponse[]): void
}>()

const {isDraggingProject} = useProjectDragState()

const projectList = useProjects()
const updateMutation = useUpdateProjectMutation()

// Vue draggable will modify the projects list as it changes their position which will not work on a prop.
// Hence, we'll clone the prop and work on the clone.
const availableProjects = ref<ProjectResponse[]>([])
// Mid-drag, Sortable has moved the dragged item's DOM node, possibly into another list. Patching
// the list then anchors on that node, throws NotFoundError and leaves the sidebar half patched.
let projectsChangedDuringDrag: ProjectResponse[] | null = null
watch(
	() => props.modelValue,
	projects => {
		if (isDraggingProject.value) {
			projectsChangedDuringDrag = projects || []
			return
		}
		availableProjects.value = projects || []
	},
	{immediate: true},
)
watch(isDraggingProject, dragging => {
	if (dragging || projectsChangedDuringDrag === null) {
		return
	}
	availableProjects.value = projectsChangedDuringDrag
	projectsChangedDuringDrag = null
})

const projectUpdating = ref<Record<number, boolean>>({})

async function saveProjectPosition(e: SortableEvent) {
	isDraggingProject.value = false
	if (!e.newIndex && e.newIndex !== 0) return

	const projectsActive = availableProjects.value
	// If the project was dragged to the last position, Safari will report e.newIndex as the size of the projectsActive
	// array instead of using the position. Because the index is wrong in that case, dragging the project will fail.
	// To work around that we're explicitly checking that case here and decrease the index.
	const newIndex = e.newIndex === projectsActive.length ? e.newIndex - 1 : e.newIndex

	const projectIdStr = e.item.dataset.projectId
	if (!projectIdStr) return

	const projectId = parseInt(projectIdStr)
	const project = projectList.projects[projectId]
	if (!project) return

	const parentNode = e.to.parentNode as HTMLElement | null
	const parentProjectIdFromDom = parentNode?.dataset?.projectId ? parseInt(parentNode.dataset.projectId) : 0
	const parentProjectId = projectList.getEffectiveParentProjectId(project, parentProjectIdFromDom)
	const projectBefore = projectsActive[newIndex - 1] ?? null
	const projectAfter = projectsActive[newIndex + 1] ?? null
	projectUpdating.value[project.id] = true

	const position = calculateItemPosition(
		projectBefore !== null ? projectBefore.position : null,
		projectAfter !== null ? projectAfter.position : null,
	)

	try {
		await updateMutation.mutateAsync({
			...project,
			position,
			parent_project_id: parentProjectId,
		})
		emit('update:modelValue', availableProjects.value)
	} finally {
		projectUpdating.value[project.id] = false
	}
}
</script>
