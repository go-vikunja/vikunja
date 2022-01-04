<template>
	<div class="control repeat-after-input">
		<div class="buttons has-addons is-centered mt-2">
			<x-button variant="secondary" class="is-small" @click="() => setRepeatAfter(1, 'days')">{{ $t('task.repeat.everyDay') }}</x-button>
			<x-button variant="secondary" class="is-small" @click="() => setRepeatAfter(1, 'weeks')">{{ $t('task.repeat.everyWeek') }}</x-button>
			<x-button variant="secondary" class="is-small" @click="() => setRepeatAfter(1, 'months')">{{ $t('task.repeat.everyMonth') }}</x-button>
		</div>
		<div class="is-flex is-align-items-center mb-2">
			<label for="repeatMode" class="is-fullwidth">
				{{ $t('task.repeat.mode') }}:
			</label>
			<div class="control">
				<div class="select">
					<select @change="updateData" v-model="task.repeatMode" id="repeatMode">
						<option :value="repeatModes.REPEAT_MODE_DEFAULT">{{ $t('misc.default') }}</option>
						<option :value="repeatModes.REPEAT_MODE_MONTH">{{ $t('task.repeat.monthly') }}</option>
						<option :value="repeatModes.REPEAT_MODE_FROM_CURRENT_DATE">{{ $t('task.repeat.fromCurrentDate') }}</option>
					</select>
				</div>
			</div>
		</div>
		<div class="is-flex" v-if="task.repeatMode !== repeatModes.REPEAT_MODE_MONTH">
			<p class="pr-4">
				{{ $t('task.repeat.each') }}
			</p>
			<div class="field has-addons is-fullwidth">
				<div class="control">
					<input
						:disabled="disabled || null"
						@change="updateData"
						class="input"
						:placeholder="$t('task.repeat.specifyAmount')"
						v-model="repeatAfter.amount"
						type="number"
					/>
				</div>
				<div class="control">
					<div class="select">
						<select :disabled="disabled || null" @change="updateData" v-model="repeatAfter.type">
							<option value="hours">{{ $t('task.repeat.hours') }}</option>
							<option value="days">{{ $t('task.repeat.days') }}</option>
							<option value="weeks">{{ $t('task.repeat.weeks') }}</option>
							<option value="months">{{ $t('task.repeat.months') }}</option>
							<option value="years">{{ $t('task.repeat.years') }}</option>
						</select>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>

<script>
import repeatModes from '@/models/constants/taskRepeatModes'

export default {
	name: 'repeatAfter',
	data() {
		return {
			task: {},
			repeatAfter: {
				amount: 0,
				type: '',
			},
			repeatModes,
		}
	},
	props: {
		modelValue: {
			default: () => {},
			required: true,
		},
		disabled: {
			default: false,
		},
	},
	emits: ['update:modelValue', 'change'],
	watch: {
		modelValue: {
			handler(value) {
				this.task = value
				if (typeof value.repeatAfter !== 'undefined') {
					this.repeatAfter = value.repeatAfter
				}
			},
			immediate: true,
		},
	},
	methods: {
		updateData() {
			if (this.task.repeatMode !== repeatModes.REPEAT_MODE_DEFAULT && this.repeatAfter.amount === 0) {
				return
			}
			
			this.task.repeatAfter = this.repeatAfter
			this.$emit('update:modelValue', this.task)
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