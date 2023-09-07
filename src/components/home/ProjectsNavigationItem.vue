<template>
	<li
		class="list-menu loader-container is-loading-small"
		:class="{'is-loading': isLoading}"
	>
		<div>
			<BaseButton
				v-if="canCollapse && childProjects?.length > 0"
				@click="childProjectsOpen = !childProjectsOpen"
				class="collapse-project-button"
			>
				<icon icon="chevron-down" :class="{ 'project-is-collapsed': !childProjectsOpen }"/>
			</BaseButton>
			<BaseButton
				:to="{ name: 'project.index', params: { projectId: project.id} }"
				class="list-menu-link"
				:class="{'router-link-exact-active': currentProject?.id === project.id}"
			>
				<span
					v-if="!canCollapse || childProjects?.length === 0"
					class="collapse-project-button-placeholder"
				></span>
				<div class="color-bubble-handle-wrapper" :class="{'is-draggable': project.id > 0}">
					<ColorBubble
						v-if="project.hexColor !== ''"
						:color="project.hexColor"
					/>
					<span v-else-if="project.id < -1" class="saved-filter-icon icon menu-item-icon">
						<icon icon="filter"/>
					</span>
					<span
						v-if="project.id > 0"
						class="icon menu-item-icon handle lines-handle"
						:class="{'has-color-bubble': project.hexColor !== ''}"
					>
						<icon icon="grip-lines"/>
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
				<icon :icon="project.isFavorite ? 'star' : ['far', 'star']"/>
			</BaseButton>
			<ProjectSettingsDropdown
				v-if="project.id > 0"
				class="menu-list-dropdown"
				:project="project"
				:level="level"
			>
				<template #trigger="{toggleOpen}">
					<BaseButton class="menu-list-dropdown-trigger" @click="toggleOpen">
						<icon icon="ellipsis-h" class="icon"/>
					</BaseButton>
				</template>
			</ProjectSettingsDropdown>
			<span class="list-setting-spacer" v-else></span>
		</div>
		<ProjectsNavigation
			v-if="canNestDeeper && childProjectsOpen && canCollapse"
			:model-value="childProjects"
			:can-edit-order="true"
			:can-collapse="canCollapse"
			:level="level + 1"
		/>
	</li>
</template>

<script setup lang="ts">
import {computed, ref} from 'vue'
import {useProjectStore} from '@/stores/projects'
import {useBaseStore} from '@/stores/base'

import type {IProject} from '@/modelTypes/IProject'

import BaseButton from '@/components/base/BaseButton.vue'
import ProjectSettingsDropdown from '@/components/project/project-settings-dropdown.vue'
import {getProjectTitle} from '@/helpers/getProjectTitle'
import ColorBubble from '@/components/misc/colorBubble.vue'
import ProjectsNavigation from '@/components/home/ProjectsNavigation.vue'
import {canNestProjectDeeper} from '@/helpers/canNestProjectDeeper'

const {
	project,
	isLoading,
	canCollapse,
	level = 0,
} = defineProps<{
	project: IProject,
	isLoading?: boolean,
	canCollapse?: boolean,
	level?: number,
}>()

const projectStore = useProjectStore()
const baseStore = useBaseStore()
const currentProject = computed(() => baseStore.currentProject)

const childProjectsOpen = ref(true)

const childProjects = computed(() => {
	if (!canNestDeeper.value) {
		return []
	}

	return projectStore.getChildProjects(project.id)
		.filter(p => !p.isArchived)
		.sort((a, b) => a.position - b.position)
})

const canNestDeeper = computed(() => canNestProjectDeeper(level))
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
</style>
