<template>
	<XButton
		variant="secondary"
		icon="filter"
		:class="{'has-filters': hasFilters}"
		@click="() => modalOpen = true"
	>
		{{ $t('filters.title') }}
	</XButton>
	<Modal
		:enabled="modalOpen"
		transition-name="fade"
		:overflow="true"
		variant="hint-modal"
		@close="() => modalOpen = false"
	>
		<Filters
			ref="filters"
			v-model="value"
			:has-title="true"
			class="filter-popup"
			:change-immediately="false"
			:filter-from-view="filterFromView"
			@showResults="showResults"
		/>
	</Modal>
</template>

<script setup lang="ts">
import {computed, ref, watch} from 'vue'

import Filters from '@/components/project/partials/Filters.vue'

import {type TaskFilterParams} from '@/services/taskCollection'
import {type IProjectView} from '@/modelTypes/IProjectView'
import {type IProject} from '@/modelTypes/IProject'
import {useProjectStore} from '@/stores/projects'

const props = defineProps<{
	modelValue: TaskFilterParams,
	projectId?: IProject['id'],
	viewId?: IProjectView['id'],
}>()

const emit = defineEmits<{
	'update:modelValue': [value: TaskFilterParams]
}>()

const projectStore = useProjectStore()

const value = ref<TaskFilterParams>({})

watch(
	() => props.modelValue,
	(modelValue: TaskFilterParams) => {
		value.value = modelValue
	},
	{
		immediate: true,
		deep: true,
	},
)

const hasFilters = computed(() => {
	return value.value.filter !== '' ||
		value.value.s !== ''
})

const modalOpen = ref(false)

function showResults() {
	emit('update:modelValue', {
		...value.value,
		filter: value.value.filter,
		s: value.value.s,
	})
	modalOpen.value = false
}

const filterFromView = computed(() => {
	if (!props.projectId || !props.viewId) {
		return
	}
	
	const project = projectStore.projects[props.projectId]
	if (!project) {
		return
	}
	const view = project.views.find(v => v.id === props.viewId)
	return view?.filter?.filter
})
</script>

<style scoped lang="scss">
.filter-popup {
	margin: 0;

	&.is-open {
		margin: 2rem 0 1rem;
	}
}

$filter-bubble-size: .75rem;
.has-filters {
	position: relative;

	&::after {
		content: '';
		position: absolute;
		inset-block-start: math.div($filter-bubble-size, -2);
		right: math.div($filter-bubble-size, -2);

		width: $filter-bubble-size;
		height: $filter-bubble-size;
		border-radius: 100%;
		background: var(--primary);
	}
}
</style>
