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
				:value="r"
			/>
			<a @click="removeReminderByIndex(index)" v-if="!disabled">
				<icon icon="times"></icon>
			</a>
		</div>
		<div class="reminder-input" v-if="showNewReminder">
			<flat-pickr
				:config="flatPickerConfig"
				:disabled="disabled"
				:value="null"
				placeholder="Add a new reminder..."
			/>
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
			showNewReminder: true,
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
			const newDate = +new Date(selectedDates[0])

			// Don't update if nothing changed
			if (newDate === this.lastReminder) {
				return
			}

			// No date selected
			if (isNaN(newDate)) {
				return
			}

			const index = parseInt(instance.input.dataset.index)
			if (isNaN(index)) {
				this.reminders.push(newDate)
				// This is a workaround to recreate the flatpicker instance which essentially resets it.
				// Even though flatpickr itself has a reset event, the Vue component does not expose it.
				this.showNewReminder = false
				this.$nextTick(() => this.showNewReminder = true)
			} else {
				this.reminders[index] = newDate
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
