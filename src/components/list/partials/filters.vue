<template>
	<div class="card filters">
		<div class="card-content">
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
							class="input"
							:config="flatPickerConfig"
							placeholder="Due Date Range"
							v-model="filters.dueDate"
							@on-close="setDueDateFilter"
					/>
				</div>
			</div>
		</div>
	</div>
</template>

<script>
	import Fancycheckbox from '../../input/fancycheckbox'
	import flatPickr from 'vue-flatpickr-component'
	import 'flatpickr/dist/flatpickr.css'

	export default {
		name: 'filters',
		components: {
			Fancycheckbox,
			flatPickr,
		},
		data() {
			return {
				params: {
					sort_by: [],
					order_by: [],
					filter_by: [],
					filter_value: [],
					filter_comparator: [],
				},
				filters: {
					done: false,
					dueDate: '',
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
			this.prepareDone()
		},
		props: {
			value: {
				required: true,
			}
		},
		watch: {
			value(newVal) {
				this.$set(this, 'params', newVal)
				this.prepareDone()
			}
		},
		methods: {
			change() {
				this.$emit('input', this.params)
				this.$emit('change', this.params)
			},
			prepareDone() {
				// Set filters.done based on params
				if(typeof this.params.filter_by !== 'undefined') {
					let foundDone = false
					this.params.filter_by.forEach((f, i) => {
						if (f === 'done') {
							foundDone = i
						}
					})
					if (foundDone === false) {
						this.filters.done = true
					}
				}
			},
			setDoneFilter() {
				if (this.filters.done) {
					for (const i in this.params.filter_by) {
						if (this.params.filter_by[i] === 'done') {
							this.params.filter_by.splice(i, 1)
							this.params.filter_comparator.splice(i, 1)
							this.params.filter_value.splice(i, 1)
							break
						}
					}
				} else {
					this.params.filter_by.push('done')
					this.params.filter_comparator.push('equals')
					this.params.filter_value.push('false')
				}
				this.change()
			},
			setDueDateFilter() {
				// Only filter if we have a start and end due date
				if (this.filters.dueDate !== '') {

					const parts = this.filters.dueDate.split(' to ')

					if(parts.length < 2) {
						return
					}

					// Check if we already have values in params and only update them if we do
					let foundStart = false
					let foundEnd = false
					this.params.filter_by.forEach((f, i) => {
						if (f === 'due_date' && this.params.filter_comparator[i] === 'greater_equals') {
							foundStart = true
							this.params.filter_value[i] = +new Date(parts[0]) / 1000
						}
						if (f === 'due_date' && this.params.filter_comparator[i] === 'less_equals') {
							foundEnd = true
							this.params.filter_value[i] = +new Date(parts[1]) / 1000
						}
					})

					if (!foundStart) {
						this.params.filter_by.push('due_date')
						this.params.filter_comparator.push('greater_equals')
						this.params.filter_value.push(+new Date(parts[0]) / 1000)
					}
					if (!foundEnd) {
						this.params.filter_by.push('due_date')
						this.params.filter_comparator.push('less_equals')
						this.params.filter_value.push(+new Date(parts[1]) / 1000)
					}
					this.change()
				}
			},
		},
	}
</script>
