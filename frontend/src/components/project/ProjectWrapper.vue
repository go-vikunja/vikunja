<template>
	<div
		class="loader-container"
		:class="{
			'is-loading': isLoadingProject,
			'is-archived': currentProject.is_archived,
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
				<template #trigger="{ toggleOpen, open }">
					<SwitchView
						:is="BaseButton"
						class="switch-view-dropdown-trigger"
						:aria-expanded="open"
						@click="toggleOpen"
					>
						{{ activeViewTitle }}
						<Icon
							icon="chevron-down"
							class="dropdown-icon"
						/>
					</SwitchView>
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
			<SwitchView
				v-if="views.length > 1"
				ref="switchViewRef"
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
			</SwitchView>
			<slot name="header" />
		</div>
		<CustomTransition name="fade">
			<Message
				v-if="currentProject.is_archived"
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
import {unrefElement, useResizeObserver} from '@vueuse/core'
import {useI18n} from 'vue-i18n'

import BaseButton from '@/components/base/BaseButton.vue'
import Dropdown from '@/components/misc/Dropdown.vue'
import DropdownItem from '@/components/misc/DropdownItem.vue'
import Icon from '@/components/misc/Icon'
import Message from '@/components/misc/Message.vue'
import CustomTransition from '@/components/misc/CustomTransition.vue'
import SwitchView from '@/components/misc/SwitchView.vue'

import {getProjectTitle} from '@/helpers/getProjectTitle'
import {useTitle} from '@/composables/useTitle'

import {useViewFiltersStore} from '@/stores/viewFilters'
import {useCurrentProject} from '@/composables/useCurrentProject'

import type {ProjectView} from '@/client/generated'
import {normalizeProject} from '@/client/queries/projects'

const props = defineProps<{
	isLoadingProject: boolean,
	projectId: number,
	viewId: number,
}>()

const {t} = useI18n()

const viewFiltersStore = useViewFiltersStore()
const {currentProject: queriedProject} = useCurrentProject()

const switchViewContainerRef = ref<HTMLElement>()
const switchViewRef = ref<InstanceType<typeof SwitchView>>()
const isOverflowing = ref(false)
const overflowChecked = ref(false)

function checkOverflow() {
	const switchView = unrefElement(switchViewRef)
	if (!switchView || !switchViewContainerRef.value) {
		return
	}
	const buttonsWidth = switchView.scrollWidth
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

const currentProject = computed(() => queriedProject.value ?? normalizeProject({id: 0}))
useTitle(() => currentProject.value?.id ? getProjectTitle(currentProject.value) : '')

const views = computed(() => currentProject.value.views)

const activeViewTitle = computed(() => {
	const activeView = views.value.find((view: ProjectView) => view.id === props.viewId)
	return activeView ? getViewTitle(activeView) : ''
})

// Re-check overflow when views change
watch(views, () => {
	nextTick(() => checkOverflow())
})

function getViewTitle(view: ProjectView) {
	switch (view.title) {
		case 'List':
			return t('project.list.title')
		case 'Gantt':
			return t('project.gantt.title')
		case 'Table':
			return t('project.table.title')
		case 'Kanban':
			return t('project.kanban.title')
		case 'Calendar':
		case 'Календарь':
			return view.view_kind === 'calendar' ? t('project.calendar.title') : view.title
	}

	return view.title ?? ''
}

function getViewRoute(view: ProjectView) {
	const viewId = view.id ?? 0
	const storedQuery = viewFiltersStore.getViewQuery(viewId)
	return {
		name: 'project.view',
		params: {projectId: props.projectId, viewId},
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
	background: var(--switch-view-active-background);
}

.dropdown-icon {
	font-size: .6rem;
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
