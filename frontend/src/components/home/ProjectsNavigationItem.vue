<template>
	<li
		class="list-menu loader-container is-loading-small"
		:class="{'is-loading': isLoading}"
	>
		<div class="navigation-item">
			<BaseButton
				v-if="canCollapse && childProjects?.length > 0"
				class="collapse-project-button"
				@click="childProjectsOpen = !childProjectsOpen"
			>
				<Icon
					icon="chevron-down"
					:class="{ 'project-is-collapsed': !childProjectsOpen }"
				/>
			</BaseButton>
			<BaseButton
				:to="{ name: 'project.index', params: { projectId: project.id} }"
				class="list-menu-link"
				:class="{'router-link-exact-active': currentProject?.id === project.id}"
			>
				<span
					v-if="!canCollapse || childProjects?.length === 0"
					class="collapse-project-button-placeholder"
				/>
				<div
					class="color-bubble-handle-wrapper"
					:class="{'is-draggable': project.id > 0}"
				>
					<ColorBubble
						v-if="project.hexColor !== ''"
						:color="project.hexColor"
					/>
					<span
						v-else-if="project.id < -1"
						class="saved-filter-icon icon menu-item-icon"
					>
						<Icon icon="filter" />
					</span>
					<span
						v-if="project.id > 0"
						class="icon menu-item-icon handle"
						:class="{'has-color-bubble': project.hexColor !== ''}"
					>
						<Icon icon="grip-lines" />
					</span>
				</div>
				<span class="project-menu-title">{{ getProjectTitle(project) }}</span>
			</BaseButton>
			<BaseButton
				v-if="project.id > 0"
				class="favorite"
				:class="{'is-favorite': project.isFavorite}"
				@click="projectStore.toggleProjectFavorite(project)"
			>
				<Icon :icon="project.isFavorite ? 'star' : ['far', 'star']" />
			</BaseButton>
			<ProjectSettingsDropdown
				class="menu-list-dropdown"
				:project="project"
			>
				<template #trigger="{toggleOpen}">
					<BaseButton
						class="menu-list-dropdown-trigger"
						@click="toggleOpen"
					>
						<Icon
							icon="ellipsis-h"
							class="icon"
						/>
					</BaseButton>
				</template>
			</ProjectSettingsDropdown>
		</div>
		<ProjectsNavigation
			v-if="childProjectsOpen && canCollapse"
			:model-value="childProjects"
			:can-edit-order="true"
			:can-collapse="canCollapse"
		/>
	</li>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useProjectStore} from '@/stores/projects'
import {useBaseStore} from '@/stores/base'
import {useStorage} from '@vueuse/core'

import type {IProject} from '@/modelTypes/IProject'

import BaseButton from '@/components/base/BaseButton.vue'
import ProjectSettingsDropdown from '@/components/project/ProjectSettingsDropdown.vue'
import {getProjectTitle} from '@/helpers/getProjectTitle'
import ColorBubble from '@/components/misc/ColorBubble.vue'
import ProjectsNavigation from '@/components/home/ProjectsNavigation.vue'

const props = defineProps<{
	project: IProject,
	isLoading?: boolean,
	canCollapse?: boolean,
}>()

const projectStore = useProjectStore()
const baseStore = useBaseStore()
const currentProject = computed(() => baseStore.currentProject)

// Persist open state across browser reloads. Using a separate ref for the state 
// allows us to use only one entry in local storage instead of one for every project id.
type OpenState = { [key: number]: boolean }
const childProjectsOpenState = useStorage<OpenState>('navigation-child-projects-open', {})
const childProjectsOpen = computed({
	get() {
		return childProjectsOpenState.value[props.project.id] ?? true
	},
	set(open) {
		childProjectsOpenState.value[props.project.id] = open
	},
})

const childProjects = computed(() => {
	return projectStore.getChildProjects(props.project.id)
		.filter(p => !p.isArchived)
		.sort((a, b) => a.position - b.position)
})
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

.list-menu:hover > div > .favorite {
	opacity: 1;
}

.list-menu:hover > div > a > .color-bubble-handle-wrapper.is-draggable > {
	.saved-filter-icon,
	.color-bubble {
		opacity: 0;
	}
}

.is-touch .color-bubble {
	opacity: 1 !important;
}

.color-bubble-handle-wrapper {
	position: relative;
	width: 1rem;
	height: 1rem;
	display: flex;
	align-items: center;
	justify-content: flex-start;
	margin-right: .25rem;
	flex-shrink: 0;

	.color-bubble, .icon {
		transition: all $transition;
		position: absolute;
		width: 12px;
		margin: 0 !important;
		padding: 0 !important;
	}
}

.project-menu-title {
	overflow: hidden;
	text-overflow: ellipsis;
}

.saved-filter-icon {
	color: var(--grey-300) !important;
	font-size: .75rem;
}

.is-touch .handle.has-color-bubble {
	display: none !important;
}

.navigation-item:has(*:focus-visible) {
	box-shadow: 0 0 0 2px hsla(var(--primary-hsl), 0.5);
	background-color: var(--white);

	.favorite, .menu-list-dropdown {
		opacity: 1;
	}
}

.navigation-item a:focus-visible {
	// The focus ring is already added to the navigation-item, so we don't need to add it again.
	box-shadow: none;
}
</style>
