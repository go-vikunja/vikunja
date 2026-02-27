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
			ref="switchViewContainerRef"
			class="switch-view-container d-print-none"
			:class="{'is-justify-content-flex-end': views.length === 1}"
		>
			<!-- Dropdown mode when buttons overflow -->
			<Dropdown
				v-if="isOverflowing && views.length > 1"
				class="switch-view-dropdown"
			>
				<template #trigger="{ toggleOpen }">
					<BaseButton
						class="switch-view switch-view-dropdown-trigger"
						@click="toggleOpen"
					>
						{{ activeViewTitle }}
						<Icon
							icon="chevron-down"
							class="dropdown-icon"
						/>
					</BaseButton>
				</template>
				<template #default="{ close }">
					<div @click="close">
						<DropdownItem
							v-for="view in views"
							:key="view.id"
							:to="getViewRoute(view)"
							:class="{'is-active': view.id === viewId}"
						>
							{{ getViewTitle(view) }}
						</DropdownItem>
					</div>
				</template>
			</Dropdown>

			<!-- Inline buttons, hidden when overflowing but kept in DOM for width measurement -->
			<div
				v-if="views.length > 1"
				ref="switchViewRef"
				class="switch-view"
				:class="{'switch-view--hidden': isOverflowing || !overflowChecked}"
				:aria-hidden="isOverflowing || undefined"
			>
				<BaseButton
					v-for="view in views"
					:key="view.id"
					class="switch-view-button"
					:class="{'is-active': view.id === viewId}"
					:to="getViewRoute(view)"
					:tabindex="isOverflowing ? -1 : undefined"
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
import {computed, ref, watch, nextTick, onMounted} from 'vue'
import {useResizeObserver} from '@vueuse/core'
import {useI18n} from 'vue-i18n'

import BaseButton from '@/components/base/BaseButton.vue'
import Dropdown from '@/components/misc/Dropdown.vue'
import DropdownItem from '@/components/misc/DropdownItem.vue'
import Icon from '@/components/misc/Icon'
import Message from '@/components/misc/Message.vue'
import CustomTransition from '@/components/misc/CustomTransition.vue'

import {getProjectTitle} from '@/helpers/getProjectTitle'
import {useTitle} from '@/composables/useTitle'

import {useBaseStore} from '@/stores/base'
import {useProjectStore} from '@/stores/projects'
import {useViewFiltersStore} from '@/stores/viewFilters'

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
const viewFiltersStore = useViewFiltersStore()

const switchViewContainerRef = ref<HTMLElement>()
const switchViewRef = ref<HTMLElement>()
const isOverflowing = ref(false)
const overflowChecked = ref(false)

function checkOverflow() {
	if (!switchViewRef.value || !switchViewContainerRef.value) {
		return
	}
	const buttonsWidth = switchViewRef.value.scrollWidth
	const containerWidth = switchViewContainerRef.value.clientWidth
	isOverflowing.value = buttonsWidth > containerWidth
	overflowChecked.value = true
}

onMounted(() => {
	checkOverflow()
})

useResizeObserver(switchViewContainerRef, () => {
	requestAnimationFrame(() => checkOverflow())
})

const currentProject = computed<IProject>(() => {
	return baseStore.currentProject || {
		id: 0,
		title: '',
		isArchived: false,
		maxPermission: null,
	}
})
useTitle(() => currentProject.value?.id ? getProjectTitle(currentProject.value) : '')

const views = computed(() => projectStore.projects[props.projectId]?.views)

const activeViewTitle = computed(() => {
	const activeView = views.value?.find((v: IProjectView) => v.id === props.viewId)
	return activeView ? getViewTitle(activeView) : ''
})

// Re-check overflow when views change
watch(views, () => {
	nextTick(() => checkOverflow())
})

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

function getViewRoute(view: IProjectView) {
	const storedQuery = viewFiltersStore.getViewQuery(view.id)
	return {
		name: 'project.view',
		params: {projectId: props.projectId, viewId: view.id},
		query: storedQuery,
	}
}
</script>

<style lang="scss" scoped>
.switch-view-container {
	position: relative;
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

.switch-view--hidden {
	position: absolute;
	visibility: hidden;
	pointer-events: none;
	white-space: nowrap;
	inset-inline-start: 0;
	inset-inline-end: 0;
	overflow: hidden;
}

.switch-view-dropdown-trigger {
	cursor: pointer;
	display: inline-flex;
	align-items: center;
	gap: .25rem;
	font-weight: bold;
	color: var(--switch-view-color);
	background: var(--primary);
}

.dropdown-icon {
	font-size: .6rem;
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
