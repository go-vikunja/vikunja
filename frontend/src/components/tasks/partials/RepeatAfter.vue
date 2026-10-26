<template>
	<div class="control repeat-after-input">
		<p
			v-if="isComplex"
			class="repeat-advanced-note"
		>
			{{ $t('task.repeat.advancedReadonly') }}
		</p>
		<div class="button-group">
			<XButton
				variant="secondary"
				class="is-small"
				:disabled="disabled || isComplex"
				@click="() => setQuickRepeat('daily')"
			>
				{{ $t('task.repeat.everyDay') }}
			</XButton>
			<XButton
				variant="secondary"
				class="is-small"
				:disabled="disabled || isComplex"
				@click="() => setQuickRepeat('weekly')"
			>
				{{ $t('task.repeat.everyWeek') }}
			</XButton>
			<XButton
				variant="secondary"
				class="is-small"
				:disabled="disabled || isComplex"
				@click="() => setQuickRepeat('monthly')"
			>
				{{ $t('task.repeat.everyMonth') }}
			</XButton>
			<XButton
				variant="secondary"
				class="is-small"
				:disabled="disabled || isComplex"
				@click="() => setQuickRepeat('yearly')"
			>
				{{ $t('task.repeat.everyYear') }}
			</XButton>
		</div>
		<div class="repeat-from-current-date">
			<FancyCheckbox
				v-model="repeatsFromCurrentDate"
				:disabled="disabled"
				@update:modelValue="updateFromCurrentDate"
			>
				{{ $t('task.repeat.fromCurrentDate') }}
			</FancyCheckbox>
		</div>
		<div
			v-if="!hasUnmappableFreq"
			class="repeat-custom-row"
		>
			<label
				for="repeatInterval"
				class="repeat-custom-label"
			>
				{{ $t('task.repeat.each') }}
			</label>
			<div class="field has-addons repeat-interval-controls">
				<div class="control repeat-interval-amount">
					<input
						id="repeatInterval"
						v-model.number="repeatInterval"
						:disabled="disabled || isComplex || undefined"
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
							:disabled="disabled || isComplex || undefined"
							:aria-label="$t('task.repeat.intervalUnit')"
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
						:disabled="disabled || isComplex || undefined"
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
import {computed, ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'

import {error} from '@/message'

import type {Task as ITask} from '@/client/generated'
import {
	freqToUiFreq,
	isComplexRepeat,
	isMappableFreq,
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
	'update:modelValue': [value: ITask],
}>()

const {t} = useI18n({useScope: 'global'})

const repeatInterval = ref(1)
const repeatFrequency = ref<RepeatFrequency>('days')
const repeatByMonthDay = ref(0)
const repeatsFromCurrentDate = ref(false)

// Editing a rule the simple controls can't represent would drop its advanced parts; it can still be removed.
const isComplex = computed(() => isComplexRepeat(props.modelValue?.repeat))

// A freq without a unit in the dropdown (e.g. MINUTELY) would show a wrong value, so hide the row.
const hasUnmappableFreq = computed(() => props.modelValue?.repeat != null
	&& !isMappableFreq(props.modelValue.repeat.freq))

watch(
	() => props.modelValue,
	(value) => {
		if (!value) {
			return
		}
		repeatsFromCurrentDate.value = value.repeats_from_current_date ?? false
		if (value.repeat) {
			repeatInterval.value = value.repeat.interval || 1
			repeatFrequency.value = freqToUiFreq(value.repeat.freq ?? '')
			repeatByMonthDay.value = value.repeat.by_month_day?.[0] ?? 0
		}
	},
	{
		immediate: true,
		deep: true,
	},
)

function emitTask(changes: Pick<ITask, 'repeat' | 'repeats_from_current_date'>) {
	if (!props.modelValue) {
		return
	}
	emit('update:modelValue', {
		...props.modelValue,
		...changes,
	})
}

// The completion-date flag is orthogonal to the rule: changing it sends the existing rule back
// whole, never one rebuilt from the controls, which would drop parts they don't show. Without a
// rule yet, the controls build one as for any other edit.
function updateFromCurrentDate() {
	if (props.modelValue?.repeat) {
		emitTask({repeats_from_current_date: repeatsFromCurrentDate.value})
		return
	}
	updateData()
}

function updateData() {
	// A complex rule keeps the completion-date flag editable and is sent back whole.
	if (isComplex.value) {
		emitTask({repeats_from_current_date: repeatsFromCurrentDate.value})
		return
	}

	if (repeatInterval.value < 1) {
		error({message: t('task.repeat.invalidAmount')})
		return
	}

	const byMonthDay = repeatFrequency.value === 'months' && repeatByMonthDay.value > 0
		? repeatByMonthDay.value
		: undefined
	emitTask({
		repeat: repeatFromSettings(repeatInterval.value, repeatFrequency.value, byMonthDay) ?? undefined,
		repeats_from_current_date: repeatsFromCurrentDate.value,
	})
}

function setQuickRepeat(freq: string) {
	if (isComplex.value) {
		return
	}

	repeatInterval.value = 1
	repeatFrequency.value = freqToUiFreq(freq)
	repeatByMonthDay.value = 0
	emitTask({
		repeat: {
			freq,
			interval: 1,
		},
		repeats_from_current_date: repeatsFromCurrentDate.value,
	})
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

.repeat-advanced-note {
	margin: 0;
	font-size: .85rem;
	color: var(--grey-500);
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
