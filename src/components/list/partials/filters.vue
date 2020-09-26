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
		</div>
	</div>
</template>

<script>
import Fancycheckbox from '../../input/fancycheckbox'
import flatPickr from 'vue-flatpickr-component'
import 'flatpickr/dist/flatpickr.css'

import {formatISO} from 'date-fns'
import PrioritySelect from '@/components/tasks/partials/prioritySelect'
import PercentDoneSelect from '@/components/tasks/partials/percentDoneSelect'

export default {
	name: 'filters',
	components: {
		PrioritySelect,
		Fancycheckbox,
		flatPickr,
		PercentDoneSelect,
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
			},
			flatPickerConfig: {
				altFormat: 'j M Y H:i',
				altInput: true,
				dateFormat: 'Y-m-d H:i',
				enableTime: true,
				time_24hr: true,
				mode: 'range',
			},
		}
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
			this.prepareDueDate()
			this.prepareStartDate()
			this.prepareEndDate()
			this.preparePriority()
			this.preparePercentDone()
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
		setSingleValueFilter(filterName, variableName, useVariableName) {
			if (!this.filters[useVariableName]) {
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
				this.params.filter_comparator.push('equals')
				this.params.filter_value.push(this.filters[variableName])
			}

			this.change()
		},
		prepareSingleValue(filterName, variableName, useVariableName, isNumber = false) {
			let found = false
			for (const i in this.params.filter_by) {
				if (this.params.filter_by[i] === filterName) {
					found = i
					break
				}
			}

			if (found === false) {
				this.filters[useVariableName] = false
				return
			}

			if (isNumber) {
				this.filters[variableName] = Number(this.params.filter_value[found])
			} else {
				this.filters[variableName] = this.params.filter_value[found]
			}

			this.filters[useVariableName] = true
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
		prepareDueDate() {
			this.prepareDate('due_date', 'dueDate')
		},
		preparePriority() {
			this.prepareSingleValue('priority', 'priority', 'usePriority', true)
		},
		prepareStartDate() {
			this.prepareDate('start_date', 'startDate')
		},
		prepareEndDate() {
			this.prepareDate('end_date', 'endDate')
		},
		preparePercentDone() {
			this.prepareSingleValue('percent_done', 'percentDone', 'usePercentDone', true)
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
