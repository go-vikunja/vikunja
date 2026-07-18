<script setup lang="ts">
import {computed, onBeforeMount, ref} from 'vue'

import type {IProjectView} from '@/modelTypes/IProjectView'
import type {IFilters} from '@/modelTypes/ISavedFilter'

import {hasFilterQuery, transformFilterStringForApi, transformFilterStringFromApi} from '@/helpers/filters'
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

const labelStore = useLabelStore()
const projectStore = useProjectStore()

onBeforeMount(() => {
	const transformFilterFromApi = (filterInput: IFilters): IFilter => {
		const filterString = transformFilterStringFromApi(
			filterInput.filter,
			labelId => labelStore.getLabelById(labelId)?.title || null,
			projectId => projectStore.projects[projectId]?.title || null,
		)
		
		const filter: IFilters = {
			filter: '',
			s: '',
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
		filter: transformFilterFromApi(props.modelValue.filter),
		bucketConfiguration: props.modelValue.bucketConfiguration.map(bc => ({
			title: bc.title,
			filter: transformFilterFromApi(bc.filter),
		})),
		// modelValue can originate from the (readonly) project store; clone these
		// so in-place edits (v-model on array indices, push/splice) aren't
		// silently blocked by Vue's readonly guard.
		bucketSortBy: [...(props.modelValue.bucketSortBy || [])],
		bucketSortOrder: [...(props.modelValue.bucketSortOrder || [])],
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
		const filter: IFilters = {
			filter_include_nulls: filterInput?.filter_include_nulls ?? false,
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
		filter: transformFilterForApi(view.value?.filter),
		bucketConfiguration: view.value?.bucketConfiguration.map(bc => ({
			title: bc.title,
			filter: transformFilterForApi(bc.filter),
		})),
	})
}

const BUCKET_SORT_FIELDS = ['priority', 'due_date', 'created', 'updated', 'title']

const BUCKET_SORT_FIELD_LABEL_KEYS: Record<string, string> = {
	priority: 'project.kanban.sortPriority',
	due_date: 'project.kanban.sortDueDate',
	created: 'project.kanban.sortCreated',
	updated: 'project.kanban.sortUpdated',
	title: 'project.kanban.sortTitle',
}

const BUCKET_SORT_DIRECTION_LABEL_KEYS: Record<string, {asc: string, desc: string}> = {
	priority: {asc: 'project.kanban.sortDirections.priority.asc', desc: 'project.kanban.sortDirections.priority.desc'},
	due_date: {asc: 'project.kanban.sortDirections.dueDate.asc', desc: 'project.kanban.sortDirections.dueDate.desc'},
	created: {asc: 'project.kanban.sortDirections.created.asc', desc: 'project.kanban.sortDirections.created.desc'},
	updated: {asc: 'project.kanban.sortDirections.updated.asc', desc: 'project.kanban.sortDirections.updated.desc'},
	title: {asc: 'project.kanban.sortDirections.title.asc', desc: 'project.kanban.sortDirections.title.desc'},
}

const BUCKET_SORT_DEFAULT_ORDER: Record<string, 'asc' | 'desc'> = {
	priority: 'desc',
	due_date: 'asc',
	created: 'asc',
	updated: 'asc',
	title: 'asc',
}

function bucketSortFieldLabelKey(field: string) {
	return BUCKET_SORT_FIELD_LABEL_KEYS[field]
}

function bucketSortDirectionLabelKey(field: string, direction: string) {
	return BUCKET_SORT_DIRECTION_LABEL_KEYS[field]?.[direction as 'asc' | 'desc']
}

function availableBucketSortFields(index: number) {
	const usedElsewhere = view.value?.bucketSortBy.filter((f, i) => i !== index) || []
	return BUCKET_SORT_FIELDS.filter(f => f === view.value?.bucketSortBy[index] || !usedElsewhere.includes(f))
}

const canAddBucketSort = computed(() => (view.value?.bucketSortBy.length || 0) < BUCKET_SORT_FIELDS.length)

function addBucketSort() {
	if (!view.value) {
		return
	}

	const used = new Set(view.value.bucketSortBy)
	const nextField = BUCKET_SORT_FIELDS.find(f => !used.has(f))
	if (!nextField) {
		return
	}

	// view.value can originate from the (readonly) project store; reassign
	// new arrays instead of mutating in place so the write isn't silently
	// blocked by Vue's readonly guard.
	view.value.bucketSortBy = [...view.value.bucketSortBy, nextField]
	view.value.bucketSortOrder = [...view.value.bucketSortOrder, BUCKET_SORT_DEFAULT_ORDER[nextField] || 'asc']
}

function changeBucketSortField(index: number, field: string) {
	if (!view.value) {
		return
	}

	const by = [...view.value.bucketSortBy]
	const order = [...view.value.bucketSortOrder]
	by[index] = field
	order[index] = BUCKET_SORT_DEFAULT_ORDER[field] || 'asc'
	view.value.bucketSortBy = by
	view.value.bucketSortOrder = order
}

function removeBucketSort(index: number) {
	if (!view.value) {
		return
	}

	view.value.bucketSortBy = view.value.bucketSortBy.filter((_, i) => i !== index)
	view.value.bucketSortOrder = view.value.bucketSortOrder.filter((_, i) => i !== index)
}

function moveBucketSort(index: number, delta: number) {
	if (!view.value) {
		return
	}

	const newIndex = index + delta
	if (newIndex < 0 || newIndex >= view.value.bucketSortBy.length) {
		return
	}

	const by = [...view.value.bucketSortBy]
	const order = [...view.value.bucketSortOrder]
	;[by[index], by[newIndex]] = [by[newIndex], by[index]]
	;[order[index], order[newIndex]] = [order[newIndex], order[index]]
	view.value.bucketSortBy = by
	view.value.bucketSortOrder = order
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
						@click="() => view.bucketConfiguration.push({title: '', filter: {filter: '', filter_include_nulls: false}})"
					>
						{{ $t('project.kanban.addBucket') }}
					</XButton>
				</div>
			</div>
		</div>
		<div
			v-if="view.viewKind === 'kanban'"
			class="field"
		>
			<label class="label">
				{{ $t('project.kanban.sortBy') }}
			</label>
			<p
				v-if="view.bucketSortBy.length === 0"
				class="help"
			>
				{{ $t('project.kanban.sortEmptyHint') }}
			</p>
			<div class="control">
				<div
					v-for="(field, index) in view.bucketSortBy"
					:key="'bucket_sort_'+index"
					class="bucket-sort-row"
				>
					<div class="select">
						<select
							:value="field"
							@change="changeBucketSortField(index, ($event.target as HTMLSelectElement).value)"
						>
							<option
								v-for="opt in availableBucketSortFields(index)"
								:key="opt"
								:value="opt"
							>
								{{ $t(bucketSortFieldLabelKey(opt)) }}
							</option>
						</select>
					</div>
					<div class="select">
						<select v-model="view.bucketSortOrder[index]">
							<option value="asc">
								{{ $t(bucketSortDirectionLabelKey(field, 'asc')) }}
							</option>
							<option value="desc">
								{{ $t(bucketSortDirectionLabelKey(field, 'desc')) }}
							</option>
						</select>
					</div>
					<XButton
						variant="secondary"
						icon="chevron-up"
						:disabled="index === 0"
						@click.prevent="moveBucketSort(index, -1)"
					/>
					<XButton
						variant="secondary"
						icon="chevron-down"
						:disabled="index === view.bucketSortBy.length - 1"
						@click.prevent="moveBucketSort(index, 1)"
					/>
					<button
						class="is-danger"
						@click.prevent="removeBucketSort(index)"
					>
						<Icon icon="trash-alt" />
					</button>
				</div>
				<div class="is-flex is-justify-content-end">
					<XButton
						v-if="canAddBucketSort"
						variant="secondary"
						icon="plus"
						@click.prevent="addBucketSort"
					>
						{{ $t('project.kanban.addSort') }}
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
.bucket-sort-row {
	display: flex;
	align-items: center;
	gap: .5rem;
	margin-block-end: .5rem;

	.select {
		flex: 1 1 auto;

		select {
			inline-size: 100%;
		}
	}

	> button {
		background: transparent;
		border: none;
		color: var(--danger);
		cursor: pointer;
	}
}

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
