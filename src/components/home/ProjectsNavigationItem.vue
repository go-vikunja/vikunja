<template>
	<section>
		<BaseButton
			v-if="canCollapse"
			@click="emit('collapse')"
			class="collapse-project-button"
		>
			<icon icon="chevron-down" :class="{ 'project-is-collapsed': isCollapsed }"/>
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
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useProjectStore} from '@/stores/projects'
import {useBaseStore} from '@/stores/base'

import type {IProject} from '@/modelTypes/IProject'

import BaseButton from '@/components/base/BaseButton.vue'
import ProjectSettingsDropdown from '@/components/project/project-settings-dropdown.vue'
import {getProjectTitle} from '@/helpers/getProjectTitle'
import ColorBubble from '@/components/misc/colorBubble.vue'

defineProps<{
	project: IProject,
	isCollapsed: boolean,
	canCollapse: boolean,
}>()

const emit = defineEmits(['collapse'])

const projectStore = useProjectStore()
const baseStore = useBaseStore()
const currentProject = computed(() => baseStore.currentProject)
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
