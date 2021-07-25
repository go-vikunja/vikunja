<template>
	<div class="reminders">
		<div
			v-for="(r, index) in reminders"
			:key="index"
			:class="{ 'overdue': r < new Date()}"
			class="reminder-input"
		>
			<datepicker
				v-model="reminders[index]"
				:disabled="disabled"
				@close-on-change="() => addReminderDate(index)"
			/>
			<a @click="removeReminderByIndex(index)" v-if="!disabled" class="remove">
				<icon icon="times"></icon>
			</a>
		</div>
		<div class="reminder-input" v-if="!disabled">
			<datepicker
				v-model="newReminder"
				@close-on-change="() => addReminderDate()"
				:choose-date-label="$t('task.addReminder')"
			/>
		</div>
	</div>
</template>

<script>
import datepicker from '@/components/input/datepicker.vue'

export default {
	name: 'reminders',
	data() {
		return {
			newReminder: null,
			reminders: [],
		}
	},
	props: {
		value: {
			default: () => [],
			validator: prop => {
				// This allows arrays of Dates and strings
				if (!(prop instanceof Array)) {
					return false
				}

				for (const e of prop) {
					const isDate = e instanceof Date
					const isString = typeof e === 'string'
					if (!isDate && !isString) {
						console.log('validation failed', e, e instanceof Date)
						return false
					}
				}

				return true
			},
		},
		disabled: {
			default: false,
		},
	},
	components: {
		datepicker,
	},
	mounted() {
		this.reminders = this.value
	},
	watch: {
		value(newVal) {
			for (const i in newVal) {
				if (typeof newVal[i] === 'string') {
					newVal[i] = new Date(newVal[i])
				}
			}
			this.reminders = newVal
		},
	},
	methods: {
		updateData() {
			this.$emit('input', this.reminders)
			this.$emit('change')
		},
		addReminderDate(index = null) {
			// New Date
			if (index === null) {
				if (this.newReminder === null) {
					return
				}
				this.reminders.push(new Date(this.newReminder))
				this.newReminder = null
			} else if(this.reminders[index] === null) {
				return
			}

			this.updateData()
		},
		removeReminderByIndex(index) {
			this.reminders.splice(index, 1)
			this.updateData()
		},
	},
}
</script>
