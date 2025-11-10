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
			class="tw-text-sm mbe-2"
		>
			{{ $t('filters.fromView') }}
			<code>{{ filterFromView }}</code><br>
			{{ $t('filters.fromViewBoth') }}
		</div>

		<div class="field is-flex is-flex-direction-column">
			<FancyCheckbox
				v-model="params.filter_include_nulls"
				@change="() => change('always')"
			>
				{{ $t('filters.attributes.includeNulls') }}
			</FancyCheckbox>
		</div>

		<div class="field">
			<label class="label">{{ $t('task.show.sortBy') }}</label>
			<div class="field has-addons">
				<div class="control is-expanded">
					<div class="select is-fullwidth">
						<select
							:value="sortField"
							@change="setSortField($event.target.value)"
						>
							<option value="due_date">
								{{ $t('task.attributes.dueDate') }}
							</option>
							<option value="priority">
								{{ $t('task.attributes.priority') }}
							</option>
							<option value="title">
								{{ $t('task.attributes.title') }}
							</option>
							<option value="start_date">
								{{ $t('task.attributes.startDate') }}
							</option>
							<option value="end_date">
								{{ $t('task.attributes.endDate') }}
							</option>
							<option value="done">
								{{ $t('task.attributes.done') }}
							</option>
							<option
								v-if="showPositionSort"
								value="position"
							>
								{{ $t('task.attributes.position') }}
							</option>
						</select>
					</div>
				</div>
				<div class="control">
					<button
						type="button"
						class="button"
						@click.prevent="toggleSortOrder"
					>
						{{ sortOrderLabel }}
					</button>
				</div>
			</div>
		</div>

		<FilterInputDocs />

		<template
			v-if="hasFooter"
			#footer
		>
			<XButton
				variant="secondary"
				class="mie-2"
				:disabled="filterQuery === ''"
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

<script lang="ts">
export const ALPHABETICAL_SORT = 'title'
</script>

<script setup lang="ts">
import {computed, ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import {useRoute} from 'vue-router'
import type {TaskFilterParams} from '@/services/taskCollection'
import {useLabelStore} from '@/stores/labels'
import {useProjectStore} from '@/stores/projects'
import {
	hasFilterQuery,
	transformFilterStringForApi,
} from '@/helpers/filters'
import FilterInputDocs from '@/components/input/filter/FilterInputDocs.vue'
import FilterInput from '@/components/input/filter/FilterInput.vue'

const props = withDefaults(defineProps<{
	modelValue: TaskFilterParams,
	hasTitle?: boolean,
	hasFooter?: boolean,
	changeImmediately?: boolean,
	filterFromView?: string,
	showClose?: boolean,
	showPositionSort?: boolean,
}>(), {
	hasTitle: false,
	hasFooter: true,
	changeImmediately: false,
	filterFromView: undefined,
	showClose: false,
	showPositionSort: true,
})

const emit = defineEmits<{
	'update:modelValue': [value: TaskFilterParams],
	'showResults': [],
	'close': [],
}>()

const {t} = useI18n({useScope: 'global'})

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

// Local state for sorting that doesn't trigger immediate updates
const localSortBy = ref<string>('')
const localOrderBy = ref<string>('')

const sortField = computed(() => localSortBy.value || params.value.sort_by?.[0] || 'due_date')
const sortOrder = computed(() => localOrderBy.value || params.value.order_by?.[0] || 'asc')

// Sync local state when params change from outside
watch(
	() => [params.value.sort_by?.[0], params.value.order_by?.[0]],
	([sortBy, orderBy]) => {
		localSortBy.value = sortBy || ''
		localOrderBy.value = orderBy || ''
	},
	{immediate: true},
)

const sortOrderLabel = computed(() => {
	const isAsc = sortOrder.value === 'asc'

	switch (sortField.value) {
		case 'due_date':
		case 'start_date':
		case 'end_date':
			return isAsc ? t('task.sort.earliestFirst') : t('task.sort.latestFirst')
		case 'priority':
			return isAsc ? t('task.sort.lowPriorityFirst') : t('task.sort.highPriorityFirst')
		case 'title':
			return isAsc ? t('task.sort.aToZ') : t('task.sort.zToA')
		case 'done':
			return isAsc ? t('task.sort.undoneFirst') : t('task.sort.doneFirst')
		case 'position':
			return isAsc ? t('task.sort.firstToLast') : t('task.sort.lastToFirst')
		default:
			return isAsc ? t('misc.ascending') : t('misc.descending')
	}
})

function setSortField(field: string) {
	localSortBy.value = field
	// Initialize order if not set
	if (!localOrderBy.value) {
		localOrderBy.value = 'asc'
	}
	// If changeImmediately is true, update params and emit right away
	if (props.changeImmediately) {
		params.value.sort_by = [field as TaskFilterParams['sort_by'][number], 'id']
		if (!params.value.order_by || params.value.order_by.length === 0) {
			params.value.order_by = ['asc', 'desc']
		}
		change('always')
	}
}

function toggleSortOrder() {
	localOrderBy.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
	// If changeImmediately is true, update params and emit right away
	if (props.changeImmediately) {
		params.value.order_by = [localOrderBy.value as TaskFilterParams['order_by'][number], 'desc']
		change('always')
	}
}

const labelStore = useLabelStore()
const projectStore = useProjectStore()

const filterInputRef = ref()

// Using watchDebounced to prevent the filter re-triggering itself.
watch(
	() => props.modelValue,
	(value: TaskFilterParams) => {
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
		labelTitle => labelStore.getLabelByExactTitle(labelTitle)?.id || null,
		projectTitle => {
			const found = projectStore.findProjectByExactname(projectTitle)
			return found?.id || null
		},
	)

	let s = ''

	// When the filter does not contain any filter tokens, assume a simple search and redirect the input
	if (!hasFilterQuery(filter)) {
		s = filter
	}

	// Apply local sort state to params before emitting
	if (localSortBy.value) {
		params.value.sort_by = [localSortBy.value as TaskFilterParams['sort_by'][number], 'id']
	}
	if (localOrderBy.value) {
		params.value.order_by = [localOrderBy.value as TaskFilterParams['order_by'][number], 'desc']
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

function focusFilterInput() {
	filterInputRef.value?.focus()
}

defineExpose({
	focusFilterInput,
})
</script>
