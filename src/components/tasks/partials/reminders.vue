<template>
	<div class="reminders">
		<div
			:class="{ 'overdue': (r < nowUnix && index !== (reminders.length - 1))}"
			:key="index"
			class="reminder-input"
			v-for="(r, index) in reminders">
			<flat-pickr
				:config="flatPickerConfig"
				:data-index="index"
				:disabled="disabled"
				:id="'taskreminderdate' + index"
				:v-model="reminders"
				:value="r"
				placeholder="Add a new reminder..."
			>
			</flat-pickr>
			<a @click="removeReminderByIndex(index)" v-if="index !== (reminders.length - 1) && !disabled">
				<icon icon="times"></icon>
			</a>
		</div>
	</div>
</template>

<script>
import flatPickr from 'vue-flatpickr-component'
import 'flatpickr/dist/flatpickr.css'

export default {
	name: 'reminders',
	data() {
		return {
			reminders: [],
			lastReminder: 0,
			nowUnix: new Date(),
			flatPickerConfig: {
				altFormat: 'j M Y H:i',
				altInput: true,
				dateFormat: 'Y-m-d H:i',
				enableTime: true,
				onOpen: this.updateLastReminderDate,
				onClose: this.addReminderDate,
			},
		}
	},
	props: {
		value: {
			default: () => [],
			type: Array,
		},
		disabled: {
			default: false,
		},
	},
	components: {
		flatPickr,
	},
	mounted() {
		this.reminders = this.value
	},
	watch: {
		value(newVal) {
			this.reminders = newVal
		},
	},
	methods: {
		updateData() {
			this.$emit('input', this.reminders)
			this.$emit('change')
		},
		updateLastReminderDate(selectedDates) {
			this.lastReminder = +new Date(selectedDates[0])
		},
		addReminderDate(selectedDates, dateStr, instance) {
			let newDate = +new Date(selectedDates[0])

			// Don't update if nothing changed
			if (newDate === this.lastReminder) {
				return
			}

			let index = parseInt(instance.input.dataset.index)
			this.reminders[index] = newDate

			let lastIndex = this.reminders.length - 1
			// put a new null at the end if we changed something
			if (lastIndex === index && !isNaN(newDate)) {
				this.reminders.push(null)
			}

			this.updateData()
		},
		removeReminderByIndex(index) {
			this.reminders.splice(index, 1)
			// Reset the last to 0 to have the "add reminder" button
			this.reminders[this.reminders.length - 1] = null

			this.updateData()
		},
	},
}
</script>
