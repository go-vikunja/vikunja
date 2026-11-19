<template>
	<div>
		<Popup
			placement="bottom-start"
			:anchor="triggerEl"
			:prefer-larger-side="compact"
			sheet-on-mobile
			:sheet-title="$t('task.attributes.reminders')"
			@update:open="showFormSwitch = null"
		>
			<template #trigger="{toggle}">
				<SimpleButton
					ref="trigger"
					v-tooltip="reminder.reminder && reminder.relative_to !== '' ? formatDisplayDate(reminder.reminder) : null"
					:disabled="disabled"
					@click.prevent.stop="toggle()"
				>
					{{ reminderText }}
				</SimpleButton>
			</template>
			<template
				v-if="activeForm === 'absolute'"
				#header-action
			>
				<BaseButton
					class="bottom-sheet__header-action"
					@click.stop="reminderDate = null"
				>
					{{ $t('input.datepicker.clear') }}
				</BaseButton>
			</template>
			<template #content="{isOpen, close}">
				<div
					class="reminder-options-popup"
					:class="{'is-open': isOpen}"
				>
					<template v-if="activeForm === 'absolute'">
						<DatepickerPanel
							v-if="isOpen"
							v-model="reminderDate"
							:confirm-disabled="reminderDate === null"
							:compact="compact"
							@confirm="confirmAndClose(close)"
						/>
					</template>
					<Card
						v-else
						class="reminder-options"
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
								:class="{'currently-active': p.relative_period === modelValue?.relative_period
									&& modelValue?.relative_to === p.relative_to}"
								@click="setReminderFromPreset(p, close)"
							>
								{{ formatReminder(p) }}
							</SimpleButton>
							<SimpleButton
								class="option-button"
								:class="{'currently-active': typeof modelValue !== 'undefined'
									&& modelValue?.relative_to !== ''
									&& presets.find(p => p.relative_period === modelValue?.relative_period
										&& modelValue?.relative_to === p.relative_to) === undefined}"
								@click="showFormSwitch = 'relative'"
							>
								{{ $t('task.reminder.custom') }}
							</SimpleButton>
							<SimpleButton
								v-if="allowAbsolute"
								class="option-button"
								:class="{'currently-active': modelValue?.relative_to === ''}"
								@click="showFormSwitch = 'absolute'"
							>
								{{ $t('task.reminder.dateAndTime') }}
							</SimpleButton>
						</div>

						<template v-if="activeForm === 'relative'">
							<ReminderPeriod
								v-model="reminder"
								:lock-relative-to="lockedRelativeTo"
							/>
							<XButton
								class="reminder__close-button"
								:shadow="false"
								@click="confirmAndClose(close)"
							>
								{{ $t('misc.confirm') }}
							</XButton>
						</template>
					</Card>
				</div>
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
import type {TaskReminder as ITaskReminder} from '@/client/generated'
import {formatDisplayDate} from '@/helpers/time/formatDate'
import {createReminderDraft} from '@/helpers/task'

import BaseButton from '@/components/base/BaseButton.vue'
import DatepickerPanel from '@/components/input/DatepickerPanel.vue'
import ReminderPeriod from '@/components/tasks/partials/ReminderPeriod.vue'
import Popup from '@/components/misc/Popup.vue'

import Card from '@/components/misc/Card.vue'
import SimpleButton from '@/components/input/SimpleButton.vue'

const props = withDefaults(defineProps<{
	modelValue?: ITaskReminder,
	clearAfterUpdate?: boolean,
	defaultRelativeTo?: IReminderPeriodRelativeTo | null,
	allowAbsolute?: boolean,
	disabled?: boolean,
	compact?: boolean,
}>(), {
	modelValue: () => createReminderDraft(),
	clearAfterUpdate: false,
	defaultRelativeTo: REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE,
	allowAbsolute: true,
	disabled: false,
	compact: false,
})

const emit = defineEmits<{
	'update:modelValue': [value: ITaskReminder | undefined],
}>()

const {t} = useI18n({useScope: 'global'})

const trigger = ref<InstanceType<typeof SimpleButton> | null>(null)
const triggerEl = computed<HTMLElement | null>(() => trigger.value?.$el ?? null)

const reminder = ref<ITaskReminder>(createReminderDraft())

const presets = computed(() => [
	{reminder: '', relative_period: 0, relative_to: props.defaultRelativeTo},
	{reminder: '', relative_period: -2 * SECONDS_A_HOUR, relative_to: props.defaultRelativeTo},
	{reminder: '', relative_period: -1 * SECONDS_A_DAY, relative_to: props.defaultRelativeTo},
	{reminder: '', relative_period: -1 * SECONDS_A_DAY * 3, relative_to: props.defaultRelativeTo},
	{reminder: '', relative_period: -1 * SECONDS_A_DAY * 7, relative_to: props.defaultRelativeTo},
	{reminder: '', relative_period: -1 * SECONDS_A_DAY * 30, relative_to: props.defaultRelativeTo},
] as ITaskReminder[])
const reminderDate = ref<Date | null>(null)

const showFormSwitch = ref<null | 'relative' | 'absolute'>(null)

const activeForm = computed(() => {
	if (props.defaultRelativeTo === null && props.allowAbsolute) {
		return 'absolute'
	}

	return showFormSwitch.value
})

const lockedRelativeTo = computed(() => {
	if (props.allowAbsolute) {
		return null
	}
	return props.defaultRelativeTo
})

const reminderText = computed(() => {
	if (reminder.value.relative_to) {
		return formatReminder(reminder.value)
	}

	if (reminder.value.reminder) {
		return formatDisplayDate(reminder.value.reminder)
	}

	return t('task.addReminder')
})

watch(
	() => props.modelValue,
	(newReminder) => {
		if (newReminder) {
			reminder.value = {...newReminder}

			if (!newReminder.relative_to && newReminder.reminder) {
				reminderDate.value = new Date(newReminder.reminder)
			}

			return
		}

		reminder.value = createReminderDraft()
	},
	{immediate: true},
)

function updateData() {
	emit('update:modelValue', reminder.value)

	if (props.clearAfterUpdate) {
		reminder.value = createReminderDraft()
	}
}

function setReminderDateAndClose(close: () => void) {
	reminder.value.reminder = reminderDate.value === null
		? ''
		: new Date(reminderDate.value).toISOString()
	reminder.value.relative_to = ''
	reminder.value.relative_period = 0
	updateDataAndMaybeCloseNow(close)
}


function setReminderFromPreset(preset: ITaskReminder, close: () => void) {
	reminder.value = {...preset}
	updateData()
	close()
}

function updateDataAndMaybeCloseNow(close: () => void) {
	updateData()
	if (props.clearAfterUpdate) {
		close()
	}
}

function confirmAndClose(close: () => void) {
	if (activeForm.value === 'absolute') {
		setReminderDateAndClose(close)
	} else {
		updateDataAndMaybeCloseNow(close)
	}
}

function formatReminder(reminder: ITaskReminder) {
	const period = secondsToPeriod(reminder.relative_period ?? 0)

	if (period.amount === 0) {
		switch (reminder.relative_to) {
			case REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE:
				return t('task.reminder.onDueDate')
			case REMINDER_PERIOD_RELATIVE_TO_TYPES.STARTDATE:
				return t('task.reminder.onStartDate')
			case REMINDER_PERIOD_RELATIVE_TO_TYPES.ENDDATE:
				return t('task.reminder.onEndDate')
		}
	}

	const amountAbs = Math.abs(period.amount)

	let relative_to = ''
	switch (reminder.relative_to) {
		case REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE:
			relative_to = t('task.attributes.dueDate')
			break
		case REMINDER_PERIOD_RELATIVE_TO_TYPES.STARTDATE:
			relative_to = t('task.attributes.startDate')
			break
		case REMINDER_PERIOD_RELATIVE_TO_TYPES.ENDDATE:
			relative_to = t('task.attributes.endDate')
			break
	}

	if ((reminder.relative_period ?? 0) <= 0) {
		return t('task.reminder.before', {
			amount: amountAbs,
			unit: translateUnit(amountAbs, period.unit),
			type: relative_to,
		})
	}

	return t('task.reminder.after', {
		amount: amountAbs,
		unit: translateUnit(amountAbs, period.unit),
		type: relative_to,
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

.reminder-options {
	inline-size: 310px;
	z-index: 99;

	.bottom-sheet & {
		inline-size: 100%;
		border: 0;
		box-shadow: none;
		border-radius: 0;
		padding-block-end: .5rem;

		.option-button {
			inline-size: 100%;
			padding: .75rem 1rem;
			font-size: 1rem;
		}
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
