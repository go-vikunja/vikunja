<script setup lang="ts">
import {onBeforeMount, ref, watch} from 'vue'

import type {ProjectView, ProjectViewWritable, TaskCollection} from '@/client/generated'
import type {EditableTaskCollection} from '@/types/EditableTaskCollection'
import {
	createProjectViewDraft,
	createProjectViewUpdate,
	type ProjectViewDraft,
} from '@/client/queries/projectViews'

import {hasFilterQuery, transformFilterStringForApi, transformFilterStringFromApi} from '@/helpers/filters'
import {useLabels} from '@/composables/useLabels'
import {useProjectNavigation} from '@/composables/useProjectNavigation'

import XButton from '@/components/input/Button.vue'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import FilterInputDocs from '@/components/input/filter/FilterInputDocs.vue'
import FilterInput from '@/components/input/filter/FilterInput.vue'
import FormField from '@/components/input/FormField.vue'

const props = withDefaults(defineProps<{
	modelValue: ProjectViewFormValue,
	loading?: boolean,
	showSaveButtons?: boolean,
}>(), {
	loading: false,
	showSaveButtons: false,
})

const emit = defineEmits<{
	'update:modelValue': [value: ProjectViewFormValue],
	'cancel': [],
}>()

type ProjectViewFormValue = ProjectViewWritable & Pick<ProjectView, 'id' | 'project_id'>
type LoadedProjectView = Omit<ProjectViewDraft, 'filter' | 'bucket_configuration'> &
	Pick<ProjectView, 'id' | 'project_id'> & {
		filter: EditableTaskCollection
		bucket_configuration: Array<{title: string, filter: EditableTaskCollection}>
	}

const {isPending, getLabelByExactTitle, getLabelById} = useLabels()
const projectNavigation = useProjectNavigation()

const transformFilterFromApi = (filterInput?: TaskCollection): EditableTaskCollection => {
	const filterString = transformFilterStringFromApi(
		filterInput?.filter ?? '',
		labelId => getLabelById(labelId)?.title || null,
		projectId => projectNavigation.projects[projectId]?.title || null,
	)

	const filter: EditableTaskCollection = {
		sort_by: filterInput?.sort_by ?? [],
		order_by: filterInput?.order_by ?? [],
		filter: '',
		filter_include_nulls: false,
		s: '',
	}
	if (hasFilterQuery(filterString)) {
		filter.filter = filterString
	} else {
		filter.s = filterString
	}

	if (filter.s === '') {
		filter.s = filterInput?.s ?? ''
	}

	if (filter.filter === '') {
		filter.filter = filter.s
	}

	filter.filter_include_nulls = filterInput?.filter_include_nulls ?? false

	return filter
}

function transformViewFromApi(modelValue: ProjectViewFormValue): LoadedProjectView {
	const draft = createProjectViewUpdate(modelValue)
	return {
		...modelValue,
		...draft,
		filter: transformFilterFromApi(modelValue.filter),
		bucket_configuration: (modelValue.bucket_configuration ?? []).map(bucket => ({
			title: bucket.title ?? '',
			filter: transformFilterFromApi(bucket.filter),
		})),
	}
}

const view = ref<LoadedProjectView>(transformViewFromApi(createProjectViewDraft()))

onBeforeMount(() => {
	const transformed = transformViewFromApi(props.modelValue)

	if (JSON.stringify(view.value) !== JSON.stringify(transformed)) {
		view.value = transformed
	}

	const initialFilter = transformed.filter.filter
	const initialBuckets = transformed.bucket_configuration.map(bucket => ({
		title: bucket.title,
		filter: bucket.filter.filter,
	}))

	watch(isPending, pending => {
		if (pending || !view.value) {
			return
		}

		const resolved = transformViewFromApi(props.modelValue)
		if (view.value.filter.filter === initialFilter) {
			view.value.filter = {
				...view.value.filter,
				filter: resolved.filter.filter,
			}
		}

		view.value.bucket_configuration = view.value.bucket_configuration.map((bucket, index) => {
			const initial = initialBuckets[index]
			const resolvedBucket = resolved.bucket_configuration[index]
			if (
				!initial ||
				!resolvedBucket ||
				bucket.title !== initial.title ||
				bucket.filter.filter !== initial.filter
			) {
				return bucket
			}

			return {
				...bucket,
				filter: {
					...bucket.filter,
					filter: resolvedBucket.filter.filter,
				},
			}
		})
	}, {immediate: true})

	watch(() => view.value?.view_kind, kind => {
		if (kind === 'kanban' && view.value?.bucket_configuration_mode === 'none') {
			view.value.bucket_configuration_mode = 'manual'
		}
	}, {immediate: true})
})

function save() {
	if (!view.value) {
		return
	}

	const transformFilterForApi = (filterInput?: EditableTaskCollection): EditableTaskCollection => {
		const filterString = transformFilterStringForApi(
			filterInput?.filter || '',
			labelTitle => getLabelByExactTitle(labelTitle)?.id || null,
			projectTitle => {
				const found = projectNavigation.findProjectByExactname(projectTitle)
				return found?.id || null
			},
		)
		const filter: EditableTaskCollection = {
			sort_by: filterInput?.sort_by ?? [],
			order_by: filterInput?.order_by ?? [],
			filter: '',
			filter_include_nulls: filterInput?.filter_include_nulls ?? false,
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
		filter: transformFilterForApi(view.value.filter),
		bucket_configuration: view.value.bucket_configuration.map(bucket => ({
			title: bucket.title,
			filter: transformFilterForApi(bucket.filter),
		})),
	})
}

const titleValid = ref(true)

function validateTitle() {
	titleValid.value = view.value?.title !== ''
}

function handleBubbleSave(event: FocusEvent) {
	const form = event.currentTarget as HTMLFormElement | null
	if (props.showSaveButtons || form?.contains(event.relatedTarget as Node | null)) {
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
						v-model="view.view_kind"
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
			:project-id="view.project_id"
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
			v-if="view.view_kind === 'kanban'"
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
						v-model="view.bucket_configuration_mode"
						type="radio"
						name="configMode"
						value="manual"
					>
					{{ $t('project.views.bucketConfigManual') }}
				</label>
				<label class="radio">
					<input
						v-model="view.bucket_configuration_mode"
						type="radio"
						name="configMode"
						value="filter"
					>
					{{ $t('project.views.filter') }}
				</label>
			</div>
		</div>

		<div
			v-if="view.view_kind === 'kanban' && view.bucket_configuration_mode === 'filter'"
			class="field"
		>
			<label class="label">
				{{ $t('project.views.bucketConfig') }}
			</label>
			<div class="control">
				<div
					v-for="(b, index) in view.bucket_configuration"
					:key="'bucket_'+index"
					class="filter-bucket"
				>
					<button
						class="is-danger"
						@click.prevent="() => view.bucket_configuration.splice(index, 1)"
					>
						<Icon icon="trash-alt" />
					</button>
					<div class="filter-bucket-form">
						<FormField
							:id="'bucket_'+index+'_title'"
							v-model="view.bucket_configuration[index].title"
							:label="$t('project.views.title')"
							:placeholder="$t('project.share.links.namePlaceholder')"
						/>

						<FilterInput
							v-model="view.bucket_configuration[index].filter.filter"
							:project-id="view.project_id"
							:input-label="$t('project.views.filter')"
							class="mbe-2"
						/>

						<div class="is-size-7 mbe-2">
							<FilterInputDocs />
						</div>

						<div class="field mbe-3">
							<FancyCheckbox
								v-model="view.bucket_configuration[index].filter.filter_include_nulls"
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
						@click="() => view.bucket_configuration.push({title: '', filter: {sort_by: [], order_by: [], filter: '', filter_include_nulls: false, s: ''}})"
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
