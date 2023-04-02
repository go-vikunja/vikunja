<template>
	<li
		class="list-menu loader-container is-loading-small"
		:class="{'is-loading': isLoading}"
		:data-project-id="project.id"
	>
		<section>
			<BaseButton
				v-if="childProjects?.length > 0"
				@click="childProjectsOpen = !childProjectsOpen"
				class="collapse-project-button"
			>
				<icon icon="chevron-down" :class="{ 'project-is-collapsed': !childProjectsOpen }"/>
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
			v-if="childProjectsOpen"
			v-model="childProjects"
			:can-edit-order="true"
		/>
	</li>
</template>

<script setup lang="ts">
import {computed, watch, ref} from 'vue'
import {useProjectStore} from '@/stores/projects'
import {useBaseStore} from '@/stores/base'

import type {IProject} from '@/modelTypes/IProject'

import BaseButton from '@/components/base/BaseButton.vue'
import ProjectSettingsDropdown from '@/components/project/project-settings-dropdown.vue'
import {getProjectTitle} from '@/helpers/getProjectTitle'
import ColorBubble from '@/components/misc/colorBubble.vue'
import ProjectsNavigation from '@/components/home/ProjectsNavigation.vue'

const props = defineProps<{
	project: IProject,
	isLoading?: boolean,
}>()

const projectStore = useProjectStore()
const baseStore = useBaseStore()
const currentProject = computed(() => baseStore.currentProject)

const childProjectsOpen = ref(true)

const childProjects = computed(() => {
	return projectStore.getChildProjects(props.project.id)
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

.list-menu:hover > section > .favorite {
	opacity: 1;
}
</style>
