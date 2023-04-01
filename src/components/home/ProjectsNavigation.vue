<template>
	<draggable
		v-model="availableProjects"
		v-bind="dragOptions"
		group="projects"
		@start="() => drag = true"
		@end="saveProjectPosition"
		handle=".handle"
		tag="ul"
		item-key="id"
		:disabled="!canEditOrder"
		:component-data="{
			type: 'transition-group',
			name: !drag ? 'flip-list' : null,
			class: [
				'menu-list can-be-hidden',
				{ 'dragging-disabled': !canEditOrder }
			]
		}"
	>
		<template #item="{element: project}">
			<li
				class="list-menu loader-container is-loading-small"
				:class="{'is-loading': projectUpdating[project.id]}"
				:data-project-id="project.id"
			>
				<ProjectsNavigationItem
					:project="project"
					:can-collapse="childProjects[project.id]?.length > 0"
					:is-collapsed="collapsedProjects[project.id] || false"
					@collapse="collapsedProjects[project.id] = !collapsedProjects[project.id]"
				/>
				<ProjectsNavigation
					v-if="!collapsedProjects[project.id]"
					v-model="childProjects[project.id]"
					:can-edit-order="true"
				/>
			</li>
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
}>()
const emit = defineEmits(['update:modelValue'])

const drag = ref(false)
const dragOptions = {
	animation: 100,
	ghostClass: 'ghost',
}

const projectStore = useProjectStore()

// Vue draggable will modify the projects list as it changes their position which will not work on a prop.
// Hence, we'll clone the prop and work on the clone.
// FIXME: cloning does not work when loading the page initially
// TODO: child projects
const collapsedProjects = ref<{ [id: IProject['id']]: boolean }>({})
const availableProjects = ref<IProject[]>([])
const childProjects = ref<{ [id: IProject['id']]: boolean }>({})
watch(
	() => props.modelValue,
	projects => {
		availableProjects.value = projects || []
		projects?.forEach(p => {
			collapsedProjects.value[p.id] = false
			childProjects.value[p.id] = projectStore.getChildProjects(p.id)
				.sort((a, b) => a.position - b.position)
		})
	},
	{immediate: true},
)

const projectUpdating = ref<{ [id: IProject['id']]: boolean }>({})

async function saveProjectPosition(e: SortableEvent) {
	if (!e.newIndex && e.newIndex !== 0) return

	const projectsActive = availableProjects.value
	// If the project was dragged to the last position, Safari will report e.newIndex as the size of the projectsActive
	// array instead of using the position. Because the index is wrong in that case, dragging the project will fail.
	// To work around that we're explicitly checking that case here and decrease the index.
	const newIndex = e.newIndex === projectsActive.length ? e.newIndex - 1 : e.newIndex

	const projectId = parseInt(e.item.dataset.projectId)
	const project = projectStore.getProjectById(projectId)

	const parentProjectId = e.to.parentNode.dataset.projectId ? parseInt(e.to.parentNode.dataset.projectId) : 0
	const projectBefore = projectsActive[newIndex - 1] ?? null
	const projectAfter = projectsActive[newIndex + 1] ?? null
	projectUpdating.value[project.id] = true

	const position = calculateItemPosition(
		projectBefore !== null ? projectBefore.position : null,
		projectAfter !== null ? projectAfter.position : null,
	)

	if (project.parentProjectId !== parentProjectId && project.parentProjectId > 0) {
		const parentProject = projectStore.getProjectById(project.parentProjectId)
		const childProjectIndex = parentProject.childProjectIds.findIndex(pId => pId === project.id)
		parentProject.childProjectIds.splice(childProjectIndex, 1)
	}

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
