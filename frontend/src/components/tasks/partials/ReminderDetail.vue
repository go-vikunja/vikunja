<template>
	<div>
		<Popup @update:open="showFormSwitch = null">
			<template #trigger="{toggle}">
				<SimpleButton
					v-tooltip="reminder.reminder && reminder.relativeTo !== null ? formatDisplayDate(reminder.reminder) : null"
					@click.prevent.stop="toggle()"
				>
					{{ reminderText }}
				</SimpleButton>
			</template>
			<template #content="{isOpen, close}">
				<Card
					class="reminder-options-popup"
					:class="{'is-open': isOpen}"
					:padding="false"
				>
					<div
						v-if="activeForm === null"
						class="options"
					>
						<SimpleButton
							v-for="(p, k) in presets"
							:key="k"
							class="option-button"
							:class="{'currently-active': p.relativePeriod === modelValue?.relativePeriod && modelValue?.relativeTo === p.relativeTo}"
							@click="setReminderFromPreset(p, close)"
						>
							{{ formatReminder(p) }}
						</SimpleButton>
						<SimpleButton
							class="option-button"
							:class="{'currently-active': typeof modelValue !== 'undefined' && modelValue?.relativeTo !== null && presets.find(p => p.relativePeriod === modelValue?.relativePeriod && modelValue?.relativeTo === p.relativeTo) === undefined}"
							@click="showFormSwitch = 'relative'"
						>
							{{ $t('task.reminder.custom') }}
						</SimpleButton>
						<SimpleButton
							class="option-button"
							:class="{'currently-active': modelValue?.relativeTo === null}"
							@click="showFormSwitch = 'absolute'"
						>
							{{ $t('task.reminder.dateAndTime') }}
						</SimpleButton>
					</div>

					<ReminderPeriod
						v-if="activeForm === 'relative'"
						v-model="reminder"
					/>

					<DatepickerInline
						v-else-if="activeForm === 'absolute'"
						v-model="reminderDate"
						@update:modelValue="setReminderDateAndClose(close)"
					/>

					<XButton
						v-if="showFormSwitch !== null"
						class="reminder__close-button"
						:shadow="false"
						@click="updateDataAndMaybeCloseNow(close)"
					>
						{{ $t('misc.confirm') }}
					</XButton>
				</Card>
			</template>
		</Popup>
	</div>
</template>

<script setup lang="ts">
import {computed, ref, watch} from 'vue'
import {SECONDS_A_DAY, SECONDS_A_HOUR} from '@/constants/date'
import {type IReminderPeriodRelativeTo, REMINDER_PERIOD_RELATIVE_TO_TYPES} from '@/types/IReminderPeriodRelativeTo'
import {useI18n} from 'vue-i18n'

import {type PeriodUnit, secondsToPeriod} from '@/helpers/time/period'
import type {ITaskReminder} from '@/modelTypes/ITaskReminder'
import {formatDisplayDate} from '@/helpers/time/formatDate'

import DatepickerInline from '@/components/input/DatepickerInline.vue'
import ReminderPeriod from '@/components/tasks/partials/ReminderPeriod.vue'
import Popup from '@/components/misc/Popup.vue'

import TaskReminderModel from '@/models/taskReminder'
import Card from '@/components/misc/Card.vue'
import SimpleButton from '@/components/input/SimpleButton.vue'
import {useDebounceFn} from '@vueuse/core'

const props = withDefaults(defineProps<{
	modelValue?: ITaskReminder,
	clearAfterUpdate?: boolean,
	defaultRelativeTo?: IReminderPeriodRelativeTo | null,
}>(), {
	modelValue: () => new TaskReminderModel() as ITaskReminder,
	clearAfterUpdate: false,
	defaultRelativeTo: REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE,
})

const emit = defineEmits<{
	'update:modelValue': [value: ITaskReminder | undefined],
}>()

const {t} = useI18n({useScope: 'global'})

const reminder = ref<ITaskReminder>(new TaskReminderModel())

const presets = computed(() => [
	{reminder: null, relativePeriod: 0, relativeTo: props.defaultRelativeTo},
	{reminder: null, relativePeriod: -2 * SECONDS_A_HOUR, relativeTo: props.defaultRelativeTo},
	{reminder: null, relativePeriod: -1 * SECONDS_A_DAY, relativeTo: props.defaultRelativeTo},
	{reminder: null, relativePeriod: -1 * SECONDS_A_DAY * 3, relativeTo: props.defaultRelativeTo},
	{reminder: null, relativePeriod: -1 * SECONDS_A_DAY * 7, relativeTo: props.defaultRelativeTo},
	{reminder: null, relativePeriod: -1 * SECONDS_A_DAY * 30, relativeTo: props.defaultRelativeTo},
] as ITaskReminder[])
const reminderDate = ref<Date | null>(null)

const showFormSwitch = ref<null | 'relative' | 'absolute'>(null)

const activeForm = computed(() => {
	if (props.defaultRelativeTo === null) {
		return 'absolute'
	}

	return showFormSwitch.value
})

const reminderText = computed(() => {
	if (reminder.value.relativeTo !== null) {
		return formatReminder(reminder.value)
	}

	if (reminder.value.reminder !== null) {
		return formatDisplayDate(reminder.value.reminder)
	}

	return t('task.addReminder')
})

watch(
	() => props.modelValue,
	(newReminder) => {
		if (newReminder) {
			reminder.value = newReminder

			if (newReminder.relativeTo === null && newReminder.reminder !== null) {
				reminderDate.value = new Date(newReminder.reminder)
			}

			return
		}

		reminder.value = new TaskReminderModel()
	},
	{immediate: true},
)

function updateData() {
	emit('update:modelValue', reminder.value)

	if (props.clearAfterUpdate) {
		reminder.value = new TaskReminderModel()
	}
}

function setReminderDateAndClose(close: () => void) {
	reminder.value.reminder = reminderDate.value === null
		? null
		: new Date(reminderDate.value)
	reminder.value.relativeTo = null
	reminder.value.relativePeriod = 0
	updateDataAndMaybeClose(close)
}


function setReminderFromPreset(preset: ITaskReminder, close: () => void) {
	reminder.value = preset
	updateData()
	close()
}

const updateDataAndMaybeClose = useDebounceFn(updateDataAndMaybeCloseNow, 500)

function updateDataAndMaybeCloseNow(close: () => void) {
	updateData()
	if (props.clearAfterUpdate) {
		close()
	}
}

function formatReminder(reminder: ITaskReminder) {
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
		case 'years':
			return t('time.units.years', amount)
		default:
			throw new Error(`Unknown unit: ${unit}`)
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
	inset-block-start: unset;
}

.reminder-options-popup {
	inline-size: 310px;
	z-index: 99;

	@media screen and (max-width: ($tablet)) {
		inline-size: calc(100vw - 5rem);
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
	inline-size: calc(100% - 1rem);
}

.currently-active {
	color: var(--primary);
}
</style>
