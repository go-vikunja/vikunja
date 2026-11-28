<template>
	<Card
		class="filters has-overflow"
		:title="hasTitle ? $t('filters.title') : ''"
		role="search"
		:show-close="showClose"
		@close="$emit('close')"
	>
		<FilterEditor
			ref="filterInputRef"
			v-model="params.filter"
			v-model:search="params.s"
			:project-id="projectId"
			class="mbe-2"
			@change="editorChanged"
			@update:search="change('always')"
		>
			<template
				v-if="filterFromView"
				#hint
			>
				<div class="has-text-grey tw:text-sm mbe-3">
					{{ $t('filters.fromView') }}
					<code>{{ filterFromView }}</code><br>
					{{ $t('filters.fromViewBoth') }}
				</div>
			</template>
		</FilterEditor>

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
				@update:modelValue="change('always')"
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
import {useIncludeSubprojects} from '@/composables/useIncludeSubprojects'
import {useProjects} from '@/composables/useProjects'
import {getSavedFilterIdFromProjectId, type ProjectResponse} from '@/client/queries/projects'
import {savedFilterQuery} from '@/client/queries/savedFilters'
import type {ProjectView} from '@/client/generated'
import {hasFilterQuery} from '@/helpers/filters'
import FilterEditor from '@/components/input/filter/FilterEditor.vue'

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

// A subproject's task has no bucket in a kanban view, so the api ignores the flag there.
const showIncludeSubprojectsToggle = computed(() => (props.view?.project_id ?? 0) > 0 && props.view?.view_kind !== 'kanban')

const {
	includeSubprojects,
	fromView: includeSubprojectsFromView,
	overridesView,
} = useIncludeSubprojects(() => props.view)

const hasActiveFilters = computed(() => params.value.filter !== ''
	|| params.value.s !== ''
	|| params.value.filter_include_nulls
	|| overridesView.value)

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
		// params.filter holds the api form, which is how saved filters are stored,
		// so it needs no transform. change() moves a plain search over to `s`.
		params.value.filter = savedFilter.filters.filter || savedFilter.filters.s || ''
		params.value.s = ''
		params.value.filter_include_nulls = !!savedFilter.filters.filter_include_nulls
		change('always')
	} finally {
		loadingSavedFilterProjectId.value = 0
	}
}

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

function change(event: 'modelValue' | 'always') {
	if (event === 'modelValue' && !props.changeImmediately) {
		return
	}

	let filter = params.value.filter ?? ''
	let s = params.value.s ?? ''
	if (filter !== '' && !hasFilterQuery(filter)) {
		s = filter
		filter = ''
	}

	const newParams = {
		...props.modelValue,
		...params.value,
		filter,
		s,
	}

	if (JSON.stringify(props.modelValue) === JSON.stringify(newParams)) {
		return
	}

	emit('update:modelValue', newParams)
}

// The Advanced input shows `filter || s`, so it owns both.
function editorChanged(mode: 'simple' | 'advanced') {
	if (mode === 'advanced') {
		params.value.s = ''
	}
	change(mode === 'simple' ? 'always' : 'modelValue')
}

function changeAndEmitButton() {
	change('always')
	emit('showResults')
}

function clearFiltersAndEmit() {
	params.value.filter = ''
	params.value.s = ''
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
