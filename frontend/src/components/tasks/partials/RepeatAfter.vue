<template>
	<div class="control repeat-after-input">
		<div class="button-group">
			<XButton
				variant="secondary"
				class="is-small"
				@click="() => setQuickRepeat('daily', 1)"
			>
				{{ $t('task.repeat.everyDay') }}
			</XButton>
			<XButton
				variant="secondary"
				class="is-small"
				@click="() => setQuickRepeat('weekly', 1)"
			>
				{{ $t('task.repeat.everyWeek') }}
			</XButton>
			<XButton
				variant="secondary"
				class="is-small"
				@click="() => setQuickRepeat('monthly', 1)"
			>
				{{ $t('task.repeat.everyMonth') }}
			</XButton>
			<XButton
				variant="secondary"
				class="is-small"
				@click="() => setQuickRepeat('yearly', 1)"
			>
				{{ $t('task.repeat.everyYear') }}
			</XButton>
		</div>
		<div class="repeat-from-current-date">
			<FancyCheckbox
				v-model="repeatsFromCurrentDate"
				:disabled="disabled"
				@update:modelValue="updateData"
			>
				{{ $t('task.repeat.fromCurrentDate') }}
			</FancyCheckbox>
		</div>
		<div class="repeat-custom-row">
			<span class="repeat-custom-label">
				{{ $t('task.repeat.each') }}
			</span>
			<div class="field has-addons repeat-interval-controls">
				<div class="control repeat-interval-amount">
					<input
						v-model.number="repeatInterval"
						:disabled="disabled || undefined"
						class="input"
						:placeholder="$t('task.repeat.specifyAmount')"
						type="number"
						min="1"
						@change="updateData"
					>
				</div>
				<div class="control repeat-interval-unit">
					<div class="select">
						<select
							v-model="repeatFrequency"
							:disabled="disabled || undefined"
							@change="updateData"
						>
							<option value="hours">
								{{ $t('task.repeat.hours') }}
							</option>
							<option value="days">
								{{ $t('task.repeat.days') }}
							</option>
							<option value="weeks">
								{{ $t('task.repeat.weeks') }}
							</option>
							<option value="months">
								{{ $t('task.repeat.months') }}
							</option>
							<option value="years">
								{{ $t('task.repeat.years') }}
							</option>
						</select>
					</div>
				</div>
			</div>
		</div>
		<div
			v-if="repeatFrequency === 'months'"
			class="repeat-month-day-row"
		>
			<label
				for="repeatDay"
				class="repeat-month-day-label"
			>
				{{ $t('task.repeat.onDay') }}
			</label>
			<div class="control repeat-month-day-control">
				<div class="select">
					<select
						id="repeatDay"
						v-model.number="repeatByMonthDay"
						@change="updateData"
					>
						<option :value="0">
							{{ $t('task.repeat.sameDay') }}
						</option>
						<option
							v-for="day in 31"
							:key="day"
							:value="day"
						>
							{{ day }}
						</option>
					</select>
				</div>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import {ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'

import {error} from '@/message'

import type {ITask} from '@/modelTypes/ITask'
import TaskModel from '@/models/task'
import {
	freqToUiFreq,
	repeatFromSettings,
	type RepeatFrequency,
} from '@/helpers/rrule'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'

const props = withDefaults(defineProps<{
	modelValue: ITask | undefined,
	disabled?: boolean
}>(), {
	disabled: false,
})

const emit = defineEmits<{
	'update:modelValue': [value: ITask | undefined],
}>()

const {t} = useI18n({useScope: 'global'})

const task = ref<ITask>(new TaskModel())
const repeatInterval = ref(1)
const repeatFrequency = ref<RepeatFrequency>('days')
const repeatByMonthDay = ref(0)
const repeatsFromCurrentDate = ref(false)

watch(
	() => props.modelValue,
	(value: ITask | undefined) => {
		if (!value) {
			return
		}
		task.value = value
		repeatsFromCurrentDate.value = value.repeatsFromCurrentDate || false

		// Parse the existing repeat config if present
		if (value.repeat) {
			repeatInterval.value = value.repeat.interval || 1
			repeatFrequency.value = freqToUiFreq(value.repeat.freq)
			repeatByMonthDay.value = value.repeat.byMonthDay?.[0] || 0
		}
	},
	{
		immediate: true,
		deep: true,
	},
)

function updateData() {
	if (!task.value) {
		return
	}

	if (repeatInterval.value < 1) {
		error({message: t('task.repeat.invalidAmount')})
		return
	}

	// Build structured repeat object
	const bymonthday = repeatFrequency.value === 'months' && repeatByMonthDay.value > 0
		? repeatByMonthDay.value
		: undefined
	task.value.repeat = repeatFromSettings(repeatInterval.value, repeatFrequency.value, bymonthday)
	task.value.repeatsFromCurrentDate = repeatsFromCurrentDate.value

	emit('update:modelValue', task.value)
}

function setQuickRepeat(freq: string, interval: number) {
	if (!task.value) {
		return
	}

	// Update local state
	repeatInterval.value = interval
	repeatFrequency.value = freqToUiFreq(freq)
	repeatByMonthDay.value = 0

	task.value.repeat = {freq, interval}
	task.value.repeatsFromCurrentDate = repeatsFromCurrentDate.value
	emit('update:modelValue', task.value)
}
</script>

<style lang="scss" scoped>
.repeat-after-input {
	container-type: inline-size;
	display: flex;
	flex-direction: column;
	gap: .5rem;
	margin-block-start: .25rem;
	max-inline-size: 100%;
}

.button-group {
	display: flex;
	flex-wrap: wrap;
	gap: .25rem;
}

.repeat-from-current-date {
	display: flex;
	align-items: center;
	min-block-size: 1.25rem;
}

.repeat-custom-row,
.repeat-month-day-row {
	display: flex;
	align-items: center;
	gap: .5rem;
	min-inline-size: 0;
}

.repeat-custom-label,
.repeat-month-day-label {
	display: flex;
	align-items: center;
	flex: 0 0 auto;
	min-block-size: 2.25rem;
	line-height: 1.2;
	white-space: nowrap;
}

.repeat-interval-controls {
	display: flex;
	flex: 0 1 12.5rem;
	max-inline-size: 12.5rem;
	min-inline-size: 0;
	margin-block-end: 0;
}

.repeat-interval-amount {
	flex: 0 0 4rem;
}

.repeat-interval-unit {
	flex: 0 1 8.5rem;
	min-inline-size: 7.5rem;
	max-inline-size: 8.5rem;
}

.repeat-month-day-control {
	flex: 1 1 4.5rem;
	min-inline-size: 4.5rem;
	max-inline-size: 8rem;
}

.input {
	min-inline-size: 0;
}

.repeat-interval-amount .input,
.repeat-interval-unit .select,
.repeat-interval-unit select,
.repeat-month-day-control .select,
.repeat-month-day-control select {
	inline-size: 100%;
}

@container (max-width: 15rem) {
	.repeat-interval-controls {
		display: grid;
		grid-template-columns: minmax(0, 1fr);
		gap: .25rem;
	}

	.repeat-interval-controls > .control {
		margin-inline-end: 0;
	}

	.repeat-interval-amount {
		flex: none;
		inline-size: 4rem;
	}

	.repeat-interval-unit {
		flex: none;
		inline-size: 100%;
	}

	.repeat-interval-amount .input {
		border-start-end-radius: $radius;
		border-end-end-radius: $radius;
	}

	.repeat-interval-unit .select select {
		border-start-start-radius: $radius;
		border-end-start-radius: $radius;
	}
}
</style>
