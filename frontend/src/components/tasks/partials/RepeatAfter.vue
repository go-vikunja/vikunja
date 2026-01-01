<template>
	<div class="control repeat-after-input">
		<div class="buttons has-addons is-centered mbs-2">
			<XButton
				variant="secondary"
				class="is-small"
				@click="() => setQuickRepeat('FREQ=DAILY;INTERVAL=1')"
			>
				{{ $t('task.repeat.everyDay') }}
			</XButton>
			<XButton
				variant="secondary"
				class="is-small"
				@click="() => setQuickRepeat('FREQ=WEEKLY;INTERVAL=1')"
			>
				{{ $t('task.repeat.everyWeek') }}
			</XButton>
			<XButton
				variant="secondary"
				class="is-small"
				@click="() => setQuickRepeat('FREQ=DAILY;INTERVAL=30')"
			>
				{{ $t('task.repeat.every30d') }}
			</XButton>
			<XButton
				variant="secondary"
				class="is-small"
				@click="() => setQuickRepeat('FREQ=MONTHLY;INTERVAL=1')"
			>
				{{ $t('task.repeat.everyMonth') }}
			</XButton>
			<XButton
				variant="secondary"
				class="is-small"
				@click="() => setQuickRepeat('FREQ=MONTHLY;INTERVAL=3')"
			>
				{{ $t('task.repeat.everyQuarter') }}
			</XButton>
			<XButton
				variant="secondary"
				class="is-small"
				@click="() => setQuickRepeat('FREQ=MONTHLY;INTERVAL=6')"
			>
				{{ $t('task.repeat.everySixMonths') }}
			</XButton>
			<XButton
				variant="secondary"
				class="is-small"
				@click="() => setQuickRepeat('FREQ=YEARLY;INTERVAL=1')"
			>
				{{ $t('task.repeat.everyYear') }}
			</XButton>
		</div>
		<div class="is-flex is-align-items-center mbe-2">
			<label class="is-fullwidth">
				<input
					v-model="repeatsFromCurrentDate"
					type="checkbox"
					@change="updateData"
				>
				{{ $t('task.repeat.fromCurrentDate') }}
			</label>
		</div>
		<div class="is-flex">
			<p class="pis-4">
				{{ $t('task.repeat.each') }}
			</p>
			<div class="field has-addons is-fullwidth">
				<div class="control">
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
				<div class="control">
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
			class="is-flex is-align-items-center mbe-2"
		>
			<label
				for="repeatDay"
				class="is-fullwidth"
			>
				{{ $t('task.repeat.onDay') }}:
			</label>
			<div class="control">
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
	parseRRule,
	repeatSettingsToRRule,
	rruleFreqToUiFreq,
	type RepeatFrequency,
} from '@/helpers/rrule'

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

		// Parse the existing RRULE if present
		if (value.repeats) {
			const parsed = parseRRule(value.repeats)
			if (parsed) {
				repeatInterval.value = parsed.interval
				repeatFrequency.value = rruleFreqToUiFreq(parsed.freq)
				repeatByMonthDay.value = parsed.bymonthday || 0
			}
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

	// Generate the RRULE string
	const bymonthday = repeatFrequency.value === 'months' && repeatByMonthDay.value > 0
		? repeatByMonthDay.value
		: undefined
	const rrule = repeatSettingsToRRule(repeatInterval.value, repeatFrequency.value, bymonthday)

	task.value.repeats = rrule
	task.value.repeatsFromCurrentDate = repeatsFromCurrentDate.value

	emit('update:modelValue', task.value)
}

function setQuickRepeat(rrule: string) {
	if (!task.value) {
		return
	}

	// Parse and update local state
	const parsed = parseRRule(rrule)
	if (parsed) {
		repeatInterval.value = parsed.interval
		repeatFrequency.value = rruleFreqToUiFreq(parsed.freq)
		repeatByMonthDay.value = parsed.bymonthday || 0
	}

	task.value.repeats = rrule
	emit('update:modelValue', task.value)
}
</script>

<style lang="scss" scoped>
p {
	padding-block-start: 6px;
}

.input {
	min-inline-size: 2rem;
}
</style>
