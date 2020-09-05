<template>
	<div :class="{'is-loading': taskService.loading}" class="defer-task loading-container">
		<label class="label">Defer due date</label>
		<div class="defer-days">
			<button @click="() => deferDays(1)" class="button is-outlined is-primary has-no-shadow">1 day</button>
			<button @click="() => deferDays(3)" class="button is-outlined is-primary has-no-shadow">3 days</button>
			<button @click="() => deferDays(7)" class="button is-outlined is-primary has-no-shadow">1 week</button>
		</div>
		<flat-pickr
			:class="{ 'disabled': taskService.loading}"
			:config="flatPickerConfig"
			:disabled="taskService.loading"
			class="input"
			v-model="dueDate"
		/>
	</div>
</template>

<script>
import TaskService from '../../../services/task'
import flatPickr from 'vue-flatpickr-component'

export default {
	name: 'defer-task',
	data() {
		return {
			taskService: TaskService,
			task: null,
			// We're saving the due date seperately to prevent null errors in very short periods where the task is null.
			dueDate: null,
			lastValue: null,
			changeInterval: null,

			flatPickerConfig: {
				altFormat: 'j M Y H:i',
				altInput: true,
				dateFormat: 'Y-m-d H:i',
				enableTime: true,
				time_24hr: true,
				inline: true,
			},
		}
	},
	components: {
		flatPickr,
	},
	props: {
		value: {
			required: true,
		},
	},
	created() {
		this.taskService = new TaskService()
	},
	mounted() {
		this.task = this.value
		this.dueDate = this.task.dueDate
		this.lastValue = this.dueDate

		// Because we don't really have other ways of handling change since if we let flatpickr
		// change events trigger updates, it would trigger a flatpickr change event which would trigger
		// an update which would trigger a change event and so on...
		// This is either a bug in flatpickr or in the vue component of it.
		// To work around that, we're only updating if something changed and check each second and when closing the popup.
		if (this.changeInterval) {
			clearInterval(this.changeInterval)
		}

		this.changeInterval = setInterval(this.updateDueDate, 1000)
	},
	beforeDestroy() {
		if (this.changeInterval) {
			clearInterval(this.changeInterval)
		}
		this.updateDueDate()
	},
	watch: {
		value(newVal) {
			this.task = newVal
			this.dueDate = this.task.dueDate
			this.lastValue = this.dueDate
		},
	},
	methods: {
		deferDays(days) {
			this.dueDate = new Date(this.dueDate)
			this.dueDate = this.dueDate.setDate(this.dueDate.getDate() + days)
			this.updateDueDate()
		},
		updateDueDate() {
			if (!this.dueDate) {
				return
			}

			if (+new Date(this.dueDate) === +this.lastValue) {
				return
			}

			this.task.dueDate = new Date(this.dueDate)
			this.taskService.update(this.task)
				.then(r => {
					this.lastValue = r.dueDate
					this.task = r
					this.$emit('input', r)
				})
				.catch(e => {
					this.error(e, this)
				})
		},
	},
}
</script>
