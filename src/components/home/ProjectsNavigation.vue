<template>
	<!--								v-if="projectsVisible[n.id] ?? true"-->
	<!--		:disabled="n.id < 0 || undefined"-->
	<!--		:modelValue="p"-->
	<!--		@update:modelValue="(projects) => updateActiveProjects(n, projects)"-->
	<!--		v-for="(p, pk) in projects"-->
	<!--		:key="p.id"-->
	<!--		:data-project-id="p.id"-->
	<!--		:data-project-index="pk"-->
	<draggable
		v-model="availableProjects"
		v-bind="dragOptions"
		group="namespace-lists"
		@start="() => drag = true"
		@end="saveProjectPosition"
		handle=".handle"
		tag="ul"
		item-key="id"
		:component-data="{
			type: 'transition-group',
			name: !drag ? 'flip-list' : null,
			class: [
				'menu-list can-be-hidden',
				{ 'dragging-disabled': false }
			]
		}"
	>
		<template #item="{element: p}">
			<li
				class="list-menu loader-container is-loading-small"
				:class="{'is-loading': projectUpdating[p.id]}"
			>
				<section>
					<BaseButton
						v-if="p.childProjects.length > 0"
						@click="collapsedProjects[p.id] = !collapsedProjects[p.id]"
						class="collapse-project-button"
					>
						<icon icon="chevron-down" :class="{ 'project-is-collapsed': collapsedProjects[p.id] }"/>
					</BaseButton>
					<span class="collapse-project-button-placeholder" v-else></span>
					<BaseButton
						:to="{ name: 'project.index', params: { projectId: p.id} }"
						class="list-menu-link"
						:class="{'router-link-exact-active': currentProject.id === p.id}"
					>
						<span class="icon menu-item-icon handle">
							<icon icon="grip-lines"/>
						</span>
						<ColorBubble
							v-if="p.hexColor !== ''"
							:color="p.hexColor"
							class="mr-1"
						/>
						<span class="list-menu-title">{{ getProjectTitle(p) }}</span>
					</BaseButton>
					<BaseButton
						v-if="p.id > 0"
						class="favorite"
						:class="{'is-favorite': p.isFavorite}"
						@click="projectStore.toggleProjectFavorite(l)"
					>
						<icon :icon="p.isFavorite ? 'star' : ['far', 'star']"/>
					</BaseButton>
					<ProjectSettingsDropdown class="menu-list-dropdown" :project="p" v-if="p.id > 0">
						<template #trigger="{toggleOpen}">
							<BaseButton class="menu-list-dropdown-trigger" @click="toggleOpen">
								<icon icon="ellipsis-h" class="icon"/>
							</BaseButton>
						</template>
					</ProjectSettingsDropdown>
					<span class="list-setting-spacer" v-else></span>
				</section>
				<ProjectsNavigation
					v-if="p.childProjects.length > 0 && !collapsedProjects[p.id]"
					:projects="p.childProjects"
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
	projects: IProject[],
}>()
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
watch(
	() => props.projects,
	projects => {
		availableProjects.value = projects
		projects.forEach(p => collapsedProjects.value[p.id] = false)
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

	const project = projectsActive[newIndex]
	const projectBefore = projectsActive[newIndex - 1] ?? null
	const projectAfter = projectsActive[newIndex + 1] ?? null
	projectUpdating.value[project.id] = true

	const position = calculateItemPosition(
		projectBefore !== null ? projectBefore.position : null,
		projectAfter !== null ? projectAfter.position : null,
	)

	console.log({
		position,
		newIndex,
		project: project.id,
		projectBefore: projectBefore?.id,
		projectAfter: projectAfter?.id,
	})

	try {
		// create a copy of the project in order to not violate pinia manipulation
		await projectStore.updateProject({
			...project,
			position,
		})
	} finally {
		projectUpdating.value[project.id] = false
	}
}
</script>

<style lang="scss" scoped>
.list-setting-spacer {
	width: 2.5rem;
	flex-shrink: 0;
}

.project-is-collapsed {
	transform: rotate(-90deg);
}
</style>