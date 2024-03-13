<template>
	<x-button
		variant="secondary"
		icon="filter"
		:class="{'has-filters': hasFilters}"
		@click="() => modalOpen = true"
	>
		{{ $t('filters.title') }}
	</x-button>
	<modal
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
			@update:modelValue="emitChanges"
			@showResultsButtonClicked="() => modalOpen = false"
		/>
	</modal>
</template>

<script setup lang="ts">
import {computed, ref, watch} from 'vue'

import Filters from '@/components/project/partials/filters.vue'

import {getDefaultTaskFilterParams, type TaskFilterParams} from '@/services/taskCollection'
import {useRouteQuery} from '@vueuse/router'

const modelValue = defineModel<TaskFilterParams>({})

const value = ref<TaskFilterParams>({})
const filter = useRouteQuery('filter')

watch(
	() => modelValue.value,
	(modelValue: TaskFilterParams) => {
		value.value = modelValue
		if (value.value.filter !== '' && value.value.filter !== getDefaultTaskFilterParams().filter) {
			filter.value = value.value.filter
		}
	},
	{immediate: true},
)

watch(
	() => filter.value,
	val => {
		if (modelValue.value?.filter === val || typeof val === 'undefined') {
			return
		}

		modelValue.value.filter = val
	},
	{immediate: true},
)

function emitChanges(newValue: TaskFilterParams) {
	filter.value = newValue.filter
	if (modelValue.value?.filter === newValue.filter && modelValue.value?.s === newValue.s) {
		return
	}

	modelValue.value.filter = newValue.filter
	modelValue.value.s = newValue.s
}

const hasFilters = computed(() => {
	return value.value.filter !== '' ||
		value.value.s !== ''
})

const modalOpen = ref(false)
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
		top: math.div($filter-bubble-size, -2);
		right: math.div($filter-bubble-size, -2);

		width: $filter-bubble-size;
		height: $filter-bubble-size;
		border-radius: 100%;
		background: var(--primary);
	}
}
</style>
