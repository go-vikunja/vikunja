<template>
	<div class="control repeat-after-input">
		<div class="buttons has-addons is-centered mt-2">
			<button class="button is-small" @click="() => setRepeatAfter(1, 'days')">Every Day</button>
			<button class="button is-small" @click="() => setRepeatAfter(1, 'weeks')">Every Week</button>
			<button class="button is-small" @click="() => setRepeatAfter(1, 'months')">Every Month</button>
		</div>
		<div class="columns is-align-items-center">
			<div class="is-flex column">
				<p class="pr-4">
					Each
				</p>
				<div class="field has-addons is-fullwidth">
					<div class="control">
						<input
							:disabled="disabled"
							@change="updateData"
							class="input"
							placeholder="Specify an amount..."
							v-model="repeatAfter.amount"/>
					</div>
					<div class="control">
						<div class="select">
							<select :disabled="disabled" @change="updateData" v-model="repeatAfter.type">
								<option value="hours">Hours</option>
								<option value="days">Days</option>
								<option value="weeks">Weeks</option>
								<option value="months">Months</option>
								<option value="years">Years</option>
							</select>
						</div>
					</div>
				</div>
			</div>
			<fancycheckbox
				:disabled="disabled"
				@change="updateData"
				class="column"
				v-model="task.repeatFromCurrentDate"
				v-tooltip="'When marking the task as done, all dates will be set relative to the current date rather than the date they had before.'"
			>
				Repeat from current date
			</fancycheckbox>
		</div>
	</div>
</template>

<script>
import Fancycheckbox from '../../input/fancycheckbox'

export default {
	name: 'repeatAfter',
	components: {Fancycheckbox},
	data() {
		return {
			task: {},
			repeatAfter: {
				amount: 0,
				type: '',
			},
		}
	},
	props: {
		value: {
			default: () => {
			},
			required: true,
		},
		disabled: {
			default: false,
		},
	},
	watch: {
		value(newVal) {
			this.task = newVal
			if (typeof newVal.repeatAfter !== 'undefined') {
				this.repeatAfter = newVal.repeatAfter
			}
		},
	},
	mounted() {
		this.task = this.value
		if (typeof this.value.repeatAfter !== 'undefined') {
			this.repeatAfter = this.value.repeatAfter
		}
	},
	methods: {
		updateData() {
			this.task.repeatAfter = this.repeatAfter
			this.$emit('input', this.task)
			this.$emit('change')
		},
		setRepeatAfter(amount, type) {
			this.repeatAfter.amount = amount
			this.repeatAfter.type = type
			this.updateData()
		},
	},
}
</script>

<style lang="scss" scoped>
p {
	padding-top: 6px;
}

.input {
	min-width: 2rem;
}

.fancycheckbox {
	padding: 0;
}
</style>