<template>
	<div
		class="loader-container"
		:class="{
			'is-loading': isLoadingProject,
			'is-archived': currentProject?.isArchived,
		}"
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
				class="mbe-4"
			>
				{{ $t('project.archivedMessage') }}
			</Message>
		</CustomTransition>

		<slot v-if="!isLoadingProject" />
	</div>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useI18n} from 'vue-i18n'

import BaseButton from '@/components/base/BaseButton.vue'
import Message from '@/components/misc/Message.vue'
import CustomTransition from '@/components/misc/CustomTransition.vue'

import {getProjectTitle} from '@/helpers/getProjectTitle'
import {useTitle} from '@/composables/useTitle'

import {useBaseStore} from '@/stores/base'
import {useProjectStore} from '@/stores/projects'

import type {IProject} from '@/modelTypes/IProject'
import type {IProjectView} from '@/modelTypes/IProjectView'

const props = defineProps<{
	isLoadingProject: boolean,
	projectId: IProject['id'],
	viewId: IProjectView['id'],
}>()

const {t} = useI18n()

const baseStore = useBaseStore()
const projectStore = useProjectStore()

const currentProject = computed<IProject>(() => {
	return typeof baseStore.currentProject === 'undefined' ? {
		id: 0,
		title: '',
		isArchived: false,
		maxPermission: null,
	} : baseStore.currentProject
})
useTitle(() => currentProject.value?.id ? getProjectTitle(currentProject.value) : '')

const views = computed(() => projectStore.projects[props.projectId]?.views)

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
	min-block-size: $switch-view-height;
	margin-block-end: 1rem;
	
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
		margin-inline-end: .5rem;
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
	margin-block-end: 1rem;
}

.project-title-print {
	display: none;
	font-size: 1.75rem;
	text-align: center;
	margin-block-end: .5rem;

	@media print {
		display: block;
	}
}
</style>
