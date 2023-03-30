<template>
	<div
		v-if="!!reminder?.relativeTo"	
		class="reminder-period"
	>
		<Popup>
			<template #trigger="{toggle}">
				<BaseButton
					@click="toggle"
					:disabled="disabled"
					class="show"
				>
					{{ formatDuration(reminder.relativePeriod) }} {{ reminder.relativePeriod <= 0 ? '&le;' : '&gt;' }}
					{{ formatRelativeTo(reminder.relativeTo) }}
					<span class="icon"><icon icon="chevron-down"/></span>
				</BaseButton>
			</template>

			<template #content>
				<div class="mt-2">
					<div class="control is-flex is-align-items-center">
						<label>
							<input
								:disabled="disabled"
								class="input"
								:placeholder="$t('task.reminder.daysShort')"
								v-model="periodInput.duration.days"
								type="number"
								min="0"
							/> {{ $t('task.reminder.days') }}
						</label>
						<input
							:disabled="disabled"
							class="input"
							:placeholder="$t('task.reminder.hoursShort')"
							v-model="periodInput.duration.hours"
							type="number"
							min="0"
						/>:
						<input
							:disabled="disabled"
							class="input"
							:placeholder="$t('task.reminder.minutesShort')"
							v-model="periodInput.duration.minutes"
							type="number"
							min="0"
						/>

						<div class="select">
							<select :disabled="disabled" v-model.number="periodInput.sign">
								<option value="-1">&le;</option>
								<option value="1">&gt;</option>
							</select>
						</div>

						<div class="select">
							<select :disabled="disabled" v-model="periodInput.relativeTo">
								<option :value="REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE">{{ $t('task.attributes.dueDate') }}</option>
								<option :value="REMINDER_PERIOD_RELATIVE_TO_TYPES.STARTDATE">{{ $t('task.attributes.startDate')}}</option>
								<option :value="REMINDER_PERIOD_RELATIVE_TO_TYPES.ENDDATE">{{ $t('task.attributes.endDate') }}</option>
							</select>
						</div>
					</div>

					<div class="control">
						<x-button
							:disabled="disabled"
							class="close-button"
							:shadow="false"
							@click="submitForm"
						>
							{{ $t('misc.confirm') }}
						</x-button>
					</div>
				</div>
			</template>
		</Popup>
	</div>
</template>

<script setup lang="ts">
import {reactive, ref, watch, type PropType, computed} from 'vue'
import {useI18n} from 'vue-i18n'

import BaseButton from '@/components/base/BaseButton.vue'
import Popup from '@/components/misc/popup.vue'

import {periodToSeconds, secondsToPeriod} from '@/helpers/time/period'

import TaskReminderModel from '@/models/taskReminder'

import type {ITaskReminder} from '@/modelTypes/ITaskReminder'
import {REMINDER_PERIOD_RELATIVE_TO_TYPES, type IReminderPeriodRelativeTo} from '@/types/IReminderPeriodRelativeTo'

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

const emit = defineEmits(['update:modelValue'])

const reminder = ref<ITaskReminder>()

const periodInput = reactive({
	duration: {
		days: 0,
		hours: 0,
		minutes: 0,
		seconds: 0
	},
	relativeTo: REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE,
	sign: -1,
})

watch(
		props.modelValue,
		(value) => {
			reminder.value = value
			if (value && value.relativeTo != null) {
				Object.assign(periodInput.duration, secondsToPeriod(Math.abs(value.relativePeriod)))
				periodInput.relativeTo = value.relativeTo
				periodInput.sign = value.relativePeriod <= 0 ? -1 : 1
			} else {
				reminder.value = new TaskReminderModel()
				isShowForm.value = true
			}
		},
		{immediate: true},
)


function updateData() {
	changed.value = true
	if (reminder.value) {
		reminder.value.relativePeriod = periodInput.sign * periodToSeconds(periodInput.duration.days, periodInput.duration.hours, periodInput.duration.minutes, 0)
		reminder.value.relativeTo = periodInput.relativeTo
		reminder.value.reminder = null
	}
	emit('update:modelValue', reminder.value)
}

function submitForm() {
	updateData()
	close()
}

const changed = ref(false)

function close() {
	setTimeout(() => {
		isShowForm.value = false
		if (changed.value) {
			changed.value = false
		}
	}, 200)
}

function formatDuration(reminderPeriod: number): string {
	if (Math.abs(reminderPeriod) < 60) {
		return '00:00'
	}
	const duration = secondsToPeriod(Math.abs(reminderPeriod))
	return (duration.days > 0 ? `${duration.days} ${t('task.reminder.days')} `: '') +
			('' + duration.hours).padStart(2, '0') + ':' +
			('' + duration.minutes).padStart(2, '0')
}

const relativeToOptions = {
	[REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE]: t('task.attributes.dueDate'),
	[REMINDER_PERIOD_RELATIVE_TO_TYPES.STARTDATE]: t('task.attributes.startDate'),
	[REMINDER_PERIOD_RELATIVE_TO_TYPES.ENDDATE]: t('task.attributes.endDate'),
} as const

const relativeTo = computed(() => relativeToOptions[periodInput.relativeTo]))



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
.input {
	max-width: 5rem;
	width: 4rem;
}

.close-button {
	margin: 0.5rem;
	width: calc(100% - 1rem);
}

</style>