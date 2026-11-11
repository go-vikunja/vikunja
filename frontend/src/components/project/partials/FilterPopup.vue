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
		:overflow="true"
		variant="hint-modal"
		:aria-label="$t('filters.title')"
		@close="() => modalOpen = false"
	>
		<Filters
			ref="filtersRef"
			v-model="value"
			v-model:include-subprojects="includeSubprojects"
			:has-title="true"
			class="filter-popup"
			:change-immediately="false"
			:filter-from-view="filterFromView"
			:show-include-subprojects-toggle="supportsIncludeSubprojects"
			:include-subprojects-from-view="includeSubprojectsFromView"
			show-close
			@close="modalOpen = false"
			@showResults="showResults"
		/>
	</Modal>
</template>

<script setup lang="ts">
import {computed, ref, watch, nextTick, type DeepReadonly} from 'vue'

import Filters from '@/components/project/partials/Filters.vue'

import type {EditableTaskCollection} from '@/types/EditableTaskCollection'
import {type TaskFilterParams} from '@/client/queries/tasks'
import type {ProjectView} from '@/client/generated'
import {useIncludeSubprojects} from '@/composables/useIncludeSubprojects'

const props = defineProps<{
	modelValue: TaskFilterParams,
	view?: ProjectView | DeepReadonly<ProjectView>,
}>()

const emit = defineEmits<{
	'update:modelValue': [value: TaskFilterParams]
}>()

const value = ref<EditableTaskCollection>({
	sort_by: [],
	order_by: [],
	filter: '',
	filter_include_nulls: false,
	s: '',
})
const filtersRef = ref()

watch(
	() => props.modelValue,
	(modelValue: TaskFilterParams) => {
		value.value = {
			sort_by: modelValue.sort_by ?? [],
			order_by: modelValue.order_by ?? [],
			filter: modelValue.filter ?? '',
			filter_include_nulls: modelValue.filter_include_nulls ?? false,
			s: modelValue.q ?? '',
		}
	},
	{
		immediate: true,
		deep: true,
	},
)

const hasFilters = computed(() => {
	return value.value.filter !== '' ||
		value.value.s !== '' ||
		(includeSubprojects.value && !includeSubprojectsFromView.value)
})

const modalOpen = ref(false)

// Auto-focus filter input when modal opens
watch(modalOpen, (isOpen) => {
	if (isOpen) {
		nextTick(() => {
			filtersRef.value?.focusFilterInput()
		})
	}
})

function showResults() {
	const {s, ...rest} = value.value
	emit('update:modelValue', {...props.modelValue, ...rest, q: s})
	modalOpen.value = false
}

// A subproject's task has no bucket in a kanban view, so the api ignores the flag there.
const supportsIncludeSubprojects = computed(() => (props.view?.project_id ?? 0) > 0 && props.view?.view_kind !== 'kanban')

const includeSubprojects = useIncludeSubprojects(() => props.view)

// The api ors the view's own setting in, so the query param cannot turn it back off.
const includeSubprojectsFromView = computed(() => props.view?.filter?.include_subprojects ?? false)

const filterFromView = computed(() => props.view?.filter?.filter)
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
		inset-inline-end: math.div($filter-bubble-size, -2);

		inline-size: $filter-bubble-size;
		block-size: $filter-bubble-size;
		border-radius: 100%;
		background: var(--primary);
	}
}
</style>
