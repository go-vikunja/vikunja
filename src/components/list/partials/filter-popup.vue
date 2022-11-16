<template>
	<x-button
		v-if="hasFilters"
		variant="secondary"
		@click="clearFilters"
	>
		{{ $t('filters.clear') }}
	</x-button>
	<x-button
		@click="() => modalOpen = true"
		variant="secondary"
		icon="filter"
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
		<filters
			:has-title="true"
			v-model="value"
			ref="filters"
			class="filter-popup"
		/>
	</modal>
</template>

<script setup lang="ts">
import {computed, ref, watch} from 'vue'

import Filters from '@/components/list/partials/filters.vue'

import {getDefaultParams} from '@/composables/useTaskList'

const	props = defineProps({
	modelValue: {
		required: true,
	},
})
const emit = defineEmits(['update:modelValue'])

const value = computed({
	get() {
		return props.modelValue
	},
	set(value) {
		emit('update:modelValue', value)
	},
})

watch(
	() => props.modelValue,
	(modelValue) => {
		value.value = modelValue
	},
	{immediate: true},
)
		
const hasFilters = computed(() => {
	// this.value also contains the page parameter which we don't want to include in filters
	// eslint-disable-next-line no-unused-vars
	const {filter_by, filter_value, filter_comparator, filter_concat, s} = value.value
	const def = {...getDefaultParams()}

	const params = {filter_by, filter_value, filter_comparator, filter_concat, s}
	const defaultParams = {
		filter_by: def.filter_by,
		filter_value: def.filter_value,
		filter_comparator: def.filter_comparator,
		filter_concat: def.filter_concat,
		s: s ? def.s : undefined,
	}

	return JSON.stringify(params) !== JSON.stringify(defaultParams)
})

const modalOpen = ref(false)

function clearFilters() {
	value.value = {...getDefaultParams()}
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
