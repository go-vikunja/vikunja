<template>
	<div class="card filters">
		<div class="card-content">
			<fancycheckbox v-model="params.filter_include_nulls">
				Include Tasks which don't have a value set
			</fancycheckbox>
			<fancycheckbox
				v-model="filters.requireAllFilters"
				@change="setFilterConcat()"
			>
				Require all filters to be true for a task to show up
			</fancycheckbox>
			<div class="field">
				<label class="label">Show Done Tasks</label>
				<div class="control">
					<fancycheckbox @change="setDoneFilter" v-model="filters.done">
						Show Done Tasks
					</fancycheckbox>
				</div>
			</div>
			<div class="field">
				<label class="label">Priority</label>
				<div class="control single-value-control">
					<priority-select
						:disabled="!filters.usePriority"
						v-model.number="filters.priority"
						@change="setPriority"
					/>
					<fancycheckbox
						v-model="filters.usePriority"
						@change="setPriority"
					>
						Enable Filter By Priority
					</fancycheckbox>
				</div>
			</div>
			<div class="field">
				<label class="label">Percent Done</label>
				<div class="control single-value-control">
					<percent-done-select
						v-model.number="filters.percentDone"
						@change="setPercentDoneFilter"
						:disabled="!filters.usePercentDone"
					/>
					<fancycheckbox
						v-model="filters.usePercentDone"
						@change="setPercentDoneFilter"
					>
						Enable Filter By Percent Done
					</fancycheckbox>
				</div>
			</div>
			<div class="field">
				<label class="label">Due Date</label>
				<div class="control">
					<flat-pickr
						:config="flatPickerConfig"
						@on-close="setDueDateFilter"
						class="input"
						placeholder="Due Date Range"
						v-model="filters.dueDate"
					/>
				</div>
			</div>
			<div class="field">
				<label class="label">Start Date</label>
				<div class="control">
					<flat-pickr
						:config="flatPickerConfig"
						@on-close="setStartDateFilter"
						class="input"
						placeholder="Start Date Range"
						v-model="filters.startDate"
					/>
				</div>
			</div>
			<div class="field">
				<label class="label">End Date</label>
				<div class="control">
					<flat-pickr
						:config="flatPickerConfig"
						@on-close="setEndDateFilter"
						class="input"
						placeholder="End Date Range"
						v-model="filters.endDate"
					/>
				</div>
			</div>
			<div class="field">
				<label class="label">Reminders</label>
				<div class="control">
					<flat-pickr
						:config="flatPickerConfig"
						@on-close="setReminderFilter"
						class="input"
						placeholder="Reminder Date Range"
						v-model="filters.reminders"
					/>
				</div>
			</div>

			<div class="field">
				<label class="label">Assignees</label>
				<div class="control">
					<multiselect
						:clear-on-select="true"
						:close-on-select="true"
						:hide-selected="true"
						:internal-search="true"
						:loading="usersService.loading"
						:multiple="true"
						:options="foundusers"
						:options-limit="300"
						:searchable="true"
						:showNoOptions="false"
						:taggable="false"
						@search-change="query => find('users', query)"
						@select="() => add('users', 'assignees')"
						@remove="() => remove('users', 'assignees')"
						label="username"
						placeholder="Type to search for a user..."
						track-by="id"
						v-model="users"
					>
						<template slot="clear" slot-scope="props">
							<div
								@mousedown.prevent.stop="clear('users', props.search)"
								class="multiselect__clear"
								v-if="users.length"></div>
						</template>
					</multiselect>
				</div>
			</div>

			<div class="field">
				<label class="label">Labels</label>
				<div class="control">
					<multiselect
						:clear-on-select="true"
						:close-on-select="false"
						:hide-selected="true"
						:internal-search="true"
						:loading="labelService.loading"
						:multiple="true"
						:options="foundLabels"
						:options-limit="300"
						:searchable="true"
						:showNoOptions="false"
						@search-change="findLabels"
						@select="label => addLabel(label)"
						label="title"
						placeholder="Type to search for a label..."
						track-by="id"
						v-model="labels"
					>
						<template
							slot="tag"
							slot-scope="{ option }">
							<span
								:style="{'background': option.hexColor, 'color': option.textColor}"
								class="tag mr-2 mb-2">
								<span>{{ option.title }}</span>
								<a @click="removeLabel(option)" class="delete is-small"></a>
							</span>
						</template>
						<template slot="clear" slot-scope="props">
							<div
								@mousedown.prevent.stop="clearLabels(props.search)"
								class="multiselect__clear"
								v-if="labels.length"></div>
						</template>
					</multiselect>
				</div>
			</div>

			<template v-if="$route.name === 'filters.create' || $route.name === 'list.edit'">
				<div class="field">
					<label class="label">Lists</label>
					<div class="control">
						<multiselect
							:clear-on-select="true"
							:close-on-select="true"
							:hide-selected="true"
							:internal-search="true"
							:loading="listsService.loading"
							:multiple="true"
							:options="foundlists"
							:options-limit="300"
							:searchable="true"
							:showNoOptions="false"
							:taggable="false"
							@search-change="query => find('lists', query)"
							@select="() => add('lists', 'list_id')"
							@remove="() => remove('lists', 'list_id')"
							label="title"
							placeholder="Type to search for a list..."
							track-by="id"
							v-model="lists"
						>
							<template slot="clear" slot-scope="props">
								<div
									@mousedown.prevent.stop="clear('lists', props.search)"
									class="multiselect__clear"
									v-if="lists.length"></div>
							</template>
						</multiselect>
					</div>
				</div>
				<div class="field">
					<label class="label">Namespaces</label>
					<div class="control">
						<multiselect
							:clear-on-select="true"
							:close-on-select="true"
							:hide-selected="true"
							:internal-search="true"
							:loading="namespaceService.loading"
							:multiple="true"
							:options="foundnamespace"
							:options-limit="300"
							:searchable="true"
							:showNoOptions="false"
							:taggable="false"
							@search-change="query => find('namespace', query)"
							@select="() => add('namespace', 'namespace')"
							@remove="() => remove('namespace', 'namespace')"
							label="title"
							placeholder="Type to search for a namespace..."
							track-by="id"
							v-model="namespace"
						>
							<template slot="clear" slot-scope="props">
								<div
									@mousedown.prevent.stop="clear('namespace', props.search)"
									class="multiselect__clear"
									v-if="namespace.length"></div>
							</template>
						</multiselect>
					</div>
				</div>
			</template>
		</div>
	</div>
</template>

<script>
import Fancycheckbox from '../../input/fancycheckbox'
import flatPickr from 'vue-flatpickr-component'
import 'flatpickr/dist/flatpickr.css'
import Multiselect from 'vue-multiselect'

import {formatISO} from 'date-fns'
import differenceWith from 'lodash/differenceWith'

import PrioritySelect from '@/components/tasks/partials/prioritySelect'
import PercentDoneSelect from '@/components/tasks/partials/percentDoneSelect'

import UserService from '@/services/user'
import LabelService from '@/services/label'
import ListService from '@/services/list'
import NamespaceService from '@/services/namespace'

export default {
	name: 'filters',
	components: {
		PrioritySelect,
		Fancycheckbox,
		flatPickr,
		PercentDoneSelect,
		Multiselect,
	},
	data() {
		return {
			params: {
				sort_by: [],
				order_by: [],
				filter_by: [],
				filter_value: [],
				filter_comparator: [],
				filter_include_nulls: true,
				filter_concat: 'or',
			},
			filters: {
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
			},
			flatPickerConfig: {
				altFormat: 'j M Y H:i',
				altInput: true,
				dateFormat: 'Y-m-d H:i',
				enableTime: true,
				time_24hr: true,
				mode: 'range',
			},

			usersService: UserService,
			foundusers: [],
			users: [],

			labelService: LabelService,
			foundLabels: [],
			labels: [],

			listsService: ListService,
			foundlists: [],
			lists: [],

			namespaceService: NamespaceService,
			foundnamespace: [],
			namespace: [],
		}
	},
	created() {
		this.usersService = new UserService()
		this.labelService = new LabelService()
		this.listsService = new ListService()
		this.namespaceService = new NamespaceService()
	},
	mounted() {
		this.params = this.value
		this.filters.requireAllFilters = this.params.filter_concat === 'and'
		this.prepareFilters()
	},
	props: {
		value: {
			required: true,
		},
	},
	watch: {
		value(newVal) {
			this.$set(this, 'params', newVal)
			this.prepareFilters()
		},
	},
	methods: {
		change() {
			this.$emit('input', this.params)
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
			this.prepareRelatedObjectFilter('labels', 'labels', 'label')
			this.prepareRelatedObjectFilter('lists', 'list_id')
			this.prepareRelatedObjectFilter('namespace')
		},
		removePropertyFromFilter(propertyName) {
			for (const i in this.params.filter_by) {
				if (this.params.filter_by[i] === propertyName) {
					this.params.filter_by.splice(i, 1)
					this.params.filter_comparator.splice(i, 1)
					this.params.filter_value.splice(i, 1)
					break
				}
			}
		},
		setDateFilter(filterName, variableName) {
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
						this.$set(this.params.filter_value, i, formatISO(new Date(parts[0])))
					}
					if (f === filterName && this.params.filter_comparator[i] === 'less_equals') {
						foundEnd = true
						this.$set(this.params.filter_value, i, formatISO(new Date(parts[1])))
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
			}
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
					this.$set(this.params.filter_value, i, this.filters[variableName])
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

			let foundDone = false
			this.params.filter_by.forEach((f, i) => {
				if (f === 'done') {
					foundDone = i
				}
			})
			if (foundDone === false) {
				this.$set(this.filters, 'done', true)
			}
		},
		prepareRelatedObjectFilter(kind, filterName = null, servicePrefix = null) {
			if (filterName === null) {
				filterName = kind
			}

			if (servicePrefix === null) {
				servicePrefix = kind
			}

			this.prepareSingleValue(filterName)
			if (this.filters[filterName] !== '') {
				this[`${servicePrefix}Service`].getAll({}, {s: this.filters[filterName]})
					.then(r => {
						this.$set(this, kind, r)
					})
					.catch(e => this.error(e, this))
			}
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
			this.$set(this, `found${kind}`, [])
		},
		find(kind, query) {

			if (query === '') {
				this.clear(kind)
			}

			this[`${kind}Service`].getAll({}, {s: query})
				.then(response => {
					// Filter the results to not include users who are already assigneid
					this.$set(this, `found${kind}`, differenceWith(response, this[kind], (first, second) => {
						return first.id === second.id
					}))
				})
				.catch(e => {
					this.error(e, this)
				})
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

			this.$set(this.filters, filterName, ids.join(','))
			this.setSingleValueFilter(filterName, filterName, '', 'in')
		},
		clearLabels() {
			this.$set(this, 'foundLabels', [])
		},
		findLabels(query) {

			if (query === '') {
				this.clearLabels()
			}

			this.labelService.getAll({}, {s: query})
				.then(response => {
					// Filter the results to not include labels already selected
					this.$set(this, 'foundLabels', differenceWith(response, this.labels, (first, second) => {
						return first.id === second.id
					}))
				})
				.catch(e => {
					this.error(e, this)
				})
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

			this.$set(this.filters, 'labels', labelIDs.join(','))
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
