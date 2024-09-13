<template>
	<Card
		class="filters has-overflow"
		:title="hasTitle ? $t('filters.title') : ''"
		role="search"
	>
		<FilterInput
			v-model="filterQuery"
			:project-id="projectId"
			@update:modelValue="() => change('modelValue')"
			@blur="() => change('blur')"
		/>

		<div class="field is-flex is-flex-direction-column">
			<FancyCheckbox
				v-model="params.filter_include_nulls"
				@change="() => change('always')"
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
	changeImmediately?: boolean,
}>(), {
	hasTitle: false,
	hasFooter: true,
	changeImmediately: false,
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
			labelId => labelStore.getLabelById(labelId)?.title || null,
			projectId => projectStore.projects[projectId]?.title || null,
		)
		params.value = val
	},
	{
		immediate: true,
		deep: true,
	},
)

function change(event: 'blur' | 'modelValue' | 'always') {
	if (event !== 'always') {
		// The filter edit setting needs to save immediately, but the filter query edit in project views should 
		// only change on blur, or it will show the filter replaced for api when the query is not yet complete. 
		// This is highly confusing UX, hence we want to avoid that.
		// The approach taken here allows us to either toggle on blur or immediately, depending on the prop
		// value provided. This probably is a hacky way to do this, but it is also the most effective.
		if (props.changeImmediately && event === 'blur') {
			return
		}

		if (!props.changeImmediately && event === 'modelValue') {
			return
		}
	}

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
