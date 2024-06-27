<template>
	<div
		:class="{ 'is-loading': projectService.loading, 'is-archived': currentProject?.isArchived}"
		class="loader-container"
	>
		<h1 class="project-title-print">
			{{ getProjectTitle(currentProject) }}
		</h1>

		<div
			class="switch-view-container d-print-none"
			:class="{'is-justify-content-flex-end': views.length === 1}"
		>
			<div
				v-if="views.length > 1"
				class="switch-view"
			>
				<BaseButton
					v-for="view in views"
					:key="view.id"
					class="switch-view-button"
					:class="{'is-active': view.id === viewId}"
					:to="{ name: 'project.view', params: { projectId, viewId: view.id } }"
				>
					{{ getViewTitle(view) }}
				</BaseButton>
			</div>
			<slot name="header" />
		</div>
		<CustomTransition name="fade">
			<Message
				v-if="currentProject?.isArchived"
				variant="warning"
				class="mb-4"
			>
				{{ $t('project.archivedMessage') }}
			</Message>
		</CustomTransition>

		<slot v-if="loadedProjectId" />
	</div>
</template>

<script setup lang="ts">
import {computed, ref, watch} from 'vue'
import {useRoute} from 'vue-router'

import BaseButton from '@/components/base/BaseButton.vue'
import Message from '@/components/misc/Message.vue'
import CustomTransition from '@/components/misc/CustomTransition.vue'

import ProjectModel from '@/models/project'
import ProjectService from '@/services/project'

import {getProjectTitle} from '@/helpers/getProjectTitle'
import {saveProjectToHistory} from '@/modules/projectHistory'
import {useTitle} from '@/composables/useTitle'

import {useBaseStore} from '@/stores/base'
import {useProjectStore} from '@/stores/projects'
import type {IProject} from '@/modelTypes/IProject'
import type {IProjectView} from '@/modelTypes/IProjectView'
import {useI18n} from 'vue-i18n'

const props = defineProps<{
	projectId: IProject['id'],
	viewId: IProjectView['id'],
}>()

const route = useRoute()
const {t} = useI18n()

const baseStore = useBaseStore()
const projectStore = useProjectStore()
const projectService = ref(new ProjectService())
const loadedProjectId = ref(0)

const currentProject = computed<IProject>(() => {
	return typeof baseStore.currentProject === 'undefined' ? {
		id: 0,
		title: '',
		isArchived: false,
		maxRight: null,
	} : baseStore.currentProject
})
useTitle(() => currentProject.value?.id ? getProjectTitle(currentProject.value) : '')

const views = computed(() => projectStore.projects[props.projectId]?.views)

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
				projectIdToLoad === currentProject.value?.id
			)
			&& typeof currentProject.value !== 'undefined' && currentProject.value.maxRight !== null
		) {
			loadedProjectId.value = projectIdToLoad
			return
		}

		console.debug('Loading project, $route.params =', route.params, `, loadedProjectId = ${loadedProjectId.value}, currentProject = `, currentProject.value)

		// Set the current project to the one we're about to load so that the title is already shown at the top
		loadedProjectId.value = 0
		const projectFromStore = projectStore.projects[projectData.id]
		if (projectFromStore) {
			baseStore.handleSetCurrentProject({project: projectFromStore})
		}

		// We create an extra project object instead of creating it in project.value because that would trigger a ui update which would result in bad ux.
		const project = new ProjectModel(projectData)
		try {
			const loadedProject = await projectService.value.get(project)
			baseStore.handleSetCurrentProject({project: loadedProject})
		} finally {
			loadedProjectId.value = projectIdToLoad
		}
	},
	{immediate: true},
)

function getViewTitle(view: IProjectView) {
	switch (view.title) {
		case 'List':
			return t('project.list.title')
		case 'Gantt':
			return t('project.gantt.title')
		case 'Table':
			return t('project.table.title')
		case 'Kanban':
			return t('project.kanban.title')
	}
	
	return view.title
}
</script>

<style lang="scss" scoped>
.switch-view-container {
	min-height: $switch-view-height;
	margin-bottom: 1rem;
	
	display: flex;
	justify-content: space-between;
	align-items: center;	
	gap: 1rem;
	
	@media screen and (max-width: $tablet) {
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

.project-title-print {
	display: none;
	font-size: 1.75rem;
	text-align: center;
	margin-bottom: .5rem;

	@media print {
		display: block;
	}
}
</style>