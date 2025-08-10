<script setup lang="ts">
import {onBeforeMount, ref} from 'vue'

import type {IProjectView} from '@/modelTypes/IProjectView'
import type {IFilters} from '@/modelTypes/ISavedFilter'

import {hasFilterQuery, transformFilterStringForApi, transformFilterStringFromApi} from '@/helpers/filters'
import {useLabelStore} from '@/stores/labels'
import {useProjectStore} from '@/stores/projects'

import XButton from '@/components/input/Button.vue'
import FilterInputDocs from '@/components/input/filter/FilterInputDocs.vue'
import FilterInput from '@/components/input/filter/FilterInput.vue'

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

		return filter
	}

	const transformed = {
		...props.modelValue,
		filter: transformFilterFromApi(props.modelValue.filter),
		bucketConfiguration: props.modelValue.bucketConfiguration.map(bc => ({
			title: bc.title,
			filter: transformFilterFromApi(bc.filter),
		})),
	}

	if (JSON.stringify(view.value) !== JSON.stringify(transformed)) {
		view.value = transformed
	}
})

function save() {
	const transformFilterForApi = (filterQuery: string): IFilters => {
		const filterString = transformFilterStringForApi(
			filterQuery,
			labelTitle => labelStore.getLabelByExactTitle(labelTitle)?.id || null,
			projectTitle => {
				const found = projectStore.findProjectByExactname(projectTitle)
				return found?.id || null
			},
		)
		const filter: IFilters = {}
		if (hasFilterQuery(filterString)) {
			filter.filter = filterString
		} else {
			filter.s = filterString
		}

		return filter
	}

	emit('update:modelValue', {
		...view.value,
		filter: transformFilterForApi(view.value?.filter?.filter || ''),
		bucketConfiguration: view.value?.bucketConfiguration.map(bc => ({
			title: bc.title,
			filter: transformFilterForApi(bc.filter?.filter || ''),
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
	<form @focusout="handleBubbleSave">
		<div class="field">
			<label
				class="label"
				for="title"
			>
				{{ $t('project.views.title') }}
			</label>
			<div class="control">
				<input
					id="title"
					v-model="view.title"
					v-focus
					class="input"
					:placeholder="$t('project.share.links.namePlaceholder')"
					@blur="validateTitle"
				>
			</div>
			<p
				v-if="!titleValid"
				class="help is-danger"
			>
				{{ $t('project.views.titleRequired') }}
			</p>
		</div>

		<div class="field">
			<label
				class="label"
				for="kind"
			>
				{{ $t('project.views.kind') }}
			</label>
			<div class="control">
				<div class="select">
					<select
						id="kind"
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
			</div>
		</div>

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

		<div class="is-size-7 mbe-3">
			<FilterInputDocs />
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
						<div class="field">
							<label
								class="label"
								:for="'bucket_'+index+'_title'"
							>
								{{ $t('project.views.title') }}
							</label>
							<div class="control">
								<input
									:id="'bucket_'+index+'_title'"
									v-model="view.bucketConfiguration[index].title"
									class="input"
									:placeholder="$t('project.share.links.namePlaceholder')"
								>
							</div>
						</div>

						<FilterInput
							v-model="view.bucketConfiguration[index].filter.filter"
							:project-id="view.projectId"
							:input-label="$t('project.views.filter')"
							class="mbe-2"
						/>

						<div class="is-size-7">
							<FilterInputDocs />
						</div>
					</div>
				</div>
				<div class="is-flex is-justify-content-end">
					<XButton
						variant="secondary"
						icon="plus"
						@click="() => view.bucketConfiguration.push({title: '', filter: {filter: ''}})"
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
				@click="save"
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
</style>
