<template>
	<div
		:class="{ 'is-loading': projectService.loading, 'is-archived': currentProject.isArchived}"
		class="loader-container"
	>
		<div class="switch-view-container">
			<div class="switch-view">
				<BaseButton
					v-shortcut="'g l'"
					:title="$t('keyboardShortcuts.project.switchToListView')"
					class="switch-view-button"
					:class="{'is-active': viewName === 'project'}"
					:to="{ name: 'project.list',   params: { projectId } }"
				>
					{{ $t('project.list.title') }}
				</BaseButton>
				<BaseButton
					v-shortcut="'g g'"
					:title="$t('keyboardShortcuts.project.switchToGanttView')"
					class="switch-view-button"
					:class="{'is-active': viewName === 'gantt'}"
					:to="{ name: 'project.gantt',  params: { projectId } }"
				>
					{{ $t('project.gantt.title') }}
				</BaseButton>
				<BaseButton
					v-shortcut="'g t'"
					:title="$t('keyboardShortcuts.project.switchToTableView')"
					class="switch-view-button"
					:class="{'is-active': viewName === 'table'}"
					:to="{ name: 'project.table',  params: { projectId } }"
				>
					{{ $t('project.table.title') }}
				</BaseButton>
				<BaseButton
					v-shortcut="'g k'"
					:title="$t('keyboardShortcuts.project.switchToKanbanView')"
					class="switch-view-button"
					:class="{'is-active': viewName === 'kanban'}"
					:to="{ name: 'project.kanban', params: { projectId } }"
				>
					{{ $t('project.kanban.title') }}
				</BaseButton>
			</div>
			<slot name="header" />
		</div>
		<CustomTransition name="fade">
			<Message variant="warning" v-if="currentProject.isArchived" class="mb-4">
				{{ $t('project.archivedText') }}
			</Message>
		</CustomTransition>

		<slot v-if="loadedProjectId"/>
	</div>
</template>

<script setup lang="ts">
import {ref, computed, watch} from 'vue'
import {useRoute} from 'vue-router'

import BaseButton from '@/components/base/BaseButton.vue'
import Message from '@/components/misc/message.vue'
import CustomTransition from '@/components/misc/CustomTransition.vue'

import ProjectModel from '@/models/project'
import ProjectService from '@/services/project'

import {getProjectTitle} from '@/helpers/getProjectTitle'
import {saveProjectToHistory} from '@/modules/projectHistory'
import {useTitle} from '@/composables/useTitle'

import {useBaseStore} from '@/stores/base'
import {useProjectStore} from '@/stores/projects'

const props = defineProps({
	projectId: {
		type: Number,
		required: true,
	},
	viewName: {
		type: String,
		required: true,
	},
})

const route = useRoute()

const baseStore = useBaseStore()
const projectStore = useProjectStore()
const projectService = ref(new ProjectService())
const loadedProjectId = ref(0)

const currentProject = computed(() => {
	return typeof baseStore.currentProject === 'undefined' ? {
		id: 0,
		title: '',
		isArchived: false,
		maxRight: null,
	} : baseStore.currentProject
})
useTitle(() => currentProject.value.id ? getProjectTitle(currentProject.value) : '')

// watchEffect would be called every time the prop would get a value assigned, even if that value was the same as before.
// This resulted in loading and setting the project multiple times, even when navigating away from it.
// This caused wired bugs where the project background would be set on the home page but only right after setting a new 
// project background and then navigating to home. It also highlighted the project in the menu and didn't allow changing any
// of it, most likely due to the rights not being properly populated.
watch(
	() => props.projectId,
	// loadProject
	async (projectIdToLoad: number) => {
		const projectData = {id: projectIdToLoad}
		saveProjectToHistory(projectData)

		// Don't load the project if we either already loaded it or aren't dealing with a project at all currently and
		// the currently loaded project has the right set.
		if (
			(
				projectIdToLoad === loadedProjectId.value ||
				typeof projectIdToLoad === 'undefined' ||
				projectIdToLoad === currentProject.value.id
			)
			&& typeof currentProject.value !== 'undefined' && currentProject.value.maxRight !== null
		) {
			loadedProjectId.value = props.projectId
			return
		}

		console.debug(`Loading project, props.viewName = ${props.viewName}, $route.params =`, route.params, `, loadedProjectId = ${loadedProjectId.value}, currentProject = `, currentProject.value)

		// Set the current project to the one we're about to load so that the title is already shown at the top
		loadedProjectId.value = 0
		const projectFromStore = projectStore.getProjectById(projectData.id)
		if (projectFromStore !== null) {
			baseStore.setBackground(null)
			baseStore.setBlurHash(null)
			baseStore.handleSetCurrentProject({project: projectFromStore})
		}

		// We create an extra project object instead of creating it in project.value because that would trigger a ui update which would result in bad ux.
		const project = new ProjectModel(projectData)
		try {
			const loadedProject = await projectService.value.get(project)
			baseStore.handleSetCurrentProject({project: loadedProject})
		} finally {
			loadedProjectId.value = props.projectId
		}
	},
	{immediate: true},
)
</script>

<style lang="scss" scoped>
.switch-view-container {
  @media screen and (max-width: $tablet) {
    display: flex;
    justify-content: center;
    flex-direction: column;
  }
}

.switch-view {
  background: var(--white);
  display: inline-flex;
  border-radius: $radius;
  font-size: .75rem;
  box-shadow: var(--shadow-sm);
  height: $switch-view-height;
  margin: 0 auto 1rem;
  padding: .5rem;
}

.switch-view-button {
	padding: .25rem .5rem;
	display: block;
	border-radius: $radius;
	transition: all 100ms;

	&:not(:last-child) {
		margin-right: .5rem;
	}

	&:hover {
		color: var(--switch-view-color);
		background: var(--primary);
	}

	&.is-active {
		color: var(--switch-view-color);
		background: var(--primary);
		font-weight: bold;
		box-shadow: var(--shadow-xs);
	}
}

// FIXME: this should be in notification and set via a prop
.is-archived .notification.is-warning {
  margin-bottom: 1rem;
}
</style>