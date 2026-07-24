<script setup lang="ts">
import {computed, onBeforeMount, ref} from 'vue'
import {useI18n} from 'vue-i18n'

import type {IProjectView} from '@/modelTypes/IProjectView'
import type {IFilters} from '@/modelTypes/ISavedFilter'

import {hasFilterQuery, transformFilterStringForApi, transformFilterStringFromApi} from '@/helpers/filters'
import {
	decodeSortSelection,
	encodeSortSelection,
	sortByFromViewFilter,
	viewFilterSortFromSortBy,
} from '@/helpers/viewSort'
import {useLabelStore} from '@/stores/labels'
import {useProjectStore} from '@/stores/projects'

import XButton from '@/components/input/Button.vue'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import FilterInputDocs from '@/components/input/filter/FilterInputDocs.vue'
import FilterInput from '@/components/input/filter/FilterInput.vue'
import FormField from '@/components/input/FormField.vue'

const props = withDefaults(defineProps<{
	modelValue: IProjectView,
	loading?: boolean,
	showSaveButtons?: boolean,
}>(), {
	loading: false,
	showSaveButtons: false,
})

const emit = defineEmits<{
	'update:modelValue': [value: IProjectView],
	'cancel': [],
}>()

const view = ref<IProjectView>()
const {t} = useI18n({useScope: 'global'})

const labelStore = useLabelStore()
const projectStore = useProjectStore()

const MANUAL_SORT = 'position:asc'

const sortOptions = computed(() => {
	const manual = {value: MANUAL_SORT, label: t('sorting.manually')}
	const rest = [
		{value: 'title:asc', label: t('sorting.options.titleAsc')},
		{value: 'title:desc', label: t('sorting.options.titleDesc')},
		{value: 'priority:desc', label: t('sorting.options.priorityDesc')},
		{value: 'priority:asc', label: t('sorting.options.priorityAsc')},
		{value: 'due_date:asc', label: t('sorting.options.dueDateAsc')},
		{value: 'due_date:desc', label: t('sorting.options.dueDateDesc')},
		{value: 'start_date:asc', label: t('sorting.options.startDateAsc')},
		{value: 'start_date:desc', label: t('sorting.options.startDateDesc')},
		{value: 'end_date:asc', label: t('sorting.options.endDateAsc')},
		{value: 'end_date:desc', label: t('sorting.options.endDateDesc')},
		{value: 'percent_done:desc', label: t('sorting.options.percentDoneDesc')},
		{value: 'percent_done:asc', label: t('sorting.options.percentDoneAsc')},
		{value: 'created:desc', label: t('sorting.options.createdDesc')},
		{value: 'created:asc', label: t('sorting.options.createdAsc')},
		{value: 'updated:desc', label: t('sorting.options.updatedDesc')},
		{value: 'updated:asc', label: t('sorting.options.updatedAsc')},
	].sort((a, b) => a.label.localeCompare(b.label))

	return [manual, ...rest]
})

const showDefaultSort = computed(() =>
	view.value?.viewKind === 'list' || view.value?.viewKind === 'table',
)

const defaultSortSelection = computed({
	get() {
		return encodeSortSelection(sortByFromViewFilter(view.value?.filter), MANUAL_SORT)
	},
	set(value: string) {
		if (!view.value?.filter) {
			return
		}
		const sort = decodeSortSelection(value)
		const {sort_by, order_by} = viewFilterSortFromSortBy(sort)
		view.value.filter.sort_by = sort_by
		view.value.filter.order_by = order_by
	},
})

function readSortArrays(filterInput: IFilters): Pick<IFilters, 'sort_by' | 'order_by'> {
	const raw = filterInput as IFilters & {
		sortBy?: IFilters['sort_by']
		orderBy?: IFilters['order_by']
	}
	return {
		sort_by: raw.sort_by ?? raw.sortBy ?? [],
		order_by: raw.order_by ?? raw.orderBy ?? [],
	}
}

onBeforeMount(() => {
	const transformFilterFromApi = (filterInput: IFilters): IFilters => {
		const filterString = transformFilterStringFromApi(
			filterInput.filter,
			labelId => labelStore.getLabelById(labelId)?.title || null,
			projectId => projectStore.projects[projectId]?.title || null,
		)
		
		const {sort_by, order_by} = readSortArrays(filterInput)
		const filter: IFilters = {
			filter: '',
			s: '',
			sort_by,
			order_by,
			filter_include_nulls: false,
		}
		if (hasFilterQuery(filterString)) {
			filter.filter = filterString
		} else {
			filter.s = filterString
		}
		
		if (filter.s === '') {
			filter.s = filterInput.s
		}
		
		if (filter.filter === '') {
			filter.filter = filter.s
		}

		// AbstractModel.assignData() runs objectToCamelCase recursively on all
		// nested objects, which converts filter_include_nulls to filterIncludeNulls
		// inside the filter object. IFilters intentionally uses snake_case keys to
		// match the API query param format. We check both key forms here to handle
		// data coming from either the API response (camelCased by assignData) or
		// from a freshly constructed filter object (snake_case).
		filter.filter_include_nulls = filterInput.filter_include_nulls
			?? (filterInput as Record<string, unknown>).filterIncludeNulls as boolean
			?? false

		return filter
	}

	const transformed = {
		...props.modelValue,
		filter: transformFilterFromApi(props.modelValue.filter ?? {
			sort_by: [],
			order_by: [],
			filter: '',
			filter_include_nulls: false,
			s: '',
		}),
		bucketConfiguration: (props.modelValue.bucketConfiguration || []).map(bc => ({
			title: bc.title,
			filter: transformFilterFromApi(bc.filter),
		})),
	}

	if (JSON.stringify(view.value) !== JSON.stringify(transformed)) {
		view.value = transformed
	}
})

function save() {
	const transformFilterForApi = (filterInput: IFilters): IFilters => {
		const filterString = transformFilterStringForApi(
			filterInput?.filter || '',
			labelTitle => labelStore.getLabelByExactTitle(labelTitle)?.id || null,
			projectTitle => {
				const found = projectStore.findProjectByExactname(projectTitle)
				return found?.id || null
			},
		)
		const {sort_by, order_by} = readSortArrays(filterInput || {
			sort_by: [],
			order_by: [],
			filter: '',
			filter_include_nulls: false,
			s: '',
		})
		const filter: IFilters = {
			filter_include_nulls: filterInput?.filter_include_nulls ?? false,
			sort_by,
			order_by,
			filter: '',
			s: '',
		}
		if (hasFilterQuery(filterString)) {
			filter.filter = filterString
		} else {
			filter.s = filterString
		}

		return filter
	}

	emit('update:modelValue', {
		...view.value,
		filter: transformFilterForApi(view.value?.filter as IFilters),
		bucketConfiguration: view.value?.bucketConfiguration.map(bc => ({
			title: bc.title,
			filter: transformFilterForApi(bc.filter),
		})),
	})
}

const titleValid = ref(true)

function validateTitle() {
	titleValid.value = view.value?.title !== ''
}

function handleBubbleSave() {
	if (props.showSaveButtons) {
		return
	}

	save()
}
</script>

<template>
	<form
		@focusout="handleBubbleSave"
		@submit.prevent="save"
	>
		<FormField
			id="title"
			v-model="view.title"
			v-focus
			:label="$t('project.views.title')"
			:placeholder="$t('project.share.links.namePlaceholder')"
			:error="titleValid ? null : $t('project.views.titleRequired')"
			@blur="validateTitle"
		/>

		<FormField :label="$t('project.views.kind')">
			<template #default="{ id }">
				<div class="select">
					<select
						:id="id"
						v-model="view.viewKind"
					>
						<option value="list">
							{{ $t('project.list.title') }}
						</option>
						<option value="gantt">
							{{ $t('project.gantt.title') }}
						</option>
						<option value="table">
							{{ $t('project.table.title') }}
						</option>
						<option value="kanban">
							{{ $t('project.kanban.title') }}
						</option>
					</select>
				</div>
			</template>
		</FormField>

		<label
			class="label"
			for="filter"
		>
			{{ $t('project.views.filter') }}
		</label>
		<FilterInput
			id="filter"
			v-model="view.filter.filter"
			:project-id="view.projectId"
			class="mbe-1"
		/>

		<div class="is-size-7 mbe-2">
			<FilterInputDocs />
		</div>

		<div class="field mbe-3">
			<FancyCheckbox
				v-model="view.filter.filter_include_nulls"
			>
				{{ $t('filters.attributes.includeNulls') }}
			</FancyCheckbox>
		</div>

		<div
			v-if="showDefaultSort"
			class="field mbe-3"
		>
			<label
				class="label"
				for="defaultSort"
			>
				{{ $t('project.views.defaultSort') }}
			</label>
			<p class="is-size-7 has-text-grey mbe-2">
				{{ $t('project.views.defaultSortDescription') }}
			</p>
			<div class="select is-fullwidth">
				<select
					id="defaultSort"
					v-model="defaultSortSelection"
				>
					<option
						v-for="o in sortOptions"
						:key="o.value"
						:value="o.value"
					>
						{{ o.label }}
					</option>
				</select>
			</div>
		</div>

		<div
			v-if="view.viewKind === 'kanban'"
			class="field"
		>
			<label
				class="label"
				for="configMode"
			>
				{{ $t('project.views.bucketConfigMode') }}
			</label>
			<div
				id="configMode"
				class="control"
			>
				<label class="radio">
					<input
						v-model="view.bucketConfigurationMode"
						type="radio"
						name="configMode"
						value="manual"
					>
					{{ $t('project.views.bucketConfigManual') }}
				</label>
				<label class="radio">
					<input
						v-model="view.bucketConfigurationMode"
						type="radio"
						name="configMode"
						value="filter"
					>
					{{ $t('project.views.filter') }}
				</label>
			</div>
		</div>

		<div
			v-if="view.viewKind === 'kanban' && view.bucketConfigurationMode === 'filter'"
			class="field"
		>
			<label class="label">
				{{ $t('project.views.bucketConfig') }}
			</label>
			<div class="control">
				<div
					v-for="(b, index) in view.bucketConfiguration"
					:key="'bucket_'+index"
					class="filter-bucket"
				>
					<button
						class="is-danger"
						@click.prevent="() => view.bucketConfiguration.splice(index, 1)"
					>
						<Icon icon="trash-alt" />
					</button>
					<div class="filter-bucket-form">
						<FormField
							:id="'bucket_'+index+'_title'"
							v-model="view.bucketConfiguration[index].title"
							:label="$t('project.views.title')"
							:placeholder="$t('project.share.links.namePlaceholder')"
						/>

						<FilterInput
							v-model="view.bucketConfiguration[index].filter.filter"
							:project-id="view.projectId"
							:input-label="$t('project.views.filter')"
							class="mbe-2"
						/>

						<div class="is-size-7 mbe-2">
							<FilterInputDocs />
						</div>

						<div class="field mbe-3">
							<FancyCheckbox
								v-model="view.bucketConfiguration[index].filter.filter_include_nulls"
							>
								{{ $t('filters.attributes.includeNulls') }}
							</FancyCheckbox>
						</div>
					</div>
				</div>
				<div class="is-flex is-justify-content-end">
					<XButton
						variant="secondary"
						icon="plus"
						@click="() => view.bucketConfiguration.push({title: '', filter: {filter: '', filter_include_nulls: false, sort_by: [], order_by: [], s: ''}})"
					>
						{{ $t('project.kanban.addBucket') }}
					</XButton>
				</div>
			</div>
		</div>
		<div
			v-if="showSaveButtons"
			class="is-flex is-justify-content-end"
		>
			<XButton
				variant="tertiary"
				class="mie-2"
				@click="emit('cancel')"
			>
				{{ $t('misc.cancel') }}
			</XButton>
			<XButton
				:loading="loading"
				type="submit"
			>
				{{ $t('misc.save') }}
			</XButton>
		</div>
	</form>
</template>

<style scoped lang="scss">
.filter-bucket {
	display: flex;

	button {
		background: transparent;
		border: none;
		color: var(--danger);
		padding-inline-end: .75rem;
		cursor: pointer;
	}

	&-form {
		margin-block-end: .5rem;
		padding: .5rem;
		border: 1px solid var(--grey-200);
		border-radius: $radius;
		inline-size: 100%;
	}
}

// Ported from bulma-css-variables/sass/form/checkbox-radio.sass
// (the %checkbox-radio placeholder plus the .radio + .radio sibling rule),
// scoped to this component so we can drop the global Bulma import.
label.radio {
	cursor: pointer;
	display: inline-block;
	line-height: 1.25;
	position: relative;

	input {
		cursor: pointer;
	}

	&:hover {
		color: var(--input-hover-color);
	}

	&[disabled],
	input[disabled] {
		color: var(--input-disabled-color);
		cursor: not-allowed;
	}

	& + .radio {
		margin-inline-start: .5em;
	}
}
</style>
