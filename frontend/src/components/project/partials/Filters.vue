<template>
	<Card
		class="filters has-overflow"
		:title="hasTitle ? $t('filters.title') : ''"
		role="search"
		:show-close="showClose"
		@close="$emit('close')"
	>
		<FilterInput
			ref="filterInputRef"
			v-model="filterQuery"
			:project-id="projectId"
			class="mbe-2"
			@update:modelValue="() => change('modelValue')"
		/>
		<div
			v-if="filterFromView"
			class="tw:text-sm mbe-2"
		>
			{{ $t('filters.fromView') }}
			<code>{{ filterFromView }}</code><br>
			{{ $t('filters.fromViewBoth') }}
		</div>

		<div
			v-if="showSavedFilters && savedFilters.length > 0"
			class="mbe-2"
		>
			<p class="tw:text-sm mbe-1">
				{{ $t('filters.savedFilters') }}
			</p>
			<XButton
				v-for="savedFilter in savedFilters"
				:key="savedFilter.id"
				variant="secondary"
				class="mie-2 mbe-1"
				icon="filter"
				:loading="loadingSavedFilterProjectId === savedFilter.id"
				:disabled="loadingSavedFilterProjectId !== 0"
				@click.prevent.stop="selectSavedFilter(savedFilter)"
			>
				{{ savedFilter.title }}
			</XButton>
		</div>

		<div class="field is-flex is-flex-direction-column">
			<FancyCheckbox
				v-model="params.filter_include_nulls"
				@change="() => change('always')"
			>
				{{ $t('filters.attributes.includeNulls') }}
			</FancyCheckbox>
			<FancyCheckbox
				v-if="showIncludeSubprojectsToggle"
				v-model="includeSubprojects"
				v-tooltip="includeSubprojectsFromView
					? $t('project.views.includeSubprojectsFromView')
					: $t('project.views.includeSubprojectsHint')"
				:disabled="includeSubprojectsFromView"
			>
				{{ $t('project.views.includeSubprojects') }}
			</FancyCheckbox>
		</div>

		<FilterInputDocs />

		<template
			v-if="hasFooter"
			#footer
		>
			<XButton
				variant="secondary"
				class="mie-2"
				:disabled="!hasActiveFilters"
				@click.prevent.stop="clearFiltersAndEmit"
			>
				{{ $t('filters.clear') }}
			</XButton>
			<XButton
				variant="primary"
				@click.prevent.stop="changeAndEmitButton"
			>
				{{ $t('filters.showResults') }}
			</XButton>
		</template>
	</Card>
</template>

<script setup lang="ts" generic="T extends EditableTaskCollection">
import {computed, ref, watch, type DeepReadonly} from 'vue'
import {useQueryClient} from '@tanstack/vue-query'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import {useRoute} from 'vue-router'
import type {EditableTaskCollection} from '@/types/EditableTaskCollection'
import {useLabels} from '@/composables/useLabels'
import {useProjects} from '@/composables/useProjects'
import {useIncludeSubprojects} from '@/composables/useIncludeSubprojects'
import {getSavedFilterIdFromProjectId, type ProjectResponse} from '@/client/queries/projects'
import {savedFilterQuery} from '@/client/queries/savedFilters'
import type {ProjectView} from '@/client/generated'
import {
	hasFilterQuery,
	transformFilterStringForApi,
	transformFilterStringFromApi,
} from '@/helpers/filters'
import FilterInputDocs from '@/components/input/filter/FilterInputDocs.vue'
import FilterInput from '@/components/input/filter/FilterInput.vue'

const props = withDefaults(defineProps<{
	modelValue: T,
	hasTitle?: boolean,
	hasFooter?: boolean,
	changeImmediately?: boolean,
	filterFromView?: string,
	showClose?: boolean,
	view?: ProjectView | DeepReadonly<ProjectView>,
	// Cross-project saved filters only make sense in cross-project contexts
	// (e.g. the planner), so single-project views keep them hidden.
	showSavedFilters?: boolean,
}>(), {
	hasTitle: false,
	hasFooter: true,
	changeImmediately: false,
	filterFromView: undefined,
	showClose: false,
	view: undefined,
	showSavedFilters: false,
})

const emit = defineEmits<{
	'update:modelValue': [value: T],
	'showResults': [],
	'close': [],
}>()

const route = useRoute()
const projectId = computed(() => {
	if (typeof route.name === 'string' && route.name.startsWith('project.')) {
		return Number(route.params.projectId)
	}

	return undefined
})

const params = ref<EditableTaskCollection>({
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

// A subproject's task has no bucket in a kanban view, so the api ignores the flag there.
const showIncludeSubprojectsToggle = computed(() => (props.view?.project_id ?? 0) > 0 && props.view?.view_kind !== 'kanban')

const {
	includeSubprojects,
	fromView: includeSubprojectsFromView,
	overridesView,
} = useIncludeSubprojects(() => props.view)

const hasActiveFilters = computed(() => filterQuery.value !== '' || params.value.filter_include_nulls || overridesView.value)

const {getLabelByExactTitle, getLabelById} = useLabels()
const projectList = useProjects()

const filterInputRef = ref()

const queryClient = useQueryClient()

const savedFilters = computed(() => projectList.savedFilterProjects)
const loadingSavedFilterProjectId = ref(0)

async function selectSavedFilter(project: ProjectResponse) {
	loadingSavedFilterProjectId.value = project.id
	try {
		const savedFilter = await queryClient.ensureQueryData(
			savedFilterQuery(getSavedFilterIdFromProjectId(project.id)),
		)
		filterQuery.value = transformFilterStringFromApi(
			savedFilter.filters.filter || savedFilter.filters.s || '',
			labelId => getLabelById(labelId)?.title || null,
			filterProjectId => projectList.projects[filterProjectId]?.title || null,
		)
		params.value.filter_include_nulls = !!savedFilter.filters.filter_include_nulls
		change('always')
	} finally {
		loadingSavedFilterProjectId.value = 0
	}
}

// Using watchDebounced to prevent the filter re-triggering itself.
watch(
	() => props.modelValue,
	(value: T) => {
		params.value = {...value}
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
		labelTitle => getLabelByExactTitle(labelTitle)?.id || null,
		projectTitle => {
			const found = projectList.findProjectByExactname(projectTitle)
			return found?.id || null
		},
	)

	let s = ''

	// When the filter does not contain any filter tokens, assume a simple search and redirect the input
	if (!hasFilterQuery(filter)) {
		s = filter
	}

	const newParams = {
		...props.modelValue,
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
	change('always')
	emit('showResults')
}

function clearFiltersAndEmit() {
	filterQuery.value = ''
	params.value.filter_include_nulls = false
	if (overridesView.value) {
		includeSubprojects.value = false
	}
	changeAndEmitButton()
}

function focusFilterInput() {
	filterInputRef.value?.focus()
}

defineExpose({
	focusFilterInput,
})
</script>
