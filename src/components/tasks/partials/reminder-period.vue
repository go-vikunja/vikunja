<template>
	<div class="datepicker">
		<BaseButton class="show" v-if="!!reminder?.relativeTo" @click.stop="togglePeriodPopup"
								:disabled="disabled || undefined">
			{{ formatDuration(reminder.relativePeriod) }} <span v-html="formatBeforeAfter(reminder.relativePeriod)"></span>
			{{ formatRelativeTo(reminder.relativeTo) }}
		</BaseButton>
		<CustomTransition name="fade">
			<div v-if="show" class="control is-flex is-align-items-center mb-2">
				<input
						:disabled="disabled || undefined"
						class="input"
						placeholder="d"
						v-model="periodInput.duration.days"
						type="number"
						min="0"
				/> d
				<input
						:disabled="disabled || undefined"
						class="input"
						placeholder="HH"
						v-model="periodInput.duration.hours"
						type="number"
						min="0"
				/>:
				<input
						:disabled="disabled || undefined"
						class="input"
						placeholder="MM"
						v-model="periodInput.duration.minutes"
						type="number"
						min="0"
				/>
				<div class="select">
					<select v-model="periodInput.sign" id="sign">
						<option value="-1">&le;</option>
						<option value="1">&gt;</option>
					</select>
				</div>
				<div class="control">
					<div class="select">
						<select v-model="periodInput.relativeTo" id="relativeTo">
							<option :value="REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE">{{ $t('task.attributes.dueDate') }}</option>
							<option :value="REMINDER_PERIOD_RELATIVE_TO_TYPES.STARTDATE">{{
									$t('task.attributes.startDate')
								}}
							</option>
							<option :value="REMINDER_PERIOD_RELATIVE_TO_TYPES.ENDDATE">{{ $t('task.attributes.endDate') }}</option>
						</select>
					</div>
				</div>

				<x-button
						class="datepicker__close-button"
						:shadow="false"
						@click="close"
						v-cy="'closeDatepicker'"
				>
					{{ $t('misc.confirm') }}
				</x-button>

			</div>
		</CustomTransition>
	</div>
</template>

<script setup lang="ts">
import { reactive, ref, watch, type PropType } from 'vue'

import BaseButton from '@/components/base/BaseButton.vue'
import CustomTransition from '@/components/misc/CustomTransition.vue'
import { periodToSeconds, secondsToPeriod } from '@/helpers/time/period'
import TaskReminderModel from '@/models/taskReminder'
import type { ITaskReminder } from '@/modelTypes/ITaskReminder'
import { REMINDER_PERIOD_RELATIVE_TO_TYPES, type IReminderPeriodRelativeTo } from '@/types/IReminderPeriodRelativeTo'
import { useI18n } from 'vue-i18n'

const {t} = useI18n({useScope: 'global'})

const props = defineProps({
	modelValue: {
		type: Object as PropType<ITaskReminder>,
		required: false,
	},
	disabled: {
		type: Boolean,
		default: false,
	},
})

const emit = defineEmits(['update:modelValue', 'close', 'close-on-change'])

const reminder = ref<ITaskReminder>()
const show = ref(false)

const periodInput = reactive({
	duration: {days: 0, hours: 0, minutes: 0, seconds: 0},
	relativeTo: REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE,
	sign: -1,
})


watch(
		() => props.modelValue,
		(value) => {
			console.log('reminders-period.watch', value)
			reminder.value = value
			if (value && value.relativeTo != null) {
				Object.assign(periodInput.duration, secondsToPeriod(Math.abs(value.relativePeriod)))
				periodInput.relativeTo = value.relativeTo
				periodInput.sign = value.relativePeriod <= 0 ? -1 : 1
			} else {
				reminder.value = new TaskReminderModel()
				show.value = true
			}
		},
		{immediate: true},
)


function updateData() {
	changed.value = true
	reminder.value.relativePeriod = parseInt(periodInput.sign) * periodToSeconds(periodInput.duration.days, periodInput.duration.hours, periodInput.duration.minutes, 0)
	reminder.value.relativeTo = periodInput.relativeTo
	reminder.value.reminder = null
	console.log('reminders-period.updateData', reminder.value)
	emit('update:modelValue', reminder.value)
}

function togglePeriodPopup() {
	if (props.disabled) {
		return
	}

	show.value = !show.value
}

const changed = ref(false)

function close() {
	// Kind of dirty, but the timeout allows us to enter a time and click on "confirm" without
	// having to click on another input field before it is actually used.
	updateData()
	setTimeout(() => {
		show.value = false
		emit('close', changed.value)
		if (changed.value) {
			changed.value = false
			emit('close-on-change', changed.value)
		}
	}, 200)
}


function formatDuration(reminderPeriod: number): string {
	if (Math.abs(reminderPeriod) < 60) {
		return '00:00'
	}
	const duration = secondsToPeriod(Math.abs(reminderPeriod))
	return (duration.days > 0 ? duration.days + ' d ' : '') +
			('' + duration.hours).padStart(2, '0') + ':' +
			('' + duration.minutes).padStart(2, '0')
}

function formatBeforeAfter(reminderPeriod: number): string {
	if (reminderPeriod <= 0) {
		return '&le;'
	}
	return '&gt;'
}

function formatRelativeTo(relativeTo: IReminderPeriodRelativeTo | null): string | null {
	switch (relativeTo) {
		case REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE:
			return t('task.attributes.dueDate')
		case REMINDER_PERIOD_RELATIVE_TO_TYPES.STARTDATE:
			return t('task.attributes.startDate')
		case REMINDER_PERIOD_RELATIVE_TO_TYPES.ENDDATE:
			return t('task.attributes.endDate')
		default:
			return relativeTo
	}
}
</script>

<style lang="scss" scoped>
.reminders {
	.reminder-input {
		display: flex;
		align-items: center;

		&.overdue :deep(.datepicker .show) {
			color: var(--danger);
		}

		&:last-child {
			margin-bottom: 0.75rem;
		}

		.remove {
			color: var(--danger);
			padding-left: .5rem;
		}
	}
}

.input {
	max-width: 70px;
	width: 70px;
}

.datepicker__close-button {
	margin: 1rem;
	width: calc(100% - 2rem);
}


</style>