<template>
	<card class="filters has-overflow" :title="hasTitle ? $t('filters.title') : ''">
		<div class="field is-flex is-flex-direction-column">
			<fancycheckbox
				v-model="params.filter_include_nulls"
				@update:model-value="change()"
			>
				{{ $t('filters.attributes.includeNulls') }}
			</fancycheckbox>
			<fancycheckbox
				v-model="filters.requireAllFilters"
				@update:model-value="setFilterConcat()"
			>
				{{ $t('filters.attributes.requireAll') }}
			</fancycheckbox>
			<fancycheckbox
				v-model="filters.done"
				@update:model-value="setDoneFilter"
			>
				{{ $t('filters.attributes.showDoneTasks') }}
			</fancycheckbox>
			<fancycheckbox
				v-if="!['project.kanban', 'project.table'].includes($route.name as string)"
				v-model="sortAlphabetically"
				@update:model-value="change()"
			>
				{{ $t('filters.attributes.sortAlphabetically') }}
			</fancycheckbox>
		</div>
		<div class="field">
			<label class="label">{{ $t('misc.search') }}</label>
			<div class="control">
				<input
					class="input"
					:placeholder="$t('misc.search')"
					v-model="params.s"
					@blur="change()"
					@keyup.enter="change()"
				/>
			</div>
		</div>
		<div class="field">
			<label class="label">{{ $t('task.attributes.priority') }}</label>
			<div class="control single-value-control">
				<priority-select
					v-model.number="filters.priority"
					@update:model-value="setPriority"
					:disabled="!filters.usePriority || undefined"
				/>
				<fancycheckbox
					v-model="filters.usePriority"
					@update:model-value="setPriority"
				>
					{{ $t('filters.attributes.enablePriority') }}
				</fancycheckbox>
			</div>
		</div>
		<div class="field">
			<label class="label">{{ $t('task.attributes.percentDone') }}</label>
			<div class="control single-value-control">
				<percent-done-select
					v-model.number="filters.percentDone"
					@update:model-value="setPercentDoneFilter"
					:disabled="!filters.usePercentDone || undefined"
				/>
				<fancycheckbox
					v-model="filters.usePercentDone"
					@update:model-value="setPercentDoneFilter"
				>
					{{ $t('filters.attributes.enablePercentDone') }}
				</fancycheckbox>
			</div>
		</div>
		<div class="field">
			<label class="label">{{ $t('task.attributes.dueDate') }}</label>
			<div class="control">
				<datepicker-with-range
					v-model="filters.dueDate"
					@update:model-value="values => setDateFilter('due_date', values)"
				>
					<template #trigger="{toggle, buttonText}">
						<x-button @click.prevent.stop="toggle()" variant="secondary" :shadow="false" class="mb-2">
							{{ buttonText }}
						</x-button>
					</template>
				</datepicker-with-range>
			</div>
		</div>
		<div class="field">
			<label class="label">{{ $t('task.attributes.startDate') }}</label>
			<div class="control">
				<datepicker-with-range
					v-model="filters.startDate"
					@update:model-value="values => setDateFilter('start_date', values)"
				>
					<template #trigger="{toggle, buttonText}">
						<x-button @click.prevent.stop="toggle()" variant="secondary" :shadow="false" class="mb-2">
							{{ buttonText }}
						</x-button>
					</template>
				</datepicker-with-range>
			</div>
		</div>
		<div class="field">
			<label class="label">{{ $t('task.attributes.endDate') }}</label>
			<div class="control">
				<datepicker-with-range
					v-model="filters.endDate"
					@update:model-value="values => setDateFilter('end_date', values)"
				>
					<template #trigger="{toggle, buttonText}">
						<x-button @click.prevent.stop="toggle()" variant="secondary" :shadow="false" class="mb-2">
							{{ buttonText }}
						</x-button>
					</template>
				</datepicker-with-range>
			</div>
		</div>
		<div class="field">
			<label class="label">{{ $t('task.attributes.reminders') }}</label>
			<div class="control">
				<datepicker-with-range
					v-model="filters.reminders"
					@update:model-value="values => setDateFilter('reminders', values)"
				>
					<template #trigger="{toggle, buttonText}">
						<x-button @click.prevent.stop="toggle()" variant="secondary" :shadow="false" class="mb-2">
							{{ buttonText }}
						</x-button>
					</template>
				</datepicker-with-range>
			</div>
		</div>

		<div class="field">
			<label class="label">{{ $t('task.attributes.assignees') }}</label>
			<div class="control">
				<SelectUser
					v-model="entities.users"
					@select="changeMultiselectFilter('users', 'assignees')"
					@remove="changeMultiselectFilter('users', 'assignees')"
				/>
			</div>
		</div>

		<div class="field">
			<label class="label">{{ $t('task.attributes.labels') }}</label>
			<div class="control labels-list">
				<edit-labels
					:creatable="false"
					v-model="entities.labels"
					@update:model-value="changeLabelFilter"
				/>
			</div>
		</div>

		<template
			v-if="['filters.create', 'project.edit', 'filter.settings.edit'].includes($route.name as string)"
		>
			<div class="field">
				<label class="label">{{ $t('project.projects') }}</label>
				<div class="control">
					<SelectProject
						v-model="entities.projects"
						@select="changeMultiselectFilter('projects', 'project_id')"
						@remove="changeMultiselectFilter('projects', 'project_id')"
					/>
				</div>
			</div>
		</template>
	</card>
</template>

<script lang="ts">
export const ALPHABETICAL_SORT = 'title'
</script>

<script setup lang="ts">
import {computed, nextTick, onMounted, reactive, ref, shallowReactive, toRefs, watch} from 'vue'
import {camelCase} from 'camel-case'

import type {ILabel} from '@/modelTypes/ILabel'
import type {IUser} from '@/modelTypes/IUser'
import type {IProject} from '@/modelTypes/IProject'

import {useLabelStore} from '@/stores/labels'

import DatepickerWithRange from '@/components/date/datepickerWithRange.vue'
import PrioritySelect from '@/components/tasks/partials/prioritySelect.vue'
import PercentDoneSelect from '@/components/tasks/partials/percentDoneSelect.vue'
import EditLabels from '@/components/tasks/partials/editLabels.vue'
import Fancycheckbox from '@/components/input/fancycheckbox.vue'
import SelectUser from '@/components/input/SelectUser.vue'
import SelectProject from '@/components/input/SelectProject.vue'

import {parseDateOrString} from '@/helpers/time/parseDateOrString'
import {dateIsValid, formatISO} from '@/helpers/time/formatDate'
import {objectToSnakeCase} from '@/helpers/case'

import UserService from '@/services/user'
import ProjectService from '@/services/project'

// FIXME: do not use this here for now. instead create new version from DEFAULT_PARAMS
import {getDefaultParams} from '@/composables/useTaskList'

// FIXME: merge with DEFAULT_PARAMS in taskProject.js
const DEFAULT_PARAMS = {
	sort_by: [],
	order_by: [],
	filter_by: [],
	filter_value: [],
	filter_comparator: [],
	filter_include_nulls: true,
	filter_concat: 'or',
	s: '',
} as const

const DEFAULT_FILTERS = {
	done: false,
	dueDate: '',
	requireAllFilters: false,
	priority: 0,
	usePriority: false,
	startDate: '',
	endDate: '',
	percentDone: 0,
	usePercentDone: false,
	reminders: '',
	assignees: '',
	labels: '',
	project_id: '',
} as const

const props = defineProps({
	modelValue: {
		required: true,
	},
	hasTitle: {
		type: Boolean,
		default: false,
	},
})

const emit = defineEmits(['update:modelValue'])

const {modelValue} = toRefs(props)

const labelStore = useLabelStore()

const params = ref({...DEFAULT_PARAMS})
const filters = ref({...DEFAULT_FILTERS})

const services = {
	users: shallowReactive(new UserService()),
	projects: shallowReactive(new ProjectService()),
}

interface Entities {
	users: IUser[]
	labels: ILabel[]
	projects: IProject[]
}

type EntityType = 'users' | 'labels' | 'projects'

const entities: Entities = reactive({
	users: [],
	labels: [],
	projects: [],
})

onMounted(() => {
	filters.value.requireAllFilters = params.value.filter_concat === 'and'
})

watch(
	modelValue,
	(value) => {
		// FIXME: filters should only be converted to snake case in
		// the last moment
		params.value = objectToSnakeCase(value)
		prepareFilters()
	},
	{immediate: true},
)

const sortAlphabetically = computed({
	get() {
		return params.value?.sort_by?.find(sortBy => sortBy === ALPHABETICAL_SORT) !== undefined
	},
	set(sortAlphabetically) {
		params.value.sort_by = sortAlphabetically
			? [ALPHABETICAL_SORT]
			: getDefaultParams().sort_by

		change()
	},
})

function change() {
	const newParams = {...params.value}
	newParams.filter_value = newParams.filter_value.map(v => v instanceof Date ? v.toISOString() : v)
	emit('update:modelValue', newParams)
}

function prepareFilters() {
	prepareDone()
	prepareDate('due_date', 'dueDate')
	prepareDate('start_date', 'startDate')
	prepareDate('end_date', 'endDate')
	prepareSingleValue('priority', 'priority', 'usePriority', true)
	prepareSingleValue('percent_done', 'percentDone', 'usePercentDone', true)
	prepareDate('reminders')
	prepareRelatedObjectFilter('users', 'assignees')
	prepareProjectsFilter()

	prepareSingleValue('labels')

	const newLabels = typeof filters.value.labels === 'string'
		? filters.value.labels
		: ''
	const labelIds = newLabels.split(',').map(i => parseInt(i))

	entities.labels = labelStore.getLabelsByIds(labelIds)
}

function removePropertyFromFilter(filterName) {
	// Because of the way arrays work, we can only ever remove one element at once.
	// To remove multiple filter elements of the same name this function has to be called multiple times.
	for (const i in params.value.filter_by) {
		if (params.value.filter_by[i] === filterName) {
			params.value.filter_by.splice(i, 1)
			params.value.filter_comparator.splice(i, 1)
			params.value.filter_value.splice(i, 1)
			break
		}
	}
}

function setDateFilter(filterName, {dateFrom, dateTo}) {
	dateFrom = parseDateOrString(dateFrom, null)
	dateTo = parseDateOrString(dateTo, null)

	// Only filter if we have a date
	if (dateFrom !== null && dateTo !== null) {

		// Check if we already have values in params and only update them if we do
		let foundStart = false
		let foundEnd = false
		params.value.filter_by.forEach((f, i) => {
			if (f === filterName && params.value.filter_comparator[i] === 'greater_equals') {
				foundStart = true
				params.value.filter_value[i] = dateFrom
			}
			if (f === filterName && params.value.filter_comparator[i] === 'less_equals') {
				foundEnd = true
				params.value.filter_value[i] = dateTo
			}
		})

		if (!foundStart) {
			params.value.filter_by.push(filterName)
			params.value.filter_comparator.push('greater_equals')
			params.value.filter_value.push(dateFrom)
		}
		if (!foundEnd) {
			params.value.filter_by.push(filterName)
			params.value.filter_comparator.push('less_equals')
			params.value.filter_value.push(dateTo)
		}

		filters.value[camelCase(filterName)] = {
			// Passing the dates as string values avoids an endless loop between values changing 
			// in the datepicker (bubbling up to here) and changing here and bubbling down to the 
			// datepicker (because there's a new date instance every time this function gets called).
			// See https://kolaente.dev/vikunja/frontend/issues/2384
			dateFrom: dateIsValid(dateFrom) ? formatISO(dateFrom) : dateFrom,
			dateTo: dateIsValid(dateTo) ? formatISO(dateTo) : dateTo,
		}
		change()
		return
	}

	removePropertyFromFilter(filterName)
	removePropertyFromFilter(filterName)
	change()
}

function prepareDate(filterName, variableName) {
	if (typeof params.value.filter_by === 'undefined') {
		return
	}

	let foundDateStart = false
	let foundDateEnd = false
	for (const i in params.value.filter_by) {
		if (params.value.filter_by[i] === filterName && params.value.filter_comparator[i] === 'greater_equals') {
			foundDateStart = i
		}
		if (params.value.filter_by[i] === filterName && params.value.filter_comparator[i] === 'less_equals') {
			foundDateEnd = i
		}

		if (foundDateStart !== false && foundDateEnd !== false) {
			break
		}
	}

	if (foundDateStart !== false && foundDateEnd !== false) {
		const startDate = new Date(params.value.filter_value[foundDateStart])
		const endDate = new Date(params.value.filter_value[foundDateEnd])
		filters.value[variableName] = {
			dateFrom: !isNaN(startDate)
				? `${startDate.getFullYear()}-${startDate.getMonth() + 1}-${startDate.getDate()}`
				: params.value.filter_value[foundDateStart],
			dateTo: !isNaN(endDate)
				? `${endDate.getFullYear()}-${endDate.getMonth() + 1}-${endDate.getDate()}`
				: params.value.filter_value[foundDateEnd],
		}
	}
}

function setSingleValueFilter(filterName, variableName, useVariableName = '', comparator = 'equals') {
	if (useVariableName !== '' && !filters.value[useVariableName]) {
		removePropertyFromFilter(filterName)
		return
	}

	let found = false
	params.value.filter_by.forEach((f, i) => {
		if (f === filterName) {
			found = true
			params.value.filter_value[i] = filters.value[variableName]
		}
	})

	if (!found) {
		params.value.filter_by.push(filterName)
		params.value.filter_comparator.push(comparator)
		params.value.filter_value.push(filters.value[variableName])
	}

	change()
}

function prepareSingleValue(
	/** The filter name in the api. */
	filterName,
	/** The name of the variable in filters ref. */
	variableName = null,
	/** The name of the variable of the "Use this filter" variable. Will only be set if the parameter is not null. */
	useVariableName = null,
	/** Toggles if the value should be parsed as a number. */
	isNumber = false,
) {
	if (variableName === null) {
		variableName = filterName
	}

	let found = false
	for (const i in params.value.filter_by) {
		if (params.value.filter_by[i] === filterName) {
			found = i
			break
		}
	}

	if (found === false && useVariableName !== null) {
		filters.value[useVariableName] = false
		return
	}

	if (isNumber) {
		filters.value[variableName] = Number(params.value.filter_value[found])
	} else {
		filters.value[variableName] = params.value.filter_value[found]
	}

	if (useVariableName !== null) {
		filters.value[useVariableName] = true
	}
}

function prepareDone() {
	// Set filters.done based on params
	if (typeof params.value.filter_by === 'undefined') {
		return
	}

	filters.value.done = params.value.filter_by.some((f) => f === 'done') === false
}

async function prepareRelatedObjectFilter(kind: EntityType, filterName = null, servicePrefix: Omit<EntityType, 'labels'> | null = null) {
	if (filterName === null) {
		filterName = kind
	}

	if (servicePrefix === null) {
		servicePrefix = kind
	}

	prepareSingleValue(filterName)
	if (typeof filters.value[filterName] === 'undefined' || filters.value[filterName] === '') {
		return
	}

	// Don't load things if we already have something loaded.
	// This is not the most ideal solution because it prevents a re-population when filters are changed 
	// from the outside. It is still fine because we're not changing them from the outside, other than 
	// loading them initially.
	if (entities[kind].length > 0) {
		return
	}

	entities[kind] = await services[servicePrefix].getAll({}, {s: filters.value[filterName]})
}

async function prepareProjectsFilter() {
	await prepareRelatedObjectFilter('projects', 'project_id')
	entities.projects = entities.projects.filter(p => p.id > 0)
}

function setDoneFilter() {
	if (filters.value.done) {
		removePropertyFromFilter('done')
	} else {
		params.value.filter_by.push('done')
		params.value.filter_comparator.push('equals')
		params.value.filter_value.push('false')
	}
	change()
}

function setFilterConcat() {
	if (filters.value.requireAllFilters) {
		params.value.filter_concat = 'and'
	} else {
		params.value.filter_concat = 'or'
	}
	change()
}

function setPriority() {
	setSingleValueFilter('priority', 'priority', 'usePriority')
}

function setPercentDoneFilter() {
	setSingleValueFilter('percent_done', 'percentDone', 'usePercentDone')
}

async function changeMultiselectFilter(kind: EntityType, filterName) {
	await nextTick()

	if (entities[kind].length === 0) {
		removePropertyFromFilter(filterName)
		change()
		return
	}

	const ids = entities[kind].map(u => kind === 'users' ? u.username : u.id)

	filters.value[filterName] = ids.join(',')
	setSingleValueFilter(filterName, filterName, '', 'in')
}

function changeLabelFilter() {
	if (entities.labels.length === 0) {
		removePropertyFromFilter('labels')
		change()
		return
	}

	const labelIDs = entities.labels.map(u => u.id)
	filters.value.labels = labelIDs.join(',')
	setSingleValueFilter('labels', 'labels', '', 'in')
}
</script>

<style lang="scss" scoped>
.single-value-control {
	display: flex;
	align-items: center;

	.fancycheckbox {
		margin-left: .5rem;
	}
}

:deep(.datepicker-with-range-container .popup) {
	right: 0;
}
</style>
