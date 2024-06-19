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
			name: !drag ? 'flip-list' : null,
			class: [
				'menu-list can-be-hidden',
				{ 'dragging-disabled': !canEditOrder }
			],
		}"
		@start="() => drag = true"
		@end="saveProjectPosition"
	>
		<template #item="{element: project}">
			<ProjectsNavigationItem
				:class="{'drag-disabled': project.id < 0}"
				:project="project"
				:is-loading="projectUpdating[project.id]"
				:can-collapse="canCollapse"
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

async function saveProjectPosition(e: SortableEvent) {
	drag.value = false

	if (!e.newIndex && e.newIndex !== 0) return

	const projectsActive = availableProjects.value
	// If the project was dragged to the last position, Safari will report e.newIndex as the size of the projectsActive
	// array instead of using the position. Because the index is wrong in that case, dragging the project will fail.
	// To work around that we're explicitly checking that case here and decrease the index.
	const newIndex = e.newIndex === projectsActive.length ? e.newIndex - 1 : e.newIndex

	const projectId = parseInt(e.item.dataset.projectId)
	const project = projectStore.projects[projectId]

	const parentProjectId = e.to.parentNode.dataset.projectId ? parseInt(e.to.parentNode.dataset.projectId) : 0
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
		})
		emit('update:modelValue', availableProjects.value)
	} finally {
		projectUpdating.value[project.id] = false
	}
}
</script>
