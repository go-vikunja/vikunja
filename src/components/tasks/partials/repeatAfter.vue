<template>
	<div class="control repeat-after-input">
		<div class="buttons has-addons is-centered mt-2">
			<x-button type="secondary" class="is-small" @click="() => setRepeatAfter(1, 'days')">Every Day</x-button>
			<x-button type="secondary" class="is-small" @click="() => setRepeatAfter(1, 'weeks')">Every Week</x-button>
			<x-button type="secondary" class="is-small" @click="() => setRepeatAfter(1, 'months')">Every Month</x-button>
		</div>
		<div class="is-flex is-align-items-center mb-2">
			<label for="repeatMode" class="is-fullwidth">
				Repeat mode:
			</label>
			<div class="control">
				<div class="select">
					<select @change="updateData" v-model="task.repeatMode" id="repeatMode">
						<option :value="repeatModes.REPEAT_MODE_DEFAULT">Default</option>
						<option :value="repeatModes.REPEAT_MODE_MONTH">Monthly</option>
						<option :value="repeatModes.REPEAT_MODE_FROM_CURRENT_DATE">From Current Date</option>
					</select>
				</div>
			</div>
		</div>
		<div class="is-flex" v-if="task.repeatMode !== repeatModes.REPEAT_MODE_MONTH">
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
						v-model="repeatAfter.amount"
						type="number"
					/>
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
	</div>
</template>

<script>
import repeatModes from '@/models/taskRepeatModes'

export default {
	name: 'repeatAfter',
	data() {
		return {
			task: {},
			repeatAfter: {
				amount: 0,
				type: '',
			},
			repeatModes: repeatModes,
		}
	},
	props: {
		value: {
			default: () => {},
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
</style>