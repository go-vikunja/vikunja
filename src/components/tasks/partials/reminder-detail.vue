<template>
	<div>
		<Popup @close="showFormSwitch = null">
			<template #trigger="{toggle}">
				<SimpleButton
					@click.prevent.stop="toggle()"
				>
					{{ reminderText }}
				</SimpleButton>
			</template>
			<template #content="{isOpen, toggle}">
				<Card class="reminder-options-popup" :class="{'is-open': isOpen}" :padding="false">
					<div class="options" v-if="showFormSwitch === null">
						<SimpleButton
							class="option-button"
							v-for="p in presets"
						>
							{{ formatReminder(p) }}
						</SimpleButton>
						<SimpleButton @click="showFormSwitch = 'relative'" class="option-button">
							Custom
						</SimpleButton>
						<SimpleButton @click="showFormSwitch = 'absolute'" class="option-button">
							Date
						</SimpleButton>
					</div>

					<ReminderPeriod
						v-if="showFormSwitch === 'relative'"
						v-model="reminder"
						@update:modelValue="emit('update:modelValue', reminder)"
					/>

					<DatepickerInline
						v-if="showFormSwitch === 'absolute'"
						v-model="reminderDate"
						@update:modelValue="setReminderDate"
					/>

					<x-button
						v-if="showFormSwitch !== null"
						class="reminder__close-button"
						:shadow="false"
						@click="toggle"
					>
						{{ $t('misc.confirm') }}
					</x-button>
				</Card>
			</template>
		</Popup>
	</div>
</template>

<script setup lang="ts">
import {computed, ref, watch, type PropType} from 'vue'
import {toRef} from '@vueuse/core'
import {SECONDS_A_DAY} from '@/constants/date'
import {REMINDER_PERIOD_RELATIVE_TO_TYPES} from '@/types/IReminderPeriodRelativeTo'
import {useI18n} from 'vue-i18n'

import {PeriodUnit, secondsToPeriod} from '@/helpers/time/period'
import type {ITaskReminder} from '@/modelTypes/ITaskReminder'
import {formatDateShort} from '@/helpers/time/formatDate'

import BaseButton from '@/components/base/BaseButton.vue'
import DatepickerInline from '@/components/input/datepickerInline.vue'
import ReminderPeriod from '@/components/tasks/partials/reminder-period.vue'
import Popup from '@/components/misc/popup.vue'

import TaskReminderModel from '@/models/taskReminder'
import Card from '@/components/misc/card.vue'
import SimpleButton from '@/components/input/SimpleButton.vue'

const {t} = useI18n({useScope: 'global'})

const props = defineProps({
	modelValue: {
		type: Object as PropType<ITaskReminder>,
		required: false,
	},
	disabled: {
		default: false,
	},
})

const emit = defineEmits(['update:modelValue'])

const reminder = ref<ITaskReminder>(new TaskReminderModel())

const presets: TaskReminderModel[] = [
	{relativePeriod: SECONDS_A_DAY, relativeTo: REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE},
	{relativePeriod: SECONDS_A_DAY * 3, relativeTo: REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE},
	{relativePeriod: SECONDS_A_DAY * 7, relativeTo: REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE},
	{relativePeriod: SECONDS_A_DAY * 30, relativeTo: REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE},
]
const reminderDate = ref(null)

const showFormSwitch = ref<null | 'relative' | 'absolute'>(null)

const reminderText = computed(() => {

	if (reminder.value.relativeTo !== null) {
		return formatReminder(reminder.value)
	}

	if (reminder.value.reminder !== null) {
		return formatDateShort(reminder.value.reminder)
	}

	return t('task.addReminder')
})

const modelValue = toRef(props, 'modelValue')
watch(
	modelValue,
	(newReminder) => {
		reminder.value = newReminder || new TaskReminderModel()
	},
	{immediate: true},
)

function setReminderDate() {
	reminder.value.reminder = reminderDate.value === null
		? null
		: new Date(reminderDate.value)
	emit('update:modelValue', reminder.value)
}

function formatReminder(reminder: TaskReminderModel) {

	const period = secondsToPeriod(reminder.relativePeriod)

	if (period.amount === 0) {
		switch (reminder.relativeTo) {
			case REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE:
				return t('task.reminder.onDueDate')
			case REMINDER_PERIOD_RELATIVE_TO_TYPES.STARTDATE:
				return t('task.reminder.onStartDate')
			case REMINDER_PERIOD_RELATIVE_TO_TYPES.ENDDATE:
				return t('task.reminder.onEndDate')
		}
	}

	const amountAbs = Math.abs(period.amount)

	let relativeTo = ''
	switch (reminder.relativeTo) {
		case REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE:
			relativeTo = t('task.attributes.dueDate')
			break
		case REMINDER_PERIOD_RELATIVE_TO_TYPES.STARTDATE:
			relativeTo = t('task.attributes.startDate')
			break
		case REMINDER_PERIOD_RELATIVE_TO_TYPES.ENDDATE:
			relativeTo = t('task.attributes.endDate')
			break
	}

	if (reminder.relativePeriod <= 0) {
		return t('task.reminder.before', {
			amount: amountAbs,
			unit: translateUnit(amountAbs, period.unit),
			type: relativeTo,
		})
	}

	return t('task.reminder.after', {
		amount: amountAbs,
		unit: translateUnit(amountAbs, period.unit),
		type: relativeTo,
	})
}

function translateUnit(amount: number, unit: PeriodUnit): string {
	switch (unit) {
		case 'seconds':
			return t('time.units.seconds', amount)
		case 'minutes':
			return t('time.units.minutes', amount)
		case 'hours':
			return t('time.units.hours', amount)
		case 'days':
			return t('time.units.days', amount)
		case 'weeks':
			return t('time.units.weeks', amount)
		case 'months':
			return t('time.units.months', amount)
		case 'years':
			return t('time.units.years', amount)
	}
}
</script>

<style lang="scss" scoped>
.options {
	display: flex;
	flex-direction: column;
	align-items: flex-start;
}

:deep(.popup) {
	top: unset;
}

.reminder-options-popup {
	width: 300px;
	z-index: 99;

	@media screen and (max-width: ($tablet)) {
		width: calc(100vw - 5rem);
	}

	.option-button {
		font-size: .85rem;
		border-radius: 0;
		padding: .5rem;
		margin: 0;

		&:hover {
			background: var(--grey-100);
		}
	}
}

.reminder__close-button {
	margin: .5rem;
	width: calc(100% - 1rem);
}
</style>
