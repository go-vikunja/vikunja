<template>
	<Card
		class="filters has-overflow"
		:title="hasTitle ? $t('filters.title') : ''"
		role="search"
	>
		<FilterInput
			v-model="filterQuery"
			:project-id="projectId"
			@blur="change()"
		/>

		<div class="field is-flex is-flex-direction-column">
			<FancyCheckbox
				v-model="params.filter_include_nulls"
				@blur="change()"
			>
				{{ $t('filters.attributes.includeNulls') }}
			</FancyCheckbox>
		</div>

		<FilterInputDocs />

		<template
			v-if="hasFooter"
			#footer
		>
			<x-button
				variant="secondary"
				class="mr-2"
				:disabled="filterQuery === ''"
				@click.prevent.stop="clearFiltersAndEmit"
			>
				{{ $t('filters.clear') }}
			</x-button>
			<x-button
				variant="primary"
				@click.prevent.stop="changeAndEmitButton"
			>
				{{ $t('filters.showResults') }}
			</x-button>
		</template>
	</Card>
</template>

<script lang="ts">
export const ALPHABETICAL_SORT = 'title'
</script>

<script setup lang="ts">
import {computed, ref, watch} from 'vue'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import FilterInput from '@/components/project/partials/FilterInput.vue'
import {useRoute} from 'vue-router'
import type {TaskFilterParams} from '@/services/taskCollection'
import {useLabelStore} from '@/stores/labels'
import {useProjectStore} from '@/stores/projects'
import {FILTER_OPERATORS, transformFilterStringForApi, transformFilterStringFromApi} from '@/helpers/filters'
import FilterInputDocs from '@/components/project/partials/FilterInputDocs.vue'

const props = withDefaults(defineProps<{
	modelValue: TaskFilterParams,
	hasTitle?: boolean,
	hasFooter?: boolean,
}>(), {
	hasTitle: false,
	hasFooter: true,
})

const emit = defineEmits<{
	'update:modelValue': [value: TaskFilterParams],
	'showResults': [],
}>()

const route = useRoute()
const projectId = computed(() => {
	if (route.name?.startsWith('project.')) {
		return Number(route.params.projectId)
	}

	return undefined
})

const params = ref<TaskFilterParams>({
	sort_by: [],
	order_by: [],
	filter: '',
	filter_include_nulls: false,
	s: '',
})

const filterQuery = ref('')
watch(
	() => [params.value.filter, params.value.s],
	() => {
		const filter = params.value.filter || ''
		const s = params.value.s || ''
		filterQuery.value = filter || s
	},
)

const labelStore = useLabelStore()
const projectStore = useProjectStore()

// Using watchDebounced to prevent the filter re-triggering itself.
watch(
	() => props.modelValue,
	(value: TaskFilterParams) => {
		const val = {...value}
		val.filter = transformFilterStringFromApi(
			val?.filter || '',
			labelId => labelStore.getLabelById(labelId)?.title,
			projectId => projectStore.projects[projectId]?.title || null,
		)
		params.value = val
	},
	{
		immediate: true,
		deep: true,
	},
)

function change() {
	const filter = transformFilterStringForApi(
		filterQuery.value,
		labelTitle => labelStore.getLabelByExactTitle(labelTitle)?.id || null,
		projectTitle => {
			const found = projectStore.findProjectByExactname(projectTitle)
			return found?.id || null
		},
	)

	let s = ''

	// When the filter does not contain any filter tokens, assume a simple search and redirect the input
	const hasFilterQueries = FILTER_OPERATORS.find(o => filter.includes(o)) || false
	if (!hasFilterQueries) {
		s = filter
	}

	const newParams = {
		...params.value,
		filter: s === '' ? filter : '',
		s,
	}

	if (JSON.stringify(props.modelValue) === JSON.stringify(newParams)) {
		return
	}

	emit('update:modelValue', newParams)
}

function changeAndEmitButton() {
	change()
	emit('showResults')
}

function clearFiltersAndEmit() {
	filterQuery.value = ''
	changeAndEmitButton()
}
</script>
