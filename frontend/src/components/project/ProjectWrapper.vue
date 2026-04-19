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
		>
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

import Message from '@/components/misc/Message.vue'
import CustomTransition from '@/components/misc/CustomTransition.vue'

import {getProjectTitle} from '@/helpers/getProjectTitle'
import {useTitle} from '@/composables/useTitle'

import {useBaseStore} from '@/stores/base'

import type {IProject} from '@/modelTypes/IProject'

const props = defineProps<{
	isLoadingProject: boolean,
	projectId: IProject['id'],
	viewId: number,
}>()

const baseStore = useBaseStore()

const currentProject = computed<IProject>(() => {
	return baseStore.currentProject || {
		id: 0,
		title: '',
		isArchived: false,
		maxPermission: null,
	}
})
useTitle(() => currentProject.value?.id ? getProjectTitle(currentProject.value) : '')
</script>

<style lang="scss" scoped>
.switch-view-container {
	position: relative;
	min-block-size: $switch-view-height;
	margin-block-end: 1rem;
	
	display: flex;
	justify-content: flex-end;
	align-items: center;	
	gap: 1rem;
	
	@media screen and (max-width: $tablet) {
		justify-content: center;
		flex-direction: column;
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
