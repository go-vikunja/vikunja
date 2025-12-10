<template>
	<li
		class="list-menu loader-container is-loading-small"
		:class="{
			'is-loading': isLoading,
			'is-drop-target': isDropTarget,
		}"
		:data-project-id="project.id"
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
			<span
				v-if="canEditOrder && project.id > 0 && project.maxPermission !== null && project.maxPermission > PERMISSIONS.READ"
				class="icon menu-item-icon handle drag-handle-standalone"
				@mousedown.stop
				@click.stop.prevent
				@touchstart.stop
			>
				<Icon icon="grip-lines" />
			</span>
			<BaseButton
				:to="{ name: 'project.index', params: { projectId: project.id} }"
				class="list-menu-link"
				:class="{'router-link-exact-active': currentProject?.id === project.id}"
			>
				<span
					v-if="!canCollapse || childProjects?.length === 0"
					class="collapse-project-button-placeholder"
				/>
				<div class="color-bubble-wrapper">
					<ColorBubble
						v-if="project.hexColor !== ''"
						:color="project.hexColor"
						:aria-label="$t('project.color')"
					/>
					<span
						v-else-if="project.id < -1"
						class="saved-filter-icon icon menu-item-icon"
					>
						<Icon icon="filter" />
					</span>
				</div>
				<span class="project-menu-title">{{ getProjectTitle(project) }}</span>
			</BaseButton>
			<BaseButton
				v-if="project.id > 0 && project.maxPermission !== null && project.maxPermission > PERMISSIONS.READ"
				class="favorite"
				:class="{'is-favorite': project.isFavorite}"
				@click="projectStore.toggleProjectFavorite(project)"
			>
				<span class="is-sr-only">{{ project.isFavorite ? $t('project.unfavorite') : $t('project.favorite') }}</span>
				<Icon :icon="project.isFavorite ? 'star' : ['far', 'star']" />
			</BaseButton>
			<ProjectSettingsDropdown
				v-if="project.maxPermission !== null && project.maxPermission > PERMISSIONS.READ"
				class="menu-list-dropdown"
				:project="project"
				:simple="true"
			>
				<template #trigger="{toggleOpen}">
					<BaseButton
						class="menu-list-dropdown-trigger"
						@click="toggleOpen"
					>
						<span class="is-sr-only">{{ $t('project.openSettingsMenu') }}</span>
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
import {computed, ref, onUnmounted, watch} from 'vue'
import {useProjectStore} from '@/stores/projects'
import {useBaseStore} from '@/stores/base'
import {useTaskStore} from '@/stores/tasks'
import {useStorage} from '@vueuse/core'

import type {IProject} from '@/modelTypes/IProject'

import BaseButton from '@/components/base/BaseButton.vue'
import ProjectSettingsDropdown from '@/components/project/ProjectSettingsDropdown.vue'
import {getProjectTitle} from '@/helpers/getProjectTitle'
import ColorBubble from '@/components/misc/ColorBubble.vue'
import ProjectsNavigation from '@/components/home/ProjectsNavigation.vue'
import {PERMISSIONS} from '@/constants/permissions'

const props = defineProps<{
	project: IProject,
	isLoading?: boolean,
	canCollapse?: boolean,
	canEditOrder?: boolean,
}>()

const taskStore = useTaskStore()
const isHoveredDuringDrag = ref(false)

// Track mouse position during drag to detect hover (mouseenter doesn't fire during drag)
function handleMouseMove(e: MouseEvent) {
	if (!taskStore.draggedTask) {
		isHoveredDuringDrag.value = false
		return
	}

	const elementsUnderMouse = document.elementsFromPoint(e.clientX, e.clientY)
	const isOverThisProject = elementsUnderMouse.some(el => {
		const projectId = (el as HTMLElement).dataset?.projectId
		return projectId && parseInt(projectId) === props.project.id
	})

	isHoveredDuringDrag.value = isOverThisProject
}

// Only add the listener when a task is being dragged
// Use capture phase to receive events before Sortable.js can prevent them
watch(() => taskStore.draggedTask, (draggedTask) => {
	if (draggedTask) {
		document.addEventListener('mousemove', handleMouseMove, true)
		document.addEventListener('dragover', handleMouseMove, true)
	} else {
		document.removeEventListener('mousemove', handleMouseMove, true)
		document.removeEventListener('dragover', handleMouseMove, true)
		isHoveredDuringDrag.value = false
	}
}, {immediate: true})

onUnmounted(() => {
	document.removeEventListener('mousemove', handleMouseMove, true)
	document.removeEventListener('dragover', handleMouseMove, true)
})

// Show drop target highlight when a task is being dragged and this project is hovered
const isDropTarget = computed(() => {
	if (!taskStore.draggedTask || !isHoveredDuringDrag.value) {
		return false
	}
	// Highlight any valid project (not a pseudo project, has write permission)
	// The actual drop logic will handle the case when it's the same project (no-op)
	return props.project.id > 0
		&& props.project.maxPermission !== null
		&& props.project.maxPermission > PERMISSIONS.READ
})

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
.list-menu {
	transition: background-color $transition;
}

.list-setting-spacer {
	inline-size: 5rem;
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

.list-menu:hover > div > .drag-handle-standalone {
	opacity: 1;
}

.list-menu:hover .color-bubble-wrapper > .color-bubble {
	opacity: 0;
}

.is-touch .color-bubble {
	opacity: 1 !important;
}

.color-bubble-wrapper {
	position: relative;
	inline-size: 1rem;
	block-size: 1rem;
	display: flex;
	align-items: center;
	justify-content: flex-start;
	margin-inline-end: .25rem;
	flex-shrink: 0;

	.color-bubble, .icon {
		transition: all $transition;
		position: absolute;
		inline-size: 12px;
		margin: 0 !important;
		padding: 0 !important;
	}
}

.drag-handle-standalone {
	inline-size: 1rem;
	block-size: 1rem;
	opacity: 0;
	cursor: grab;
	transition: opacity $transition;
	z-index: 2;

	position: absolute;
	inset-inline-start: 1.75rem;

	&:active {
		cursor: grabbing;
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

@media (pointer: coarse) {
	.drag-handle-standalone {
		display: none !important;
	}
}

.navigation-item {
	position: relative;
	transition: background-color $transition, box-shadow $transition;
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

.is-drop-target {
	background-color: hsla(var(--primary-hsl), 0.15);
	border-radius: $radius;

	.navigation-item {
		background-color: hsla(var(--primary-hsl), 0.1);
		box-shadow: inset 0 0 0 2px var(--primary);
		border-radius: $radius;
	}
}
</style>
