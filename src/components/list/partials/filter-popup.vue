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
		@close="() => modalOpen = false"
		:enabled="modalOpen"
		transition-name="fade"
		:overflow="true"
		variant="hint-modal"
	>
		<filters
			:has-title="true"
			v-model="value"
			ref="filters"
			class="filter-popup"
			:class="{'is-open': isOpen}"
		/>
	</modal>
</template>

<script lang="ts">
import {defineComponent, ref} from 'vue'

import Filters from '@/components/list/partials/filters.vue'

import {getDefaultParams} from '@/composables/taskList'

export default defineComponent({
	name: 'filter-popup',
	components: {
		Filters,
	},
	props: {
		modelValue: {
			required: true,
		},
	},
	emits: ['update:modelValue'],
	computed: {
		value: {
			get() {
				return this.modelValue
			},
			set(value) {
				this.$emit('update:modelValue', value)
			},
		},
		hasFilters() {
			// this.value also contains the page parameter which we don't want to include in filters
			// eslint-disable-next-line no-unused-vars
			const {filter_by, filter_value, filter_comparator, filter_concat, s} = this.value
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
		},
	},
	watch: {
		modelValue: {
			handler(value) {
				this.value = value
			},
			immediate: true,
		},
	},
	setup() {
		const modalOpen = ref(false)

		return {
			modalOpen,
		}
	},
	methods: {
		clearFilters() {
			this.value = {...getDefaultParams()}
		},
	},
})
</script>

<style scoped lang="scss">
.filter-popup {
	margin: 0;

	&.is-open {
		margin: 2rem 0 1rem;
	}
}
</style>
