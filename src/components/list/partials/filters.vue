<template>
	<card class="filters has-overflow">
		<fancycheckbox v-model="params.filter_include_nulls">
			{{ $t('filters.attributes.includeNulls') }}
		</fancycheckbox>
		<fancycheckbox
			v-model="filters.requireAllFilters"
			@change="setFilterConcat()"
		>
			{{ $t('filters.attributes.requireAll') }}
		</fancycheckbox>
		<div class="field">
			<label class="label">
				{{ $t('filters.attributes.showDoneTasks') }}
			</label>
			<div class="control">
				<fancycheckbox @change="setDoneFilter" v-model="filters.done">
					{{ $t('filters.attributes.showDoneTasks') }}
				</fancycheckbox>
			</div>
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
					:disabled="!filters.usePriority || null"
					v-model.number="filters.priority"
					@change="setPriority"
				/>
				<fancycheckbox
					v-model="filters.usePriority"
					@change="setPriority"
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
					@change="setPercentDoneFilter"
					:disabled="!filters.usePercentDone || null"
				/>
				<fancycheckbox
					v-model="filters.usePercentDone"
					@change="setPercentDoneFilter"
				>
					{{ $t('filters.attributes.enablePercentDone') }}
				</fancycheckbox>
			</div>
		</div>
		<div class="field">
			<label class="label">{{ $t('task.attributes.dueDate') }}</label>
			<div class="control">
				<flat-pickr
					:config="flatPickerConfig"
					@on-close="setDueDateFilter"
					class="input"
					:placeholder="$t('filters.attributes.dueDateRange')"
					v-model="filters.dueDate"
				/>
			</div>
		</div>
		<div class="field">
			<label class="label">{{ $t('task.attributes.startDate') }}</label>
			<div class="control">
				<flat-pickr
					:config="flatPickerConfig"
					@on-close="setStartDateFilter"
					class="input"
					:placeholder="$t('filters.attributes.startDateRange')"
					v-model="filters.startDate"
				/>
			</div>
		</div>
		<div class="field">
			<label class="label">{{ $t('task.attributes.endDate') }}</label>
			<div class="control">
				<flat-pickr
					:config="flatPickerConfig"
					@on-close="setEndDateFilter"
					class="input"
					:placeholder="$t('filters.attributes.endDateRange')"
					v-model="filters.endDate"
				/>
			</div>
		</div>
		<div class="field">
			<label class="label">{{ $t('task.attributes.reminders') }}</label>
			<div class="control">
				<flat-pickr
					:config="flatPickerConfig"
					@on-close="setReminderFilter"
					class="input"
					:placeholder="$t('filters.attributes.reminderRange')"
					v-model="filters.reminders"
				/>
			</div>
		</div>

		<div class="field">
			<label class="label">{{ $t('task.attributes.assignees') }}</label>
			<div class="control">
				<multiselect
					:loading="usersService.loading"
					:placeholder="$t('team.edit.search')"
					@search="query => find('users', query)"
					:search-results="foundusers"
					@select="() => add('users', 'assignees')"
					label="username"
					:multiple="true"
					@remove="() => remove('users', 'assignees')"
					v-model="users"
				/>
			</div>
		</div>

		<div class="field">
			<label class="label">{{ $t('task.attributes.labels') }}</label>
			<div class="control labels-list">
				<edit-labels v-model="labels" @change="changeLabelFilter"/>
			</div>
		</div>

		<template v-if="$route.name === 'filters.create' || $route.name === 'list.edit'">
			<div class="field">
				<label class="label">{{ $t('list.lists') }}</label>
				<div class="control">
					<multiselect
						:loading="listsService.loading"
						:placeholder="$t('list.search')"
						@search="query => find('lists', query)"
						:search-results="foundlists"
						@select="() => add('lists', 'list_id')"
						label="title"
						@remove="() => remove('lists', 'list_id')"
						:multiple="true"
						v-model="lists"
					/>
				</div>
			</div>
			<div class="field">
				<label class="label">{{ $t('namespace.namespaces') }}</label>
				<div class="control">
					<multiselect
						:loading="namespaceService.loading"
						:placeholder="$t('namespace.search')"
						@search="query => find('namespace', query)"
						:search-results="foundnamespace"
						@select="() => add('namespace', 'namespace')"
						label="title"
						@remove="() => remove('namespace', 'namespace')"
						:multiple="true"
						v-model="namespace"
					/>
				</div>
			</div>
		</template>
	</card>
</template>

<script>
import Fancycheckbox from '../../input/fancycheckbox'
import flatPickr from 'vue-flatpickr-component'
import 'flatpickr/dist/flatpickr.css'

import {includesById} from '@/helpers/utils'
import {formatISO} from 'date-fns'
import PrioritySelect from '@/components/tasks/partials/prioritySelect.vue'
import PercentDoneSelect from '@/components/tasks/partials/percentDoneSelect.vue'
import Multiselect from '@/components/input/multiselect.vue'

import UserService from '@/services/user'
import ListService from '@/services/list'
import NamespaceService from '@/services/namespace'
import EditLabels from '@/components/tasks/partials/editLabels.vue'

import {objectToSnakeCase} from '@/helpers/case'

// FIXME: merge with DEFAULT_PARAMS in taskList.js
const DEFAULT_PARAMS = {
	sort_by: [],
	order_by: [],
	filter_by: [],
	filter_value: [],
	filter_comparator: [],
	filter_include_nulls: true,
	filter_concat: 'or',
	s: '',
}

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
	list_id: '',
	namespace: '',
}

export default {
	name: 'filters',
	components: {
		EditLabels,
		PrioritySelect,
		Fancycheckbox,
		flatPickr,
		PercentDoneSelect,
		Multiselect,
	},
	data() {
		return {
			params: DEFAULT_PARAMS,
			filters: DEFAULT_FILTERS,

			usersService: new UserService(),
			foundusers: [],
			users: [],

			labelQuery: '',
			labels: [],

			listsService: new ListService(),
			foundlists: [],
			lists: [],

			namespaceService: new NamespaceService(),
			foundnamespace: [],
			namespace: [],
		}
	},
	mounted() {
		this.filters.requireAllFilters = this.params.filter_concat === 'and'
	},
	props: {
		modelValue: {
			required: true,
		},
	},
	emits: ['update:modelValue', 'change'],
	watch: {
		modelValue: {
			handler(value) {
				// FIXME: filters should only be converted to snake case in
				// the last moment
				this.params = objectToSnakeCase(value)
				this.prepareFilters()
			},
			immediate: true,
		},
	},
	computed: {
		foundLabels() {
			return this.$store.getters['labels/filterLabelsByQuery'](this.labels, this.query)
		},
		flatPickerConfig() {
			return {
				altFormat: this.$t('date.altFormatLong'),
				altInput: true,
				dateFormat: 'Y-m-d H:i',
				enableTime: true,
				time_24hr: true,
				mode: 'range',
				locale: {
					firstDayOfWeek: this.$store.state.auth.settings.weekStart,
				},
			}
		},
	},
	methods: {
		change() {
			this.$emit('update:modelValue', this.params)
			this.$emit('change', this.params)
		},
		prepareFilters() {
			this.prepareDone()
			this.prepareDate('due_date', 'dueDate')
			this.prepareDate('start_date', 'startDate')
			this.prepareDate('end_date', 'endDate')
			this.prepareSingleValue('priority', 'priority', 'usePriority', true)
			this.prepareSingleValue('percent_done', 'percentDone', 'usePercentDone', true)
			this.prepareDate('reminders')
			this.prepareRelatedObjectFilter('users', 'assignees')
			this.prepareRelatedObjectFilter('lists', 'list_id')
			this.prepareRelatedObjectFilter('namespace')

			this.prepareSingleValue('labels')

			const labels = typeof this.filters.labels === 'string'
				? this.filters.labels
				: ''
			const labelIds = labels.split(',').map(i => parseInt(i))

			this.labels = this.$store.getters['labels/getLabelsByIds'](labelIds)
		},
		removePropertyFromFilter(propertyName) {
			// Because of the way arrays work, we can only ever remove one element at once.
			// To remove multiple filter elements of the same name this function has to be called multiple times.
			for (const i in this.params.filter_by) {
				if (this.params.filter_by[i] === propertyName) {
					this.params.filter_by.splice(i, 1)
					this.params.filter_comparator.splice(i, 1)
					this.params.filter_value.splice(i, 1)
					break
				}
			}
		},
		setDateFilter(filterName, variableName = null) {
			if (variableName === null) {
				variableName = filterName
			}
			
			// Only filter if we have a start and end due date
			if (this.filters[variableName] !== '') {

				const parts = this.filters[variableName].split(' to ')

				if (parts.length < 2) {
					return
				}

				// Check if we already have values in params and only update them if we do
				let foundStart = false
				let foundEnd = false
				this.params.filter_by.forEach((f, i) => {
					if (f === filterName && this.params.filter_comparator[i] === 'greater_equals') {
						foundStart = true
						this.params.filter_value[i] = formatISO(new Date(parts[0]))
					}
					if (f === filterName && this.params.filter_comparator[i] === 'less_equals') {
						foundEnd = true
						this.params.filter_value[i] = formatISO(new Date(parts[1]))
					}
				})

				if (!foundStart) {
					this.params.filter_by.push(filterName)
					this.params.filter_comparator.push('greater_equals')
					this.params.filter_value.push(formatISO(new Date(parts[0])))
				}
				if (!foundEnd) {
					this.params.filter_by.push(filterName)
					this.params.filter_comparator.push('less_equals')
					this.params.filter_value.push(formatISO(new Date(parts[1])))
				}
				this.change()
				return
			}

			this.removePropertyFromFilter(filterName)
			this.removePropertyFromFilter(filterName)
			this.change()
		},
		prepareDate(filterName, variableName) {
			if (typeof this.params.filter_by === 'undefined') {
				return
			}

			let foundDateStart = false
			let foundDateEnd = false
			for (const i in this.params.filter_by) {
				if (this.params.filter_by[i] === filterName && this.params.filter_comparator[i] === 'greater_equals') {
					foundDateStart = i
				}
				if (this.params.filter_by[i] === filterName && this.params.filter_comparator[i] === 'less_equals') {
					foundDateEnd = i
				}

				if (foundDateStart !== false && foundDateEnd !== false) {
					break
				}
			}

			if (foundDateStart !== false && foundDateEnd !== false) {
				const start = new Date(this.params.filter_value[foundDateStart])
				const end = new Date(this.params.filter_value[foundDateEnd])
				this.filters[variableName] = `${start.getFullYear()}-${start.getMonth() + 1}-${start.getDate()} to ${end.getFullYear()}-${end.getMonth() + 1}-${end.getDate()}`
			}
		},
		setSingleValueFilter(filterName, variableName, useVariableName = '', comparator = 'equals') {
			if (useVariableName !== '' && !this.filters[useVariableName]) {
				this.removePropertyFromFilter(filterName)
				return
			}

			let found = false
			this.params.filter_by.forEach((f, i) => {
				if (f === filterName) {
					found = true
					this.params.filter_value[i] = this.filters[variableName]
				}
			})

			if (!found) {
				this.params.filter_by.push(filterName)
				this.params.filter_comparator.push(comparator)
				this.params.filter_value.push(this.filters[variableName])
			}

			this.change()
		},
		/**
		 *
		 * @param filterName The filter name in the api.
		 * @param variableName The name of the variable in this.filters.
		 * @param useVariableName The name of the variable of the "Use this filter" variable. Will only be set if the parameter is not null.
		 * @param isNumber Toggles if the value should be parsed as a number.
		 */
		prepareSingleValue(filterName, variableName = null, useVariableName = null, isNumber = false) {
			if (variableName === null) {
				variableName = filterName
			}

			let found = false
			for (const i in this.params.filter_by) {
				if (this.params.filter_by[i] === filterName) {
					found = i
					break
				}
			}

			if (found === false && useVariableName !== null) {
				this.filters[useVariableName] = false
				return
			}

			if (isNumber) {
				this.filters[variableName] = Number(this.params.filter_value[found])
			} else {
				this.filters[variableName] = this.params.filter_value[found]
			}

			if (useVariableName !== null) {
				this.filters[useVariableName] = true
			}
		},
		prepareDone() {
			// Set filters.done based on params
			if (typeof this.params.filter_by === 'undefined') {
				return
			}

			this.filters.done = this.params.filter_by.some((f) => f === 'done') === false
		},
		async prepareRelatedObjectFilter(kind, filterName = null, servicePrefix = null) {
			if (filterName === null) {
				filterName = kind
			}

			if (servicePrefix === null) {
				servicePrefix = kind
			}

			this.prepareSingleValue(filterName)
			if (typeof this.filters[filterName] === 'undefined' || this.filters[filterName] === '') {
				return
			}

			this[kind] = await this[`${servicePrefix}Service`].getAll({}, {s: this.filters[filterName]})
		},
		setDoneFilter() {
			if (this.filters.done) {
				this.removePropertyFromFilter('done')
			} else {
				this.params.filter_by.push('done')
				this.params.filter_comparator.push('equals')
				this.params.filter_value.push('false')
			}
			this.change()
		},
		setFilterConcat() {
			if (this.filters.requireAllFilters) {
				this.params.filter_concat = 'and'
			} else {
				this.params.filter_concat = 'or'
			}
		},
		setDueDateFilter() {
			this.setDateFilter('due_date', 'dueDate')
		},
		setPriority() {
			this.setSingleValueFilter('priority', 'priority', 'usePriority')
		},
		setStartDateFilter() {
			this.setDateFilter('start_date', 'startDate')
		},
		setEndDateFilter() {
			this.setDateFilter('end_date', 'endDate')
		},
		setPercentDoneFilter() {
			this.setSingleValueFilter('percent_done', 'percentDone', 'usePercentDone')
		},
		setReminderFilter() {
			this.setDateFilter('reminders')
		},
		clear(kind) {
			this[`found${kind}`] = []
		},
		async find(kind, query) {

			if (query === '') {
				this.clear(kind)
			}

			const response = await this[`${kind}Service`].getAll({}, {s: query})

			// Filter users from the results who are already assigned
			this[`found${kind}`] = response.filter(({id}) => !includesById(this[kind], id))
		},
		add(kind, filterName) {
			this.$nextTick(() => {
				this.changeMultiselectFilter(kind, filterName)
			})
		},
		remove(kind, filterName) {
			this.$nextTick(() => {
				this.changeMultiselectFilter(kind, filterName)
			})
		},
		changeMultiselectFilter(kind, filterName) {
			if (this[kind].length === 0) {
				this.removePropertyFromFilter(filterName)
				this.change()
				return
			}

			let ids = []
			this[kind].forEach(u => {
				ids.push(u.id)
			})

			this.filters[filterName] = ids.join(',')
			this.setSingleValueFilter(filterName, filterName, '', 'in')
		},
		findLabels(query) {
			this.labelQuery = query
		},
		addLabel() {
			this.$nextTick(() => {
				this.changeLabelFilter()
			})
		},
		removeLabel(label) {
			this.$nextTick(() => {
				for (const l in this.labels) {
					if (this.labels[l].id === label.id) {
						this.labels.splice(l, 1)
					}
					break
				}

				this.changeLabelFilter()
			})
		},
		changeLabelFilter() {
			if (this.labels.length === 0) {
				this.removePropertyFromFilter('labels')
				this.change()
				return
			}

			let labelIDs = []
			this.labels.forEach(u => {
				labelIDs.push(u.id)
			})

			this.filters.labels = labelIDs.join(',')
			this.setSingleValueFilter('labels', 'labels', '', 'in')
		},
	},
}
</script>

<style lang="scss">
.single-value-control {
	display: flex;
	align-items: center;

	.fancycheckbox {
		margin-left: .5rem;
	}
}
</style>
