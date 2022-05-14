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
						:disabled="disabled || undefined"
						@change="updateData"
						class="input"
						:placeholder="$t('task.repeat.specifyAmount')"
						v-model="repeatAfter.amount"
						type="number"
					/>
				</div>
				<div class="control">
					<div class="select">
						<select
							v-model="repeatAfter.type"
							@change="updateData"
							:disabled="disabled || undefined"
						>
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

<script setup lang="ts">
import {ref, reactive, watch} from 'vue'
import repeatModes from '@/models/constants/taskRepeatModes'
import TaskModel from '@/models/task'

const props = defineProps({
	modelValue: {
		default: () => {},
		required: true,
	},
	disabled: {
		type: Boolean,
		default: false,
	},
})

const emit = defineEmits(['update:modelValue', 'change'])

const task = ref<TaskModel>({})
const repeatAfter = reactive({
	amount: 0,
	type: '',
})

watch(
	() => props.modelValue,
	(value) => {
		task.value = value
		if (typeof value.repeatAfter !== 'undefined') {
			Object.assign(repeatAfter, value.repeatAfter)
		}
	},
	{immediate: true},
)

function updateData() {
	if (task.value.repeatMode !== repeatModes.REPEAT_MODE_DEFAULT && repeatAfter.amount === 0) {
		return
	}
	
	Object.assign(task.value.repeatAfter, repeatAfter)
	emit('update:modelValue', task.value)
	emit('change')
}

function setRepeatAfter(amount: number, type) {
	Object.assign(repeatAfter, { amount, type})
	updateData()
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