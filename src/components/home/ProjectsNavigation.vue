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
				<section>
					<BaseButton
						v-if="childProjects[project.id]?.length > 0"
						@click="collapsedProjects[project.id] = !collapsedProjects[project.id]"
						class="collapse-project-button"
					>
						<icon icon="chevron-down" :class="{ 'project-is-collapsed': collapsedProjects[project.id] }"/>
					</BaseButton>
					<span class="collapse-project-button-placeholder" v-else></span>
					<BaseButton
						:to="{ name: 'project.index', params: { projectId: project.id} }"
						class="list-menu-link"
						:class="{'router-link-exact-active': currentProject.id === project.id}"
					>
						<span class="icon menu-item-icon handle">
							<icon icon="grip-lines"/>
						</span>
						<ColorBubble
							v-if="project.hexColor !== ''"
							:color="project.hexColor"
							class="mr-1"
						/>
						<span class="list-menu-title">{{ getProjectTitle(project) }}</span>
					</BaseButton>
					<BaseButton
						v-if="project.id > 0"
						class="favorite"
						:class="{'is-favorite': project.isFavorite}"
						@click="projectStore.toggleProjectFavorite(project)"
					>
						<icon :icon="project.isFavorite ? 'star' : ['far', 'star']"/>
					</BaseButton>
					<ProjectSettingsDropdown class="menu-list-dropdown" :project="project" v-if="project.id > 0">
						<template #trigger="{toggleOpen}">
							<BaseButton class="menu-list-dropdown-trigger" @click="toggleOpen">
								<icon icon="ellipsis-h" class="icon"/>
							</BaseButton>
						</template>
					</ProjectSettingsDropdown>
					<span class="list-setting-spacer" v-else></span>
				</section>
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
import {ref, computed, watch} from 'vue'
import draggable from 'zhyswan-vuedraggable'
import type {SortableEvent} from 'sortablejs'

import BaseButton from '@/components/base/BaseButton.vue'
import ProjectSettingsDropdown from '@/components/project/project-settings-dropdown.vue'

import {calculateItemPosition} from '@/helpers/calculateItemPosition'
import {getProjectTitle} from '@/helpers/getProjectTitle'
import type {IProject} from '@/modelTypes/IProject'
import ColorBubble from '@/components/misc/colorBubble.vue'

import {useBaseStore} from '@/stores/base'
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

const baseStore = useBaseStore()
const projectStore = useProjectStore()
const currentProject = computed(() => baseStore.currentProject)

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

<style lang="scss" scoped>
.list-setting-spacer {
	width: 5rem;
	flex-shrink: 0;
}

.project-is-collapsed {
	transform: rotate(-90deg);
}

.favorite {
	transition: opacity $transition, color $transition;
	opacity: 0;

	&:hover,
	&.is-favorite {
		opacity: 1;
		color: var(--warning);
	}
}

.list-menu:hover > section > .favorite {
	opacity: 1;
}
</style>