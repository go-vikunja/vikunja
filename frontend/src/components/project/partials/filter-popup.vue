<template>
	<x-button
		v-if="hasFilters"
		variant="secondary"
		@click="clearFilters"
	>
		{{ $t('filters.clear') }}
	</x-button>
	<x-button
		variant="secondary"
		icon="filter"
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
		/>
	</modal>
</template>

<script setup lang="ts">
import {computed, ref, watch} from 'vue'

import Filters from '@/components/project/partials/filters.vue'

import {getDefaultTaskFilterParams, type TaskFilterParams} from '@/services/taskCollection'

const modelValue = defineModel<TaskFilterParams>({})

const value = ref<TaskFilterParams>({})

watch(
	() => modelValue.value,
	(modelValue: TaskFilterParams) => {
		value.value = modelValue
	},
	{immediate: true},
)

function emitChanges(newValue: TaskFilterParams) {
	if (modelValue.value?.filter === newValue.filter && modelValue.value?.s === newValue.s) {
		return
	}

	modelValue.value.filter = newValue.filter
	modelValue.value.s = newValue.s
}

const hasFilters = computed(() => {
	// this.value also contains the page parameter which we don't want to include in filters
	// eslint-disable-next-line no-unused-vars
	const {filter, s} = value.value
	const def = {...getDefaultTaskFilterParams()}

	const params = {filter, s}
	const defaultParams = {
		filter: def.filter,
		s: s ? def.s : undefined,
	}

	return JSON.stringify(params) !== JSON.stringify(defaultParams)
})

const modalOpen = ref(false)

function clearFilters() {
	value.value = {...getDefaultTaskFilterParams()}
}
</script>

<style scoped lang="scss">
.filter-popup {
	margin: 0;

	&.is-open {
		margin: 2rem 0 1rem;
	}
}
</style>
